package os

import (
	os1 "os"
	"os/exec"
)

const (
	LINUX = "linux"
)

type linux struct {
	system
}

func (os *linux) Delimiter() string {
	return "\n"
}

func (os *linux) Command(command string) *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", command)
	if len(os.env) > 0 {
		cmd.Env = append(os1.Environ(), os.env...)
	}
	return cmd
}
