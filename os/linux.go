package os

import (
	os1 "os"
	"os/exec"
)

const (
	LINUX = "linux"
)

type linux struct {
	os        string
	framework string
	env       []string
}

func (os *linux) OS() string {
	return os.os
}

func (os *linux) Framework() string {
	return os.framework
}

func (os *linux) Delimiter() string {
	return "\n"
}

func (os *linux) init(env []string) error {
	if len(env) < 1 {
		return nil
	}
	os.env = append(os.env, env...)
	return nil
}

func (os *linux) command(command string) *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", command)
	if len(os.env) > 0 {
		cmd.Env = append(os1.Environ(), os.env...)
	}
	return cmd
}
