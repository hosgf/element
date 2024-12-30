package k8s

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"k8s.io/apimachinery/pkg/api/errors"
)

type operation interface {
	basics
	resource
	namespace
	service
	pods
	job
	storage
}

type option struct {
	isDebug bool
	err     error
}

// basics
type basics interface {
	Init(homePath string) error
	Version() (string, error)
}

// 资源
type resource interface {
}

// namespace
type namespace interface {
	GetNamespaces(ctx context.Context) ([]string, error)
	NamespaceIsExist(ctx context.Context, namespace string) (bool, error)
	CreateNamespace(ctx context.Context, namespace string) ([]string, error)
	DeleteNamespace(ctx context.Context, namespace string) error
}

// 服务
type service interface {
	GetServices(ctx context.Context, namespace string) ([]*Service, error)
	ServiceIsExist(ctx context.Context, namespace, service string) (bool, error)
	CreateService(ctx context.Context, service Service) error
	ApplyService(ctx context.Context, service Service) error
	DeleteService(ctx context.Context, namespace, service string) error
}

// pods
type pods interface {
	GetPod(ctx context.Context, namespace, appname string) ([]*Pod, error)
	GetPods(ctx context.Context, namespace string) ([]*Pod, error)
	PodIsExist(ctx context.Context, namespace, pod string) (bool, error)
	CreatePod(ctx context.Context, pod Pod) error
	ApplyPod(ctx context.Context, pod Pod) error
	DeletePod(ctx context.Context, namespace, pod string) error
	RestartPod(ctx context.Context, namespace, pod string) error
	RestartAppPods(ctx context.Context, namespace, appname string) error
}

// job
type job interface {
}

// 存储
type storage interface {
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
