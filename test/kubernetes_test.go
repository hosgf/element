package test

import (
	"context"
	"github.com/hosgf/element/client/k8s"
	"testing"
)

func Test(t *testing.T) {
	ctx := context.Background()
	kubernetes := k8s.New(true)
	kubernetes.Init("")
	kubernetes.Namespace().List(ctx)
	//kubernetes.Init()
}
