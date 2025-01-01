package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/model/resource"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type namespaceOperation struct {
	*options
}

func (o *namespaceOperation) List(ctx context.Context) ([]resource.Namespace, error) {
	if o.err != nil {
		return nil, o.err
	}
	datas, err := o.api.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	namespaces := make([]resource.Namespace, 0, len(datas.Items))
	for _, ns := range datas.Items {
		namespaces = append(namespaces, resource.Namespace{
			Name:   ns.Name,
			Status: Status(string(ns.Status.Phase)),
		})
	}
	return namespaces, nil
}

func (o *namespaceOperation) Exists(ctx context.Context, namespace string) (bool, error) {
	if o.err != nil {
		return false, o.err
	}
	_, err := o.api.CoreV1().Namespaces().Get(ctx, namespace, v1.GetOptions{})
	return o.isExist("", err, "Error occurred while fetching namespace: %v")
}

func (o *namespaceOperation) Create(ctx context.Context, namespace string) (bool, error) {
	if o.err != nil {
		return false, o.err
	}
	_, err := o.api.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: namespace}}, v1.CreateOptions{})
	if err != nil {
		return false, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create namespace: %v", err)
	}
	return true, nil
}

func (o *namespaceOperation) Delete(ctx context.Context, namespace string) error {
	if o.err != nil {
		return o.err
	}
	return o.api.CoreV1().Namespaces().Delete(ctx, namespace, v1.DeleteOptions{})
}
