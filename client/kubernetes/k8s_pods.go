package kubernetes

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/health"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Pod struct {
	Namespace string        `json:"namespace,omitempty"`
	App       string        `json:"app,omitempty"`
	Name      string        `json:"name,omitempty"`
	Status    health.Health `json:"status,omitempty"`
	Cpu       string        `json:"cpu,omitempty"`
	Memory    string        `json:"memory,omitempty"`
}

func (k *kubernetes) GetPod(ctx context.Context, namespace, appname string) ([]*Pod, error) {
	if k.err != nil {
		return nil, k.err
	}
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", appname),
	}
	return k.pods(ctx, namespace, opts)
}

func (k *kubernetes) GetPods(ctx context.Context, namespace string) ([]*Pod, error) {
	if k.err != nil {
		return nil, k.err
	}
	opts := v1.ListOptions{
		//LabelSelector: fmt.Sprintf("app=%s", name),
	}
	return k.pods(ctx, namespace, opts)
}

func (k *kubernetes) PodIsExist(ctx context.Context, namespace, pod string) (bool, error) {
	if k.err != nil {
		return false, k.err
	}
	opts := v1.GetOptions{}
	p, err := k.api.CoreV1().Pods(namespace).Get(ctx, pod, opts)
	return k.isExist(p, err, "Failed to get Pod: %v")
}

func (k *kubernetes) pods(ctx context.Context, namespace string, opts v1.ListOptions) ([]*Pod, error) {
	list, err := k.api.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get pods: %v", err)
	}
	pods := make([]*Pod, 0, len(list.Items))
	for _, p := range list.Items {
		pods = append(pods, &Pod{
			Namespace: p.Namespace,
			Name:      p.Name,
			App:       p.Labels["app"],
			Status:    Status(string(p.Status.Phase)),
		})
	}
	return pods, nil
}
