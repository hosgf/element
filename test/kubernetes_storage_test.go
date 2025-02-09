package test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hosgf/element/client/k8s"
)

func TestStorageGet(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Storage().Get(ctx, "sandbox", "sandbox-storage")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestStorageCreate(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	storage := &k8s.PersistentStorage{
		Model: k8s.Model{
			Namespace:   "sandbox",
			App:         "sandboxApp",
			Group:       "sandboxGroup",
			Owner:       "sandboxOwner",
			Scope:       "sandboxScope",
			Labels:      make(map[string]string),
			AllowUpdate: true,
		},
		Storage: toStorage(),
	}
	err := kubernetes.Storage().Apply(ctx, storage)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestStorageDelete(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	err := kubernetes.Storage().Delete(ctx, false, "sandbox", "sandbox-storage")
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}
