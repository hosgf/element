package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type Service struct {
	progress.Service
	ServiceType string          `json:"serviceType,omitempty"`
	Ports       []progress.Port `json:"ports,omitempty"`
}

func (k *kubernetes) GetServices(ctx context.Context, namespace string) ([]*Service, error) {
	if k.err != nil {
		return nil, k.err
	}
	opts := v1.ListOptions{}
	svcs, err := k.api.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get Services: %v", err)
	}
	services := make([]*Service, 0, len(svcs.Items))
	for _, svc := range svcs.Items {
		services = append(services, &Service{
			Service: progress.Service{
				Namespace: namespace,
				Name:      svc.Name,
				App:       svc.Labels[types.LabelApp.String()],
				Owners:    svc.Labels[types.LabelOwners.String()],
				GroupType: svc.Labels[types.LabelType.String()],
				Status:    Status(svc.Status.String()),
			}})
	}
	return services, nil
}

func (k *kubernetes) ServiceIsExist(ctx context.Context, namespace, service string) (bool, error) {
	if k.err != nil {
		return false, k.err
	}
	opts := v1.GetOptions{}
	svc, err := k.api.CoreV1().Services(namespace).Get(ctx, service, opts)
	return k.isExist(svc, err, "Failed to get Services: %v")
}

func (k *kubernetes) CreateService(ctx context.Context, service Service) error {
	if k.err != nil {
		return k.err
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
			Labels: map[string]string{
				types.LabelApp.String():    service.App,
				types.LabelOwners.String(): service.Owners,
				types.LabelType.String():   service.GroupType,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": service.App,
			},
			Ports: ports,
			Type:  any(len(service.ServiceType) < 1, corev1.ServiceTypeClusterIP, corev1.ServiceType(service.ServiceType)), // 默认为 ClusterIP 类型
		},
	}

	opts := v1.CreateOptions{}
	_, err := k.api.CoreV1().Services(service.Namespace).Create(ctx, svc, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create service: %v", err)
	}
	return nil
}

func (k *kubernetes) ApplyService(ctx context.Context, service Service) error {
	if k.err != nil {
		return k.err
	}
	svc := applyconfigurationscorev1.Service(service.Name, service.Namespace)
	svc.WithLabels(map[string]string{
		types.LabelApp.String():    service.App,
		types.LabelOwners.String(): service.Owners,
		types.LabelType.String():   service.GroupType,
	})
	t := any(len(service.ServiceType) < 1, corev1.ServiceTypeClusterIP, corev1.ServiceType(service.ServiceType))
	svc.Spec.Type = &t
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
	_, err := k.api.CoreV1().Services(service.Namespace).Apply(ctx, svc, opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to apply service: %v", err)
	}
	return nil
}

func (k *kubernetes) DeleteService(ctx context.Context, namespace, service string) error {
	if k.err != nil {
		return k.err
	}
	opts := v1.DeleteOptions{}
	return k.api.CoreV1().Services(namespace).Delete(ctx, service, opts)
}