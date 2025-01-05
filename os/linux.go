package os

import (
	"context"
	"fmt"
	os1 "os"
	"os/exec"
	"path/filepath"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/rcode"
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

func (os *linux) SetHosts(ctx context.Context, hosts []Host) error {
	if len(hosts) < 1 {
		return nil
	}
	etcDirectory := filepath.Join("/", "etc")
	if !gfile.Exists(etcDirectory) {
		err := gfile.Mkdir(etcDirectory)
		if err != nil {
			return gerror.NewCode(rcode.FAILURE, fmt.Sprintf("创建目录失败: %s %s", etcDirectory, err.Error()))
		}
	}
	hostsFile := filepath.Join(etcDirectory, "hosts")
	if gfile.Exists(hostsFile) {
		_ = gfile.ReadLines(hostsFile, func(line string) error {
			line = gstr.Trim(line)
			if !gstr.HasPrefix(line, "#") {
				list := gstr.Split(line, " ")
				if len(list) >= 2 {
					var flag = true
					for _, host := range hosts {
						if host.Key == gstr.Trim(list[0]) && host.Value == gstr.Trim(gstr.Join(list[1:], " ")) {
							flag = false
						}
					}
					if flag {
						hosts = append(hosts, Host{
							Key:   gstr.Trim(list[0]),
							Value: gstr.Trim(gstr.Join(list[1:], " ")),
						})
					}
				}
			}
			return nil
		})
		err := gfile.RemoveFile(hostsFile)
		if err != nil {
			return gerror.NewCode(rcode.FAILURE, fmt.Sprintf("删除文件失败: %s %s", hostsFile, err.Error()))
		}
	}
	for _, host := range hosts {
		err := gfile.PutContentsAppend(hostsFile, fmt.Sprintf("%s %s\n", gstr.Trim(host.Key), gstr.Trim(host.Value)))
		if err != nil {
			logger.Errorf(ctx, "设置hosts失败: %s %s %s %s", hostsFile, gstr.Trim(host.Key), gstr.Trim(host.Value), err.Error())
		}
	}
	return nil
}
