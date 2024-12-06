package service

import (
	"context"
	"github.com/gogf/gf/v2/os/glog"
	os1 "github.com/hosgf/element/os"
	"sync"
)

// Init 初始化
func Init(ctx context.Context, isDebug bool) {
	get().Init(ctx, isDebug)
}

// Enable 设置开机自启动
func Enable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return get().Enable(ctx, name, logger)
}

// Disable 禁止开机自启动
func Disable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return get().Disable(ctx, name, logger)
}

// Start 启动服务
func Start(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return get().Start(ctx, name, logger)
}

// Restart 重启服务
func Restart(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return get().Restart(ctx, name, logger)
}

// Stop 停止服务
func Stop(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return get().Stop(ctx, name, logger)
}

// Status 服务状态
func Status(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return get().Status(ctx, name, logger)
}

// Reload 重新加载服务配置文件
func Reload(ctx context.Context, logger *glog.Logger) (string, error) {
	return get().Reload(ctx, logger)
}

var (
	oper Operation
	mu   sync.Mutex
)

func get() Operation {
	if oper != nil {
		return oper
	}
	mu.Lock()
	defer mu.Unlock()
	if oper != nil {
		return oper
	}
	os := os1.OS()
	switch os {
	case os1.WINDOWS:
		oper = &windows{}
		break
	case os1.LINUX:
		oper = &linux{}
		break
	case os1.MACOS:
		oper = &macos{}
		break
	default:
		oper = &linux{}
		break
	}
	return oper
}

type Operation interface {
	// Init 初始化
	Init(ctx context.Context, isDebug bool)
	// Enable 启动
	Enable(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Disable 启动
	Disable(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Start 启动
	Start(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Restart 重启
	Restart(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Stop 停止
	Stop(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Status 状态查询
	Status(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Reload 重新加载
	Reload(ctx context.Context, logger *glog.Logger) (string, error)
}
