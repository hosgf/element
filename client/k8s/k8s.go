package k8s

import (
	"path/filepath"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/types"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func New(isDebug bool) *Kubernetes {
	k := &Kubernetes{}
	k.options = &options{isDebug: isDebug}
	k.nodes = &nodesOperation{k.options}
	k.namespace = &namespaceOperation{k.options}
	k.service = &serviceOperation{k.options}
	k.pods = &podsOperation{k.options}
	k.metrics = &metricsOperation{k.options}
	k.progress = &progressOperation{k8s: k, options: k.options}
	k.resource = &resourceOperation{k8s: k, options: k.options}
	return k
}

type options struct {
	isDebug    bool
	err        error
	api        *k8s.Clientset
	metricsApi *metricsv.Clientset
}

type Kubernetes struct {
	*options
	nodes     *nodesOperation
	namespace *namespaceOperation
	service   *serviceOperation
	pods      *podsOperation
	metrics   *metricsOperation
	progress  *progressOperation
	resource  *resourceOperation
}

func (k *Kubernetes) Nodes() *nodesOperation {
	return k.nodes
}

func (k *Kubernetes) Namespace() *namespaceOperation {
	return k.namespace
}

func (k *Kubernetes) Metrics() *metricsOperation {
	return k.metrics
}

func (k *Kubernetes) Service() *serviceOperation {
	return k.service
}

func (k *Kubernetes) Pod() *podsOperation {
	return k.pods
}

func (k *Kubernetes) Progress() *progressOperation {
	return k.progress
}

func (k *Kubernetes) Resource() *resourceOperation {
	return k.resource
}

func (k *Kubernetes) Init(homePath string) error {
	kubeconfig := k.config(homePath)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		k.err = gerror.NewCodef(gcode.CodeNotImplemented, "Failed to build kubeconfig: %v", err)
		return k.err
	}
	k.api, err = k8s.NewForConfig(config)
	if err != nil {
		k.err = gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create Kubernetes client: %v", err)
		return k.err
	}
	k.metricsApi, err = metricsv.NewForConfig(config)
	if err != nil {
		k.err = gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create Kubernetes client: %v", err)
		return k.err
	}
	return nil
}

func (k *Kubernetes) Version() (string, error) {
	if k.err != nil {
		return "", k.err
	}
	version, err := k.api.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func (k *Kubernetes) config(homePath string) string {
	if homePath != "" {
		return filepath.Join(homePath, ".kube", "config")
	}
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

func toServiceType(serviceType string) string {
	if len(serviceType) < 1 {
		return types.DefaultServiceType
	}
	return serviceType
}

func (o *options) isExist(value interface{}, err error, format string) (bool, error) {
	if err == nil {
		return value != nil, nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}
	return false, gerror.NewCodef(gcode.CodeNotImplemented, format, err)
}
