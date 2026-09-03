package goframe

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
	"github.com/hosgf/element/util"
)

const defaultDedupTTL = 300 * time.Second

// DedupStore 去重占位存储。dup=true 表示键已存在。
type DedupStore interface {
	TryMark(ctx context.Context, key string, ttl time.Duration) (dup bool, err error)
}

// DedupOptions 去重参数。不读配置文件，由调用方在注册时传入。
type DedupOptions struct {
	Store        DedupStore
	KeyPrefix    string
	TTL          time.Duration
	SkipPaths    []string
	RequireRedis bool
	OnDup        func(r *ghttp.Request, header, value string)
	OnError      func(r *ghttp.Request, err error)
}

// Deduper 请求去重：范围 Method+Path；同范围内 X-Req-Id > X-Req-Secret；两者皆无则跳过。
type Deduper struct {
	store        DedupStore
	opts         DedupOptions
	mu           sync.RWMutex
	excludePaths []string
}

// NewDedup 创建去重器。store 为 nil 时使用 Redis SET NX（g.Redis）。
func NewDedup(store DedupStore, o DedupOptions) *Deduper {
	if store == nil {
		store = o.Store
	}
	if store == nil {
		store = RedisStore{Require: o.RequireRedis}
	}
	if o.TTL <= 0 {
		o.TTL = defaultDedupTTL
	}
	if strings.TrimSpace(o.KeyPrefix) == "" {
		o.KeyPrefix = "request:dedup:"
	}
	if o.OnDup == nil {
		o.OnDup = defaultOnDup
	}
	if o.OnError == nil {
		o.OnError = defaultOnError
	}
	if len(o.SkipPaths) == 0 {
		o.SkipPaths = []string{"/health", "/metrics", "/metrics/*", "/debug/*", "/ping", "/favicon.ico"}
	}
	return &Deduper{store: store, opts: o}
}

// Dedup 注册用中间件。无参则 Redis SET NX + 默认前缀。
func Dedup(o ...DedupOptions) ghttp.HandlerFunc {
	var opts DedupOptions
	if len(o) > 0 {
		opts = o[0]
	}
	d := NewDedup(opts.Store, opts)
	bindDefaultDedup(d)
	return d.Middleware()
}

var (
	dedupMu        sync.Mutex
	defaultDedup   *Deduper
	pendingExclude []string
)

// DedupExclude 登记排除路径（可在 Dedup() 注册前后调用）。
func DedupExclude(paths ...string) {
	dedupMu.Lock()
	defer dedupMu.Unlock()
	if defaultDedup != nil {
		defaultDedup.Exclude(paths...)
		return
	}
	for _, p := range paths {
		if p = strings.TrimSpace(p); p != "" {
			pendingExclude = append(pendingExclude, p)
		}
	}
}

func bindDefaultDedup(d *Deduper) {
	dedupMu.Lock()
	defer dedupMu.Unlock()
	defaultDedup = d
	if len(pendingExclude) > 0 {
		d.Exclude(pendingExclude...)
		pendingExclude = nil
	}
}

// Exclude 追加排除路径（可在 server.Run 前多次调用）。
func (d *Deduper) Exclude(paths ...string) {
	if d == nil {
		return
	}
	valid := make([]string, 0, len(paths))
	for _, path := range paths {
		if path = strings.TrimSpace(path); path != "" {
			valid = append(valid, path)
		}
	}
	if len(valid) == 0 {
		return
	}
	d.mu.Lock()
	d.excludePaths = append(d.excludePaths, valid...)
	d.mu.Unlock()
}

// Middleware 返回 GoFrame 中间件。
func (d *Deduper) Middleware() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if d == nil || d.store == nil {
			r.Middleware.Next()
			return
		}
		path := r.URL.Path
		d.mu.RLock()
		excludePaths := d.excludePaths
		d.mu.RUnlock()
		if util.MatchPath(path, d.opts.SkipPaths) || util.MatchPath(path, excludePaths) {
			r.Middleware.Next()
			return
		}

		mark, ok := findMark(r)
		if !ok {
			r.Middleware.Next()
			return
		}
		dup, err := d.store.TryMark(r.Context(), d.opts.KeyPrefix+mark.key, d.opts.TTL)
		if err != nil {
			d.opts.OnError(r, err)
			return
		}
		if dup {
			d.opts.OnDup(r, mark.header, mark.value)
			return
		}
		r.Middleware.Next()
	}
}

type dedupMark struct {
	key    string
	header string
	value  string
}

func findMark(r *ghttp.Request) (dedupMark, bool) {
	scope := reqScope(r)
	// 优先 X-Req-Id：SameAuth 的 X-Req-Secret 多为 MD5(timestamp+salt)，同秒请求会撞键。
	if id := GetHeader(r, request.HeaderReqId); id != "" {
		return dedupMark{
			key:    hashKey(scope + ":" + id),
			header: request.HeaderReqId.String(),
			value:  id,
		}, true
	}
	if secret := GetHeader(r, request.HeaderSignature); secret != "" {
		return dedupMark{
			key:    hashKey(scope + ":" + secret),
			header: request.HeaderSignature.String(),
			value:  maskSecret(secret),
		}, true
	}
	return dedupMark{}, false
}

func reqScope(r *ghttp.Request) string {
	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	return fmt.Sprintf("%s:%s", r.Method, path)
}

func hashKey(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func maskSecret(value string) string {
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}

func defaultOnDup(r *ghttp.Request, header, value string) {
	logger.Warningf(r.Context(), "dedup_rejected path=%s method=%s header=%s value=%s",
		r.URL.Path, r.Method, header, value)
	result.Writer(r, result.Build(result.SC_FAILURE, "请求重复，请勿重复提交", "", nil))
}

func defaultOnError(r *ghttp.Request, err error) {
	logger.Errorf(r.Context(), "dedup_unavailable path=%s: %v", r.URL.Path, err)
	r.Response.WriteStatus(http.StatusServiceUnavailable, "请求去重服务暂不可用，请稍后重试")
	r.Exit()
}
