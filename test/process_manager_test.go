package test

import (
	"context"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/process/manager"
	"testing"
)

func GetRuntimeConfig() manager.RuntimeConfig {
	runtime := manager.RuntimeConfig{}
	return runtime
}

func TestManagerProcessStart(t *testing.T) {
	m := manager.GetDefault()
	pid, err := m.Start(context.Background(), GetRuntimeConfig(), logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pid)
}

func TestManagerProcessRestart(t *testing.T) {
	m := manager.GetDefault()
	pid, err := m.Restart(context.Background(), GetRuntimeConfig(), logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pid)
}

func TestManagerProcessStop(t *testing.T) {
	m := manager.GetDefault()
	err := m.Stop(context.Background(), GetRuntimeConfig(), logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("\r\n-------------   END")
}

func TestManagerProcessStatus(t *testing.T) {
	m := manager.GetDefault()
	health, err := m.Status(context.Background(), GetRuntimeConfig(), logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(health)
}

func TestManagerProcessClear(t *testing.T) {
	m := manager.GetDefault()
	m.Clear()
	t.Log("\r\n-------------   END")
}
