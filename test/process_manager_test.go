package test

import (
	"github.com/hosgf/element/process/manager"
	"testing"
)

func TestManagerProcess(t *testing.T) {
	m := manager.GetDefault()
	m.Clear()
}
