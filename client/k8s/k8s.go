package k8s

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hosgf/element/types"
	"github.com/hosgf/element/uerrors"
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
	ctx := context.Background()
	kubeconfig := k.config(homePath)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		k.err = uerrors.WrapKubernetesError(ctx, err, "初始化Kubernetes配置")
		return k.err
	}
	config.QPS = 50    // 每秒最大 50 个请求
	config.Burst = 100 // 突发请求 100 个
	k.c = config
	k.api, err = k8s.NewForConfig(k.c)
	if err != nil {
		k.err = uerrors.WrapKubernetesError(ctx, err, "创建Kubernetes客户端")
		return k.err
	}
	k.metricsApi, err = metricsv.NewForConfig(k.c)
	if err != nil {
		k.err = uerrors.WrapKubernetesError(ctx, err, "创建Metrics客户端")
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

// toLabelListOptions 根据标签名和值创建 ListOptions
func toLabelListOptions(label types.Label, value string) v1.ListOptions {
	opts := v1.ListOptions{}
	if value != "" {
		opts.LabelSelector = fmt.Sprintf("%s=%s", label, value)
	}
	return opts
}

func toAppListOptions(app string) v1.ListOptions {
	return toLabelListOptions(types.LabelApp, app)
}

func toGroupListOptions(group string) v1.ListOptions {
	return toLabelListOptions(types.LabelGroup, group)
}

func toServiceType(serviceType string) string {
	if serviceType == "" {
		return types.DefaultServiceType
	}
	return serviceType
}

// isExist 检查资源是否存在
// value: 资源对象，如果为nil表示资源不存在
// err: 获取资源时的错误
// operation: 操作名称，用于错误消息
func (o *options) isExist(ctx context.Context, value interface{}, err error, operation string) (bool, error) {
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, uerrors.WrapKubernetesError(ctx, err, operation)
	}
	// err == nil 时，检查 value 是否为 nil
	return value != nil, nil
}

func (o *options) failed(ctx context.Context, err error, operation string) {
	if err == nil {
		return
	}
	if errors.IsTimeout(err) {
		o.err = uerrors.NewKubernetesError(ctx, operation, "调用环境服务超时", err.Error())
		return
	}
	// 对于其他错误，包装为Kubernetes错误
	o.err = uerrors.WrapKubernetesError(ctx, err, operation)
}

func (o *options) success() {
	o.err = nil
}
