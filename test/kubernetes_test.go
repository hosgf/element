package test

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/hosgf/element/client/k8s"
	"github.com/hosgf/element/types"
	"testing"
)

func Test(t *testing.T) {
	ctx := context.Background()
	kubernetes := k8s.New(true)
	kubernetes.Init("")
	kubernetes.Namespace().List(ctx)
	//kubernetes.Init()
}

func TestNodeTop(t *testing.T) {
	ctx := context.Background()
	kubernetes := k8s.New(true)
	kubernetes.Init("")
	nodes, err := kubernetes.Nodes().Top(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(nodes)
}

func TestParse(t *testing.T) {
	fmt.Println(types.Parse("16384Mi"))
	fmt.Println(types.Parse("16384"))
	fmt.Println(types.Parse("16384u"))
	fmt.Println(types.Parse(""))
}
