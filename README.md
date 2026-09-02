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

### 2. HTTP（GoFrame 自动 OTel）

```go
import mf "github.com/hosgf/element/middleware/goframe"

mf.SetMiddleware(g.Server(), mf.Recover)
logger.Infof(r.Context(), "hello")  // 自动输出 {trace_id} <request_id>
```

### 3. HTTP（Gin 需显式 OTel）

```go
import "github.com/hosgf/element/middleware/ugin"

r.Use(ugin.Tracing("my-service"))  // 必须在 MiddlewareHeader 前
ugin.SetMiddleware(r, ugin.ExceptionHandler())
```

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
