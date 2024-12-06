package os

import (
	"github.com/gogf/gf/v2/text/gstr"
	"os/exec"
	"runtime"
	"sync"
)

// Command 执行命令
func Command(command string) *exec.Cmd {
	return get().Command(command)
}
func Delimiter() string {
	return get().Delimiter()
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
	switch os {
	case WINDOWS:
		service = &windows{os: os, framework: Framework()}
		break
	case LINUX:
		service = &linux{os: os, framework: Framework()}
		break
	case MACOS:
		service = &macos{os: os, framework: Framework()}
		break
	default:
		service = &linux{os: os, framework: Framework()}
		break
	}
	return service
}

type Service interface {
	OS() string
	Delimiter() string
	Framework() string
	Command(command string) *exec.Cmd
	init(env []string) error
}
