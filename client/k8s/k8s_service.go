package k8s

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func (s *Service) toCoreService() *corev1.Service {
	ports := make([]corev1.ServicePort, 0, len(s.Ports))
	for _, p := range s.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       p.GetName(),
			Protocol:   corev1.Protocol(p.Protocol.String()),
			Port:       p.Port,                         // 对外暴露的端口
			TargetPort: intstr.FromInt32(p.TargetPort), // Pod 内部服务监听的端口
			NodePort:   p.NodePort,
		})
	}
	return &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      s.Name,      // Service 名称
			Namespace: s.Namespace, // Service 所在的 Namespace
			Labels:    s.labels(),
		},
		Spec: corev1.ServiceSpec{
			Selector: s.toSelector(),
			Ports:    ports,
			Type:     corev1.ServiceType(toServiceType(s.ServiceType)), // 默认为 ClusterIP 类型
		},
	}
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
	if v, ok := selector[types.DefaultGroupLabel]; ok {
		s.Group = v
		s.groupLabel = types.DefaultGroupLabel
		return
	}
	for k, v := range selector {
		if gstr.Contains(k, types.DefaultGroupLabel) {
			s.Group = v
			s.groupLabel = k
			return
		}
	}
}

func (pg *ProcessGroupConfig) toServices() []*Service {
	if pg.Process == nil || len(pg.Process) < 1 {
		return nil
	}
	labels := &pg.Labels
	if len(labels.Group) < 1 {
		labels.Group = pg.GroupName
	}
	svcmap := make(map[string]*Service)
	for _, p := range pg.Process {
		ports := p.Ports
		if len(p.Service) < 1 || ports == nil || len(ports) < 1 {
			continue
		}
		model := Model{Namespace: pg.Namespace, Name: p.Service, AllowUpdate: pg.AllowUpdate}
		key := model.Key()
		svc := svcmap[key]
		if svc == nil {
			svc = &Service{Model: model, ServiceType: p.ServiceType, Ports: make([]*progress.Port, 0)}
			svc.setTypesLabels(labels)
		}
		for _, p := range ports {
			svc.Ports = append(svc.Ports, &p)
		}
		svcmap[key] = svc
	}
	svcs := make([]*Service, 0)
	for _, v := range svcmap {
		svcs = append(svcs, v)
	}
	return svcs
}

func (o *serviceOperation) List(ctx context.Context, namespace string, groups ...string) ([]*Service, error) {
	if o.err != nil {
		return nil, o.err
	}
	if groups == nil || len(groups) == 0 {
		svcs, err := o.list(ctx, namespace, "")
		return o.toServices(namespace, svcs), err
	}
	services := make([]*Service, 0)
	for _, g := range groups {
		if len(g) < 1 {
			continue
		}
		svcs, err := o.list(ctx, namespace, g)
		if err != nil {
			return nil, err
		}
		services = append(services, o.toServices(namespace, svcs)...)
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

func (o *serviceOperation) Apply(ctx context.Context, service *Service) error {
	if o.err != nil {
		return o.err
	}
	if has, err := o.Exists(ctx, service.Namespace, service.Name); has {
		if err != nil {
			return err
		}
		if service.AllowUpdate {
			return o.update(ctx, service)
		}
		return gerror.NewCodef(gcode.CodeNotImplemented, "Namespace: %s, Service: %s 已存在!", service.Namespace, service.Name)
	}
	return o.create(ctx, service)
}

//func (o *serviceOperation) Apply(ctx context.Context, service *Service) error {
//	if o.err != nil {
//		return o.err
//	}
//	svc := applyconfigurationscorev1.Service(service.Name, service.Namespace)
//	svc.WithLabels(service.labels())
//	spec := applyconfigurationscorev1.ServiceSpec()
//	spec.WithType(corev1.ServiceType(toServiceType(service.ServiceType)))
//	spec.WithSelector(service.toSelector())
//	for _, p := range service.Ports {
//		protocol := corev1.Protocol(p.Protocol.String())
//		name := p.GetName()
//		port := p.Port
//		targetPort := intstr.FromInt32(p.TargetPort)
//		nodePort := p.NodePort
//		spec.WithPorts(&applyconfigurationscorev1.ServicePortApplyConfiguration{
//			Name:       &name,
//			Protocol:   &protocol,
//			Port:       &port,       // 对外暴露的端口
//			TargetPort: &targetPort, // Pod 内部服务监听的端口
//			NodePort:   &nodePort,
//		})
//	}
//	svc.WithSpec(spec)
//	opts := v1.ApplyOptions{}
//	_, err := o.api.CoreV1().Services(service.Namespace).Apply(ctx, svc, opts)
//	if err != nil {
//		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to apply service: %v", err)
//	}
//	return nil
//}

func (o *serviceOperation) Delete(ctx context.Context, namespace, service string) error {
	if o.err != nil {
		return o.err
	}
	if has, err := o.Exists(ctx, namespace, service); has {
		if err != nil {
			return err
		}
		return o.delete(ctx, namespace, service)
	}
	return nil
}

func (o *serviceOperation) DeleteGroup(ctx context.Context, namespace string, groups ...string) error {
	if o.err != nil {
		return o.err
	}
	for _, group := range groups {
		if len(group) < 1 {
			continue
		}
		svcs, err := o.list(ctx, namespace, group)
		if err != nil {
			return err
		}
		if svcs.Items == nil || len(svcs.Items) == 0 {
			continue
		}
		for _, svc := range svcs.Items {
			if err := o.delete(ctx, namespace, svc.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *serviceOperation) create(ctx context.Context, service *Service) error {
	opts := v1.CreateOptions{}
	_, err := o.api.CoreV1().Services(service.Namespace).Create(ctx, service.toCoreService(), opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create service: %v", err)
	}
	return nil
}

func (o *serviceOperation) update(ctx context.Context, service *Service) error {
	opts := v1.UpdateOptions{}
	_, err := o.api.CoreV1().Services(service.Namespace).Update(ctx, service.toCoreService(), opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to update service: %v", err)
	}
	return nil
}

func (o *serviceOperation) delete(ctx context.Context, namespace string, service string) error {
	opts := v1.DeleteOptions{}
	err := o.api.CoreV1().Services(namespace).Delete(ctx, service, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to delete service: %v", err)
	}
	return nil
}

func (o *serviceOperation) list(ctx context.Context, namespace string, group string) (*corev1.ServiceList, error) {
	opts := v1.ListOptions{}
	if len(group) > 0 {
		opts.LabelSelector = fmt.Sprintf("%s=%s", types.LabelGroup, group)
	}
	datas, err := o.api.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return datas, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get Services: %v", err)
	}
	return datas, nil
}

func (o *serviceOperation) toServices(namespace string, svcs *corev1.ServiceList) []*Service {
	if svcs == nil {
		return nil
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
	return services
}
