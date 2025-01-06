package k8s

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/hosgf/element/model/resource"
)

type resourceOperation struct {
	*options
	k8s *Kubernetes
}

func (o *resourceOperation) Get(ctx context.Context) (*resource.Resource, error) {
	if o.err != nil {
		return nil, o.err
	}
	nodes, err := o.k8s.Nodes().Top(ctx)
	if err != nil {
		return nil, err
	}
	res := &resource.Resource{
		Env:   "k8s",
		Time:  gtime.Now().Timestamp(),
		Nodes: make([]resource.Node, 0),
	}
	for _, node := range nodes {
		res.Nodes = append(res.Nodes, node.ToNode())
	}
	res.SetStatus()
	return res, nil
}
