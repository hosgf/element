package k8s

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/process"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type podTemplateOperation struct {
	*options
}

type PodTemplate struct {
	Model
	RunningNode string          `json:"runningNode,omitempty"`
	Status      health.Health   `json:"status,omitempty"`
	Config      []types.Config  `json:"config,omitempty"`
	Storage     []types.Storage `json:"storage,omitempty"`
	Containers  []*Container    `json:"containers,omitempty"`
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
	container := &Container{
		Name:       c.Name,
		Image:      c.Image,
		PullPolicy: string(c.ImagePullPolicy),
		Command:    c.Command,
		Args:       c.Args,
		Ports:      make([]process.Port, 0, len(c.Ports)),
		Resource:   make([]process.Resource, 0),
		Env:        map[string]string{},
	}
	container.setResource(c)
	container.setMounts(p.Storage, c)
	container.setEnv(c)
	container.setPorts(c)
	p.Containers = append(p.Containers, container)
}

func (o *podTemplateOperation) Get(ctx context.Context, namespace, appname string) ([]*Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := toAppListOptions(appname)
	return o.pods(ctx, namespace, opts)
}

func (o *podTemplateOperation) List(ctx context.Context, namespace string) ([]*Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{}
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

func (o *podTemplateOperation) Apply(ctx context.Context, pod *Pod) error {
	if o.err != nil {
		return o.err
	}
	metadata := v1.ObjectMeta{
		Name:      pod.Name, // Pod 名称
		Namespace: pod.Namespace,
		Labels:    pod.labels(),
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

func (o *podTemplateOperation) Delete(ctx context.Context, namespace, pod string) error {
	if o.err != nil {
		return o.err
	}
	opts := v1.DeleteOptions{}
	return o.api.CoreV1().Pods(namespace).Delete(ctx, pod, opts)
}

func (o *podTemplateOperation) pods(ctx context.Context, namespace string, opts v1.ListOptions) ([]*Pod, error) {
	datas, err := o.api.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get pods: %v", err)
	}
	pods := make([]*Pod, 0, len(datas.Items))
	for _, p := range datas.Items {
		pod := &Pod{
			Model: Model{
				Namespace: namespace,
				Name:      p.Name,
			},
			Containers:  make([]*Container, 0),
			RunningNode: p.Spec.NodeName,
			Status:      string(p.Status.Phase),
		}
		pod.setLabels(p.Labels)
		for _, c := range p.Spec.Containers {
			pod.toContainer(c)
		}
		pods = append(pods, pod)
	}
	return pods, nil
}
