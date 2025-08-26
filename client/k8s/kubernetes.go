package k8s

import (
	"context"
	"io"
	"time"

	"github.com/hosgf/element/types"

	"github.com/hosgf/element/model/process"
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
	Process() processInterface
	Resource() resourceInterface
}

type processInterface interface {
	List(ctx context.Context, namespace string) ([]*process.Process, error)
	Running(ctx context.Context, config *ProcessGroupConfig) error
	Start(ctx context.Context, config *ProcessGroupConfig) error
	Stop(ctx context.Context, namespace string, groups ...string) error
	Destroy(ctx context.Context, namespace string, groups ...string) error
	Restart(ctx context.Context, namespace, group, process string, cmd ...string) error
	RestartGroup(ctx context.Context, namespace, group string) error
	RestartApp(ctx context.Context, namespace, appname string) error
	Command(ctx context.Context, namespace, group, process string, cmd ...string) (string, error)
	Logger(ctx context.Context, namespace, group, process string, config ProcessLogger) (io.ReadCloser, error)
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
	RestartGroup(ctx context.Context, namespace, group string) error
	RestartApp(ctx context.Context, namespace, appname string) error
	Command(ctx context.Context, namespace, group, process string, cmd ...string) (string, error)
	Exec(ctx context.Context, namespace, pod, process string, cmd ...string) (string, error)
	Logger(ctx context.Context, namespace, group, process string, config ProcessLogger) (io.ReadCloser, error)
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
	Get(ctx context.Context, namespace, name string) (*types.Storage, error)
	Exists(ctx context.Context, namespace, name string) (bool, error)
	Apply(ctx context.Context, storage *PersistentStorage) error
	BatchApply(ctx context.Context, model Model, storage []types.Storage) error
	Delete(ctx context.Context, delRes bool, namespace string, name ...string) error
	DeleteByGroup(ctx context.Context, delRes bool, namespace string, groups ...string) error
	WaitDeleted(ctx context.Context, namespace, name string, timeout time.Duration) error
	IsDeleting(ctx context.Context, namespace, name string) (bool, error)
}

type storageResourceInterface interface {
	Get(ctx context.Context, name string) (*types.Storage, error)
	Exists(ctx context.Context, name string) (bool, error)
	Apply(ctx context.Context, storage *PersistentStorageResource) error
	BatchApply(ctx context.Context, model Model, storage []types.Storage) error
	Delete(ctx context.Context, name string) error
	DeleteByGroup(ctx context.Context, groups ...string) error
	WaitDeleted(ctx context.Context, name string, timeout time.Duration) error
	IsDeleting(ctx context.Context, name string) (bool, error)
}
