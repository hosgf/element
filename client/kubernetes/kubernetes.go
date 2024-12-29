package kubernetes

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/model/progress"
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
	GetServices(ctx context.Context, namespace string) ([]*progress.Service, error)
	ServiceIsExist(ctx context.Context, namespace, service string) (bool, error)
	DeleteService(ctx context.Context, namespace, service string) error
}

// pods
type pods interface {
	GetPod(ctx context.Context, namespace, appname string) ([]string, error)
	GetPods(ctx context.Context, namespace string) ([]string, error)
	GetPodsStatus(ctx context.Context, namespace string) string
	PodIsExist(ctx context.Context, namespace, pod string) string
	RestartPod(ctx context.Context, namespace, pod string) string
	RestartAppPods(ctx context.Context, namespace, label string) string
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
