package k8s

import (
	"context"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/model/resource"
	"github.com/hosgf/element/types"
)

type operation interface {
	Init(homePath string) error
	Version() (string, error)
	Nodes() nodesInterface
	Namespace() namespaceInterface
	Service() serviceInterface
	Pod() podsInterface
	Job() jobsInterface
	Storage() storageInterface
}

type Model struct {
	Namespace string            `json:"namespace,omitempty"`
	App       string            `json:"app,omitempty"`
	Group     string            `json:"group,omitempty"`
	Owner     string            `json:"owner,omitempty"`
	Scope     string            `json:"scope,omitempty"`
	Name      string            `json:"name,omitempty"`
	Status    health.Health     `json:"status,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

func (m *Model) toLabel() map[string]string {
	labels := map[string]string{
		types.LabelApp.String():   m.App,
		types.LabelOwner.String(): m.Owner,
		types.LabelScope.String(): m.Scope,
		types.LabelGroup.String(): m.Group,
	}
	if m.Labels != nil {
		for k, v := range m.Labels {
			labels[k] = v
		}
	}
	return labels
}

func (m *Model) labels(labels map[string]string) {
	if len(labels) < 1 {
		return
	}
	m.App = labels[types.LabelApp.String()]
	m.Owner = labels[types.LabelOwner.String()]
	m.Scope = labels[types.LabelScope.String()]
	m.Group = labels[types.LabelGroup.String()]
	delete(labels, types.LabelApp.String())
	delete(labels, types.LabelOwner.String())
	delete(labels, types.LabelScope.String())
	delete(labels, types.LabelGroup.String())
	if m.Labels == nil {
		m.Labels = map[string]string{}
	}
	for k, v := range labels {
		m.Labels[k] = v
	}
}

type nodesInterface interface {
	Top(ctx context.Context) ([]Node, error)
}

type namespaceInterface interface {
	List(ctx context.Context) ([]resource.Namespace, error)
	Exists(ctx context.Context, namespace string) (bool, error)
	Create(ctx context.Context, namespace string) (bool, error)
	Delete(ctx context.Context, namespace string) error
}

type serviceInterface interface {
	List(ctx context.Context, namespace string) ([]Service, error)
	Exists(ctx context.Context, namespace, service string) (bool, error)
	Create(ctx context.Context, service Service) error
	Apply(ctx context.Context, service Service) error
	Delete(ctx context.Context, namespace, service string) error
}

type podsInterface interface {
	Get(ctx context.Context, namespace, appname string) ([]Pod, error)
	List(ctx context.Context, namespace string) ([]Pod, error)
	Exists(ctx context.Context, namespace, pod string) (bool, error)
	Create(ctx context.Context, pod Pod) error
	Apply(ctx context.Context, pod Pod) error
	Delete(ctx context.Context, namespace, pod string) error
	Restart(ctx context.Context, namespace, pod string) error
	RestartApp(ctx context.Context, namespace, appname string) error
}

type jobsInterface interface {
}

type storageInterface interface {
}
