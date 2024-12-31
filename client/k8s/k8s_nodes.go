package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/health"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node struct {
	Name    string        `json:"name,omitempty"`
	Address string        `json:"address,omitempty"`
	Role    string        `json:"role,omitempty"`
	Status  health.Health `json:"status,omitempty"`
	Cpu     Resource      `json:"cpu,omitempty"`
	Memory  Resource      `json:"memory,omitempty"`
}

type Resource struct {
	Value      string `json:"value"`
	Percentage string `json:"percentage"`
}

func (o Node) cpu(value, percentage string) {
	o.Cpu = Resource{
		Value:      value,
		Percentage: percentage,
	}
}

func (o Node) memory(value, percentage string) {
	o.Memory = Resource{
		Value:      value,
		Percentage: percentage,
	}
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
			Status: Status(string(n.Status.Phase)),
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
