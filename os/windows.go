package os

import (
	os1 "os"
	"os/exec"
)

const (
	WINDOWS = "windows"
)

type windows struct {
	os        string
	framework string
	env       []string
}

func (os *windows) OS() string {
	return os.os
}

func (os *windows) Delimiter() string {
	return "\r\n"
}

func (os *windows) Framework() string {
	return os.framework
}

func (os *windows) init(env []string) error {
	if len(env) < 1 {
		return nil
	}
	os.env = append(os.env, env...)
	exec.Command("chcp", "65001").Run()
	return nil
}

func (os *windows) command(command string) *exec.Cmd {
	cmd := exec.Command("powershell", "-Command", command)
	if len(os.env) > 0 {
		cmd.Env = append(os1.Environ(), os.env...)
	}
	return cmd
}
