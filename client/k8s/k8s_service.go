package k8s

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
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
	Model
	ServiceType string           `json:"serviceType,omitempty"`
	Status      string           `json:"status,omitempty"`
	Ports       []*progress.Port `json:"ports,omitempty"`
}

func (pg *ProcessGroupConfig) toServices() []*Service {
	if pg.Process == nil || len(pg.Process) < 1 {
		return nil
	}
	labels := &pg.Labels
	if len(labels.Group) < 1 {
		labels.Group = pg.GroupName
	}
	svcs := make([]*Service, 0)
	for _, p := range pg.Process {
		ports := p.Ports
		if len(p.Service) < 1 || ports == nil || len(ports) < 1 {
			continue
		}
		svc := &Service{
			Model:       Model{Namespace: pg.Namespace, Name: p.Service},
			ServiceType: p.ServiceType,
			Ports:       make([]*progress.Port, 0),
		}
		for _, p := range ports {
			svc.Ports = append(svc.Ports, &p)
		}

		svc.setTypesLabels(labels)
		svcs = append(svcs, svc)
	}
	return svcs
}

func (s *Service) toProgressPort() []progress.ProgressPort {
	ports := make([]progress.ProgressPort, 0)
	if len(s.Ports) == 0 {
		return ports
	}
	for _, port := range s.Ports {
		p := progress.ProgressPort{
			Name:     port.Name,
			Protocol: port.Protocol,
			Port:     port.Port,
		}
		ports = append(ports, p)
	}
	return ports
}

func (s *Service) setPorts(svc corev1.Service) {
	if len(svc.Spec.Ports) == 0 {
		return
	}
	ports := make([]*progress.Port, 0)
	for _, p := range svc.Spec.Ports {
		port := &progress.Port{
			Name:       p.Name,
			Protocol:   types.ProtocolType(p.Protocol),
			Port:       p.Port,
			TargetPort: p.TargetPort.IntVal,
			NodePort:   p.NodePort,
		}
		ports = append(ports, port)
	}
	s.Ports = ports
}

func (s *Service) setGroup(svc corev1.Service) {
	selector := svc.Spec.Selector
	if selector == nil {
		return
	}
	if v, ok := selector[types.LabelGroup.String()]; ok {
		s.Group = v
		s.groupLabel = types.LabelGroup.String()
		return
	}
	if v, ok := selector["app"]; ok {
		s.Group = v
		s.groupLabel = "app"
		return
	}
	for k, v := range selector {
		if gstr.Contains(k, "app") {
			s.Group = v
			s.groupLabel = k
			return
		}
	}
}

func (o *serviceOperation) List(ctx context.Context, namespace string) ([]*Service, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{}
	svcs, err := o.api.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get Services: %v", err)
	}
	services := make([]*Service, 0, len(svcs.Items))
	for _, svc := range svcs.Items {
		service := &Service{
			Model:       Model{Namespace: namespace, Name: svc.Name},
			Status:      health.UP.String(),
			ServiceType: string(svc.Spec.Type),
		}
		service.setPorts(svc)
		service.setGroup(svc)
		service.setLabels(svc.Labels)
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

func (o *serviceOperation) Create(ctx context.Context, service *Service) error {
	if o.err != nil {
		return o.err
	}
	ports := make([]corev1.ServicePort, 0, len(service.Ports))
	for _, p := range service.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
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
			Labels:    service.labels(),
		},
		Spec: corev1.ServiceSpec{
			Selector: service.toSelector(),
			Ports:    ports,
			Type:     corev1.ServiceType(toServiceType(service.ServiceType)), // 默认为 ClusterIP 类型
		},
	}
	opts := v1.CreateOptions{}
	_, err := o.api.CoreV1().Services(service.Namespace).Create(ctx, svc, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create service: %v", err)
	}
	return nil
}

func (o *serviceOperation) Apply(ctx context.Context, service *Service) error {
	if o.err != nil {
		return o.err
	}
	svc := applyconfigurationscorev1.Service(service.Name, service.Namespace)
	svc.WithLabels(service.labels())
	t := corev1.ServiceType(toServiceType(service.ServiceType))
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
