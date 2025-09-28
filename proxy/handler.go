package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
)

// ============================================================================
// 网关实例管理
// ============================================================================
func NewGateway(prefix string, ignore ...string) *Gateway {
	gateway := &Gateway{
		prefix:      prefix,
		name:        "",
		headers:     map[string]string{},
		routes:      map[string]*Route{},
		ignore:      map[string]interface{}{},
		middlewares: []MiddlewareItem{},
	}
	if len(ignore) > 0 {
		for _, i := range ignore {
			gateway.ignore[i] = nil
		}
	}
	return gateway
}

// ============================================================================
// 路由管理
// ============================================================================

func (gw *Gateway) Routes() struct {
	Prefix  string            `json:"prefix,omitempty"`
	Ignore  []string          `json:"ignore,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Routes  map[string]string `json:"routes,omitempty"`
} {
	routes := make(map[string]string)
	for _, route := range gw.routes {
		routes[route.Path] = route.Address
	}
	return struct {
		Prefix  string            `json:"prefix,omitempty"`
		Ignore  []string          `json:"ignore,omitempty"`
		Headers map[string]string `json:"headers,omitempty"`
		Routes  map[string]string `json:"routes,omitempty"`
	}{
		Prefix:  gw.prefix,
		Headers: gw.headers,
		Ignore:  gw.toIgnores(),
		Routes:  routes,
	}
}

func (gw *Gateway) SetName(name string) *Gateway {
	if len(name) == 0 {
		return gw
	}
	gw.name = name
	return gw
}

func (gw *Gateway) SetHeader(key, value string) *Gateway {
	if len(key) == 0 || len(value) == 0 {
		return gw
	}
	gw.headers[key] = value
	return gw
}

func (gw *Gateway) SetHeaderToRequest(t *http.Request) {
	for k, v := range gw.headers {
		t.Header.Set(k, v)
	}
}

func (gw *Gateway) CreateRoute(sameToken, name, address string, includes, excludes []string) *Gateway {
	// 验证路由参数
	if name == "" {
		logger.Errorf(context.Background(), "route name cannot be empty")
		return gw
	}
	if address == "" {
		logger.Errorf(context.Background(), "route address cannot be empty for route: %s", name)
		return gw
	}

	// 验证地址格式
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		logger.Errorf(context.Background(), "route address must start with http:// or https:// for route: %s", name)
		return gw
	}

	route := &Route{
		SameToken:   sameToken,
		Name:        name,
		Path:        fmt.Sprintf("%s/%s", gw.prefix, name),
		Address:     address,
		Includes:    includes,
		Excludes:    excludes,
		middlewares: defaultMiddlewareItems(),
	}
	gw.routes[route.Path] = route
	logger.Infof(context.Background(), "route created: %s -> %s", route.Path, address)
	return gw
}

func (gw *Gateway) match(path string) (*Route, bool) {
	// 优先精确匹配
	if route, exists := gw.routes[path]; exists {
		return route, true
	}

	// 前缀匹配，选择最长匹配
	var bestMatch *Route
	var bestPath string

	for k, route := range gw.routes {
		if strings.HasPrefix(path, k) && len(k) > len(bestPath) {
			bestMatch = route
			bestPath = k
		}
	}

	return bestMatch, bestMatch != nil
}

// ============================================================================
// 请求处理
// ============================================================================

func (gw *Gateway) Execute(r *ghttp.Request) {
	gw.handler(r, gw.execute)
}

func (gw *Gateway) execute(o *ghttp.Request, t *http.Request) {
	startTime := time.Now()
	circuitBreakerOpen := false
	success := false

	defer func() {
		latency := time.Since(startTime)
		gw.recordMetrics(success, latency, circuitBreakerOpen)
	}()

	// 设置请求头
	gw.setupRequestHeaders(o, t)

	// 从连接池获取客户端
	client := getHTTPClient()
	defer putHTTPClient(client)

	// 检查熔断器状态
	routeKey := t.URL.Host
	if !gw.isCircuitBreakerOpen(routeKey) {
		config := gw.getConfig()
		retryConfig := config.GetRetryConfig()
		resp, err := gw.doRequestWithRetry(client, t, retryConfig.MaxRetries)
		if err != nil {
			gw.response(o, resp, err)
			return
		}
		defer resp.Body.Close()

		gw.handleResponse(o, resp)
		success = true
	} else {
		// 熔断器开启，直接返回错误
		circuitBreakerOpen = true
		gw.response(o, nil, errors.New("服务暂时不可用，熔断器已开启"))
	}
}

// 设置请求头
func (gw *Gateway) setupRequestHeaders(o *ghttp.Request, t *http.Request) {
	for key, values := range o.Header {
		for _, value := range values {
			t.Header.Add(key, value)
		}
	}
	gw.SetHeaderToRequest(t)
}

// 错误响应处理
func (gw *Gateway) response(o *ghttp.Request, resp *http.Response, err error) {
	if err != nil {
		target := gw.getTargetURL(resp, o)
		code, message := gw.classifyError(err, o.Context(), target)

		res := result.NewResponse()
		res.Code = code
		res.Message = message
		o.Response.WriteJson(res)
	}
}

// 获取目标URL
func (gw *Gateway) getTargetURL(resp *http.Response, o *ghttp.Request) string {
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		return resp.Request.URL.String()
	}
	if o != nil {
		return o.URL.String()
	}
	return ""
}

// 错误分类
func (gw *Gateway) classifyError(err error, ctx context.Context, target string) (int, string) {
	// DNS/连接错误
	var ne *net.OpError
	if errors.As(err, &ne) {
		logger.Errorf(ctx, "proxy dial error: %v, target=%s", err, target)
		return SC_GATEWAY, "服务连接失败，请稍后重试"
	}

	// 超时错误
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		logger.Errorf(ctx, "proxy timeout: %v, target=%s", err, target)
		return SC_TIMEOUT, "请求超时，请稍后再试"
	}

	// 熔断器错误
	if strings.Contains(err.Error(), "熔断器已开启") {
		logger.Errorf(ctx, "circuit breaker open: %v, target=%s", err, target)
		return SC_SERVICE_ERROR, "服务暂时不可用，请稍后重试"
	}

	// 其他错误
	logger.Errorf(ctx, "proxy error: %v, target=%s", err, target)
	return SC_FAILURE, "服务器内部错误"
}

// 请求处理器
func (gw *Gateway) handler(o *ghttp.Request, finalHandler func(o *ghttp.Request, t *http.Request)) {
	path := o.URL.Path

	// 检查忽略路径
	if gw.isIgnoredPath(path) {
		return
	}

	// 匹配路由
	route, found := gw.match(path)
	if !found {
		gw.handleRouteNotFound(o)
		return
	}

	// 创建代理请求
	proxyReq, err := gw.createProxyRequest(o, route)
	if err != nil {
		gw.handleRequestCreationError(o, err)
		return
	}

	// 执行中间件链
	gw.executeMiddlewareChain(o, proxyReq, route, finalHandler)
}

// 检查是否忽略路径
func (gw *Gateway) isIgnoredPath(path string) bool {
	_, ok := gw.ignore[path]
	return ok
}

// 处理路由未找到
func (gw *Gateway) handleRouteNotFound(o *ghttp.Request) {
	res := result.NewResponse()
	res.Code = SC_NOT_FOUND
	res.Message = "未找到匹配的服务"
	o.Response.WriteJson(res)
	requestLogging(o, gerror.New("未找到匹配的服务"))
}

// 创建代理请求
func (gw *Gateway) createProxyRequest(o *ghttp.Request, route *Route) (*http.Request, error) {
	proxyURL := route.Address + gstr.TrimLeftStr(o.URL.RequestURI(), route.Path)
	return http.NewRequestWithContext(o.Context(), o.Method, proxyURL, o.Body)
}

// 处理请求创建错误
func (gw *Gateway) handleRequestCreationError(o *ghttp.Request, err error) {
	errmsg := fmt.Sprintf("创建请求失败: %v", err)
	res := result.NewResponse()
	res.Code = SC_BAD_GATEWAY
	res.Message = errmsg
	o.Response.WriteJson(res)
	requestLogging(o, gerror.New(errmsg))
}

// 执行中间件链
func (gw *Gateway) executeMiddlewareChain(o *ghttp.Request, t *http.Request, route *Route, finalHandler func(o *ghttp.Request, t *http.Request)) {
	// 合并全局中间件和路由中间件
	allMiddlewareItems := make([]MiddlewareItem, 0, len(gw.middlewares)+len(route.middlewares))
	allMiddlewareItems = append(allMiddlewareItems, gw.middlewares...)
	allMiddlewareItems = append(allMiddlewareItems, route.middlewares...)

	// 对中间件进行排序
	allMiddlewares := sortMiddlewares(allMiddlewareItems)

	if len(allMiddlewares) == 0 {
		finalHandler(o, t)
		return
	}

	var execute func(index int)
	execute = func(index int) {
		if index < len(allMiddlewares) {
			allMiddlewares[index](o, t, route, func() {
				execute(index + 1)
			})
		} else {
			finalHandler(o, t)
		}
	}
	execute(0)
}

// ============================================================================
// HTTP请求执行
// ============================================================================

// 带重试的请求执行
func (gw *Gateway) doRequestWithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Do(req)
		if err == nil {
			// 请求成功，更新熔断器状态
			gw.updateCircuitBreaker(req.URL.Host, true)
			return resp, nil
		}

		lastErr = err

		// 记录失败
		gw.updateCircuitBreaker(req.URL.Host, false)

		// 如果是最后一次重试，直接返回错误
		if i == maxRetries-1 {
			break
		}

		// 指数退避重试
		backoff := time.Duration(1<<uint(i)) * time.Second
		time.Sleep(backoff)
	}

	return nil, lastErr
}

// 处理响应
func (gw *Gateway) handleResponse(o *ghttp.Request, resp *http.Response) {
	// 复制响应头
	gw.copyResponseHeaders(o, resp)

	// 设置HTTP状态码为200
	o.Response.WriteHeader(http.StatusOK)

	// 流式复制响应体
	if err := gw.streamResponseBody(o, resp); err != nil {
		gw.handleResponseBodyError(o, resp, err)
	}
}

// 复制响应头
func (gw *Gateway) copyResponseHeaders(o *ghttp.Request, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			o.Response.Header().Set(key, value)
		}
	}
}

// 流式复制响应体
func (gw *Gateway) streamResponseBody(o *ghttp.Request, resp *http.Response) error {
	_, err := io.Copy(o.Response.Writer, resp.Body)
	if err != nil {
		logger.Errorf(o.Context(), "proxy copy body failed: %v, target=%s", err, resp.Request.URL.String())
	}
	return err
}

// 处理响应体错误
func (gw *Gateway) handleResponseBodyError(o *ghttp.Request, resp *http.Response, err error) {
	// 如果流式复制失败，尝试读取完整响应体
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		res := result.NewResponse()
		res.Code = SC_FAILURE
		res.Message = "读取响应失败"
		o.Response.WriteJson(res)
		return
	}
	// 设置HTTP状态码为200
	o.Response.WriteHeader(http.StatusOK)
	o.Response.Write(body)
}
