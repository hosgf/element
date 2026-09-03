# HTTP 中间件

GoFrame（`middleware/goframe`）与 Gin（`middleware/ugin`）提供同一套能力：Gzip、访问日志、请求去重。路径匹配见 `util.MatchPath`。

---

## 注册顺序

Gzip 必须在最外层，AccessLog / Dedup 放在内侧，才能打到明文请求/响应。

**GoFrame**（`SetMiddleware` 固定前缀）：

```go
import mf "github.com/hosgf/element/middleware/goframe"

s := g.Server()
mf.SetMiddleware(s, mf.Recover, mf.AccessLog(), mf.Dedup(mf.DedupOptions{
    KeyPrefix: "mysvc:dedup:",
}))
mf.DedupExclude("/internal/notify")
```

顺序：`MiddlewareGzip` → `MiddlewareCORS` → `MiddlewareHeader` → `MiddlewareCookies` → 传入的业务中间件。

**Gin**：

```go
import "github.com/hosgf/element/middleware/ugin"

r.Use(ugin.Tracing("my-service")) // 必须在 SetMiddleware 之前
ugin.SetMiddleware(r, ugin.ExceptionHandler(), ugin.AccessLog(), ugin.Dedup())
ugin.DedupExclude("/webhook/callback")
```

顺序：`gzip` → `MiddlewareHeader` → 传入的业务中间件。

`result.Success` / `Writer` 只写明文 JSON 到 GoFrame buffer；压缩由 `MiddlewareGzip` 按 `Accept-Encoding` 处理。

---

## AccessLog

记录方法、URL、选定请求头、参数预览、响应预览、耗时。无参即可用。默认跳过 `/health`、`/ping`（`SkipPaths == nil` 时；传空切片表示不跳过）。

```go
AccessLog(AccessLogOptions{
    SkipPaths: []string{"/health", "/metrics/*"},
    MaxParam:  4096,
    MaxResp:   2000,
})
```

默认记录的请求头：`traceparent`、`X-Req-Id`、`User-Agent`、`Content-Type`、`X-Req-App-*`、`X-Req-Client`、`timestamp`、`X-Req-Secret`、`X-Trace-Id`。密码、token 等字段会脱敏。

业务失败写入 ctx 后会打 `[异常]`：

```go
// GoFrame
r.SetCtxVar(mf.AccessFailureKey, `{"message":"..."}`)
r.SetCtxVar(mf.AccessErrorKey, err)

// Gin
c.Set(ugin.AccessFailureKey, `{"message":"..."}`)
c.Set(ugin.AccessErrorKey, err)
```

Gin 要记请求 body，需先读入 `gin.BodyBytesKey`（并重置 `c.Request.Body`），AccessLog 才会打 `[Params] body=...`。

---

## Dedup

范围是 `Method+Path`。同一范围内优先 `X-Req-Secret`，否则 `X-Req-Id`；两者皆无则跳过。默认 `g.Redis()` SET NX，失败且未设 `RequireRedis` 时降级内存。

默认跳过：`/health`、`/metrics`、`/metrics/*`、`/debug/*`、`/ping`、`/favicon.ico`。

```go
Dedup(DedupOptions{
    KeyPrefix:    "mysvc:dedup:",
    TTL:          5 * time.Minute,
    RequireRedis: false,
})
```

`DedupExclude` 可在 `Dedup()` 注册前后调用，追加排除路径。

---

## 路径匹配

`util.MatchPath` 支持精确匹配与末尾 `*` 前缀匹配，AccessLog / Dedup 的跳过路径都走它。

```go
util.MatchPath("/metrics/foo", []string{"/metrics/*"}) // true
util.JoinPath("/data", "a", "b")                       // /data/a/b
```

---

## 请求头常量

`client/request` 中与日志、去重相关的头：

| 常量 | 值 |
|------|-----|
| `HeaderUserAgent` | `User-Agent` |
| `HeaderContentType` | `Content-Type` |
| `HeaderReqId` | `X-Req-Id` |
| `HeaderSignature` | `X-Req-Secret` |
| `HeaderTraceparent` | `traceparent` |
