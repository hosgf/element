# element

Go 微服务基础库：OpenTelemetry、HTTP 中间件、统一日志、异常处理、K8s 客户端等。

## 快速开始

### 1. 初始化 OTel（可选层 2 上报）

```go
import "github.com/hosgf/element/trace"

shutdown, _ := trace.Init(ctx, trace.Config{
    Enabled:     true,
    Exporter:    "none",  // 层 1：传播与进程内 span；改 "otlp" 开启上报
    ServiceName: "my-service",
    SampleRate:  1.0,
})
defer shutdown(ctx)
```

### 2. HTTP（GoFrame）

```go
import mf "github.com/hosgf/element/middleware/goframe"

s := g.Server()
mf.SetMiddleware(s, mf.Recover, mf.AccessLog(), mf.Dedup(mf.DedupOptions{
    KeyPrefix: "mysvc:dedup:",
}))
mf.DedupExclude("/internal/notify")
```

`result.Success` / `Writer` 只把 JSON 写入 GoFrame buffer；压缩由 `MiddlewareGzip` 按 `Accept-Encoding` 处理。访问日志须注册在 Gzip **内侧**（作为 `SetMiddleware` 的后续 handler 或后 `Use`），才能打到明文请求/响应。完整说明见 [HTTP 中间件](./docs/HTTP中间件.md)。

去重：范围 `Method+Path`，同范围内 `X-Req-Id` > `X-Req-Secret`，两者皆无则跳过。默认 `g.Redis()` SET NX，失败且未设 `RequireRedis` 时降级内存。`DedupExclude` 可在注册前后调用。

业务失败写入 ctx 后 AccessLog 会打 `[异常]`：

```go
r.SetCtxVar(mf.AccessFailureKey, `{"message":"..."}`)
r.SetCtxVar(mf.AccessErrorKey, err)
```

### 3. HTTP（Gin 需显式 OTel）

```go
import "github.com/hosgf/element/middleware/ugin"

r.Use(ugin.Tracing("my-service"))  // 必须在 MiddlewareHeader 前
ugin.SetMiddleware(r, ugin.ExceptionHandler(), ugin.AccessLog(), ugin.Dedup())
ugin.DedupExclude("/webhook/callback")
```

**Gin 中间件注意**：
- Gzip 已在 `SetMiddleware` 外层处理（gin-contrib/gzip）
- `AccessLog` 自动捕获响应体，需要记录请求 body 时需先读到 `gin.BodyBytesKey`
- 业务失败写入 context：`c.Set(ugin.AccessFailureKey, msg)` / `c.Set(ugin.AccessErrorKey, err)`

### 4. 定时任务 / 消息消费

```go
import "github.com/hosgf/element/ctx"

ctx.Run(context.Background(), func(c context.Context) {
    logger.Infof(c, "tick")  // 自动创建 Root Span
})
```

### 5. 异步 goroutine

```go
ctx.Go(reqCtx, func(c context.Context) {
    logger.Infof(c, "async")
})
```

## 文档

| 文档 | 说明 |
|------|------|
| [HTTP 中间件](./docs/HTTP中间件.md) | Gzip、AccessLog、Dedup、路径匹配与请求头 |
| [链路上下文接入指南](./docs/链路上下文接入指南.md) | trace_id / request_id、HTTP / 定时器 / goroutine 接入 |
| [OpenTelemetry 接入设计](./docs/OpenTelemetry接入设计.md) | OTel 分层（层 1 传播 + 层 2 上报）、复用点、改造步骤 |
| [全局异常处理](./docs/全局异常处理.md) | GoFrame / Gin 异常中间件与 uerrors |
| [架构与代码规范](./docs/架构与代码规范.md) | 错误处理、Context 使用规范 |

## 核心特性

- ✅ **W3C traceparent 传播**：统一 OTel 标准，不再使用 `X-Trace-Id` 自定义头
- ✅ **两层设计**：层 1 必选（传播+日志）、层 2 可选（OTLP 上报）
- ✅ **GoFrame 零侵入**：复用内置 OTel，自动 SERVER/CLIENT/DB span
- ✅ **Gin 一行接入**：`ugin.Tracing(name)` 即可
- ✅ **每次出站新 X-Req-Id**：符合 Dedup 语义，不再复用线程 MDC
