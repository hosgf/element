package test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hosgf/element/client/k8s"
)

func TestProgressList(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Progress().List(ctx, "local")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestRunningProgress(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	config := &k8s.ProcessGroupConfig{}
	err := kubernetes.Progress().Running(ctx, config)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}
