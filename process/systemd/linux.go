package systemd

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/cmd"
	"github.com/hosgf/element/consts"
	"os/exec"
)

type linux struct {
	operation
}

// Install 安装服务
func (l *linux) Install(ctx context.Context, name, file string, enable bool, logger *glog.Logger) (string, error) {
	return "", gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

// Uninstall 卸载服务
func (l *linux) Uninstall(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return "", gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

// Enable 设置开机自启动
func (l *linux) Enable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return l.command(ctx, fmt.Sprintf("enable %s", name), logger)
}

// Disable 禁止开机自启动
func (l *linux) Disable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return l.command(ctx, fmt.Sprintf("disable %s", name), logger)
}

// Start 启动服务
func (l *linux) Start(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return l.command(ctx, fmt.Sprintf("start %s", name), logger)
}

// Restart 重启服务
func (l *linux) Restart(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return l.command(ctx, fmt.Sprintf("restart %s", name), logger)
}

// Stop 停止服务
func (l *linux) Stop(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return l.command(ctx, fmt.Sprintf("stop %s", name), logger)
}

// Status 服务状态
func (l *linux) Status(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	status, err := l.command(ctx, fmt.Sprintf("status %s | grep Active: | awk '{print $2}'", name), logger)
	if err != nil {
		return "", err
	}
	return gstr.Replace(gstr.Trim(status), "\r", ""), nil
}

// Reload 重新加载服务配置文件
func (l *linux) Reload(ctx context.Context, logger *glog.Logger) (string, error) {
	return l.command(ctx, "daemon-reload", logger)
}

// init 初始化systemd用于管理系统和管理服务的工具
func (l *linux) init(ctx context.Context) {
	path, err := exec.LookPath("systemctl")
	if err == nil {
		l.cmd = cmd.New(path, isDebug)
		return
	}
	l.err = gerror.NewCode(consts.FAILURE, fmt.Sprintf("[ systemctl ]命令不可用: %s", err.Error()))
}
