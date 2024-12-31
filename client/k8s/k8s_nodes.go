package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/resource"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node struct {
	Name    string           `json:"name,omitempty"`
	Address string           `json:"address,omitempty"`
	Role    string           `json:"role,omitempty"`
	Status  health.Health    `json:"status,omitempty"`
	Cpu     resource.Details `json:"cpu,omitempty"`
	Memory  resource.Details `json:"memory,omitempty"`
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
			Name:   n.Name,
			Cpu:    resource.Details{},
			Memory: resource.Details{},
			Status: Status(string(n.Status.Phase)),
		}
		for _, address := range n.Status.Addresses {
			switch address.Type {
			case corev1.NodeInternalIP:
				node.Address = address.Address
			}
		}
		if _, exists := n.Labels["node-role.kubernetes.io/master"]; exists {
			node.Role = "master"
		}
		if _, exists := n.Labels["node-role.kubernetes.io/worker"]; exists {
			node.Role = "worker"
		}
		for resourceName, quantity := range n.Status.Allocatable {
			value, unit := types.Parse(quantity.String())
			switch resourceName {
			case corev1.ResourceCPU:
				node.Cpu.Free = types.FormatCpu(value, unit)
			case corev1.ResourceMemory:
				node.Memory.Free = types.FormatMemory(value, unit)
			}
		}
		for resourceName, quantity := range n.Status.Capacity {
			value, unit := types.Parse(quantity.String())
			switch resourceName {
			case corev1.ResourceCPU:
				node.Cpu.Total = types.FormatCpu(value, unit)
			case corev1.ResourceMemory:
				node.Memory.Total = types.FormatMemory(value, unit)
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
