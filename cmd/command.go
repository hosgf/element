package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/os"
	"os/exec"
	"strings"
	"time"
)

func New(exe string, isDebug bool) *Cmd {
	return &Cmd{
		exe:     exe,
		isDebug: isDebug,
	}
}

type Cmd struct {
	Env     []string
	exe     string
	isDebug bool
}

// Convert 地址转化为IP
func (c *Cmd) Convert(address string) string {
	return gstr.SubStr(gstr.Replace(address, "http://", ""), 0, gstr.Pos(gstr.Replace(address, "http://", ""), ":"))
}

func (c *Cmd) Stream(ctx context.Context, command string, logger *glog.Logger) error {
	cmd := c.command(command)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	begin := time.Now()
	if err != nil {
		msg := gstr.TrimRight(err.Error())
		logger.Errorf(ctx, "---->\r\n\tCommand: %s %s, \r\n\ttook_time: %v, \r\n\terr: %v, \r\n\toutput: %s", c.exe, command, time.Since(begin), err, msg)
		return errors.New(msg)
	}
	if err = cmd.Start(); err != nil {
		msg := gstr.TrimRight(err.Error())
		logger.Errorf(ctx, "---->\r\n\tCommand: %s %s, \r\n\ttook_time: %v, \r\n\terr: %v, \r\n\toutput: %s", c.exe, command, time.Since(begin), err, msg)
		return errors.New(msg)
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		msg := gstr.TrimRight(err.Error())
		logger.Errorf(ctx, "---->\r\n\tCommand: %s %s, \r\n\ttook_time: %v, \r\n\terr: %v, \r\n\toutput: %s", c.exe, command, time.Since(begin), err, msg)
		return errors.New(msg)
	}
	if c.isDebug {
		logger.Debugf(ctx, "---->\r\n\tCommand: %s %s, \r\n\ttook_time: %v", c.exe, command, time.Since(begin))
	}
	return nil
}

func (c *Cmd) Command(ctx context.Context, command string, logger *glog.Logger) (string, error) {
	cmd := c.command(command)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	begin := time.Now()
	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		logger.Errorf(ctx, "---->\r\n\tCommand: %s %s, \r\n\ttook_time: %v, \r\n\terr: %v, \r\n\toutput: %s", c.exe, command, time.Since(begin), err, msg)
		return "", errors.New(msg)
	}

	out := strings.TrimSpace(stdout.String())
	if stderr.Len() < 1 {
		if c.isDebug {
			logger.Debugf(ctx, "---->\r\n\tCommand: %s %s, \r\n\ttook_time: %v, \r\n\toutput: %s", c.exe, command, time.Since(begin), out)
		}
		return out, nil
	}

	msg := strings.TrimSpace(stderr.String())
	logger.Warningf(ctx, "---->\r\n\tCommand: %s %s, \r\n\ttook_time: %v, \n\terr: %v, \r\n\toutput: %s", c.exe, command, time.Since(begin), msg, out)
	if len(msg) > 0 {
		return out, errors.New(msg)
	}
	return out, nil
}

func (c *Cmd) command(command string) *exec.Cmd {
	return os.Command(fmt.Sprintf("%s %s", c.exe, command))
}
