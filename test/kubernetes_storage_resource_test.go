package test

import (
	"context"
	"testing"

	"github.com/hosgf/element/client/k8s"
)

func TestStorageResourceCreate(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	storage := &k8s.PersistentStorageResource{
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
	err := kubernetes.StorageResource().Apply(ctx, storage)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestStorageResourceDelete(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	err := kubernetes.StorageResource().Delete(ctx, "sandbox-storage")
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}
