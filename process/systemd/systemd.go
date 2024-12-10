package systemd

import (
	"context"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/hosgf/element/cmd"
	"github.com/hosgf/element/logger"
	os1 "github.com/hosgf/element/os"
	"path/filepath"
	"sync"
)

var (
	o       Operation
	mu      sync.Mutex
	isDebug bool
)

// Enable 设置开机自启动
func Enable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return Get().Enable(ctx, name, logger)
}

// Disable 禁止开机自启动
func Disable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return Get().Disable(ctx, name, logger)
}

// Start 启动服务
func Start(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return Get().Start(ctx, name, logger)
}

// Restart 重启服务
func Restart(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return Get().Restart(ctx, name, logger)
}

// Stop 停止服务
func Stop(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return Get().Stop(ctx, name, logger)
}

// Status 服务状态
func Status(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return Get().Status(ctx, name, logger)
}

// Reload 重新加载服务配置文件
func Reload(ctx context.Context, logger *glog.Logger) (string, error) {
	return Get().Reload(ctx, logger)
}

func Get() Operation {
	if o != nil {
		return o
	}
	mu.Lock()
	defer mu.Unlock()
	if o != nil {
		return o
	}
	o = CreateInstance(context.Background(), os1.OS(), isDebug, logger.Log())
	return o
}

func CreateInstance(ctx context.Context, os string, isDebug bool, logger *glog.Logger) Operation {
	return createInstance(os, ctx, isDebug, logger)
}

func createInstance(os string, ctx context.Context, isDebug bool, logger *glog.Logger) Operation {
	operation := operation{os: os, isDebug: isDebug, logger: logger}
	var o Operation
	switch os {
	case os1.WINDOWS:
		o = &windows{operation}
		break
	case os1.LINUX:
		o = &linux{operation}
		break
	case os1.MACOS:
		o = &macos{operation}
		break
	default:
		o = &linux{operation}
		break
	}
	o.init(ctx)
	return o
}

type Operation interface {
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
	init(ctx context.Context)
}

type operation struct {
	os      string
	cmd     *cmd.Cmd
	logger  *glog.Logger
	isDebug bool
	err     error
}

func (o *operation) command(ctx context.Context, cmd string, logger *glog.Logger) (string, error) {
	if o.err != nil {
		return "", o.err
	}
	return o.cmd.Command(ctx, cmd, logger)
}

func (o *operation) getTemplatePath(name string) string {
	return filepath.Join("resource", "template", o.os, name)
}
