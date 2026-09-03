package goframe

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

var errDedupRedis = errors.New("dedup redis unavailable")

// RedisStore 使用 g.Redis() SET NX。Require=true 时 Redis 失败直接报错，否则降级内存。
type RedisStore struct {
	Require bool
}

func (s RedisStore) TryMark(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	sec := int(ttl.Seconds())
	dup, err := tryMarkRedis(ctx, key, sec)
	if err == nil {
		return dup, nil
	}
	if s.Require {
		return false, errDedupRedis
	}
	return tryMarkMemory(key, ttl), nil
}

func tryMarkRedis(ctx context.Context, key string, ttlSec int) (bool, error) {
	rds := g.Redis()
	if rds == nil {
		return false, errDedupRedis
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var (
		out interface{}
		err error
	)
	if ttlSec > 0 {
		out, err = rds.Do(timeoutCtx, "SET", key, "1", "NX", "EX", ttlSec)
	} else {
		out, err = rds.Do(timeoutCtx, "SET", key, "1", "NX")
	}
	if err != nil {
		return false, err
	}
	if out != nil && g.NewVar(out).String() == "OK" {
		return false, nil
	}
	return true, nil
}

type memEntry struct {
	expire time.Time
}

var (
	memMu    sync.Mutex
	memCache = map[string]memEntry{}
)

func tryMarkMemory(key string, ttl time.Duration) bool {
	memMu.Lock()
	defer memMu.Unlock()
	now := time.Now()
	if e, ok := memCache[key]; ok && e.expire.After(now) {
		return true
	}
	if ttl <= 0 {
		ttl = defaultDedupTTL
	}
	memCache[key] = memEntry{expire: now.Add(ttl)}
	return false
}
