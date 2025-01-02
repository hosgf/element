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
	res "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type podsOperation struct {
	*options
}

type Pod struct {
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

func (p *Pod) toLabel() map[string]string {
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

func (p *Pod) labels(labels map[string]string) {
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

func (p *Pod) containers() []corev1.Container {
	containers := make([]corev1.Container, 0, len(p.Containers))
	for _, c := range p.Containers {
		containers = append(containers, c.toContainer())
	}
	return containers
}

func (p *Pod) toContainer(c corev1.Container) {
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

type Container struct {
	Name       string              `json:"name,omitempty"`
	Image      string              `json:"image,omitempty"`
	PullPolicy string              `json:"string,omitempty"`
	Command    []string            `json:"command,omitempty"`
	Args       []string            `json:"args,omitempty"`
	Ports      []progress.Port     `json:"ports,omitempty"`
	Resource   []progress.Resource `json:"resource,omitempty"`
	Env        map[string]string   `json:"env,omitempty"`
}

func (c *Container) toContainer() corev1.Container {
	container := &corev1.Container{
		Name:    c.Name,
		Image:   c.Image,
		Command: c.Command,
		Args:    c.Args,
	}
	if len(c.PullPolicy) < 1 {
		container.ImagePullPolicy = corev1.PullIfNotPresent
	} else {
		container.ImagePullPolicy = corev1.PullPolicy(c.PullPolicy)
	}
	// 设置Port
	c.ports(container)
	// 设置资源
	c.resource(container)
	// 设置env
	c.env(container)
	// todo pvc
	return *container
}

func (c *Container) ports(container *corev1.Container) {
	list := c.Ports
	if list == nil || len(list) < 1 {
		return
	}
	ports := make([]corev1.ContainerPort, 0, len(list))
	for _, p := range list {
		ports = append(ports, corev1.ContainerPort{
			Name:          p.Name,
			Protocol:      corev1.Protocol(p.Protocol),
			ContainerPort: p.TargetPort,
		})
	}
	container.Ports = ports
}

func (c *Container) setPorts(container corev1.Container) {
	for _, port := range container.Ports {
		c.Ports = append(c.Ports, progress.Port{
			Name:       port.Name,
			TargetPort: port.ContainerPort,
			Protocol:   string(port.Protocol),
		})
	}
}

func (c *Container) resource(container *corev1.Container) {
	var (
		cpu = progress.Resource{
			Type:    types.ResourceCPU,
			Unit:    "m",
			Minimum: 500,
			Maximum: 1000,
		}
		memory = progress.Resource{
			Type:    types.ResourceMemory,
			Unit:    "Mi",
			Minimum: 30,
			Maximum: 500,
		}
	)
	if len(c.Resource) < 1 {
		for _, r := range c.Resource {
			switch r.Type {
			case types.ResourceCPU:
				cpu.Update(r)
			case types.ResourceMemory:
				memory.Update(r)
			}
		}
	}
	container.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    *res.NewQuantity(types.FormatCpu(cpu.Minimum, cpu.Unit), res.DecimalSI),
			corev1.ResourceMemory: *res.NewQuantity(types.FormatMemory(cpu.Minimum, cpu.Unit), res.DecimalSI),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    *res.NewQuantity(types.FormatCpu(cpu.Maximum, cpu.Unit), res.DecimalSI),
			corev1.ResourceMemory: *res.NewQuantity(types.FormatMemory(cpu.Maximum, cpu.Unit), res.DecimalSI),
		},
	}
}

func (c *Container) setResource(container corev1.Container) {
	resources := container.Resources
	requests := resources.Requests
	limits := resources.Limits
	cpu := progress.Resource{Type: types.ResourceCPU, Threshold: -1, Minimum: -1, Maximum: -1}
	cpu.SetMinimum(requests.Cpu().String())
	cpu.SetMaximum(limits.Cpu().String())
	memory := progress.Resource{Type: types.ResourceMemory, Threshold: -1, Minimum: -1, Maximum: -1}
	memory.SetMinimum(requests.Memory().String())
	memory.SetMaximum(limits.Memory().String())
	c.Resource = append(c.Resource, cpu, memory)
}

func (c *Container) env(container *corev1.Container) {
	env := c.Env
	if env == nil || len(env) < 1 {
		return
	}
	envVars := make([]corev1.EnvVar, 0, len(env))
	for k, v := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	container.Env = envVars
}

func (c *Container) setEnv(container corev1.Container) {
	for _, envVar := range container.Env {
		c.Env[envVar.Name] = envVar.Value
	}
}

func (o *podsOperation) Get(ctx context.Context, namespace, appname string) ([]Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
	}
	return o.pods(ctx, namespace, opts)
}

func (o *podsOperation) List(ctx context.Context, namespace string) ([]Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{
		//LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
	}
	return o.pods(ctx, namespace, opts)
}

func (o *podsOperation) Exists(ctx context.Context, namespace, pod string) (bool, error) {
	if o.err != nil {
		return false, o.err
	}
	opts := v1.GetOptions{}
	p, err := o.api.CoreV1().Pods(namespace).Get(ctx, pod, opts)
	return o.isExist(p, err, "Failed to get Pod: %v")
}

func (o *podsOperation) Create(ctx context.Context, pod Pod) error {
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

func (o *podsOperation) Apply(ctx context.Context, pod Pod) error {
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

func (o *podsOperation) Delete(ctx context.Context, namespace, pod string) error {
	if o.err != nil {
		return o.err
	}
	opts := v1.DeleteOptions{}
	return o.api.CoreV1().Pods(namespace).Delete(ctx, pod, opts)
}

func (o *podsOperation) Restart(ctx context.Context, namespace, pod string) error {
	exist, err := o.Exists(ctx, namespace, pod)
	if err != nil || !exist {
		return err
	}
	return o.api.CoreV1().Pods(namespace).Delete(ctx, pod, v1.DeleteOptions{})
}

func (o *podsOperation) RestartApp(ctx context.Context, namespace, appname string) error {
	if o.err != nil {
		return o.err
	}
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.LabelApp, appname),
	}
	corev1 := o.api.CoreV1().Pods(namespace)
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

func (o *podsOperation) pods(ctx context.Context, namespace string, opts v1.ListOptions) ([]Pod, error) {
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
