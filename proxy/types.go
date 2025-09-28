package proxy

import (
	"sort"
)

// MiddlewareItem 中间件项，包含中间件函数和排序权重
type MiddlewareItem struct {
	Middleware MiddlewareFunc
	Sort       int // 排序权重，数值越小越靠前，0表示按加入顺序
}

type Route struct {
	Name        string   `json:"name,omitempty"`
	Path        string   `json:"path,omitempty"`
	SameToken   string   `json:"sameToken,omitempty"`
	Address     string   `json:"address,omitempty"`
	Includes    []string `json:"includes,omitempty"`
	Excludes    []string `json:"excludes,omitempty"`
	middlewares []MiddlewareItem
}

type Gateway struct {
	prefix      string
	name        string
	headers     map[string]string
	routes      map[string]*Route
	ignore      map[string]interface{}
	middlewares []MiddlewareItem
}

func (gw *Gateway) toIgnores() []string {
	if len(gw.ignore) < 1 {
		return nil
	}
	ignores := make([]string, 0, len(gw.ignore))
	for key := range gw.ignore {
		ignores = append(ignores, key)
	}
	return ignores
}

func (gw *Gateway) toRoutes() []*Route {
	if len(gw.routes) < 1 {
		return nil
	}
	routes := make([]*Route, 0, len(gw.routes))
	for _, v := range gw.routes {
		routes = append(routes, v)
	}
	return routes
}

// sortMiddlewares 对中间件进行排序
func sortMiddlewares(middlewares []MiddlewareItem) []MiddlewareFunc {
	if len(middlewares) == 0 {
		return nil
	}

	// 创建副本进行排序
	sorted := make([]MiddlewareItem, len(middlewares))
	copy(sorted, middlewares)

	// 按Sort字段排序，如果Sort为0则保持原有顺序
	sort.Slice(sorted, func(i, j int) bool {
		// 如果两个都是0，保持原有顺序
		if sorted[i].Sort == 0 && sorted[j].Sort == 0 {
			return i < j
		}
		// 如果一个是0，另一个不是，0的排在后面
		if sorted[i].Sort == 0 {
			return false
		}
		if sorted[j].Sort == 0 {
			return true
		}
		// 两个都不是0，按Sort值排序
		return sorted[i].Sort < sorted[j].Sort
	})

	// 提取中间件函数
	result := make([]MiddlewareFunc, len(sorted))
	for i, item := range sorted {
		result[i] = item.Middleware
	}

	return result
}
