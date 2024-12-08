package test

import (
	"context"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/process/manager"
	"testing"
)

func TestManagerProcessStatus(t *testing.T) {
	runtime := manager.RuntimeConfig{}
	m := manager.GetDefault()
	health, err := m.Status(context.Background(), runtime, logger.Log())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(health)
}
