package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type namespaceOperation struct {
	*options
}

func (o *namespaceOperation) List(ctx context.Context) ([]string, error) {
	if o.err != nil {
		return nil, o.err
	}
	ns, err := o.api.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	namespaces := make([]string, 0, len(ns.Items))
	for _, namespace := range ns.Items {
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

func (o *namespaceOperation) IsExist(ctx context.Context, namespace string) (bool, error) {
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
