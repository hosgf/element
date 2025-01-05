package k8s

import (
	"context"

	"github.com/hosgf/element/model/resource"
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
	PodTemplate() podTemplatesInterface
}

type nodesInterface interface {
	Top(ctx context.Context) ([]Node, error)
}

type namespaceInterface interface {
	List(ctx context.Context) ([]resource.Namespace, error)
	Exists(ctx context.Context, namespace string) (bool, error)
	Create(ctx context.Context, namespace, label string) (bool, error)
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

type podTemplatesInterface interface {
	Get(ctx context.Context, namespace, appname string) ([]Pod, error)
	List(ctx context.Context, namespace string) ([]Pod, error)
	Exists(ctx context.Context, namespace, pod string) (bool, error)
	Create(ctx context.Context, pod Pod) error
	Apply(ctx context.Context, pod Pod) error
	Delete(ctx context.Context, namespace, pod string) error
}

type jobsInterface interface {
}

type storageInterface interface {
}
