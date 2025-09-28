package proxy

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hosgf/element/logger"
)

// ============================================================================
// 路由验证器
// ============================================================================

// RouteValidationError 路由验证错误
type RouteValidationError struct {
	Field   string
	Message string
}

func (e *RouteValidationError) Error() string {
	return fmt.Sprintf("route validation error: %s - %s", e.Field, e.Message)
}

// RouteValidator 路由验证器
type RouteValidator struct {
	// 允许的协议
	allowedProtocols []string
	// 路径模式验证
	pathPattern *regexp.Regexp
	// 地址模式验证
	addressPattern *regexp.Regexp
}

// NewRouteValidator 创建新的路由验证器
func NewRouteValidator() *RouteValidator {
	return &RouteValidator{
		allowedProtocols: []string{"http", "https"},
		pathPattern:      regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		addressPattern:   regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(:\d+)?(/.*)?$`),
	}
}

// ValidateRouteName 验证路由名称
func (rv *RouteValidator) ValidateRouteName(name string) error {
	if name == "" {
		return &RouteValidationError{
			Field:   "name",
			Message: "route name cannot be empty",
		}
	}

	if len(name) > 50 {
		return &RouteValidationError{
			Field:   "name",
			Message: "route name too long (max 50 characters)",
		}
	}

	if !rv.pathPattern.MatchString(name) {
		return &RouteValidationError{
			Field:   "name",
			Message: "route name contains invalid characters (only alphanumeric, underscore, and hyphen allowed)",
		}
	}

	return nil
}

// ValidateRouteAddress 验证路由地址
func (rv *RouteValidator) ValidateRouteAddress(address string) error {
	if address == "" {
		return &RouteValidationError{
			Field:   "address",
			Message: "route address cannot be empty",
		}
	}

	// 解析URL
	parsedURL, err := url.Parse(address)
	if err != nil {
		return &RouteValidationError{
			Field:   "address",
			Message: fmt.Sprintf("invalid URL format: %v", err),
		}
	}

	// 检查协议
	if parsedURL.Scheme == "" {
		return &RouteValidationError{
			Field:   "address",
			Message: "URL scheme is required",
		}
	}

	// 检查是否允许的协议
	allowed := false
	for _, protocol := range rv.allowedProtocols {
		if parsedURL.Scheme == protocol {
			allowed = true
			break
		}
	}

	if !allowed {
		return &RouteValidationError{
			Field:   "address",
			Message: fmt.Sprintf("unsupported protocol: %s (allowed: %s)", parsedURL.Scheme, strings.Join(rv.allowedProtocols, ", ")),
		}
	}

	// 检查主机名
	if parsedURL.Host == "" {
		return &RouteValidationError{
			Field:   "address",
			Message: "URL host is required",
		}
	}

	// 检查主机名格式
	if !rv.isValidHost(parsedURL.Host) {
		return &RouteValidationError{
			Field:   "address",
			Message: "invalid host format",
		}
	}

	return nil
}

// ValidateIncludePaths 验证包含路径
func (rv *RouteValidator) ValidateIncludePaths(includes []string) error {
	for i, path := range includes {
		if err := rv.validatePath(path, fmt.Sprintf("includes[%d]", i)); err != nil {
			return err
		}
	}
	return nil
}

// ValidateExcludePaths 验证排除路径
func (rv *RouteValidator) ValidateExcludePaths(excludes []string) error {
	for i, path := range excludes {
		if err := rv.validatePath(path, fmt.Sprintf("excludes[%d]", i)); err != nil {
			return err
		}
	}
	return nil
}

// 验证路径
func (rv *RouteValidator) validatePath(path, field string) error {
	if path == "" {
		return &RouteValidationError{
			Field:   field,
			Message: "path cannot be empty",
		}
	}

	if !strings.HasPrefix(path, "/") {
		return &RouteValidationError{
			Field:   field,
			Message: "path must start with '/'",
		}
	}

	if len(path) > 200 {
		return &RouteValidationError{
			Field:   field,
			Message: "path too long (max 200 characters)",
		}
	}

	return nil
}

// 验证主机名
func (rv *RouteValidator) isValidHost(host string) bool {
	// 简单的主机名验证
	if host == "" {
		return false
	}

	// 检查是否包含端口
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		if len(parts) != 2 {
			return false
		}
		// 验证端口号
		if !rv.isValidPort(parts[1]) {
			return false
		}
		host = parts[0]
	}

	// 验证主机名格式
	hostPattern := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	return hostPattern.MatchString(host)
}

// 验证端口号
func (rv *RouteValidator) isValidPort(port string) bool {
	if port == "" {
		return false
	}

	portPattern := regexp.MustCompile(`^\d+$`)
	if !portPattern.MatchString(port) {
		return false
	}

	// 这里可以添加端口范围检查
	return true
}

// ValidateRoute 验证完整路由
func (rv *RouteValidator) ValidateRoute(sameToken, name, address string, includes, excludes []string) error {
	// 验证路由名称
	if err := rv.ValidateRouteName(name); err != nil {
		return err
	}

	// 验证路由地址
	if err := rv.ValidateRouteAddress(address); err != nil {
		return err
	}

	// 验证包含路径
	if err := rv.ValidateIncludePaths(includes); err != nil {
		return err
	}

	// 验证排除路径
	if err := rv.ValidateExcludePaths(excludes); err != nil {
		return err
	}

	// 验证SameToken
	if err := rv.ValidateSameToken(sameToken); err != nil {
		return err
	}

	return nil
}

// ValidateSameToken 验证SameToken
func (rv *RouteValidator) ValidateSameToken(sameToken string) error {
	if sameToken == "" {
		return &RouteValidationError{
			Field:   "sameToken",
			Message: "sameToken cannot be empty",
		}
	}

	if len(sameToken) > 100 {
		return &RouteValidationError{
			Field:   "sameToken",
			Message: "sameToken too long (max 100 characters)",
		}
	}

	return nil
}

// 全局路由验证器实例
var routeValidator = NewRouteValidator()

// ValidateRoute 验证路由（Gateway方法）
func (gw *Gateway) ValidateRoute(sameToken, name, address string, includes, excludes []string) error {
	return routeValidator.ValidateRoute(sameToken, name, address, includes, excludes)
}

// CreateRouteWithValidation 创建路由
func (gw *Gateway) CreateRouteWithValidation(sameToken, name, address string, includes, excludes []string) error {
	// 验证路由参数
	if err := gw.ValidateRoute(sameToken, name, address, includes, excludes); err != nil {
		logger.Errorf(context.Background(), "route validation failed: %v", err)
		return err
	}

	// 创建路由
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
	logger.Infof(context.Background(), "route created with validation: %s -> %s", route.Path, address)
	return nil
}
