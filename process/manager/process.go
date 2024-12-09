package manager

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/consts"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/logger"
	os1 "github.com/hosgf/element/os"
	"sync"
)

var (
	oper    Manager
	mu      sync.Mutex
	isDebug bool
)

func GetDefault() Manager {
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

func Get(isDebug bool) Manager {
	return get(os1.OS(), context.Background(), isDebug, logger.Log())
}

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

func (o *operation) GetProcess(pid int) *gproc.Process {
	return o.manager.GetProcess(pid)
}

func (o *operation) GetPid(name string) *gproc.Process {
	return nil
}

// Start 启动服务
func (o *operation) Start(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error) {
	//process := o.manager.NewProcess(runtime.Path, runtime.Cmd, runtime.Env)
	process := gproc.NewProcessCmd(gstr.Join(runtime.Cmd, " "), runtime.Env)
	process.Manager = o.manager
	pid, err := process.Start(ctx)
	runtime.Pid = pid
	if err != nil {
		return runtime.Pid, gerror.NewCode(consts.FAILURE, fmt.Sprintf("\n -- [ %s ] 进程启动失败: %s", runtime.Name, err.Error()))
	}
	go func() {
		defer func() {
			if err := process.Wait(); err != nil {
				logger.Errorf(ctx, "\n -- [ %s ] 进程停止失败 。。。 [ PID ：%d ][ Error ：%s ]", runtime.Name, runtime.Pid, err.Error())
			}
		}()
	}()
	logger.Infof(ctx, "\n -- [ %s ] 进程启动成功 。。。 [ PID ：%d ]", runtime.Name, runtime.Pid)
	return runtime.Pid, nil
}

// Restart 重启服务
func (o *operation) Restart(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error) {
	process := o.manager.GetProcess(runtime.Pid)
	if process != nil {
		err := process.Kill()
		if err != nil {
			logger.Errorf(ctx, "\n -- [ %s ] 进程停止失败 。。。 [ PID ：%d ][ Error ：%s ]", runtime.Name, runtime.Pid, err.Error())
		}
	}
	return o.Start(ctx, runtime, logger)
}

// Stop 停止服务
func (o *operation) Stop(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) error {
	process := o.manager.GetProcess(runtime.Pid)
	if process == nil {
		return nil
	}
	err := process.Kill()
	if err == nil {
		return nil
	}
	logger.Errorf(ctx, "\n -- [ %s ] 进程停止失败 。。。 [ PID ：%d [ Error ：%s ]", runtime.Name, runtime.Pid, err.Error())
	return err
}

// Status 服务状态
func (o *operation) Status(ctx context.Context, runtime *RuntimeConfig) health.Health {
	process := o.manager.GetProcess(runtime.Pid)
	if process == nil {
		return health.UNKNOWN
	}
	state := process.ProcessState
	if state == nil {
		return health.UP
	}
	return health.DOWN
}

func (o *operation) Clear() {
	o.manager.KillAll()
}

// init
func (o *operation) init(ctx context.Context, isDebug bool, logger *glog.Logger) {
	o.manager = gproc.NewManager()
}

type Manager interface {
	// GetProcess 获取进程对象
	GetProcess(pid int) *gproc.Process
	// GetPid 获取进程ID
	GetPid(name string) *gproc.Process
	// Start 启动服务
	Start(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error)
	// Restart 重启服务
	Restart(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error)
	// Stop 停止服务
	Stop(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) error
	// Status 状态服务查询
	Status(ctx context.Context, runtime *RuntimeConfig) health.Health
	// Clear 清空进程信息
	Clear()
	// 初始化
	init(ctx context.Context, isDebug bool, logger *glog.Logger)
}

type RuntimeConfig struct {
	Pid     int      `json:"pid,omitempty"`
	Name    string   `json:"name,omitempty"`
	Restart string   `json:"restart,omitempty"`
	Cmd     []string `json:"cmd,omitempty"`
	Env     []string `json:"env,omitempty"`
}
