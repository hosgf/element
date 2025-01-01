package test

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/hosgf/element/client/k8s"
	"github.com/hosgf/element/types"
	"testing"
)

func client() *k8s.Kubernetes {
	kubernetes := k8s.New(true)
	kubernetes.Init("")
	return kubernetes
}

func TestNodeTop(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	nodes, err := kubernetes.Nodes().Top(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(nodes)
}

func TestNamespaces(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	namespaces, err := kubernetes.Namespace().List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(namespaces)
}

func TestCreateNamespace(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	isOk, err := kubernetes.Namespace().Create(ctx, "test2")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(isOk)
}

func TestParse(t *testing.T) {
	fmt.Println(types.Parse("16384Mi"))
	fmt.Println(types.Parse("16384"))
	fmt.Println(types.Parse("16384u"))
	fmt.Println(types.Parse(""))
}
