## 全局异常处理使用指南（exception + middleware）

本文档说明如何在 GoFrame 与 Gin 中接入全局异常处理，并在业务代码中像 Java 一样通过抛错统一返回给前端。

### 设计约定
- HTTP 状态码固定返回 200
- 业务错误通过 `model/result.Response` 的 `Code` 与 `Message` 表达（仅顶层）
- 非生产环境（config.Debug = true）会附带 `X-Request-ID` 响应头与 `X-Response-Time`
- 请求链路统一注入请求 ID：`X-Request-ID`

### 目录结构
- 核心能力（类型/抛错辅助）在 `exception/`
  - `types.go`: `BizError` 类型、错误类型与级别枚举、`IsBizError`
  - `errors.go`: 常用错误构造（校验/系统/网络/数据库/外部服务等）
  - `helper.go`: 便捷抛错 API（`ThrowBiz`、`ThrowValidation`、`ThrowSystem`、`Must`、`PanicIf`）
  - `requestid.go`: `GenerateRequestID()`
- 中间件在 `middleware/`
  - GoFrame: `middleware/goframe/exception.go` → `ExceptionHandler`
  - Gin: `middleware/ugin/exception.go` → `ExceptionHandler`

---

## 在 GoFrame 中使用

### 1) 注册中间件
```go
import (
    "github.com/gogf/gf/v2/net/ghttp"
    mf "github.com/hosgf/element/middleware/goframe"
)

func main() {
    s := ghttp.GetServer()
    // 绑定全局异常中间件（建议尽量靠前）
    mf.UseException(s)

    s.BindHandler("/demo", func(r *ghttp.Request) {
        // 业务中直接抛错，由中间件统一捕获并返回
        // 比如参数校验失败：
        // exception.ThrowValidation("name", "不能为空")
        // 或业务错误：
        // exception.ThrowBiz(3001, "资源不存在")
        r.Response.WriteJson(map[string]string{"ok": "true"})
    })

    s.Run()
}
```

也可直接：
```go
s.Use(mf.ExceptionHandler)
```

### 2) 业务侧抛错示例
```go
import "github.com/hosgf/element/exception"

func CreateUser(name string) {
    if len(name) == 0 {
        exception.ThrowValidation("name", "不能为空")
    }
}

func DoQuery() {
    data, err := repo.Query()
    exception.Must(err, "查询失败") // 有错自动转 BizError panic
    // ...
}

func DoBiz() {
    exception.ThrowBiz(3002, "资源冲突", "原因: 唯一索引冲突")
}
```

---

## 在 Gin 中使用

### 1) 注册中间件
```go
import (
    "github.com/gin-gonic/gin"
    mg "github.com/hosgf/element/middleware/ugin"
)

func main() {
    r := gin.New()
    // 绑定全局异常中间件（建议尽量靠前）
    r.Use(mg.ExceptionHandler())

    r.GET("/demo", func(c *gin.Context) {
        // 业务中直接抛错，由中间件统一捕获并返回
        // exception.ThrowBiz(3001, "资源不存在")
        c.JSON(200, gin.H{"ok": true})
    })

    _ = r.Run(":8080")
}
```

### 2) 业务侧抛错与 GoFrame 相同
- 同样使用 `exception.ThrowBiz/ThrowValidation/ThrowSystem/Must/PanicIf`

---

## 返回体格式（统一）
- HTTP 状态码固定 `200`
- Body 为 `model/result.Response` 结构，仅包含顶层字段：

```json
{
  "code": 3001,
  "message": "资源不存在"
}
```

说明：
- 错误类型/级别、请求 ID、时间戳等调试信息仅写入日志；响应体不再包含 `data`。
- 如需在开发环境查看详细信息，请通过日志系统或接入 `SetErrorNotifier` 的告警渠道。

---

## 常用辅助 API
- `exception.ThrowBiz(code int, message string, details ...string)`：抛出业务错误
- `exception.ThrowValidation(field, message string)`：抛出参数校验错误
- `exception.ThrowSystem(message string, details ...string)`：抛出系统错误
- `exception.Must(err error, message string)`：若 err 非空则包装为系统错误并 panic
- `exception.PanicIf(condition bool, err *exception.BizError)`：条件成立时抛出指定错误

## 自定义与扩展
- 可通过中间件内部 `SetErrorNotifier` 注入通知回调（例如上报到监控/告警系统）
- `config.Debug` 为 true 时，仍仅返回顶层 code/message；详细信息输出到日志
- 如需额外错误类型/级别，可在 `exception/types.go` 中扩展枚举

## FAQ
- Q: 为什么 HTTP 总是 200？
  - A: 统一由业务码 `code` 表达业务状态，便于前端与网关层统一处理与兜底。
- Q: 如何拿到请求 ID？
  - A: 从响应头 `X-Request-ID` 获取；在日志中也会包含该 ID 以便排障。
