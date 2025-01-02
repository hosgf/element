package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type serviceOperation struct {
	*options
}

type Service struct {
	Namespace   string            `json:"namespace,omitempty"`
	Name        string            `json:"name,omitempty"`
	ServiceType string            `json:"serviceType,omitempty"`
	App         string            `json:"app,omitempty"`
	Group       string            `json:"group,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Scope       string            `json:"scope,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Status      health.Health     `json:"status,omitempty"`
	Ports       []progress.Port   `json:"ports,omitempty"`
}

func (s *Service) toLabel() map[string]string {
	labels := map[string]string{
		types.LabelApp.String():   s.App,
		types.LabelOwner.String(): s.Owner,
		types.LabelScope.String(): s.Scope,
		types.LabelGroup.String(): s.Group,
	}
	if s.Labels != nil {
		for k, v := range s.Labels {
			labels[k] = v
		}
	}
	return labels
}

func (s *Service) labels(labels map[string]string) {
	if len(labels) < 1 {
		return
	}
	s.App = labels[types.LabelApp.String()]
	s.Owner = labels[types.LabelOwner.String()]
	s.Scope = labels[types.LabelScope.String()]
	s.Group = labels[types.LabelGroup.String()]
	delete(labels, types.LabelApp.String())
	delete(labels, types.LabelOwner.String())
	delete(labels, types.LabelScope.String())
	delete(labels, types.LabelGroup.String())
	if s.Labels == nil {
		s.Labels = map[string]string{}
	}
	for k, v := range labels {
		s.Labels[k] = v
	}
}

func (s *Service) toSelector() map[string]string {
	return map[string]string{
		types.LabelGroup.String(): s.Group,
	}
}

func (o *serviceOperation) List(ctx context.Context, namespace string) ([]Service, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{}
	svcs, err := o.api.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get Services: %v", err)
	}
	services := make([]Service, 0, len(svcs.Items))
	for _, svc := range svcs.Items {
		service := Service{
			Namespace: namespace,
			Name:      svc.Name,
			Group:     svc.Spec.Selector[types.LabelGroup.String()],
			Status:    Status(svc.Status.String()),
		}
		service.labels(svc.Labels)
		services = append(services, service)
	}
	return services, nil
}

func (o *serviceOperation) Exists(ctx context.Context, namespace, service string) (bool, error) {
	if o.err != nil {
		return false, o.err
	}
	opts := v1.GetOptions{}
	svc, err := o.api.CoreV1().Services(namespace).Get(ctx, service, opts)
	return o.isExist(svc, err, "Failed to get Services: %v")
}

func (o *serviceOperation) Create(ctx context.Context, service Service) error {
	if o.err != nil {
		return o.err
	}
	ports := make([]corev1.ServicePort, 0, len(service.Ports))
	for _, p := range service.Ports {
		ports = append(ports, corev1.ServicePort{
			Protocol:   corev1.Protocol(p.Protocol),
			Port:       p.Port,                         // 对外暴露的端口
			TargetPort: intstr.FromInt32(p.TargetPort), // Pod 内部服务监听的端口
			NodePort:   p.NodePort,
		})
	}
	svc := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      service.Name,      // Service 名称
			Namespace: service.Namespace, // Service 所在的 Namespace
			Labels:    service.toLabel(),
		},
		Spec: corev1.ServiceSpec{
			Selector: service.toSelector(),
			Ports:    ports,
			Type:     any(len(service.ServiceType) < 1, corev1.ServiceTypeClusterIP, corev1.ServiceType(service.ServiceType)), // 默认为 ClusterIP 类型
		},
	}
	opts := v1.CreateOptions{}
	_, err := o.api.CoreV1().Services(service.Namespace).Create(ctx, svc, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create service: %v", err)
	}
	return nil
}

func (o *serviceOperation) Apply(ctx context.Context, service Service) error {
	if o.err != nil {
		return o.err
	}
	svc := applyconfigurationscorev1.Service(service.Name, service.Namespace)
	svc.WithLabels(service.toLabel())
	t := any(len(service.ServiceType) < 1, corev1.ServiceTypeClusterIP, corev1.ServiceType(service.ServiceType))
	svc.Spec.Type = &t
	svc.Spec.Selector = service.toSelector()
	svc.Spec.Ports = make([]applyconfigurationscorev1.ServicePortApplyConfiguration, 0, len(service.Ports))
	for _, p := range service.Ports {
		protocol := corev1.Protocol(p.Protocol)
		port := p.Port
		targetPort := intstr.FromInt32(p.TargetPort)
		nodePort := p.NodePort
		svc.Spec.Ports = append(svc.Spec.Ports, applyconfigurationscorev1.ServicePortApplyConfiguration{
			Protocol:   &protocol,
			Port:       &port,       // 对外暴露的端口
			TargetPort: &targetPort, // Pod 内部服务监听的端口
			NodePort:   &nodePort,
		})
	}
	opts := v1.ApplyOptions{}
	_, err := o.api.CoreV1().Services(service.Namespace).Apply(ctx, svc, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to apply service: %v", err)
	}
	return nil
}

func (o *serviceOperation) Delete(ctx context.Context, namespace, service string) error {
	if o.err != nil {
		return o.err
	}
	opts := v1.DeleteOptions{}
	return o.api.CoreV1().Services(namespace).Delete(ctx, service, opts)
}
