package proxy

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/logger"
)

type MiddlewareFunc func(o *ghttp.Request, t *http.Request, route *Route, next func())

func defaultMiddlewareItems() []MiddlewareItem {
	return []MiddlewareItem{
		{Middleware: LoggerMiddleware, Sort: -99}, // 最早执行
		{Middleware: AuthMiddleware, Sort: -39},
		{Middleware: SameMiddleware, Sort: -1},
		{Middleware: ResponseMiddleware, Sort: 999}, // 最后执行
	}
}

func SameMiddleware(o *ghttp.Request, t *http.Request, route *Route, next func()) {
	t.Header.Set(request.HeaderSameToken.String(), route.SameToken)
	next()
}

func AuthMiddleware(o *ghttp.Request, t *http.Request, route *Route, next func()) {
	//token := o.Header.Get("Authorization")
	//if token == "" {
	//	res := result.NewResponse()
	//	res.Code = 401
	//	res.Message = "未授权访问"
	//	o.Response.WriteJson(res)
	//	return
	//}
	next()
}

func requestLogging(o *ghttp.Request, err error) {
	logger.RequestLogging(o, err)
}

func LoggerMiddleware(o *ghttp.Request, t *http.Request, route *Route, next func()) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)

		// 使用预分配的缓冲区减少内存分配
		logData := make([]string, 0, 8)
		logData = append(logData, fmt.Sprintf("📌 [请求处理] %s %s", t.Method, t.Header.Get("Content-Type")))

		if t != nil {
			logData = append(logData, fmt.Sprintf("Target: %s", t.URL.String()))
		}
		logData = append(logData, fmt.Sprintf("Origin: %s", o.URL.String()))

		// 只记录关键头部信息，避免内存浪费
		if t != nil {
			keyHeaders := []request.Header{request.HeaderReqToken, "User-Agent"}
			for _, key := range keyHeaders {
				if value := t.Header.Get(key.String()); value != "" {
					logData = append(logData, fmt.Sprintf("%s: %s", key, value))
				}
			}
			if value := t.Header.Get("traceparent"); value != "" {
				logData = append(logData, fmt.Sprintf("traceparent: %s", value))
			}
		}

		logData = append(logData, fmt.Sprintf("Duration: %s", duration))

		if rec := recover(); rec != nil {
			logData = append(logData, fmt.Sprintf("Response: %s", rec))
		} else {
			if t != nil && t.Response != nil {
				logData = append(logData, fmt.Sprintf("Response: %s", t.Response.Status))
			} else {
				logData = append(logData, fmt.Sprintf("Status: %d", o.Response.Status))
			}
		}

		// 使用strings.Join减少内存分配
		logger.Infof(o.Context(), "%s", strings.Join(logData, " | "))
	}()
	next()
}

func ResponseMiddleware(o *ghttp.Request, t *http.Request, route *Route, next func()) {
	next()
	// 响应处理完成
}

// ============================================================================
// 全局中间件管理
// ============================================================================

// AddMiddleware 添加全局中间件
func (gw *Gateway) AddMiddleware(middleware MiddlewareFunc) *Gateway {
	if middleware != nil {
		gw.middlewares = append(gw.middlewares, MiddlewareItem{Middleware: middleware, Sort: 0})
	}
	return gw
}

// AddMiddlewareWithSort 添加带排序权重的全局中间件
func (gw *Gateway) AddMiddlewareWithSort(middleware MiddlewareFunc, sort int) *Gateway {
	if middleware != nil {
		gw.middlewares = append(gw.middlewares, MiddlewareItem{Middleware: middleware, Sort: sort})
	}
	return gw
}

// SetMiddlewares 设置全局中间件列表（替换现有中间件）
func (gw *Gateway) SetMiddlewares(middlewares []MiddlewareFunc) *Gateway {
	gw.middlewares = make([]MiddlewareItem, 0, len(middlewares))
	for _, middleware := range middlewares {
		if middleware != nil {
			gw.middlewares = append(gw.middlewares, MiddlewareItem{Middleware: middleware, Sort: 0})
		}
	}
	return gw
}

// SetMiddlewareItems 设置全局中间件项列表（替换现有中间件）
func (gw *Gateway) SetMiddlewareItems(middlewareItems []MiddlewareItem) *Gateway {
	gw.middlewares = make([]MiddlewareItem, 0, len(middlewareItems))
	for _, item := range middlewareItems {
		if item.Middleware != nil {
			gw.middlewares = append(gw.middlewares, item)
		}
	}
	return gw
}

// GetMiddlewares 获取全局中间件列表
func (gw *Gateway) GetMiddlewares() []MiddlewareFunc {
	return sortMiddlewares(gw.middlewares)
}

// GetMiddlewareItems 获取全局中间件项列表
func (gw *Gateway) GetMiddlewareItems() []MiddlewareItem {
	return gw.middlewares
}

// ClearMiddlewares 清空全局中间件
func (gw *Gateway) ClearMiddlewares() *Gateway {
	gw.middlewares = []MiddlewareItem{}
	return gw
}

// RemoveMiddleware 移除指定位置的全局中间件
func (gw *Gateway) RemoveMiddleware(index int) *Gateway {
	if index >= 0 && index < len(gw.middlewares) {
		gw.middlewares = append(gw.middlewares[:index], gw.middlewares[index+1:]...)
	}
	return gw
}

// InsertMiddleware 在指定位置插入全局中间件
func (gw *Gateway) InsertMiddleware(index int, middleware MiddlewareFunc) *Gateway {
	if middleware == nil {
		return gw
	}

	if index < 0 {
		index = 0
	}
	if index > len(gw.middlewares) {
		index = len(gw.middlewares)
	}

	// 在指定位置插入中间件
	item := MiddlewareItem{Middleware: middleware, Sort: 0}
	gw.middlewares = append(gw.middlewares[:index], append([]MiddlewareItem{item}, gw.middlewares[index:]...)...)
	return gw
}

// InsertMiddlewareWithSort 在指定位置插入带排序权重的全局中间件
func (gw *Gateway) InsertMiddlewareWithSort(index int, middleware MiddlewareFunc, sort int) *Gateway {
	if middleware == nil {
		return gw
	}

	if index < 0 {
		index = 0
	}
	if index > len(gw.middlewares) {
		index = len(gw.middlewares)
	}

	// 在指定位置插入中间件
	item := MiddlewareItem{Middleware: middleware, Sort: sort}
	gw.middlewares = append(gw.middlewares[:index], append([]MiddlewareItem{item}, gw.middlewares[index:]...)...)
	return gw
}

// ============================================================================
// 路由中间件管理
// ============================================================================

// AddRouteMiddleware 为指定路由添加中间件
func (gw *Gateway) AddRouteMiddleware(routePath string, middleware MiddlewareFunc) *Gateway {
	if route, exists := gw.routes[routePath]; exists && middleware != nil {
		route.middlewares = append(route.middlewares, MiddlewareItem{Middleware: middleware, Sort: 0})
	}
	return gw
}

// AddRouteMiddlewareWithSort 为指定路由添加带排序权重的中间件
func (gw *Gateway) AddRouteMiddlewareWithSort(routePath string, middleware MiddlewareFunc, sort int) *Gateway {
	if route, exists := gw.routes[routePath]; exists && middleware != nil {
		route.middlewares = append(route.middlewares, MiddlewareItem{Middleware: middleware, Sort: sort})
	}
	return gw
}

// SetRouteMiddlewares 为指定路由设置中间件列表
func (gw *Gateway) SetRouteMiddlewares(routePath string, middlewares []MiddlewareFunc) *Gateway {
	if route, exists := gw.routes[routePath]; exists {
		route.middlewares = make([]MiddlewareItem, 0, len(middlewares))
		for _, middleware := range middlewares {
			if middleware != nil {
				route.middlewares = append(route.middlewares, MiddlewareItem{Middleware: middleware, Sort: 0})
			}
		}
	}
	return gw
}

// SetRouteMiddlewareItems 为指定路由设置中间件项列表
func (gw *Gateway) SetRouteMiddlewareItems(routePath string, middlewareItems []MiddlewareItem) *Gateway {
	if route, exists := gw.routes[routePath]; exists {
		route.middlewares = make([]MiddlewareItem, 0, len(middlewareItems))
		for _, item := range middlewareItems {
			if item.Middleware != nil {
				route.middlewares = append(route.middlewares, item)
			}
		}
	}
	return gw
}

// GetRouteMiddlewares 获取指定路由的中间件列表
func (gw *Gateway) GetRouteMiddlewares(routePath string) []MiddlewareFunc {
	if route, exists := gw.routes[routePath]; exists {
		return sortMiddlewares(route.middlewares)
	}
	return nil
}

// GetRouteMiddlewareItems 获取指定路由的中间件项列表
func (gw *Gateway) GetRouteMiddlewareItems(routePath string) []MiddlewareItem {
	if route, exists := gw.routes[routePath]; exists {
		return route.middlewares
	}
	return nil
}

// ClearRouteMiddlewares 清空指定路由的中间件
func (gw *Gateway) ClearRouteMiddlewares(routePath string) *Gateway {
	if route, exists := gw.routes[routePath]; exists {
		route.middlewares = []MiddlewareItem{}
	}
	return gw
}
