package os

import (
	os1 "os"
	"os/exec"
)

const (
	MACOS = "macos"
)

type macos struct {
	system
}

func (os *macos) Delimiter() string {
	return "\n"
}

func (os *macos) Command(command string) *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", command)
	if len(os.env) > 0 {
		cmd.Env = append(os1.Environ(), os.env...)
	}
	return cmd
}
