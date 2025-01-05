package systemd

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/model/result"
)

type windows struct {
	operation
}

// Install 安装服务
func (w *windows) Install(ctx context.Context, name, file string, enable bool, logger *glog.Logger) (string, error) {
	return "", gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

// Uninstall 卸载服务
func (w *windows) Uninstall(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return "", gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

// Enable 设置开机自启动
func (w *windows) Enable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return w.command(ctx, fmt.Sprintf("enable %s", name), logger)
}

// Disable 禁止开机自启动
func (w *windows) Disable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return w.command(ctx, fmt.Sprintf("disable %s", name), logger)
}

// Start 启动服务
func (w *windows) Start(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return w.command(ctx, fmt.Sprintf("start %s", name), logger)
}

// Restart 重启服务
func (w *windows) Restart(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return w.command(ctx, fmt.Sprintf("restart %s", name), logger)
}

// Stop 停止服务
func (w *windows) Stop(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return w.command(ctx, fmt.Sprintf("stop %s", name), logger)
}

// Status 服务状态
func (w *windows) Status(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	status, err := w.command(ctx, fmt.Sprintf("status %s | grep Active: | awk '{print $2}'", name), logger)
	if err != nil {
		return "", err
	}
	return gstr.Replace(gstr.Trim(status), "\r", ""), nil
}

// Reload 重新加载服务配置文件
func (w *windows) Reload(ctx context.Context, logger *glog.Logger) (string, error) {
	return w.command(ctx, "daemon-reload", logger)
}

func (w *windows) init(ctx context.Context) {
	w.err = gerror.NewCode(result.FAILURE, "没有实现的操作")
}
