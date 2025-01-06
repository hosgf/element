package k8s

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/resource"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node struct {
	Name       string                                       `json:"name,omitempty"`
	Address    string                                       `json:"address,omitempty"`
	Roles      string                                       `json:"roles,omitempty"`
	Status     health.Health                                `json:"status,omitempty"`
	Cpu        resource.Details                             `json:"cpu,omitempty"`
	Memory     resource.Details                             `json:"memory,omitempty"`
	Time       int64                                        `json:"time"`
	Indicators map[health.Indicator]health.IndicatorDetails `json:"indicators,omitempty"` // 指标
}

func (n *Node) ToNode() resource.Node {
	node := resource.Node{
		Name:   n.Name,
		Status: n.Status,
		Roles:  n.Roles,
		Time:   n.Time,
		Indicators: map[string]interface{}{
			"address": n.Address,
		},
	}
	for k, v := range n.Indicators {
		node.Indicators[k.String()] = v
	}
	return node
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
		if role, ok := n.Labels["kubernetes.io/role"]; ok {
			node.Roles = role
		} else {
			for k, _ := range n.Labels {
				if gstr.HasPrefix(k, "node-role.kubernetes.io/") {
					node.Roles = strings.TrimPrefix(k, "node-role.kubernetes.io/")
					break
				}
			}
		}

		// 资源总量
		for name, quantity := range n.Status.Capacity {
			switch name {
			case corev1.ResourceCPU:
				node.Cpu.SetTotal(quantity.String())
			case corev1.ResourceMemory:
				node.Memory.SetTotal(quantity.String())
			}
		}
		// 空闲资源
		for name, quantity := range n.Status.Allocatable {
			switch name {
			case corev1.ResourceCPU:
				node.Cpu.SetFree(quantity.String())
				node.Cpu.Usage = node.Cpu.Total - node.Cpu.Free
			case corev1.ResourceMemory:
				node.Memory.SetFree(quantity.String())
				node.Memory.Usage = node.Memory.Total - node.Memory.Free
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
				node.Time = condition.LastTransitionTime.Unix()
				node.Status = NodeStatus(status)
				node.Indicators[health.IndicatorNodeStatus] = details
			case corev1.NodeMemoryPressure:
				node.Indicators[health.IndicatorMemoryStatus] = details
			case corev1.NodeDiskPressure:
				node.Indicators[health.IndicatorDiskStatus] = details
			case corev1.NodeNetworkUnavailable:
				node.Indicators[health.IndicatorNetworkStatus] = details
			case corev1.NodePIDPressure:
				node.Indicators[health.IndicatorNodePIDPressure] = details
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
