package k8s

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type nodesOperation struct {
	*options
}

func (o *nodesOperation) Top(ctx context.Context) ([]string, error) {
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
