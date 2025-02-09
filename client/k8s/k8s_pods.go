package k8s

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/model/resource"
	"github.com/hosgf/element/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type podsOperation struct {
	*options
}

type Pod struct {
	Model
	Replicas    int32        `json:"replicas,omitempty"`
	RunningNode string       `json:"runningNode,omitempty"`
	Status      string       `json:"status,omitempty"`
	Config      []Config     `json:"config,omitempty"`
	Storage     []Storage    `json:"storage,omitempty"`
	Containers  []*Container `json:"containers,omitempty"`
}

func (pod *Pod) updateAppsDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	for k, v := range pod.labels() {
		deployment.ObjectMeta.Labels[k] = v
		deployment.Spec.Template.ObjectMeta.Labels[k] = v
	}
	deployment.Spec.Replicas = pod.replicas()
	deployment.Spec.Selector.MatchLabels = pod.toSelector()
	deployment.Spec.Template.Spec.Containers = pod.containers()
	deployment.Spec.Template.Spec.Volumes = pod.toVolumes()
	return deployment
}

func (pod *Pod) toAppsDeployment() *appsv1.Deployment {
	metadata := v1.ObjectMeta{
		Name:      pod.Name, // Pod 名称
		Namespace: pod.Namespace,
		Labels:    pod.labels(),
	}
	return &appsv1.Deployment{
		ObjectMeta: metadata,
		Spec: appsv1.DeploymentSpec{
			Replicas: pod.replicas(),
			Selector: &v1.LabelSelector{
				MatchLabels: pod.toSelector(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metadata,
				Spec: corev1.PodSpec{
					Containers: pod.containers(),
					Volumes:    pod.toVolumes(),
				},
			},
		},
	}
}

func (pod *Pod) setVolumes(vs []corev1.Volume) {
	if vs == nil || len(vs) < 1 {
		return
	}
	for _, v := range vs {
		vs := v.VolumeSource
		pvc := vs.PersistentVolumeClaim
		if pvc != nil {
			pod.Storage = append(pod.Storage, Storage{
				Name: v.Name,
				Type: types.StoragePVC,
				Item: pvc.ClaimName,
			})
			continue
		}
		ed := vs.EmptyDir
		if ed != nil {
			println(ed.Size())
			pod.Storage = append(pod.Storage, Storage{
				Name: v.Name,
				Type: types.StoragePVC,
				Item: ed.Medium,
			})
			continue
		}
	}
}

func (pod *Pod) toVolumes() []corev1.Volume {
	volumes := append(make([]corev1.Volume, 0), pod.toConfigVolumes()...)
	volumes = append(volumes, pod.toStorageVolumes()...)
	return volumes
}

func (pod *Pod) toConfigVolumes() []corev1.Volume {
	volumes := make([]corev1.Volume, 0)
	if pod.Config == nil || len(pod.Config) == 0 {
		return volumes
	}
	for _, config := range pod.Config {
		if len(config.Name) < 1 || len(config.Path) < 1 {
			continue
		}
		v := corev1.Volume{
			Name:         config.Name,
			VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: config.Item}},
		}
		volumes = append(volumes, v)
	}
	return volumes
}

func (pod *Pod) toStorageVolumes() []corev1.Volume {
	volumes := make([]corev1.Volume, 0)
	if pod.Storage == nil || len(pod.Storage) == 0 {
		return volumes
	}
	for _, storage := range pod.Storage {
		var (
			ok bool
			v  = &corev1.Volume{Name: storage.Name, VolumeSource: corev1.VolumeSource{}}
		)
		switch storage.ToStorageType() {
		case types.StoragePVC:
			ok = storage.toPvc(v)
		case types.StorageConfig:
			ok = storage.toConfig(v)
		default:
			ok = storage.toPvc(v)
		}
		if ok {
			volumes = append(volumes, *v)
		}
	}
	return volumes
}

func (s *Storage) toConfig(v *corev1.Volume) bool {
	items := gconv.Map(s.Item)
	if len(items) == 0 {
		return false
	}
	keys := make([]corev1.KeyToPath, 0, len(items))
	for k, v := range items {
		keys = append(keys, corev1.KeyToPath{Key: k, Path: gconv.String(v)})
	}
	v.VolumeSource.ConfigMap = &corev1.ConfigMapVolumeSource{Items: keys}
	return true
}

func (s *Storage) toPvc(v *corev1.Volume) bool {
	item := gconv.String(s.Item)
	if len(item) < 1 {
		return true
	}
	v.VolumeSource.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{ClaimName: item}
	return true
}

func (pod *Pod) ToProgress(svcs []*Service, metric *Metric, now int64) []*progress.Progress {
	list := make([]*progress.Progress, 0)
	cs := pod.Containers
	if len(cs) == 0 {
		return list
	}
	labels := pod.toLabels()
	var items map[string]map[types.ResourceType]MetricDetails
	if nil == metric {
		items = make(map[string]map[types.ResourceType]MetricDetails)
	} else {
		items = metric.Items
	}
	status := Status(pod.Status)
	group := progress.Details{
		Details: map[string]string{pod.Name: pod.Status},
		Status:  status,
	}
	cmap := map[string]types.StorageType{}
	if pod.Storage != nil {
		for _, c := range pod.Storage {
			cmap[c.Name] = c.Type
		}
	}
	for _, c := range cs {
		p := &progress.Progress{
			Namespace:  pod.Namespace,
			PID:        pod.Name,
			Name:       c.Name,
			Status:     status,
			Labels:     labels,
			Indicators: make(map[string]interface{}),
			Details:    make(map[string]interface{}),
			Time:       now,
		}
		p.Details["runningNode"] = pod.RunningNode
		metrics := items[c.Name]
		for _, res := range c.Resource {
			r := metrics[res.Type]
			p.Indicators[res.Type.String()] = resource.Details{
				Unit:  r.Unit,
				Free:  -1,
				Total: res.Maximum,
				Usage: r.Usage,
			}
		}
		if c.Mounts != nil {
			storage := progress.Details{
				Status:  health.UP,
				Details: map[string]string{},
			}
			for _, m := range c.Mounts {
				if d, ok := cmap[m.Name]; ok {
					storage.Details[m.Name] = d.String()
				}
			}
			p.Indicators["storage"] = storage
		}

		if len(svcs) < 1 {
			list = append(list, p)
			continue
		}
		ports := make([]progress.ProgressPort, 0)
		service := progress.Details{
			Status:  health.UNKNOWN,
			Details: map[string]string{},
		}
		for _, svc := range svcs {
			if gstr.Contains(svc.Group, p.Name) || gstr.Contains(svc.Group, pod.Group) || svc.Name == pod.Group {
				p.Service = svc.Name
				service.Details[svc.Name] = svc.Status
				service.Details["serviceType"] = svc.ServiceType
				service.Status = health.UP
				ports = append(ports, svc.toProgressPort()...)
				break
			}
		}
		if len(p.Service) > 0 {
			p.SetAddress(fmt.Sprintf("%s.%s.svc.cluster.local", p.Service, p.Namespace))
		}
		p.Indicators["group"] = group
		p.Indicators["service"] = service
		if len(ports) > 0 {
			p.Details["ports"] = ports
		}
		list = append(list, p)
	}
	return list
}

func (pod *Pod) replicas() *int32 {
	if pod.Replicas < 1 {
		pod.Replicas = 1
	}
	return &pod.Replicas
}

func (pod *Pod) containers() []corev1.Container {
	containers := make([]corev1.Container, 0, len(pod.Containers))
	for _, c := range pod.Containers {
		containers = append(containers, c.toContainer())
	}
	return containers
}

func (pod *Pod) toContainer(c corev1.Container) {
	container := &Container{
		Name:       c.Name,
		Image:      c.Image,
		PullPolicy: string(c.ImagePullPolicy),
		Command:    c.Command,
		Args:       c.Args,
		Ports:      make([]progress.Port, 0, len(c.Ports)),
		Resource:   make([]progress.Resource, 0),
		Mounts:     make([]Mount, 0),
		Env:        map[string]string{},
	}
	container.setResource(c)
	container.setMounts(pod.Storage, c)
	container.setEnv(c)
	container.setPorts(c)
	pod.Containers = append(pod.Containers, container)
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
	EnvConfig  []types.Environment `json:"envConfig,omitempty"`
	Mounts     []Mount             `json:"mounts,omitempty"`
	Probe      ProbeConfig         `json:"probe,omitempty"`
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
	// 设置env
	c.env(container)
	// 设置资源
	c.resource(container)
	// 设置存储
	c.mounts(container)
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
			Protocol:      corev1.Protocol(p.GetProtocol().String()),
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
			Protocol:   types.ProtocolType(port.Protocol),
		})
	}
}

func (c *Container) resource(container *corev1.Container) {
	var (
		cpu = progress.Resource{
			Type:    types.ResourceCPU,
			Unit:    "m",
			Minimum: types.DefaultMinimumCpu,
			Maximum: types.DefaultMaximumCpu,
		}
		memory = progress.Resource{
			Type:    types.ResourceMemory,
			Unit:    "Mi",
			Minimum: types.DefaultMinimumMemory,
			Maximum: types.DefaultMaximumMemory,
		}
	)
	if len(c.Resource) > 0 {
		for _, r := range c.Resource {
			switch r.Type {
			case types.ResourceCPU:
				cpu.Update(r)
			case types.ResourceMemory:
				memory.Update(r)
			}
		}
	}
	requests := corev1.ResourceList{}
	if cpu.Minimum > 1 {
		requests[corev1.ResourceCPU] = res.MustParse(types.ToCpuString(cpu.Minimum, cpu.Unit))
	}
	if memory.Minimum > 1 {
		requests[corev1.ResourceMemory] = res.MustParse(types.ToMemoryString(memory.Minimum, memory.Unit))
	}
	limits := corev1.ResourceList{}
	if cpu.Maximum > 1 {
		limits[corev1.ResourceCPU] = res.MustParse(types.ToCpuString(cpu.Maximum, cpu.Unit))
	}
	if memory.Maximum > 1 {
		limits[corev1.ResourceMemory] = res.MustParse(types.ToMemoryString(memory.Maximum, memory.Unit))
	}
	container.Resources = corev1.ResourceRequirements{Requests: requests, Limits: limits}
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

func (c *Container) mounts(container *corev1.Container) {
	if c.Mounts == nil || len(c.Mounts) < 1 {
		return
	}
	ms := make([]corev1.VolumeMount, 0, len(c.Mounts))
	for _, m := range c.Mounts {
		if len(m.Name) < 1 {
			continue
		}
		ms = append(ms, corev1.VolumeMount{
			Name:      m.Name,
			MountPath: m.GetPath(),
			SubPath:   m.SubPath,
		})
	}
	container.VolumeMounts = ms
}

func (c *Container) setMounts(storages []Storage, container corev1.Container) {
	vm := container.VolumeMounts
	if storages == nil || len(storages) < 1 || vm == nil || len(vm) < 1 {
		return
	}
	cmap := map[string]string{}
	for _, c := range storages {
		cmap[c.Name] = c.Scope
	}
	for _, v := range vm {
		if _, ok := cmap[v.Name]; ok {
			c.Mounts = append(c.Mounts, Mount{
				Name:    v.Name,
				Path:    v.MountPath,
				SubPath: v.SubPath,
			})
		}
	}
}

func (pg *ProcessGroupConfig) toPod() *Pod {
	if pg.Process == nil || len(pg.Process) < 1 {
		return nil
	}
	labels := &pg.Labels
	if len(labels.Group) < 1 {
		labels.Group = pg.GroupName
	}
	pg.initConfig()
	pod := &Pod{
		Model:       Model{Namespace: pg.Namespace, Name: pg.GroupName, AllowUpdate: pg.AllowUpdate},
		Replicas:    pg.Replicas,
		RunningNode: pg.RunningNode,
		Config:      pg.Config,
		Storage:     pg.Storage,
		Containers:  make([]*Container, 0),
	}
	pod.setTypesLabels(labels)
	for _, p := range pg.Process {
		c := p.toContainer(pg)
		if c == nil {
			continue
		}
		pod.Containers = append(pod.Containers, c)
	}
	return pod
}

func (p *ProcessConfig) toContainer(pg *ProcessGroupConfig) *Container {
	p.toMounts(pg)
	env, envConfig := p.toEnvConfig()
	c := &Container{
		Name:       p.Name,
		Image:      p.Source,
		PullPolicy: p.PullPolicy,
		Command:    p.Command,
		Args:       p.Args,
		Ports:      p.Ports,
		Resource:   p.Resource,
		Env:        env,
		EnvConfig:  envConfig,
		Mounts:     p.Mounts,
		Probe:      p.Probe,
	}
	return c
}

func (o *podsOperation) Get(ctx context.Context, namespace, appname string) ([]*Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := toAppListOptions(appname)
	return o.pods(ctx, namespace, opts)
}

func (o *podsOperation) List(ctx context.Context, namespace string, groups ...string) ([]*Pod, error) {
	if o.err != nil {
		return nil, o.err
	}
	if groups == nil || len(groups) == 0 {
		datas, err := o.list(ctx, namespace, "")
		return o.toPods(namespace, datas), err
	}
	pods := make([]*Pod, 0)
	for _, g := range groups {
		if len(g) < 1 {
			continue
		}
		datas, err := o.list(ctx, namespace, g)
		if err != nil {
			return nil, err
		}
		pods = append(pods, o.toPods(namespace, datas)...)
	}
	return pods, nil
}

func (o *podsOperation) Exists(ctx context.Context, namespace, pod string) (bool, error) {
	if o.err != nil {
		return false, o.err
	}
	opts := v1.GetOptions{}
	p, err := o.api.CoreV1().Pods(namespace).Get(ctx, pod, opts)
	return o.isExist(p, err, "Failed to get Pod: %v")
}

func (o *podsOperation) Apply(ctx context.Context, pod *Pod) error {
	if o.err != nil {
		return o.err
	}
	if has, datas, err := o.deploymentExists(ctx, pod.Namespace, pod.Name); has {
		if err != nil {
			return err
		}
		if pod.AllowUpdate {
			return o.update(ctx, datas, pod)
		}
		return gerror.NewCodef(gcode.CodeNotImplemented, "Namespace: %s, Pod: %s 已存在!", pod.Namespace, pod.Name)
	}
	return o.create(ctx, pod)
}

func (o *podsOperation) Delete(ctx context.Context, namespace, pod string) error {
	if o.err != nil {
		return o.err
	}
	if has, err := o.Exists(ctx, namespace, pod); has {
		if err != nil {
			return err
		}
		return o.delete(ctx, namespace, pod)
	}
	return nil
}

func (o *podsOperation) DeleteGroup(ctx context.Context, namespace string, groups ...string) error {
	if o.err != nil {
		return o.err
	}
	for _, group := range groups {
		if len(group) < 1 {
			continue
		}
		if err := o.deleteDeployment(ctx, namespace, group); err != nil {
			return err
		}
		datas, err := o.list(ctx, namespace, group)
		if err != nil {
			return err
		}
		if datas.Items == nil || len(datas.Items) == 0 {
			continue
		}
		if err := o.deleteGroup(ctx, namespace, group); err != nil {
			return err
		}
	}
	return nil
}

func (o *podsOperation) Restart(ctx context.Context, namespace, pod string) error {
	exist, err := o.Exists(ctx, namespace, pod)
	if err != nil || !exist {
		return err
	}
	opts := v1.DeleteOptions{}
	return o.api.AppsV1().Deployments(namespace).Delete(ctx, pod, opts)
}

func (o *podsOperation) RestartApp(ctx context.Context, namespace, appname string) error {
	if o.err != nil {
		return o.err
	}
	opts := toAppListOptions(appname)
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

func (o *podsOperation) deploymentExists(ctx context.Context, namespace string, group string) (bool, []appsv1.Deployment, error) {
	if o.isTest {
		return false, nil, nil
	}
	datas, err := o.api.AppsV1().Deployments(namespace).List(ctx, toGroupListOptions(group))
	items := datas.Items
	if len(items) < 1 {
		return false, items, nil
	}
	has, err := o.isExist(items, err, "Failed to get Pod: %v")
	return has, items, err
}

func (o *podsOperation) deleteDeployment(ctx context.Context, namespace string, group string) error {
	if has, _, err := o.deploymentExists(ctx, namespace, group); has {
		if err != nil {
			return err
		}
		err := o.api.AppsV1().Deployments(namespace).Delete(ctx, group, v1.DeleteOptions{})
		if err != nil {
			return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to delete apps deployment: %v", err)
		}
	}
	return nil
}

func (o *podsOperation) create(ctx context.Context, pod *Pod) error {
	deployment := pod.toAppsDeployment()
	if o.isTest {
		return nil
	}
	_, err := o.api.AppsV1().Deployments(pod.Namespace).Create(ctx, deployment, v1.CreateOptions{})
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create apps Deployment: %v", err)
	}
	return nil
}

func (o *podsOperation) update(ctx context.Context, deployments []appsv1.Deployment, pod *Pod) error {
	//err := o.api.AppsV1().ReplicaSets(pod.Namespace).DeleteCollection(ctx, v1.DeleteOptions{}, toGroupListOptions(pod.Group))
	//if err != nil {
	//	return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to update apps ReplicaSets: %v", err)
	//}
	if deployments == nil {
		return nil
	}
	for _, deployment := range deployments {
		_, err := o.api.AppsV1().Deployments(pod.Namespace).Update(ctx, pod.updateAppsDeployment(&deployment), v1.UpdateOptions{})
		if err != nil {
			return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to update apps Deployment: %v", err)
		}
	}
	return nil
}

func (o *podsOperation) delete(ctx context.Context, namespace string, pod string) error {
	opts := v1.DeleteOptions{}
	err := o.api.CoreV1().Pods(namespace).Delete(ctx, pod, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to delete pod: %v", err)
	}
	return nil
}

func (o *podsOperation) deleteGroup(ctx context.Context, namespace string, group string) error {
	err := o.api.CoreV1().Pods(namespace).DeleteCollection(ctx, v1.DeleteOptions{}, toGroupListOptions(group))
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to delete PodList: %v", err)
	}
	return nil
}

func (o *podsOperation) list(ctx context.Context, namespace string, group string) (*corev1.PodList, error) {
	datas, err := o.api.CoreV1().Pods(namespace).List(ctx, toGroupListOptions(group))
	if err != nil {
		return datas, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get pods: %v", err)
	}
	return datas, nil
}

func (o *podsOperation) pods(ctx context.Context, namespace string, opts v1.ListOptions) ([]*Pod, error) {
	datas, err := o.api.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get pods: %v", err)
	}
	return o.toPods(namespace, datas), nil
}

func (o *podsOperation) toPods(namespace string, datas *corev1.PodList) []*Pod {
	if datas == nil {
		return nil
	}
	pods := make([]*Pod, 0, len(datas.Items))
	for _, p := range datas.Items {
		pod := &Pod{
			Model:       Model{Namespace: namespace, Name: p.Name},
			Status:      string(p.Status.Phase),
			RunningNode: p.Spec.NodeName,
			Containers:  make([]*Container, 0),
			Config:      make([]Config, 0),
			Storage:     make([]Storage, 0),
		}
		pod.setLabels(p.Labels)
		pod.setVolumes(p.Spec.Volumes)
		for _, c := range p.Spec.Containers {
			pod.toContainer(c)
		}
		pods = append(pods, pod)
	}
	return pods
}
