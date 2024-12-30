package k8s

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Pod struct {
	progress.Service
	Containers []Container `json:"containers,omitempty"`
}

type Container struct {
	Name    string          `json:"name,omitempty"`
	Image   string          `json:"image,omitempty"`
	Command []string        `json:"command,omitempty"`
	Args    []string        `json:"args,omitempty"`
	Ports   []progress.Port `json:"ports,omitempty"`
	Cpu     string          `json:"cpu,omitempty"`
	Memory  string          `json:"memory,omitempty"`
}

func (k *kubernetes) GetPod(ctx context.Context, namespace, appname string) ([]*Pod, error) {
	if k.err != nil {
		return nil, k.err
	}
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
	}
	return k.pods(ctx, namespace, opts)
}

func (k *kubernetes) GetPods(ctx context.Context, namespace string) ([]*Pod, error) {
	if k.err != nil {
		return nil, k.err
	}
	opts := v1.ListOptions{
		//LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
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

func (k *kubernetes) CreatePod(ctx context.Context, pod Pod) error {
	if k.err != nil {
		return k.err
	}
	containers := make([]corev1.Container, 0, len(pod.Containers))
	for _, c := range pod.Containers {
		containers = append(containers, corev1.Container{
			Name:     c.Name,
			Port:     p.Port,
			NodePort: p.NodePort,
		})
	}
	p := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      pod.Name, // Pod 名称
			Namespace: pod.Namespace,
			Labels: map[string]string{
				types.LabelApp.String():   pod.App,
				types.LabelOwner.String(): pod.Owner,
				types.LabelType.String():  pod.GroupType,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx-container",
					Image: "nginx", // 使用 Nginx 镜像
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	opts := v1.CreateOptions{}
	_, err := k.api.CoreV1().Pods(pod.Namespace).Create(ctx, p, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create pod: %v", err)
	}
	return nil
}

func (k *kubernetes) DeletePod(ctx context.Context, namespace, pod string) error {
	if k.err != nil {
		return k.err
	}
	opts := v1.DeleteOptions{}
	return k.api.CoreV1().Pods(namespace).Delete(ctx, pod, opts)
}

func (k *kubernetes) RestartPod(ctx context.Context, namespace, pod string) error {
	exist, err := k.PodIsExist(ctx, namespace, pod)
	if err != nil || !exist {
		return err
	}
	return k.api.CoreV1().Pods(namespace).Delete(ctx, pod, v1.DeleteOptions{})
}

func (k *kubernetes) RestartAppPods(ctx context.Context, namespace, appname string) error {
	if k.err != nil {
		return k.err
	}
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
	}
	corev1 := k.api.CoreV1().Pods(namespace)
	list, err := corev1.List(ctx, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get pods: %v", err)
	}
	for _, p := range list.Items {
		if p.Name != "" {
			err = corev1.Delete(ctx, p.Name, v1.DeleteOptions{})
		}
	}
	return err
}

func (k *kubernetes) pods(ctx context.Context, namespace string, opts v1.ListOptions) ([]*Pod, error) {
	list, err := k.api.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get pods: %v", err)
	}
	pods := make([]*Pod, 0, len(list.Items))
	for _, p := range list.Items {
		pods = append(pods, &Pod{
			Service: progress.Service{
				Namespace: p.Namespace,
				Name:      p.Name,
				App:       p.Labels[types.LabelApp.String()],
				Owner:     p.Labels[types.LabelOwner.String()],
				GroupType: p.Labels[types.LabelType.String()],
				Status:    Status(string(p.Status.Phase)),
			},
		})
	}
	return pods, nil
}
