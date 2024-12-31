package k8s

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/util"
	corev1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

var Kubernetes = newKubernetes()

func newKubernetes() *kubernetes {
	k := &kubernetes{}
	k.namespaceOperation = &namespaceOperation{k.options}
	k.serviceOperation = &serviceOperation{k.options}
	k.podsOperation = &podsOperation{k.options}
	return k
}

type options struct {
	*option
	api *k8s.Clientset
}

type kubernetes struct {
	*options
	namespaceOperation *namespaceOperation
	serviceOperation   *serviceOperation
	podsOperation      *podsOperation
}

func (k *kubernetes) Namespace() *namespaceOperation {
	return k.namespaceOperation
}

func (k *kubernetes) Service() *serviceOperation {
	return k.serviceOperation
}

func (k *kubernetes) Pod() *podsOperation {
	return k.podsOperation
}

func (k *kubernetes) Init(homePath string) error {
	kubeconfig := filepath.Join(util.Any(homePath == "", homedir.HomeDir(), homePath), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		k.err = gerror.NewCodef(gcode.CodeNotImplemented, "Failed to build kubeconfig: %v", err)
		return k.err
	}
	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		k.err = gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create Kubernetes client: %v", err)
		return k.err
	}
	k.api = clientset
	k.namespaceOperation.api = k.api
	return nil
}

func (k *kubernetes) Version() (string, error) {
	if k.err != nil {
		return "", k.err
	}
	version, err := k.api.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func any(expr bool, a, b corev1.ServiceType) corev1.ServiceType {
	if expr {
		return a
	}
	return b
}
