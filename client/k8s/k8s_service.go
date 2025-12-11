package k8s

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/process"
	"github.com/hosgf/element/types"
	"github.com/hosgf/element/uerrors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type serviceOperation struct {
	*options
}

type Service struct {
	Model
	ServiceType string          `json:"serviceType,omitempty"`
	Status      string          `json:"status,omitempty"`
	Ports       []*process.Port `json:"ports,omitempty"`
}

func toServicePort(p *process.Port) corev1.ServicePort {
	port := corev1.ServicePort{
		Name:       p.GetName(),
		Protocol:   corev1.Protocol(p.GetProtocol().String()),
		Port:       p.Port,                         // 对外暴露的端口
		TargetPort: intstr.FromInt32(p.TargetPort), // Pod 内部服务监听的端口
	}
	if p.NodePort > 0 {
		port.NodePort = p.NodePort
	}
	return port
}

func (s *Service) updateCoreService(svc *corev1.Service) *corev1.Service {
	ports := make([]corev1.ServicePort, 0, len(s.Ports))
	for _, p := range s.Ports {
		ports = append(ports, toServicePort(p))
	}
	svc.Spec.Ports = ports
	svc.Spec.Selector = s.toSelector()
	svc.Spec.Type = corev1.ServiceType(toServiceType(s.ServiceType)) // 默认为 ClusterIP 类型
	for k, v := range s.labels() {
		svc.ObjectMeta.Labels[k] = v
	}
	return svc
}

func (s *Service) toCoreService() *corev1.Service {
	ports := make([]corev1.ServicePort, 0, len(s.Ports))
	for _, p := range s.Ports {
		ports = append(ports, toServicePort(p))
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

func (s *Service) toProcessPort() []process.ProcessPort {
	ports := make([]process.ProcessPort, 0)
	if len(s.Ports) == 0 {
		return ports
	}
	for _, port := range s.Ports {
		p := process.ProcessPort{
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
	ports := make([]*process.Port, 0)
	for _, p := range svc.Spec.Ports {
		port := &process.Port{
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
			svc = &Service{Model: model, ServiceType: p.ServiceType, Ports: make([]*process.Port, 0)}
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
	has, _, err := o.exists(ctx, namespace, service)
	return has, err
}

func (o *serviceOperation) Apply(ctx context.Context, service *Service) error {
	if o.err != nil {
		return o.err
	}
	if o.isTest {
		return nil
	}
	if has, svc, err := o.exists(ctx, service.Namespace, service.Name); has {
		if err != nil {
			return err
		}
		if service.AllowUpdate {
			return o.update(ctx, svc, service)
		}
		return uerrors.NewBizLogicError(uerrors.CodeResourceConflict,
			fmt.Sprintf("Service已存在: namespace=%s, service=%s", service.Namespace, service.Name))
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

	var lastErr error
	for _, group := range groups {
		if len(group) < 1 {
			continue
		}

		svcs, err := o.list(ctx, namespace, group)
		if err != nil {
			lastErr = err
			continue
		}

		if svcs != nil && len(svcs.Items) > 0 {
			for _, svc := range svcs.Items {
				if err := o.delete(ctx, namespace, svc.Name); err != nil {
					lastErr = err
				}
			}
		}
	}
	return lastErr
}

func (o *serviceOperation) create(ctx context.Context, service *Service) error {
	opts := v1.CreateOptions{}
	_, err := o.api.CoreV1().Services(service.Namespace).Create(ctx, service.toCoreService(), opts)
	if err != nil {
		return uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("创建Service: namespace=%s, service=%s", service.Namespace, service.Name))
	}
	return nil
}

func (o *serviceOperation) update(ctx context.Context, svc *corev1.Service, service *Service) error {
	if svc == nil {
		return nil
	}
	opts := v1.UpdateOptions{}
	_, err := o.api.CoreV1().Services(service.Namespace).Update(ctx, service.updateCoreService(svc), opts)
	if err != nil {
		return uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("更新Service: namespace=%s, service=%s", service.Namespace, service.Name))
	}
	return nil
}

func (o *serviceOperation) delete(ctx context.Context, namespace string, service string) error {
	opts := v1.DeleteOptions{}
	err := o.api.CoreV1().Services(namespace).Delete(ctx, service, opts)
	if err != nil {
		return uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("删除Service: namespace=%s, service=%s", namespace, service))
	}
	return nil
}

func (o *serviceOperation) exists(ctx context.Context, namespace, service string) (bool, *corev1.Service, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	opts := v1.GetOptions{}
	svc, err := o.api.CoreV1().Services(namespace).Get(ctx, service, opts)
	has, err := o.isExist(ctx, svc, err, fmt.Sprintf("检查Service是否存在: namespace=%s, service=%s", namespace, service))
	return has, svc, err
}

func (o *serviceOperation) list(ctx context.Context, namespace string, group string) (*corev1.ServiceList, error) {
	opts := toGroupListOptions(group)
	datas, err := o.api.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return datas, uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("获取Service列表: namespace=%s, group=%s", namespace, group))
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
