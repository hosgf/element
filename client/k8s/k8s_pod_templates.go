package k8s

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type podTemplateOperation struct {
	*options
}

type PodTemplate struct {
	Namespace   string            `json:"namespace,omitempty"`
	Name        string            `json:"name,omitempty"`
	App         string            `json:"app,omitempty"`
	Group       string            `json:"group,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Scope       string            `json:"scope,omitempty"`
	RunningNode string            `json:"runningNode,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Status      health.Health     `json:"status,omitempty"`
	Containers  []Container       `json:"containers,omitempty"`
}

func (p *PodTemplate) toLabel() map[string]string {
	labels := map[string]string{
		types.LabelApp.String():   p.App,
		types.LabelOwner.String(): p.Owner,
		types.LabelScope.String(): p.Scope,
		types.LabelGroup.String(): p.Group,
	}
	if p.Labels != nil {
		for k, v := range p.Labels {
			labels[k] = v
		}
	}
	return labels
}

func (p *PodTemplate) labels(labels map[string]string) {
	if len(labels) < 1 {
		return
	}
	p.App = labels[types.LabelApp.String()]
	p.Owner = labels[types.LabelOwner.String()]
	p.Scope = labels[types.LabelScope.String()]
	p.Group = labels[types.LabelGroup.String()]
	delete(labels, types.LabelApp.String())
	delete(labels, types.LabelOwner.String())
	delete(labels, types.LabelScope.String())
	delete(labels, types.LabelGroup.String())
	if p.Labels == nil {
		p.Labels = map[string]string{}
	}
	for k, v := range labels {
		p.Labels[k] = v
	}
}

func (p *PodTemplate) containers() []corev1.Container {
	containers := make([]corev1.Container, 0, len(p.Containers))
	for _, c := range p.Containers {
		containers = append(containers, c.toContainer())
	}
	return containers
}

func (p *PodTemplate) toContainer(c corev1.Container) {
	container := Container{
		Name:       c.Name,
		Image:      c.Image,
		PullPolicy: string(c.ImagePullPolicy),
		Command:    c.Command,
		Args:       c.Args,
		Ports:      make([]progress.Port, 0, len(c.Ports)),
		Resource:   make([]progress.Resource, 0),
		Env:        map[string]string{},
	}
	container.setResource(c)
	container.setEnv(c)
	container.setPorts(c)
	p.Containers = append(p.Containers, container)
}

func (o *podTemplateOperation) Get(ctx context.Context, namespace, appname string) ([]Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
	}
	return o.pods(ctx, namespace, opts)
}

func (o *podTemplateOperation) List(ctx context.Context, namespace string) ([]Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{
		//LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
	}
	return o.pods(ctx, namespace, opts)
}

func (o *podTemplateOperation) Exists(ctx context.Context, namespace, pod string) (bool, error) {
	if o.err != nil {
		return false, o.err
	}
	opts := v1.GetOptions{}
	p, err := o.api.CoreV1().Pods(namespace).Get(ctx, pod, opts)
	return o.isExist(p, err, "Failed to get Pod: %v")
}

func (o *podTemplateOperation) Create(ctx context.Context, pod Pod) error {
	if o.err != nil {
		return o.err
	}
	metadata := v1.ObjectMeta{
		Name:      pod.Name, // Pod 名称
		Namespace: pod.Namespace,
		Labels:    pod.toLabel(),
	}
	p := &corev1.PodTemplate{
		ObjectMeta: metadata,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metadata,
			Spec: corev1.PodSpec{
				Containers: pod.containers(),
			},
		},
	}
	opts := v1.CreateOptions{}
	_, err := o.api.CoreV1().PodTemplates(pod.Namespace).Create(ctx, p, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create pod: %v", err)
	}
	return nil
}

func (o *podTemplateOperation) Apply(ctx context.Context, pod Pod) error {
	if o.err != nil {
		return o.err
	}
	p := applyconfigurationscorev1.Pod(pod.Name, pod.Namespace)
	// 更新Labels
	p.WithLabels(pod.toLabel())
	// 更新容器
	p.Spec.Containers = make([]applyconfigurationscorev1.ContainerApplyConfiguration, 0, len(pod.containers()))
	for _, c := range pod.containers() {
		ports := make([]*applyconfigurationscorev1.ContainerPortApplyConfiguration, 0, len(c.Ports))
		for _, p := range c.Ports {
			port := applyconfigurationscorev1.ContainerPort()
			port.WithName(p.Name)
			port.WithProtocol(p.Protocol)
			port.WithHostPort(p.HostPort)
			port.WithHostIP(p.HostIP)
			port.WithContainerPort(p.ContainerPort)
			ports = append(ports, port)
		}
		envs := make([]*applyconfigurationscorev1.EnvVarApplyConfiguration, 0, len(c.Env))
		for _, e := range c.Env {
			env := applyconfigurationscorev1.EnvVar()
			env.WithName(e.Name)
			env.WithValue(e.Value)
			envs = append(envs, env)
		}
		container := applyconfigurationscorev1.Container()
		container.WithName(c.Name)
		container.WithImage(c.Image)
		container.WithImagePullPolicy(c.ImagePullPolicy)
		container.WithCommand(c.Command...)
		container.WithArgs(c.Args...)
		container.WithPorts(ports...)
		container.WithEnv(envs...)
		p.Spec.Containers = append(p.Spec.Containers, *container)
	}
	opts := v1.ApplyOptions{}
	_, err := o.api.CoreV1().Pods(pod.Namespace).Apply(ctx, p, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to apply pod: %v", err)
	}
	return nil
}

func (o *podTemplateOperation) Delete(ctx context.Context, namespace, pod string) error {
	if o.err != nil {
		return o.err
	}
	opts := v1.DeleteOptions{}
	return o.api.CoreV1().Pods(namespace).Delete(ctx, pod, opts)
}

func (o *podTemplateOperation) pods(ctx context.Context, namespace string, opts v1.ListOptions) ([]Pod, error) {
	datas, err := o.api.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get pods: %v", err)
	}
	pods := make([]Pod, 0, len(datas.Items))
	for _, p := range datas.Items {
		pod := Pod{
			Namespace:   namespace,
			Name:        p.Name,
			Containers:  make([]Container, 0),
			RunningNode: p.Spec.NodeName,
			Status:      Status(string(p.Status.Phase)),
		}
		pod.labels(p.Labels)
		for _, c := range p.Spec.Containers {
			pod.toContainer(c)
		}
		pods = append(pods, pod)
	}
	return pods, nil
}
