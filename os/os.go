package os

import (
	"context"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/logger"
	"os/exec"
	"runtime"
	"sync"
)

// Command 执行命令
func Command(command string) *exec.Cmd {
	return get().Command(command)
}

// Delimiter 获取标记
func Delimiter() string {
	return get().Delimiter()
}

// SetEnvironment 设置环境变量
func SetEnvironment(ctx context.Context, envs map[string]string) {
	get().SetEnvironment(ctx, envs)
}

// Init 系统初始化
func Init(env []string) error {
	return get().init(env)
}

// Framework 获取操作系统架构类型
func Framework() string {
	framework := ""
	switch runtime.GOARCH {
	case `386`:
	case `amd64`:
		framework = `x86_64`
		break
	case `arm64`:
		framework = `arm64`
		break
	}
	return framework
}

// OS 获取操作环境类型
func OS() string {
	goos := runtime.GOOS
	if gstr.Contains(goos, LINUX) {
		return LINUX
	}
	if gstr.Contains(goos, WINDOWS) {
		return WINDOWS
	}
	if gstr.Contains(goos, MACOS) {
		return MACOS
	}
	if gstr.Contains(goos, "darwin") {
		return MACOS
	}
	return LINUX
}

var (
	service Service
	mu      sync.Mutex
)

func get() Service {
	if service != nil {
		return service
	}
	mu.Lock()
	defer mu.Unlock()
	if service != nil {
		return service
	}
	os := OS()
	o := system{os: os, framework: Framework()}
	switch os {
	case WINDOWS:
		service = &windows{o}
		break
	case LINUX:
		service = &linux{o}
		break
	case MACOS:
		service = &macos{o}
		break
	default:
		service = &linux{o}
		break
	}
	return service
}

type system struct {
	os        string
	framework string
	env       []string
}

func (os *system) OS() string {
	return os.os
}

func (os *system) Framework() string {
	return os.framework
}

func (os *system) SetEnvironment(ctx context.Context, envs map[string]string) {
	if len(envs) == 0 {
		return
	}
	for key, value := range envs {
		if err := genv.Set(key, value); err != nil {
			logger.Errorf(ctx, "设置环境变量失败: %s %s %s", key, value, err.Error())
		}
	}
}

func (os *system) init(env []string) error {
	if len(env) < 1 {
		return nil
	}
	os.env = append(os.env, env...)
	return nil
}

type Service interface {
	OS() string
	Delimiter() string
	Framework() string
	SetEnvironment(ctx context.Context, envs map[string]string)
	SetHosts(ctx context.Context, hosts []Host) error
	Command(command string) *exec.Cmd
	init(env []string) error
}

type Host struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
