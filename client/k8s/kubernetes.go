package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/types"
	"k8s.io/apimachinery/pkg/api/errors"
)

type operation interface {
	Init(homePath string) error
	Version() (string, error)
	NamespaceOperation
	ServiceOperation
	PodsOperation
	JobsOperation
	StorageOperation
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

type option struct {
	isDebug bool
	err     error
}

// NamespaceOperation namespace
type NamespaceOperation interface {
	Namespace() namespaceInterface
}

type namespaceInterface interface {
	List(ctx context.Context) ([]string, error)
	IsExist(ctx context.Context, namespace string) (bool, error)
	Create(ctx context.Context, namespace string) ([]string, error)
	Delete(ctx context.Context, namespace string) error
}

// ServiceOperation 服务
type ServiceOperation interface {
	Service() serviceInterface
}

type serviceInterface interface {
	List(ctx context.Context, namespace string) ([]*Service, error)
	IsExist(ctx context.Context, namespace, service string) (bool, error)
	Create(ctx context.Context, service Service) error
	Apply(ctx context.Context, service Service) error
	Delete(ctx context.Context, namespace, service string) error
}

// PodsOperation pods
type PodsOperation interface {
	Pod() podsInterface
}

type podsInterface interface {
	Get(ctx context.Context, namespace, appname string) ([]*Pod, error)
	List(ctx context.Context, namespace string) ([]*Pod, error)
	IsExist(ctx context.Context, namespace, pod string) (bool, error)
	Create(ctx context.Context, pod Pod) error
	Apply(ctx context.Context, pod Pod) error
	Delete(ctx context.Context, namespace, pod string) error
	Restart(ctx context.Context, namespace, pod string) error
	RestartApp(ctx context.Context, namespace, appname string) error
}

// JobsOperation job
type JobsOperation interface {
	Job() jobsInterface
}

type jobsInterface interface {
}

// StorageOperation 存储
type StorageOperation interface {
	Storage() storageInterface
}

type storageInterface interface {
}

func (k *option) isExist(value interface{}, err error, format string) (bool, error) {
	if err == nil {
		return value != nil, nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}
	return false, gerror.NewCodef(gcode.CodeNotImplemented, format, err)
}
