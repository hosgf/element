# 项目架构与代码规范

## 异常处理统一规范

### 1. 统一使用 `uerrors.BizError`

项目中所有错误处理应统一使用 `uerrors.BizError`，不再使用：
- ❌ `gerror` (GoFrame错误)
- ❌ `ProxyError` (代理错误，应迁移到BizError)
- ❌ 直接返回 `error` (应包装为BizError)

### 2. 错误创建方式

#### 基础错误创建
```go
// 系统错误
err := uerrors.NewSystemError("操作失败", "详细描述")

// 业务逻辑错误
err := uerrors.NewBizLogicError(code, "消息", "详细信息")

// 验证错误
err := uerrors.NewValidationError("字段名", "验证消息")
```

#### Kubernetes错误包装
```go
// 包装Kubernetes API错误
err := uerrors.WrapKubernetesError(ctx, k8sErr, "操作名称")

// 创建Kubernetes错误
err := uerrors.NewKubernetesError(ctx, "操作名称", "错误消息", "详细信息")
```

#### 带上下文的错误
```go
// 自动从context获取RequestID和UserID
err := uerrors.WrapKubernetesError(ctx, k8sErr, "操作名称")
// err.RequestID 和 err.UserID 会自动填充
```

### 3. 错误处理最佳实践

```go
// ✅ 正确：使用uerrors包装错误
func (o *operation) DoSomething(ctx context.Context) error {
    result, err := o.api.DoSomething(ctx)
    if err != nil {
        return uerrors.WrapKubernetesError(ctx, err, "执行操作")
    }
    return nil
}

// ❌ 错误：直接返回gerror
func (o *operation) DoSomething(ctx context.Context) error {
    result, err := o.api.DoSomething(ctx)
    if err != nil {
        return gerror.NewCodef(gcode.CodeNotImplemented, "Failed: %v", err)
    }
    return nil
}
```

## Context 使用规范

### 1. 所有异步操作必须接受 context.Context

```go
// ✅ 正确
func (o *operation) List(ctx context.Context, namespace string) ([]Item, error)

// ❌ 错误
func (o *operation) List(namespace string) ([]Item, error)
```

### 2. Context 传递规范

- 所有函数的第一参数应该是 `context.Context`
- Context 应该从请求开始一直传递到最底层
- 不要创建新的 context，除非需要设置超时或取消

### 3. Context 辅助函数

使用 `uerrors` 包提供的辅助函数：

```go
// 获取RequestID
requestID := uerrors.GetRequestID(ctx)

// 设置RequestID
ctx = uerrors.WithRequestID(ctx, requestID)

// 获取UserID
userID := uerrors.GetUserID(ctx)

// 设置UserID
ctx = uerrors.WithUserID(ctx, userID)
```

### 4. Context 在错误处理中的使用

```go
// 错误会自动从context获取RequestID和UserID
err := uerrors.WrapKubernetesError(ctx, k8sErr, "操作名称")
// err.RequestID 和 err.UserID 已自动填充
```

## 迁移计划

### 阶段1：基础设施 ✅
- [x] 创建 `exception/context.go` 辅助函数
- [x] 创建 `exception/helper.go` 中的Kubernetes错误包装函数
- [x] 更新 `client/k8s/k8s.go` 基础错误处理

### 阶段2：Kubernetes客户端 (进行中)
- [x] 更新 `k8s_nodes.go`
- [x] 更新 `k8s_metrics.go`
- [x] 更新 `k8s_namespace.go`
- [x] 更新 `k8s_pod_templates.go`
- [x] 更新 `k8s_service.go` (部分)
- [ ] 更新 `k8s_pods.go` (剩余gerror使用)
- [ ] 更新 `k8s_storage.go` (剩余gerror使用)
- [ ] 更新 `k8s_storage_resource.go` (剩余gerror使用)
- [ ] 更新 `k8s_process.go` (剩余gerror使用)

### 阶段3：代理模块
- [ ] 将 `ProxyError` 迁移到 `BizError`
- [ ] 更新 `proxy/error_handler.go`
- [ ] 更新所有使用 `ProxyError` 的地方

### 阶段4：其他模块
- [ ] 检查并更新其他模块的错误处理
- [ ] 确保所有异步操作都接受 context

## 代码检查清单

在提交代码前，请检查：

- [ ] 所有错误都使用 `uerrors.BizError` 或其包装函数
- [ ] 所有异步操作都接受 `context.Context` 作为第一参数
- [ ] Context 正确传递，没有丢失
- [ ] 错误消息包含足够的上下文信息（namespace, name等）
- [ ] 使用 `uerrors.WrapKubernetesError` 包装Kubernetes错误
- [ ] 移除了所有 `gerror` 和 `gcode` 的直接使用

## 示例代码

### 完整的函数示例

```go
func (o *nodesOperation) Top(ctx context.Context) ([]*Node, error) {
    if o.err != nil {
        return nil, o.err
    }
    
    // 使用context传递
    datas, err := o.api.CoreV1().Nodes().List(ctx, v1.ListOptions{})
    if err != nil {
        // 使用uerrors包装错误，自动获取RequestID
        return nil, uerrors.WrapKubernetesError(ctx, err, "获取节点列表")
    }
    
    // ... 处理逻辑 ...
    
    return nodes, nil
}
```

### 错误检查示例

```go
func (o *operation) isExist(ctx context.Context, value interface{}, err error, operation string) (bool, error) {
    if err != nil {
        if errors.IsNotFound(err) {
            return false, nil
        }
        // 使用uerrors包装，包含操作上下文
        return false, uerrors.WrapKubernetesError(ctx, err, operation)
    }
    return value != nil, nil
}
```
