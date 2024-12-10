package process

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/hosgf/element/cmd"
	"github.com/hosgf/element/consts"
	"github.com/hosgf/element/os"
)

type RuntimeConfig struct {
	Name  string            `json:"name,omitempty"`
	Label string            `json:"label,omitempty"`
	Cmd   []string          `json:"cmd,omitempty"`
	Envs  map[string]string `json:"envs,omitempty"`
	Hosts []os.Host         `json:"hosts,omitempty"`
}

type Context struct {
	cmd *cmd.Cmd
}

func (c *Context) SetEnvironment(ctx context.Context, envs map[string]string) {
	os.SetEnvironment(ctx, envs)
}

func (c *Context) SetHosts(ctx context.Context, hosts []os.Host) error {
	return os.SetHosts(ctx, hosts)
}

func (c *Context) command(ctx context.Context, command string, logger *glog.Logger) (string, error) {
	return c.cmd.Command(ctx, command, logger)
}

// Operation 服务操作接口
type Operation interface {
	// PID 获取PID
	PID(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error)
	// Start 启动
	Start(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error)
	// Stop 停止
	Stop(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (string, error)
	// Restart 重启
	Restart(ctx context.Context, config RuntimeConfig, logger *glog.Logger) (bool, error)
	// Status 状态
	Status(ctx context.Context, config RuntimeConfig, logger *glog.Logger) bool
	// SetEnvironment 设置环境变量
	SetEnvironment(ctx context.Context, envs map[string]string)
	// SetHosts 设置Hosts
	SetHosts(ctx context.Context, hosts []os.Host) error
}

func GetDefault() (Operation, error) {
	return Factory.GetDefault(false)
}

func Get(os string) (Operation, error) {
	return Factory.Get(os, false)
}

var Factory = &factory{
	cache:   gcache.New(),
	context: context.Background(),
}

type factory struct {
	cache   *gcache.Cache
	context context.Context
}

func (factory *factory) GetDefault(isDebug bool) (Operation, error) {
	return factory.Get(os.OS(), isDebug)
}

func (factory *factory) Get(o string, isDebug bool) (Operation, error) {
	return factory.getInstance(o, isDebug)
}

func (factory *factory) getInstance(o string, isDebug bool) (Operation, error) {
	if len(o) < 1 {
		o = os.OS()
	}
	value, err := factory.cache.Get(factory.context, o)
	if err != nil {
		return nil, err
	}
	if value != nil {
		return value.Val().(Operation), nil
	}
	service, err := factory.createInstance(o, isDebug)
	if err != nil {
		return nil, err
	}
	factory.cache.Set(factory.context, o, service, 0)
	return service, nil
}

func (factory *factory) createInstance(env string, isDebug bool) (Operation, error) {
	switch env {
	case os.LINUX:
		return newLinux(isDebug), nil
	case os.MACOS:
		return newLinux(isDebug), nil
	case os.WINDOWS:
		return newWindows(isDebug), nil
	default:
		return nil, gerror.NewCode(consts.FAILURE, "没有实现的操作")
	}
}
