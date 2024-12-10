package process

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/cmd"
	"time"
)

func newWindows(isDebug bool) *windows {
	return &windows{Context{cmd: cmd.New("", isDebug)}}
}

type windows struct {
	Context
}

func (w *windows) PID(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	command := fmt.Sprintf("Get-WmiObject Win32_Process | Where-Object { $_.CommandLine -like \"*%s*\" -and $_.CommandLine -like \"*%s*\" } | Select-Object ProcessId", config.Label, config.Name)
	return w.command(ctx, command, logger)
}

func (w *windows) Start(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	if w.Status(ctx, config, logger) {
		logger.Debugf(ctx, "%s Is Running...", config.Name)
		return w.PID(ctx, config, logger)
	}
	return w.command(ctx, gstr.Join(config.Cmd, " "), logger)
}

func (w *windows) Stop(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	pid, err := w.PID(ctx, config, logger)
	if nil != err {
		return pid, err
	}
	if len(pid) > 0 {
		command := fmt.Sprintf(" taskkill /PID %s /F", pid)
		res, err := w.command(ctx, command, logger)
		if nil != err {
			return res, err
		}
		var isOk = false
		for !isOk {
			isOk = w.Status(ctx, config, logger)
		}
	}
	return pid, nil
}

func (w *windows) Restart(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (bool, error) {
	_, err := w.Stop(ctx, config, logger)
	if nil != err {
		return false, err
	}
	time.Sleep(1)
	_, err = w.Start(ctx, config, logger)
	if nil != err {
		return false, err
	}
	return true, nil
}

func (w *windows) Status(ctx context.Context, config RuntimeConfig, logger *glog.Logger) bool {
	pid, err := w.PID(ctx, config, logger)
	if nil != err {
		return false
	}
	return len(pid) > 0
}
