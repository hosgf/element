package systemd

import (
	"context"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/hosgf/element/logger"
	os1 "github.com/hosgf/element/os"
	"sync"
)

var (
	oper    Systemd
	mu      sync.Mutex
	isDebug bool
)

// Enable 设置开机自启动
func Enable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return GetDefault().Enable(ctx, name, logger)
}

// Disable 禁止开机自启动
func Disable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return GetDefault().Disable(ctx, name, logger)
}

// Start 启动服务
func Start(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return GetDefault().Start(ctx, name, logger)
}

// Restart 重启服务
func Restart(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return GetDefault().Restart(ctx, name, logger)
}

// Stop 停止服务
func Stop(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return GetDefault().Stop(ctx, name, logger)
}

// Status 服务状态
func Status(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return GetDefault().Status(ctx, name, logger)
}

// Reload 重新加载服务配置文件
func Reload(ctx context.Context, logger *glog.Logger) (string, error) {
	return GetDefault().Reload(ctx, logger)
}

func GetDefault() Systemd {
	if oper != nil {
		return oper
	}
	mu.Lock()
	defer mu.Unlock()
	if oper != nil {
		return oper
	}
	oper = Get(isDebug)
	return oper
}

func Get(isDebug bool) Systemd {
	return get(os1.OS(), context.Background(), isDebug, logger.Log())
}

func get(os string, ctx context.Context, isDebug bool, logger *glog.Logger) Systemd {
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
	oper.init(ctx, isDebug, logger)
	return oper
}

type Systemd interface {
	// Install 安装服务
	Install(ctx context.Context, name, file string, enable bool, logger *glog.Logger) (string, error)
	// Uninstall 卸载服务
	Uninstall(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Enable 设置开机自启动
	Enable(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Disable 禁止开机自启动
	Disable(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Start 启动服务
	Start(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Restart 重启服务
	Restart(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Stop 停止服务
	Stop(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Status 状态服务查询
	Status(ctx context.Context, name string, logger *glog.Logger) (string, error)
	// Reload 重新加载
	Reload(ctx context.Context, logger *glog.Logger) (string, error)
	// Systemd 初始化
	init(ctx context.Context, isDebug bool, logger *glog.Logger)
}
