package os

import (
	os1 "os"
	"os/exec"
)

const (
	MACOS = "macos"
)

type macos struct {
	os        string
	framework string
	env       []string
}

func (os *macos) OS() string {
	return os.os
}

func (os *macos) Framework() string {
	return os.framework
}

func (os *macos) Delimiter() string {
	return "\n"
}

func (os *macos) init(env []string) error {
	if len(env) < 1 {
		return nil
	}
	os.env = append(os.env, env...)
	return nil
}

func (os *macos) command(command string) *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", command)
	if len(os.env) > 0 {
		cmd.Env = append(os1.Environ(), os.env...)
	}
	return cmd
}
