# element

Go 微服务基础库：HTTP 中间件、链路上下文、统一日志、异常处理、K8s 客户端等。

## 文档

| 文档 | 说明 |
|------|------|
| [链路上下文接入指南](./docs/链路上下文接入指南.md) | trace_id / request_id、HTTP / 定时器 / goroutine 接入 |
| [全局异常处理](./docs/全局异常处理.md) | GoFrame / Gin 异常中间件与 uerrors |
| [架构与代码规范](./docs/架构与代码规范.md) | 错误处理、Context 使用规范 |

## 快速开始

### HTTP（GoFrame）

```go
mf.SetMiddleware(g.Server(), mf.Recover)
logger.Infof(r.Context(), "hello")
```

### 定时器 / 监听

```go
ctx.Run(context.Background(), func(c context.Context) {
    logger.Infof(c, "tick")
})
```

### 异步 goroutine

```go
ctx.Go(reqCtx, func(c context.Context) {
    logger.Infof(c, "async")
})
```

完整说明见 [链路上下文接入指南](./docs/链路上下文接入指南.md)。
