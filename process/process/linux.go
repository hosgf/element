package process

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/cmd"
	"github.com/hosgf/element/model/result"
)

func newLinux(isDebug bool) *linux {
	return &linux{Context{cmd: cmd.New("", isDebug)}}
}

type linux struct {
	Context
}

func (l *linux) PID(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	command := fmt.Sprintf("ps -ef | grep %s | grep %s |  grep -v grep | awk '{print $2}'", config.Label, config.Name)
	return l.command(ctx, command, logger)
}

func (l *linux) Start(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	if l.Status(ctx, config, logger) {
		return l.PID(ctx, config, logger)
	}
	command := gstr.Trim(gstr.Join(config.Cmd, " "))
	if len(command) < 1 {
		return "", gerror.NewCode(result.FAILURE, fmt.Sprintf("启动 [%s] 的命令脚本不能为空！！！", config.Name))
	}
	return l.command(ctx, command, logger)
}

func (l *linux) Stop(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	pid, err := l.PID(ctx, config, logger)
	if nil != err {
		return pid, err
	}
	if len(pid) > 0 {
		command := fmt.Sprintf("kill -term %s", pid)
		res, err := l.command(ctx, command, logger)
		if nil != err {
			return res, err
		}
		var isOk = false
		for !isOk {
			isOk = l.Status(ctx, config, logger)
		}
	}
	return pid, nil
}

func (l *linux) Restart(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (bool, error) {
	_, err := l.Stop(ctx, config, logger)
	if nil != err {
		return false, err
	}
	time.Sleep(1)
	_, err = l.Start(ctx, config, logger)
	if nil != err {
		return false, err
	}
	return true, nil
}

func (l *linux) Status(ctx context.Context, config RuntimeConfig, logger *glog.Logger) bool {
	pid, err := l.PID(ctx, config, logger)
	if nil != err {
		return false
	}
	return len(pid) > 0
}
