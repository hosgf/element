# 网关代理服务使用文档

## 概述

网关代理服务是一个高性能的HTTP代理网关，支持路由转发、熔断器、重试机制和性能监控。

## 主要特性

- 🚀 **高性能**: HTTP连接池复用，支持高并发
- 🛡️ **熔断器**: 自动故障隔离，提高稳定性
- 🔄 **重试机制**: 智能重试，提高成功率
- 📊 **性能监控**: 实时指标和健康检查
- ⚙️ **可配置**: 灵活的参数配置
- 🔍 **路由匹配**: 精确和前缀匹配
- 📝 **日志记录**: 详细的请求和错误日志

## 快速开始

### 1. 基本使用

```go
package main

import (
    "wallet-service/internal/service/proxy"
    "wallet-service/internal/web/api"
    "github.com/gogf/gf/v2/net/ghttp"
)

func main() {
    s := ghttp.GetServer()
    
    // 注册代理API
    api.Proxy.RegisterRouter(s)
    
    // 创建路由
    proxy.Proxy.CreateRoute("token123", "user-service", "http://localhost:8081", []string{}, []string{})
    proxy.Proxy.CreateRoute("token123", "order-service", "http://localhost:8082", []string{}, []string{})
    
    s.Run()
}
```

### 2. 配置网关

```go
import "wallet-service/internal/service/proxy"

// 自定义配置
config := &proxy.GatewayConfig{
    Timeout:        30 * time.Second,
    MaxRetries:     3,
    FailureThreshold: 5,
    RecoveryTimeout: 30 * time.Second,
}

// 设置全局配置
proxy.SetConfig(config)
```

## 配置说明

### 主要配置参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `timeout` | time.Duration | 30s | HTTP请求超时时间 |
| `max_retries` | int | 3 | 最大重试次数 |
| `failure_threshold` | int | 5 | 熔断器失败阈值 |
| `recovery_timeout` | time.Duration | 30s | 熔断器恢复超时时间 |

## API 接口

### 1. 代理请求

所有匹配路由的请求会自动代理到对应的后端服务。

```
GET /api/user-service/users
POST /api/order-service/orders
PUT /api/user-service/users/123
DELETE /api/order-service/orders/456
```

### 2. 性能监控

#### 获取性能指标

```http
GET /proxy/metrics
```

响应示例：

```json
{
  "request_count": 1000,
  "success_count": 950,
  "failure_count": 50,
  "success_rate": "95.00%",
  "average_latency": "150ms"
}
```

#### 健康检查

```http
GET /proxy/health
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": 1640995200
}
```

## 路由管理

### 创建路由

```go
// 基本路由
proxy.Proxy.CreateRoute("token123", "user-service", "http://localhost:8081", []string{}, []string{})

// 带包含路径的路由
proxy.Proxy.CreateRoute("token123", "api-service", "http://localhost:8083", 
    []string{"v1/users", "v1/orders"}, []string{})

// 带排除路径的路由
proxy.Proxy.CreateRoute("token123", "public-api", "http://localhost:8084", 
    []string{}, []string{"admin", "internal"})
```

### 路由参数说明

- `sameToken`: 共享令牌，用于身份验证
- `name`: 路由名称，会生成路径 `/api/{name}`
- `address`: 后端服务地址
- `includes`: 包含的路径列表（可选）
- `excludes`: 排除的路径列表（可选）

### 路由匹配规则

1. **精确匹配**: 优先匹配完全相同的路径
2. **前缀匹配**: 如果没有精确匹配，则进行前缀匹配
3. **最长匹配**: 多个前缀匹配时，选择最长的匹配

## 中间件

### 默认中间件

网关内置了以下中间件：

1. **日志中间件**: 记录请求和响应信息
2. **认证中间件**: 处理身份验证
3. **令牌中间件**: 添加共享令牌到请求头
4. **响应中间件**: 处理响应数据

### 全局中间件管理

```go
// 添加全局中间件
gateway.AddMiddleware(CustomMiddleware)

// 添加带排序的全局中间件
gateway.AddMiddlewareWithSort(CORSMiddleware, 10)

// 设置全局中间件
gateway.SetMiddlewares([]MiddlewareFunc{CustomMiddleware1, CustomMiddleware2})

// 获取全局中间件
middlewares := gateway.GetMiddlewares()

// 清除全局中间件
gateway.ClearMiddlewares()
```

### 路由中间件管理

```go
// 为特定路由添加中间件
gateway.AddRouteMiddleware("user-service", CustomMiddleware)

// 为特定路由添加带排序的中间件
gateway.AddRouteMiddlewareWithSort("user-service", CORSMiddleware, 5)

// 设置路由中间件
gateway.SetRouteMiddlewares("user-service", []MiddlewareFunc{CustomMiddleware1, CustomMiddleware2})

// 获取路由中间件
routeMiddlewares := gateway.GetRouteMiddlewares("user-service")

// 清除路由中间件
gateway.ClearRouteMiddlewares("user-service")
```

### 中间件排序

中间件按 `Sort` 字段排序执行：

- **负数**: 最早执行（如 `Sort: -1`）
- **正数**: 按数值大小排序
- **零值**: 保持插入顺序，排在非零值之后

### 自定义中间件

```go
func CustomMiddleware(o *ghttp.Request, t *http.Request, route *Route, next func()) {
    // 前置处理
    t.Header.Set("X-Custom-Header", "custom-value")
    
    // 调用下一个中间件
    next()
    
    // 后置处理
    // ...
}
```

## 错误处理

### 错误类型

| 错误码 | 说明 | 处理方式 |
|--------|------|----------|
| 200 | 成功 | 正常返回 |
| 400 | 网关错误 | 连接失败，建议重试 |
| 408 | 超时错误 | 请求超时，建议重试 |
| 502 | 服务错误 | 熔断器开启，稍后重试 |
| 500 | 内部错误 | 服务器内部错误 |

### 错误响应格式

```json
{
  "code": 400,
  "message": "服务连接失败，请稍后重试",
  "data": null
}
```

**注意**: 所有HTTP响应状态码都为200，具体错误码在响应体的 `code` 字段中。

## 最佳实践

### 1. 路由设计

- 使用有意义的路由名称
- 避免路由冲突
- 合理使用包含和排除路径

### 2. 配置管理

- 根据环境调整配置参数
- 定期评估和优化配置

### 3. 监控告警

- 定期检查 `/proxy/metrics` 接口
- 关注成功率和平均延迟
- 设置告警机制

