package manager

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/hosgf/element/consts"
	"github.com/hosgf/element/health"
	os1 "github.com/hosgf/element/os"
	"sync"
)

var (
	oper    Manager
	mu      sync.Mutex
	isDebug bool
)

func get(os string, ctx context.Context, isDebug bool, logger *glog.Logger) Manager {
	switch os {
	case os1.WINDOWS:
		oper = &operation{}
		break
	case os1.LINUX:
		oper = &operation{}
		break
	case os1.MACOS:
		oper = &operation{}
		break
	default:
		oper = &operation{}
		break
	}
	oper.init(ctx, isDebug, logger)
	return oper
}

type operation struct {
	manager *gproc.Manager
	err     error
}

// Start 启动服务
func (o *operation) Start(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) (int, error) {
	process := o.manager.NewProcess(runtime.Path, runtime.Cmd, runtime.Env)
	pid, err := process.Start(ctx)
	if err != nil {
		return -1, gerror.NewCode(consts.FAILURE, fmt.Sprintf("\n -- [ %s ] 进程启动失败: %s", runtime.Name, err.Error()))
	}
	logger.Infof(ctx, "\n -- [ %s ] 进程启动成功 。。。 [ PID ：%d ][ PATH ：%s ]", runtime.Name, pid, runtime.Path)
	return pid, nil
}

// Restart 重启服务
func (o *operation) Restart(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) (int, error) {
	process := o.manager.GetProcess(runtime.Pid)
	if process != nil {
		err := process.Kill()
		if err != nil {
			logger.Error(ctx, "\n -- [ %s ] 进程停止失败 。。。 [ PID ：%d ][ PATH ：%s ]", runtime.Name, runtime.Pid, runtime.Path)
		}
		pid, err := process.Start(ctx)
		if err == nil {
			return pid, nil
		}
	}
	return o.Start(ctx, runtime, logger)
}

// Stop 停止服务
func (o *operation) Stop(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) error {
	process := o.manager.GetProcess(runtime.Pid)
	if process == nil {
		return nil
	}
	err := process.Kill()
	if err == nil {
		return nil
	}
	logger.Error(ctx, "\n -- [ %s ] 进程停止失败 。。。 [ PID ：%d ][ PATH ：%s ]", runtime.Name, runtime.Pid, runtime.Path)
	return err
}

// Status 服务状态
func (o *operation) Status(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) (health.Health, error) {
	process := o.manager.GetProcess(runtime.Pid)
	if process == nil {
		return health.UNKNOWN, nil
	}
	state := process.ProcessState
	if state == nil {
		return health.UNKNOWN, nil
	}
	if state.Success() {
		return health.UP, nil
	}
	if state.ExitCode() < 0 {
		return health.DOWN, nil
	}
	return health.UP, nil
}

func (o *operation) Clear() {
	o.manager.Clear()
}

// init
func (o *operation) init(ctx context.Context, isDebug bool, logger *glog.Logger) {
	o.manager = gproc.NewManager()
}

type Manager interface {
	// Start 启动服务
	Start(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) (int, error)
	// Restart 重启服务
	Restart(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) (int, error)
	// Stop 停止服务
	Stop(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) error
	// Status 状态服务查询
	Status(ctx context.Context, runtime RuntimeConfig, logger *glog.Logger) (health.Health, error)
	// Clear 清空进程信息
	Clear()
	// 初始化
	init(ctx context.Context, isDebug bool, logger *glog.Logger)
}

type RuntimeConfig struct {
	Pid     int      `json:"pid,omitempty"`
	Name    string   `json:"name,omitempty"`
	Path    string   `json:"path,omitempty"`
	Restart string   `json:"restart,omitempty"`
	Cmd     []string `json:"cmd,omitempty"`
	Env     []string `json:"env,omitempty"`
}
