# OpenTelemetry 接入设计

> 状态：**Phase 1b 层 1 已完成**（2026-09-02）。  
> 目标：在保留 `element` 现有 `ctx`、middleware、HTTP client 和 logger 使用方式的前提下，接入可上报的 OpenTelemetry Trace。

## 1. 背景

`element` 当前已经提供一套统一链路上下文：

```text
HTTP / 定时任务 / 消息监听
        ↓
ctx.ContinueTrace / ctx.Run
        ↓
context 中的 trace_id / request_id
        ↓
logger、异常处理、出站 HTTP
```

现有实现能够生成和透传 `trace_id`，但还没有完成真正的 OTel Trace：

- 没有初始化带 SpanProcessor 的 `TracerProvider`；
- 没有创建 OTLP exporter；
- `ctx.Run` 只生成 ID，没有创建和结束 Span；
- `trace/otel.go` 通过自定义 TraceID 构造 Remote SpanContext，但该上下文不是可导出的 Span；
- Gin 和 `proxy` 的原生 HTTP client 尚未接入 OTel instrumentation。

因此，仅在 Kubernetes 中增加 `OTEL_*` 环境变量不会自动上报 Trace。应用还需要在启动时安装 SDK，并由入口、客户端或业务代码创建 Span。

## 2. 设计目标

### 2.1 目标

- 保留业务现有的 `ctx.Run`、`ctx.Go`、`ctx.GetTraceId`、`logger.Infof(ctx, ...)` 等调用方式；
- 复用 GoFrame 已有的 HTTP Server、HTTP Client 和数据库自动埋点；
- 统一使用 W3C `traceparent`/`tracestate` 传播 OTel 上下文；
- 日志中的 `trace_id` 来自当前真实 Span；
- SDK 初始化代码集中复用，但由上层应用显式启动和关闭；
- 未启用 OTel SDK 时，基础组件仍可运行，不影响非观测业务功能；
- 支持 HTTP、定时任务、消息消费、异步任务和代理转发等现有场景。

### 2.2 非目标

第一阶段不处理以下内容：

- 不改造为 OTel Logs SDK；日志继续输出到 stdout/stderr，由 Alloy 采集到 Loki；
- 不要求一次性给所有业务函数增加手工 Span；
- 不在第一阶段统一改造 JSON 日志格式；
- 不使用 `trace_id` 作为 Loki stream label；
- 不替换现有异常、请求 ID、租户和用户上下文体系。

## 3. 核心原则

### 3.1 TraceID 唯一权威来源

有效 OTel SpanContext 是 TraceID 的唯一权威来源：

```text
当前 SpanContext.TraceID()
        ↓
ctx.GetTraceId(ctx)
        ↓
日志 trace_id
        ↓
兼容响应头 X-Trace-Id
```

禁止使用以下方式创建正式 Trace：

- Agent 或 logger 随机生成 TraceID；
- 使用 `request_id` 代替 TraceID；
- 根据客户端传入的 `X-Trace-Id` 伪造 Remote SpanContext；
- 只生成 TraceID 而不创建、结束 Span。

### 3.2 标准传播与兼容头分离

跨服务 OTel 传播使用：

```http
traceparent: 00-<trace-id>-<span-id>-01
tracestate: ...
baggage: ...
```

现有 Header 继续保留，但职责调整为：

| Header | 职责 |
|--------|------|
| `traceparent` | OTel 父子关系和采样状态的正式传播协议 |
| `tracestate` | OTel 厂商或采样扩展状态 |
| `baggage` | 经审核允许传播的业务上下文 |
| `X-Trace-Id` | 迁移兼容和响应展示，不作为父上下文来源 |
| `X-Req-Id` | HTTP 请求、幂等或问题定位标识，不是 TraceID |

### 3.3 基础组件与业务系统分层

```text
上层业务系统
├─ 决定 service.name / version / environment
├─ 启动并关闭 OTel SDK
├─ 决定采样配置
└─ 创建有业务语义的 Span

element 基础组件
├─ 提供统一 SDK 初始化函数
├─ 读取和传播 SpanContext
├─ 提供框架、客户端和后台任务埋点
├─ 将 TraceID 关联到日志
└─ 保留旧 API 的兼容行为
```

SDK 初始化实现可以放在 `element/trace` 中复用，但不得通过包 `init()` 隐式启动。最终应用必须显式调用并负责 Shutdown。

## 4. 当前可直接复用的能力

| 当前模块 | 已有能力 | 接入处理 |
|----------|----------|----------|
| `ctx/` | 统一 context 入口和传递规范 | 保留公共 API，替换内部 Trace 实现 |
| `trace/` | TraceID 生成、延续和兼容 | 改为 OTel SpanContext 优先 |
| `middleware/goframe` | Header、request_id、异常处理 | 复用；不重复创建 Server Span |
| `middleware/ugin` | Header、request_id、异常处理 | 复用；在其前面增加 `otelgin` |
| `client/httputil` | 基于 GoFrame `gclient` | 复用 `gclient` 自带 Client Span |
| `client/request` | 业务 Header 和请求 ID 透传 | 增加标准 OTel Propagator 注入 |
| `logger` | 自动输出 context 元数据 | 改为 OTel TraceID 优先，旧值兜底 |
| `proxy` | 路由、连接池、重试、熔断 | 全部保留，仅包装 HTTP Transport |
| GoFrame `ghttp` | 默认创建 HTTP Server Span | 直接复用 |
| GoFrame `gclient` | 默认创建 Client Span 并注入传播头 | 直接复用 |
| GoFrame `gdb` | 默认创建数据库 Span | 直接复用 |

## 5. 目标调用链

### 5.1 GoFrame HTTP

```text
应用 trace.Init
    ↓
GoFrame 内置 Server tracing
    ├─ 从 traceparent 提取父上下文
    └─ 创建 Server Span
    ↓
element MiddlewareHeader
    ├─ 将真实 TraceID 同步到兼容 context key
    └─ 绑定 request_id / tenant / user 等字段
    ↓
业务 Handler
    ├─ logger 从当前 Span 读取 TraceID
    ├─ gclient 创建下游 Client Span
    └─ gdb 创建数据库 Span
    ↓
Server Span.End
    ↓
BatchSpanProcessor → OTLP → Collector/Alloy → Tempo
```

GoFrame 已有 Server tracing，不再增加第二套 GoFrame OTel 中间件。

### 5.2 Gin HTTP

```text
应用 trace.Init
    ↓
otelgin Middleware
    ├─ 提取 traceparent
    └─ 创建 Server Span
    ↓
gzip
    ↓
element MiddlewareHeader
    ↓
ExceptionHandler
    ↓
业务 Handler
```

`otelgin` 必须先于 `MiddlewareHeader` 执行，否则 Header 和 logger 获取不到当前真实 Server Span。

### 5.3 定时任务

每次定时触发创建一个新 Root Span：

```go
ctx.RunNamed(parent, "cron.settlement", func(c context.Context) {
    logger.Infof(c, "开始执行结算")
    // 业务处理
})
```

Span 必须在回调返回后统一结束；业务代码不负责手工调用 `End()`。

### 5.4 消息消费

消息存在 OTel carrier 时，先提取生产者上下文，再创建 Consumer Span：

```text
message headers
    ↓ Extract(traceparent)
remote parent context
    ↓ Start(SpanKindConsumer)
consumer span
    ↓
handler / logger / DB / HTTP
```

没有传播头时才创建新的 Root Consumer Span。消息消费不能继续把任意 `X-Trace-Id` 当作远端父上下文。

### 5.5 异步 goroutine

现有 `ctx.Go` 和 `ctx.Fork` 继续负责：

- 继承当前 context value 和 SpanContext；
- 使用 `context.WithoutCancel` 避免请求结束立即取消后台任务。

需要可观测的独立异步步骤时，新增带名称的辅助函数：

```go
ctx.GoNamed(parent, "notify.send", func(c context.Context) {
    logger.Infof(c, "发送异步通知")
})
```

`GoNamed` 在 goroutine 内创建并结束子 Span。现有 `ctx.Go` 保持兼容，不强制自动创建无业务名称的 Span。

## 6. 基础组件改造设计

### 6.1 `trace/`：SDK 初始化

拟增加：

```text
trace/provider.go
trace/span.go
```

建议接口：

```go
type Config struct {
    ServiceName    string
    ServiceVersion string
    Environment    string
}

type Shutdown func(context.Context) error

func Init(ctx context.Context, cfg Config) (Shutdown, error)
```

`Init` 负责：

1. 创建 OTLP trace exporter；
2. 构造 Resource；
3. 创建 BatchSpanProcessor；
4. 创建 TracerProvider；
5. 注册全局 TracerProvider；
6. 注册 `TraceContext + Baggage` Propagator；
7. 返回可重复安全调用的 Shutdown 函数。

约束：

- 不在 `init()` 中建立网络连接；
- 不在基础组件中硬编码 service name、环境和 OTLP 地址；
- 初始化失败由业务启动流程决定是否终止应用；
- Shutdown 应使用带超时的独立 context；
- 测试允许注入内存 exporter 或自定义 TracerProvider。

### 6.2 `trace.Continue`

保留函数签名，修改优先级：

```text
有效 SpanContext
    → 使用真实 TraceID并同步旧 context key
否则旧 trace_id / hint
    → 只做兼容保存，不伪造 Remote SpanContext
否则
    → OTel 未启用场景生成兼容 TraceID
```

当前 `withOTEL` 构造 Remote SpanContext 的实现应删除或改为仅通过标准 Propagator 提取远端上下文。

### 6.3 `ctx.GetTraceId`

保留函数名和调用方式：

```go
func GetTraceId(ctx context.Context) string {
    sc := oteltrace.SpanContextFromContext(ctx)
    if sc.IsValid() {
        return sc.TraceID().String()
    }
    return strings.TrimSpace(GetValue(ctx, types.TraceIdKey))
}
```

`ctx.WithTraceId` 保留兼容，但标记 deprecated；它不能创建正式 OTel 父上下文。

### 6.4 `ctx.Run`

保留当前调用：

```go
ctx.Run(parent, func(c context.Context) {
    // existing code
})
```

内部改为创建、结束真实 Root Span。新增：

```go
ctx.RunNamed(parent, name, fn)
ctx.GoNamed(parent, name, fn)
```

旧 `Run` 可以使用稳定的通用 Span 名称兼容，但新业务应优先使用 `RunNamed`。

`ctx.StartTrace` 当前只返回 context，无法可靠结束 Span。接入 OTel 后应标记 deprecated，引导调用者使用 `Run`、`RunNamed` 或显式 `tracer.Start/defer span.End`。

### 6.5 `client/request`

`request.Inject` 增加标准上下文注入，同时保留旧业务头：

```go
otel.GetTextMapPropagator().Inject(
    ctx,
    propagation.HeaderCarrier(header),
)

// 兼容 X-Trace-Id / X-Req-Id
```

`X-Trace-Id` 的值必须来自当前 SpanContext；没有有效 Span 时才读取旧 context value。

### 6.6 `middleware/goframe`

保留：

- `SetMiddleware`；
- CORS、Header、Cookies；
- `Recover`；
- request_id 和响应头写回。

调整：

- `bindTrace` 不再根据 `X-Trace-Id` 替换活动 SpanContext；
- 响应 `X-Trace-Id` 取当前 Server Span 的 TraceID；
- `Recover` 捕获 panic 时，对当前 Span 调用 `RecordError` 并设置 Error Status。

### 6.7 `middleware/ugin`

保留现有 `SetMiddleware` 签名，新增可复用包装：

```go
func Tracing(service string) gin.HandlerFunc
```

业务调用：

```go
r.Use(ugin.Tracing("bssvc"))
ugin.SetMiddleware(r, ugin.ExceptionHandler())
```

异常处理同样需要将 panic 或系统错误记录到当前 Span。

### 6.8 `client/httputil`

继续复用 GoFrame `gclient` 内置 OTel instrumentation，不再增加第二个 Client Span。

现有 `MiddlewarePropagate` 只负责：

- `X-Req-Id`；
- 兼容 `X-Trace-Id`；
- 其他 element 业务 Header。

标准 `traceparent` 由 `gclient` 内部 OTel middleware 注入。

### 6.9 `proxy`

保留连接池、超时、重试、熔断和路由逻辑，只包装 Transport：

```go
baseTransport := &http.Transport{
    // 当前已有连接池配置
}

client := &http.Client{
    Timeout:   timeout,
    Transport: otelhttp.NewTransport(baseTransport),
}
```

这样每次转发都产生 Client Span，并根据目标请求当前 context 注入正确的传播头。

不能直接复制入站 `traceparent` 作为出站最终值；出站 Header 应由 Client Span 对应的上下文重新注入。

### 6.10 `logger`

保留现有 logger 接口和输出格式：

```go
logger.Infof(ctx, "处理完成")
```

元数据优先级：

```text
当前有效 SpanContext.TraceID
    ↓
旧 context trace_id 兜底
```

`request_id` 继续输出，不与 TraceID 合并。第一阶段不要求业务日志手工增加 `trace_id=`。

## 7. 上层业务接入约定

### 7.1 应用启动

每个进程在创建服务器和后台任务前初始化：

```go
shutdown, err := elementtrace.Init(ctx, elementtrace.Config{
    ServiceName:    "bssvc",
    ServiceVersion: version,
    Environment:    "test",
})
if err != nil {
    return err
}

defer func() {
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = shutdown(shutdownCtx)
}()
```

不同进程必须使用不同 `service.name`，例如：

```text
aro-http
aro-scan
aro-sub
aro-cron
bssvc
chain
risk
```

### 7.2 Kubernetes 环境变量

```yaml
env:
  - name: OTEL_SERVICE_NAME
    value: "bssvc"
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "http://diag-otlp.diag.svc.cluster.local:4318"
  - name: OTEL_EXPORTER_OTLP_PROTOCOL
    value: "http/protobuf"
  - name: OTEL_TRACES_EXPORTER
    value: "otlp"
  - name: OTEL_PROPAGATORS
    value: "tracecontext,baggage"
  - name: OTEL_RESOURCE_ATTRIBUTES
    value: "service.namespace=hiit,deployment.environment.name=test"
```

SDK 初始化实现必须明确支持这些环境变量；仅设置环境变量但不调用 `trace.Init` 不会自动上报。

### 7.3 业务代码

普通 HTTP 业务代码无需改写：

```go
func handle(ctx context.Context) error {
    logger.Infof(ctx, "开始处理")
    return service.Execute(ctx)
}
```

只有需要表达业务步骤时才增加手工 Span：

```go
ctx, span := tracer.Start(ctx, "order.confirm")
defer span.End()
```

Span 名称使用稳定的低基数名称，订单号、用户 ID、交易哈希等应作为 attribute，而不是拼进 Span 名称。

## 8. 兼容策略

### 8.1 保持不变

- `ctx.GetTraceId`；
- `ctx.GetReqId`；
- `ctx.Run`；
- `ctx.Go`；
- `ctx.Fork`；
- `logger.Infof/Errorf/...`；
- GoFrame/Gin 现有 Header 和异常处理中间件；
- `X-Req-Id`；
- `X-Trace-Id` 响应头和出站兼容头。

### 8.2 行为变化

- `trace_id` 优先来自真实 OTel Span；
- 入站 `X-Trace-Id` 不再决定 OTel 父上下文；
- 标准跨服务传播改为 `traceparent`；
- `ctx.Run` 从“只生成 ID”升级为“创建并结束 Span”；
- GoFrame 日志不再出现自定义 TraceID 与 OTel TraceID 两套值。

### 8.3 后续废弃候选

稳定运行一个迁移周期后再评估：

- `ctx.WithTraceId`；
- `ctx.StartTrace`；
- 自定义 `trace/id.go`；
- 将 `X-Trace-Id` 作为出站请求头的兼容行为。

第一阶段不直接删除，避免破坏已有业务系统。

## 9. 实施阶段

### 阶段一：基础能力

1. 增加 `trace.Init/Shutdown`；
2. 修正 `trace.Continue` 和 `ctx.GetTraceId`；
3. 改造 `ctx.Run`，增加 `RunNamed`；
4. 调整 logger TraceID 优先级；
5. 增加单元测试和内存 exporter 测试。

完成后，GoFrame HTTP、`gclient` 和 `gdb` 应能形成真实链路。

### 阶段二：补齐入口和客户端

1. Gin 增加 `otelgin` 包装；
2. `proxy` 增加 `otelhttp.Transport`；
3. `request.Inject` 增加标准传播；
4. 异常处理中记录 Span Error；
5. 增加 HTTP 跨服务集成测试。

### 阶段三：业务 Span

按实际价值逐步增加：

- 定时任务；
- 消息生产和消费；
- 区块/交易扫描；
- 关键订单、支付和风控步骤；
- 未被框架自动覆盖的 Redis、MQ 或第三方客户端。

不要求一次性改造所有函数。

### 阶段四：日志与 Grafana 关联

1. 确认日志中的 TraceID 来自当前 Span；
2. Alloy 将 `trace_id` 写入 structured metadata；
3. Grafana 配置 Logs → Trace；
4. Tempo 配置 Trace → Logs；
5. 再评估 JSON 日志和 `span_id`。

## 10. 测试方案

### 10.1 单元测试

- 有有效 SpanContext 时，`ctx.GetTraceId` 返回 Span TraceID；
- 旧 context value 仅在没有 SpanContext 时生效；
- `trace.Continue` 不覆盖已有活动 Span；
- `ctx.Run` 创建 Span 并在回调结束后导出；
- `request.Inject` 同时注入 `traceparent`、`X-Trace-Id` 和 `X-Req-Id`；
- OTel 未初始化时所有公共 API 不 panic。

### 10.2 集成测试

使用内存 exporter 或测试 Collector 验证：

```text
HTTP Server Span
├─ 业务 Span
├─ gclient Client Span
│  └─ 下游 Server Span
└─ gdb Span
```

同时验证 Gin、GoFrame、proxy、cron 和消息消费场景。

### 10.3 回归测试

- 现有 HTTP Header、request_id 和异常响应不变化；
- 未配置 OTel 时应用仍可启动；
- exporter 不可用时不阻塞请求主流程；
- Shutdown 能在超时范围内刷新剩余 Span；
- 不产生同一次调用的重复 Server/Client Span。

## 11. 验收标准

- 每个业务进程使用正确的 `service.name`；
- HTTP 入站产生 Server Span；
- `gclient`、proxy 和数据库调用产生正确的子 Span；
- 定时任务和消息消费具有明确的 Root/Consumer Span；
- 跨服务使用 `traceparent` 保持同一个 TraceID；
- 日志 `trace_id` 与 Tempo 中 TraceID 一致；
- panic 和关键错误记录到对应 Span；
- 应用退出时执行 Provider Shutdown；
- 业务现有 `ctx` 和 logger 调用无需整体重写；
- Loki 不把 `trace_id` 设置为 stream label。

## 12. 最终边界

```text
element
├─ 提供可复用的 OTel 初始化和 instrumentation
├─ 保持 ctx/logger/middleware/client 公共入口稳定
└─ 不决定具体业务身份和业务 Span 语义

上层业务系统
├─ 显式初始化和关闭 SDK
├─ 提供 service.name / version / environment
├─ 配置 exporter 和采样策略
└─ 为关键业务步骤增加命名 Span 和 attributes
```

该方案的实施重点是替换现有 TraceID 底层语义并补齐缺口，而不是重新设计整套基础组件。
