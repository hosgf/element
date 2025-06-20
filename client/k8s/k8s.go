package k8s

import (
	"fmt"
	"path/filepath"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/types"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func New(isDebug, isTest bool) *Kubernetes {
	k := &Kubernetes{}
	k.options = &options{isDebug: isDebug, isTest: isTest}
	k.nodes = &nodesOperation{k.options}
	k.namespace = &namespaceOperation{k.options}
	k.service = &serviceOperation{k.options}
	k.pods = &podsOperation{k.options}
	k.storage = &storageOperation{k8s: k, options: k.options}
	k.storageResource = &storageResourceOperation{k.options}
	k.metrics = &metricsOperation{k.options}
	k.process = &processOperation{k8s: k, options: k.options}
	k.resource = &resourceOperation{k8s: k, options: k.options}
	return k
}

type options struct {
	isDebug    bool
	isTest     bool
	err        error
	c          *rest.Config
	api        *k8s.Clientset
	metricsApi *metricsv.Clientset
}

type Kubernetes struct {
	*options
	nodes           *nodesOperation
	namespace       *namespaceOperation
	service         *serviceOperation
	pods            *podsOperation
	storage         *storageOperation
	storageResource *storageResourceOperation
	metrics         *metricsOperation
	process         *processOperation
	resource        *resourceOperation
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

func (k *Kubernetes) Storage() *storageOperation {
	return k.storage
}

func (k *Kubernetes) StorageResource() *storageResourceOperation {
	return k.storageResource
}

func (k *Kubernetes) Process() *processOperation {
	return k.process
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
	//if k.isDebug {
	//	klog.InitFlags(nil)
	//	logr := klog.NewKlogr()
	//	klog.SetLogger(logr)
	//	klog.ContextualLogger(true)
	//	klog.EnableContextualLogging(true)
	//	klog.LogToStderr(true)
	//	klog.SetOutput(os.Stdout)
	//}
	config.QPS = 50    // 每秒最大 50 个请求
	config.Burst = 100 // 突发请求 100 个
	k.c = config
	k.api, err = k8s.NewForConfig(k.c)
	if err != nil {
		k.err = gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create Kubernetes client: %v", err)
		return k.err
	}
	k.metricsApi, err = metricsv.NewForConfig(k.c)
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

func toAppListOptions(app string) v1.ListOptions {
	opts := v1.ListOptions{}
	if len(app) > 0 {
		opts.LabelSelector = fmt.Sprintf("%s=%s", types.LabelApp, app)
	}
	return opts
}

func toGroupListOptions(group string) v1.ListOptions {
	opts := v1.ListOptions{}
	if len(group) > 0 {
		opts.LabelSelector = fmt.Sprintf("%s=%s", types.LabelGroup, group)
	}
	return opts
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

func (o *options) failed(err error) {
	if err == nil {
		return
	}
	if errors.IsTimeout(err) {
		o.err = gerror.NewCodef(gcode.CodeNotImplemented, "调用环境服务超时: %+v", err)
		return
	}
}

func (o *options) success() {
	o.err = nil
}
