package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/logger"
	os1 "github.com/hosgf/element/os"
	"github.com/hosgf/element/process/systemd"
	"github.com/hosgf/element/rcode"
)

var (
	o       Operation
	mu      sync.Mutex
	isDebug bool
)

func Get() Operation {
	if o != nil {
		return o
	}
	mu.Lock()
	defer mu.Unlock()
	if o != nil {
		return o
	}
	o = CreateInstance(isDebug)
	return o
}

func CreateInstance(isDebug bool) Operation {
	return createInstance(os1.OS(), context.Background(), isDebug, logger.Log())
}

func createInstance(os string, ctx context.Context, isDebug bool, logger *glog.Logger) Operation {
	o := &global{operation{os: os, isDebug: isDebug, logger: logger}}
	o.init(ctx)
	return o
}

type global struct {
	operation
}

func (o *operation) GetProcess(name string) *gproc.Process {
	pid := o.GetPid(name)
	if pid < 1 {
		return nil
	}
	return o.manager.GetProcess(pid)
}

func (o *operation) GetPid(name string) int {
	if o.err != nil {
		logger.Errorf(context.Background(), "get process %s failed ,error: %s", name, o.err.Error())
		return -1
	}
	if len(name) < 1 {
		return -1
	}
	return o.mapping.Get(name)
}

// Start 启动服务
func (o *operation) Start(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error) {
	process := gproc.NewProcessCmd(gstr.Join(runtime.Cmd, " "), runtime.Env)
	process.Manager = o.manager
	pid, err := process.Start(ctx)
	if err != nil {
		return pid, gerror.NewCode(rcode.FAILURE, fmt.Sprintf("\n -- [ %s ] 进程启动失败: %s", runtime.Name, err.Error()))
	}
	go func() {
		defer func() {
			if err := process.Wait(); err != nil {
				logger.Errorf(ctx, "\n -- [ %s ] 进程启动失败 。。。 [ PID ：%d ][ Error ：%s ]", runtime.Name, pid, err.Error())
			}
		}()
	}()
	o.mapping.Set(runtime.Name, pid)
	logger.Infof(ctx, "\n -- [ %s ] 进程启动命令提交成功 。。。 [ PID ：%d ]", runtime.Name, pid)
	return pid, nil
}

// Restart 重启服务
func (o *operation) Restart(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error) {
	if o.err != nil {
		return -1, o.err
	}
	name := runtime.Name
	if len(name) < 1 {
		return -1, gerror.NewCode(rcode.FAILURE, "\n -- 进程重启失败，进程名称不能为空")
	}
	pid := o.GetPid(runtime.Name)
	process := o.manager.GetProcess(pid)
	if process != nil {
		err := process.Kill()
		if err != nil {
			logger.Errorf(ctx, "\n -- [ %s ] 进程停止失败 。。。 [ PID ：%d ][ Error ：%s ]", runtime.Name, pid, err.Error())
		}
	}
	return o.Start(ctx, runtime, logger)
}

// Stop 停止服务
func (o *operation) Stop(ctx context.Context, name string, logger *glog.Logger) error {
	if o.err != nil {
		return o.err
	}
	pid := o.GetPid(name)
	process := o.manager.GetProcess(pid)
	if process == nil {
		return nil
	}
	err := process.Kill()
	if err == nil {
		return nil
	}
	o.mapping.Remove(name)
	logger.Errorf(ctx, "\n -- [ %s ] 进程停止失败 。。。 [ PID ：%d [ Error ：%s ]", name, pid, err.Error())
	return err
}

// Status 服务状态
func (o *operation) Status(name string) health.Health {
	if o.err != nil {
		return health.UNKNOWN
	}
	pid := o.GetPid(name)
	process := o.manager.GetProcess(pid)
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
	if o.err != nil {
		return
	}
	if err := o.manager.KillAll(); err == nil {
		o.mapping.Clear()
	}
}

// init
func (o *operation) init(ctx context.Context) {
	o.mapping = gmap.NewStrIntMap(true)
	o.manager = gproc.NewManager()
	o.sys = systemd.CreateInstance(ctx, o.os, o.isDebug, o.logger)
}

type Operation interface {
	// GetProcess 获取进程对象
	GetProcess(name string) *gproc.Process
	// GetPid 获取进程ID
	GetPid(name string) int
	// Start 启动服务
	Start(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error)
	// Restart 重启服务
	Restart(ctx context.Context, runtime *RuntimeConfig, logger *glog.Logger) (int, error)
	// Stop 停止服务
	Stop(ctx context.Context, name string, logger *glog.Logger) error
	// Status 状态服务查询
	Status(name string) health.Health
	// Clear 清空进程信息
	Clear()
	// 初始化
	init(ctx context.Context)
}

type operation struct {
	os      string
	mapping *gmap.StrIntMap
	manager *gproc.Manager
	sys     systemd.Operation
	logger  *glog.Logger
	isDebug bool
	err     error
}

type RuntimeConfig struct {
	Name    string     `json:"name,omitempty"`
	Restart string     `json:"restart,omitempty"`
	Cmd     []string   `json:"cmd,omitempty"`
	Env     []string   `json:"env,omitempty"`
	Hosts   []os1.Host `json:"hosts,omitempty"`
}
