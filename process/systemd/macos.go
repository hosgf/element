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
	"github.com/hosgf/element/os"
	"os/exec"
	"path/filepath"
)

type macos struct {
	cmd *cmd.Cmd
	err error
}

// Install 安装服务
func (m *macos) Install(ctx context.Context, name, file string, enable bool, logger *glog.Logger) (string, error) {
	return "", gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

// Uninstall 卸载服务
func (m *macos) Uninstall(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return "", gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

// Enable 设置开机自启动
func (m *macos) Enable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return m.command(ctx, fmt.Sprintf("load -w  %s", name), logger)
}

// Disable 禁止开机自启动
func (m *macos) Disable(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return m.command(ctx, fmt.Sprintf("unload -w  %s", name), logger)
}

// Start 启动服务
func (m *macos) Start(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return m.command(ctx, fmt.Sprintf("start %s", name), logger)
}

// Restart 重启服务
func (m *macos) Restart(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return m.command(ctx, fmt.Sprintf("restart %s", name), logger)
}

// Stop 停止服务
func (m *macos) Stop(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	return m.command(ctx, fmt.Sprintf("stop %s", name), logger)
}

// Status 服务状态
func (m *macos) Status(ctx context.Context, name string, logger *glog.Logger) (string, error) {
	status, err := m.command(ctx, fmt.Sprintf("list %s | grep Active: | awk '{print $2}'", name), logger)
	if err != nil {
		return "", err
	}
	return gstr.Replace(gstr.Trim(status), "\r", ""), nil
}

// Reload 重新加载服务配置文件
func (m *macos) Reload(ctx context.Context, logger *glog.Logger) (string, error) {
	return m.command(ctx, "daemon-reload", logger)
}

// init 初始化launchctl用于管理系统和管理服务的工具
func (m *macos) init(ctx context.Context, isDebug bool, logger *glog.Logger) {
	path, err := exec.LookPath("launchctl")
	if err == nil {
		m.cmd = cmd.New(path, isDebug)
		return
	}
	m.err = gerror.NewCode(consts.FAILURE, fmt.Sprintf("[ launchctl ]命令不可用: %s", err.Error()))
	logger.Errorf(ctx, "%d %s", gerror.Code(m.err).Code(), m.err.Error())
}

func (m *macos) command(ctx context.Context, cmd string, logger *glog.Logger) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.cmd.Command(ctx, cmd, logger)
}

func (m *macos) getTemplatePath(name string) string {
	return filepath.Join("resource", "template", os.MACOS, name)
}
