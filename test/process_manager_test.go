package test

import (
	"context"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/process/manager"
	"testing"
	"time"
)

func GetRuntimeConfig() *manager.RuntimeConfig {
	return &manager.RuntimeConfig{
		Name:    "match-data-platform",
		Cmd:     []string{"java -jar D:\\UE\\project\\intelligent-match-platform\\match-data-platform\\target\\match-data-platform.jar -Dmatch-data-platform -Dfile.encoding=UTF-8 &"},
		Env:     []string{"name=match-data-platform"},
		Restart: "always",
	}
}

func TestManagerProcess(t *testing.T) {
	ctx := context.Background()
	config := GetRuntimeConfig()
	log := logger.Log()
	m := manager.Get()
	pid, err := m.Start(ctx, config, log)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	gtimer.Add(ctx, time.Second*5, func(ctx context.Context) {
		if count > 10 {
			count = 0
			//nowPid, err := m.Restart(ctx, config, log)
			//t.Log(count, "pid: ", pid, "---->", "nowPid: ", nowPid, err)
		} else {
			health := m.Status(config.Name)
			t.Log(count, "pid: ", pid, "---->", health)
		}
		count++
	})
	select {}
}

func TestManagerProcessStart(t *testing.T) {
	ctx := context.Background()
	config := GetRuntimeConfig()
	m := manager.Get()
	pid, err := m.Start(ctx, config, logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	gtimer.Add(ctx, time.Second+3, func(ctx context.Context) {
		t.Log(pid)
		health := m.Status(config.Name)
		t.Log(health)
	})

}

func TestManagerProcessRestart(t *testing.T) {
	m := manager.Get()
	pid, err := m.Restart(context.Background(), GetRuntimeConfig(), logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pid)
}

func TestManagerProcessStop(t *testing.T) {
	m := manager.Get()
	config := GetRuntimeConfig()
	err := m.Stop(context.Background(), config.Name, logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("\r\n-------------   END")
}

func TestManagerProcessStatus(t *testing.T) {
	m := manager.Get()
	config := GetRuntimeConfig()
	health := m.Status(config.Name)
	t.Log(health)
}

func TestManagerProcessClear(t *testing.T) {
	m := manager.Get()
	m.Clear()
	t.Log("\r\n-------------   END")
}
