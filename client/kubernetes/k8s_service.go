package kubernetes

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/model/progress"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *kubernetes) GetServices(ctx context.Context, namespace string) ([]*progress.Service, error) {
	if k.err != nil {
		return nil, k.err
	}
	opts := v1.ListOptions{}
	svcs, err := k.api.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get Services: %v", err)
	}
	services := make([]*progress.Service, 0, len(svcs.Items))
	for _, svc := range svcs.Items {
		services = append(services, &progress.Service{
			Namespace: namespace,
			Name:      svc.Name,
			App:       svc.Labels["app"],
			Status:    Status(svc.Status.String()),
		})
	}
	return services, nil
}

func (k *kubernetes) ServiceIsExist(ctx context.Context, namespace, service string) (bool, error) {
	if k.err != nil {
		return false, k.err
	}
	opts := v1.GetOptions{}
	svc, err := k.api.CoreV1().Services(namespace).Get(ctx, service, opts)
	return k.isExist(svc, err, "Failed to get Services: %v")
}

func (k *kubernetes) DeleteService(ctx context.Context, namespace, service string) error {
	if k.err != nil {
		return k.err
	}
	opts := v1.DeleteOptions{}
	return k.api.CoreV1().Services(namespace).Delete(ctx, service, opts)
}
