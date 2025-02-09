package k8s

import (
	"context"

	"github.com/hosgf/element/types"

	"github.com/hosgf/element/model/progress"
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
	StorageResource() storageResourceInterface
	Metrics() metricsInterface
	PodTemplate() podTemplatesInterface
	Progress() progressInterface
	Resource() resourceInterface
}

type progressInterface interface {
	List(ctx context.Context, namespace string) ([]*progress.Progress, error)
	Running(ctx context.Context, config *ProcessGroupConfig) error
	Start(ctx context.Context, config *ProcessGroupConfig) error
	Stop(ctx context.Context, namespace string, groups ...string) error
	Destroy(ctx context.Context, namespace string, groups ...string) error
}

type resourceInterface interface {
	Get(ctx context.Context) (*resource.Resource, error)
}

type nodesInterface interface {
	Top(ctx context.Context) ([]*Node, error)
}

type metricsInterface interface {
	List(ctx context.Context, namespace string) ([]*Metric, error)
}

type namespaceInterface interface {
	List(ctx context.Context) ([]*types.Namespace, error)
	Exists(ctx context.Context, namespace string) (bool, error)
	Apply(ctx context.Context, namespace, label string) (bool, error)
	Delete(ctx context.Context, namespace string) error
}

type serviceInterface interface {
	List(ctx context.Context, namespace string, groups ...string) ([]*Service, error)
	Exists(ctx context.Context, namespace, service string) (bool, error)
	Apply(ctx context.Context, service *Service) error
	Delete(ctx context.Context, namespace, service string) error
	DeleteGroup(ctx context.Context, namespace string, groups ...string) error
}

type podsInterface interface {
	Get(ctx context.Context, namespace, appname string) ([]*Pod, error)
	List(ctx context.Context, namespace string, groups ...string) ([]*Pod, error)
	Exists(ctx context.Context, namespace, pod string) (bool, error)
	Apply(ctx context.Context, pod *Pod) error
	Delete(ctx context.Context, namespace, pod string) error
	DeleteGroup(ctx context.Context, namespace string, groups ...string) error
	Restart(ctx context.Context, namespace, pod string) error
	RestartApp(ctx context.Context, namespace, appname string) error
}

type podTemplatesInterface interface {
	Get(ctx context.Context, namespace, appname string) ([]*Pod, error)
	List(ctx context.Context, namespace string) ([]*Pod, error)
	Exists(ctx context.Context, namespace, pod string) (bool, error)
	Apply(ctx context.Context, pod *Pod) error
	Delete(ctx context.Context, namespace, pod string) error
}

type jobsInterface interface {
}

type storageInterface interface {
	Get(ctx context.Context, namespace, name string) (*Storage, error)
	Exists(ctx context.Context, namespace, name string) (bool, error)
	Apply(ctx context.Context, storage *PersistentStorage) error
	BatchApply(ctx context.Context, model Model, storage []Storage) error
	Delete(ctx context.Context, delRes bool, namespace string, name ...string) error
	DeleteByGroup(ctx context.Context, delRes bool, namespace string, groups ...string) error
}

type storageResourceInterface interface {
	Get(ctx context.Context, name string) (*Storage, error)
	Exists(ctx context.Context, name string) (bool, error)
	Apply(ctx context.Context, storage *PersistentStorageResource) error
	BatchApply(ctx context.Context, model Model, storage []Storage) error
	Delete(ctx context.Context, name string) error
	DeleteByGroup(ctx context.Context, groups ...string) error
}
