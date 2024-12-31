package k8s

import (
	"context"
	"github.com/hosgf/element/health"
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
	Namespace string        `json:"namespace,omitempty"`
	App       string        `json:"app,omitempty"`
	Owner     string        `json:"owner,omitempty"`
	Scope     string        `json:"scope,omitempty"`
	Name      string        `json:"name,omitempty"`
	Status    health.Health `json:"status,omitempty"`
}

func (m Model) toLabel() map[string]string {
	return map[string]string{
		types.LabelApp.String():   m.App,
		types.LabelOwner.String(): m.Owner,
		types.LabelScope.String(): m.Scope,
	}
}

func (m Model) labels(labels map[string]string) {
	if len(labels) < 1 {
		return
	}
	m.App = labels[types.LabelApp.String()]
	m.Owner = labels[types.LabelOwner.String()]
	m.Scope = labels[types.LabelScope.String()]
}

type nodesInterface interface {
	Top(ctx context.Context) ([]Node, error)
}

type namespaceInterface interface {
	List(ctx context.Context) ([]string, error)
	IsExist(ctx context.Context, namespace string) (bool, error)
	Create(ctx context.Context, namespace string) (bool, error)
	Delete(ctx context.Context, namespace string) error
}

type serviceInterface interface {
	List(ctx context.Context, namespace string) ([]Service, error)
	IsExist(ctx context.Context, namespace, service string) (bool, error)
	Create(ctx context.Context, service Service) error
	Apply(ctx context.Context, service Service) error
	Delete(ctx context.Context, namespace, service string) error
}

type podsInterface interface {
	Get(ctx context.Context, namespace, appname string) ([]Pod, error)
	List(ctx context.Context, namespace string) ([]Pod, error)
	IsExist(ctx context.Context, namespace, pod string) (bool, error)
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
