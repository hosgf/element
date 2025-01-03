package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/resource"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type Node struct {
	Name       string                                       `json:"name,omitempty"`
	Address    string                                       `json:"address,omitempty"`
	Roles      string                                       `json:"roles,omitempty"`
	Status     health.Health                                `json:"status,omitempty"`
	Cpu        resource.Details                             `json:"cpu,omitempty"`
	Memory     resource.Details                             `json:"memory,omitempty"`
	Indicators map[health.Indicator]health.IndicatorDetails `json:"indicators,omitempty"` // 指标
}

type nodesOperation struct {
	*options
}

func (o *nodesOperation) Top(ctx context.Context) ([]Node, error) {
	if o.err != nil {
		return nil, o.err
	}
	datas, err := o.api.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get nodes: %v", err)
	}
	nodes := make([]Node, 0, len(datas.Items))
	for _, n := range datas.Items {
		node := Node{
			Name:       n.Name,
			Cpu:        resource.Details{},
			Memory:     resource.Details{},
			Indicators: map[health.Indicator]health.IndicatorDetails{},
		}
		for _, address := range n.Status.Addresses {
			switch address.Type {
			case corev1.NodeInternalIP:
				node.Address = address.Address
			}
		}
		for k, _ := range n.Labels {
			if gstr.HasPrefix(k, "node-role.kubernetes.io/") {
				node.Roles = strings.TrimPrefix(k, "node-role.kubernetes.io/")
			}
		}
		// 资源总量
		for name, quantity := range n.Status.Capacity {
			value, unit := types.Parse(quantity.String())
			switch name {
			case corev1.ResourceCPU:
				node.Cpu.Total = types.FormatCpu(value, unit)
			case corev1.ResourceMemory:
				node.Memory.Total = types.FormatMemory(value, unit)
			}
		}
		// 空闲资源
		for name, quantity := range n.Status.Allocatable {
			value, unit := types.Parse(quantity.String())
			switch name {
			case corev1.ResourceCPU:
				node.Cpu.Free = types.FormatCpu(value, unit)
			case corev1.ResourceMemory:
				node.Memory.Free = types.FormatMemory(value, unit)
			}
		}
		// 状态
		for _, condition := range n.Status.Conditions {
			status := string(condition.Status)
			details := health.IndicatorDetails{
				Status:  status,
				Reason:  condition.Reason,
				Message: condition.Message,
			}
			switch condition.Type {
			case corev1.NodeReady:
				node.Status = NodeStatus(status)
				node.Indicators[health.IndicatorNodeStatus] = details
			case corev1.NodeMemoryPressure:
				node.Indicators[health.IndicatorMemoryStatus] = details
			case corev1.NodeDiskPressure:
				node.Indicators[health.IndicatorDiskStatus] = details
			case corev1.NodeNetworkUnavailable:
				node.Indicators[health.IndicatorNetworkStatus] = details
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
