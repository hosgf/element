package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *kubernetes) GetNamespaces(ctx context.Context) ([]string, error) {
	if k.err != nil {
		return nil, k.err
	}
	ns, err := k.api.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	namespaces := make([]string, 0, len(ns.Items))
	for _, namespace := range ns.Items {
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

func (k *kubernetes) NamespaceIsExist(ctx context.Context, namespace string) (bool, error) {
	if k.err != nil {
		return false, k.err
	}
	_, err := k.api.CoreV1().Namespaces().Get(ctx, namespace, v1.GetOptions{})
	return k.isExist("", err, "Error occurred while fetching namespace: %v")
}

func (k *kubernetes) CreateNamespace(ctx context.Context, namespace string) (bool, error) {
	if k.err != nil {
		return false, k.err
	}
	_, err := k.api.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: namespace}}, v1.CreateOptions{})
	if err != nil {
		return false, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create namespace: %v", err)
	}
	return true, nil
}

func (k *kubernetes) DeleteNamespace(ctx context.Context, namespace string) error {
	if k.err != nil {
		return k.err
	}
	return k.api.CoreV1().Namespaces().Delete(ctx, namespace, v1.DeleteOptions{})
}
