package service

import (
	"context"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/hosgf/element/cmd"
)

func newWindows(isDebug bool) *windows {
	return &windows{Context{cmd: cmd.New("", isDebug)}}
}

type windows struct {
	Context
}

func (w *windows) PID(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	return "", nil
}

func (w *windows) Start(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	return "", nil
}

func (w *windows) Shutdown(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error) {
	return "", nil
}

func (w *windows) Restart(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (bool, error) {
	return true, nil
}

func (w *windows) Status(ctx context.Context, config RuntimeConfig, logger *glog.Logger) bool {
	return true
}
