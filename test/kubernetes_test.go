package test

import (
	"context"
	"fmt"
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

func TestParse(t *testing.T) {
	fmt.Println(types.Parse("16384Mi"))
	fmt.Println(types.Parse("16384"))
	fmt.Println(types.Parse("16384u"))
	fmt.Println(types.Parse(""))
}
