package k8s

import (
	"context"
	"fmt"

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
	Metrics() metricsInterface
	PodTemplate() podTemplatesInterface
	Progress() progressInterface
	Resource() resourceInterface
}

type progressInterface interface {
	List(ctx context.Context, namespace string) ([]*progress.Progress, error)
	Running(ctx context.Context, config *ProcessGroupConfig) error
	Start(ctx context.Context, config *ProcessGroupConfig) error
	Stop(ctx context.Context, config *ProcessGroupConfig) error
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
}

type Model struct {
	AllowUpdate bool              `json:"allowUpdate,omitempty"` // 是否允许更新,进程存在则更新
	Namespace   string            `json:"namespace,omitempty"`
	Name        string            `json:"name,omitempty"`
	App         string            `json:"app,omitempty"`
	Group       string            `json:"group,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Scope       string            `json:"scope,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	groupLabel  string            `json:"groupLabel,omitempty"`
}

func (m *Model) Key() string {
	return fmt.Sprintf("%s.%s", m.Name, m.Namespace)
}

func (m *Model) toSelector() map[string]string {
	return map[string]string{
		types.LabelGroup.String(): m.Group,
	}
}

func (m *Model) toLabels() *types.Labels {
	return &types.Labels{
		App:    m.App,
		Group:  m.Group,
		Owner:  m.Owner,
		Scope:  m.Scope,
		Labels: m.Labels,
	}
}

func (m *Model) setTypesLabels(labels *types.Labels) {
	if nil == labels {
		return
	}
	m.App = labels.App
	m.Owner = labels.Owner
	m.Scope = labels.Scope
	m.Group = labels.Group
	m.Labels = labels.Labels
}

func (m *Model) labels() map[string]string {
	labels := make(map[string]string)
	if len(m.App) > 0 {
		labels[types.LabelApp.String()] = m.App
	}
	if len(m.Owner) > 0 {
		labels[types.LabelOwner.String()] = m.Owner
	}
	if len(m.Scope) > 0 {
		labels[types.LabelScope.String()] = m.Scope
	}
	if len(m.Group) > 0 {
		labels[types.LabelGroup.String()] = m.Group
	}
	if m.Labels != nil {
		for k, v := range m.Labels {
			labels[k] = v
		}
	}
	return labels
}

func (m *Model) setLabels(labels map[string]string) {
	if len(labels) < 1 {
		return
	}
	m.App = labels[types.LabelApp.String()]
	delete(labels, types.LabelApp.String())

	m.Owner = labels[types.LabelOwner.String()]
	delete(labels, types.LabelOwner.String())

	m.Scope = labels[types.LabelScope.String()]
	delete(labels, types.LabelScope.String())

	if len(m.Group) < 1 {
		m.Group = labels[types.LabelGroup.String()]
		m.groupLabel = types.LabelGroup.String()
		delete(labels, m.groupLabel)
	}

	if len(m.Group) < 1 {
		m.Group = labels[types.DefaultGroupLabel]
		m.groupLabel = types.DefaultGroupLabel
		delete(labels, m.groupLabel)
	}

	delete(labels, "pod-template-hash")

	if labels == nil || len(labels) < 1 {
		return
	}

	if m.Labels == nil {
		m.Labels = map[string]string{}
	}

	for k, v := range labels {
		m.Labels[k] = v
	}
}

// ProcessGroupConfig 进程组配置对象
type ProcessGroupConfig struct {
	Namespace   string          `json:"namespace,omitempty"`   // 运行进程的资源空间
	GroupName   string          `json:"groupName,omitempty"`   // 进程组名称
	Labels      types.Labels    `json:"labels,omitempty"`      // 进程组标签
	RunningNode string          `json:"runningNode,omitempty"` // 运行节点,可为空
	Replicas    int32           `json:"replicas,omitempty"`    // 节点数，以进程组为纬度 默认为 1
	AllowUpdate bool            `json:"allowUpdate,omitempty"` // 是否允许更新,进程存在则更新
	Secret      string          `json:"secret,omitempty"`      // pull镜像时使用的secret
	Process     []ProcessConfig `json:"process,omitempty"`     // 进程组下的进程信息
}

// ProcessConfig 进程对象
type ProcessConfig struct {
	Name         string              `json:"name,omitempty"`         // 进程名称
	Service      string              `json:"service,omitempty"`      // 服务名
	ServiceType  string              `json:"serviceType,omitempty"`  // 服务的访问方式
	Source       string              `json:"source,omitempty"`       // 镜像
	PullPolicy   string              `json:"pullPolicy,omitempty"`   // 镜像拉取策略
	Command      []string            `json:"command,omitempty"`      // 运行命令
	Args         []string            `json:"args,omitempty"`         // 运行参数
	Ports        []progress.Port     `json:"ports,omitempty"`        // 服务端口信息
	Resource     []progress.Resource `json:"resource,omitempty"`     // 进程运行所需的资源
	Env          []types.Environment `json:"env,omitempty"`          // 环境变量
	Storage      Storage             `json:"storage,omitempty"`      // 存储
	VolumeMounts []VolumeMount       `json:"volumeMounts,omitempty"` // 卷挂载
	Probe        ProbeConfig         `json:"probe,omitempty"`        // 探针
}

func (p *ProcessConfig) ToEnv() map[string]string {
	env := make(map[string]string)
	if p.Env == nil || len(p.Env) < 1 {
		return env
	}
	for _, e := range p.Env {
		if len(e.Name) > 0 {
			continue
		}
		if e.Items == nil || len(e.Items) < 1 {
			continue
		}
		for k, v := range e.Items {
			if len(v) > 0 {
				env[k] = v
			}
		}
	}
	return env
}

func (p *ProcessConfig) ToEnvConfig() (map[string]string, []types.Environment) {
	env := make(map[string]string)
	config := make([]types.Environment, 0)
	if p.Env == nil || len(p.Env) < 1 {
		return env, config
	}
	for _, e := range p.Env {
		if len(e.Name) > 0 {
			config = append(config, e)
			continue
		}
		if e.Items == nil || len(e.Items) < 1 {
			continue
		}
		for k, v := range e.Items {
			if len(v) > 0 {
				env[k] = v
			}
		}
	}
	return env, config
}

// VolumeMount 挂载卷,存储关系映射
type VolumeMount struct {
	Name    string `json:"name,omitempty"`    // 挂载设备名称,关联存储的名字
	Path    string `json:"path,omitempty"`    // 挂载目录,应用要使用的目录
	SubPath string `json:"subPath,omitempty"` // 子目录
}

// Storage 存储 对象
type Storage struct {
	Name      string `json:"name,omitempty"`      // 存储名称
	Type      string `json:"type,omitempty"`      // 存储类型
	Size      string `json:"size,omitempty"`      // 存储大小
	ClaimName string `json:"claimName,omitempty"` // 存储分类名称
}

// ProbeConfig 探针配置
type ProbeConfig struct {
	Enabled             bool   `json:"enabled"`                                     // 是否启用
	ProbeType           string `json:"probeType" v:"required-if:enabled,true"`      // exec http tcp
	ExecCommand         string `json:"execCommand" v:"required-if:probeType,exec"`  // 监控命令
	HttpGetPath         string `json:"httpGetPath" v:"required-if:probeType,http"`  // api路径
	HttpGetPort         int    `json:"httpGetPort" v:"required-if:probeType,http"`  // 端口号
	TcpSocketPort       int    `json:"tcpSocketPort" v:"required-if:probeType,tcp"` // 端口号
	InitialDelaySeconds int    `json:"initialDelaySeconds"`                         // 容器启动后多久开始探测 默认值 300
	TimeoutSeconds      int    `json:"timeoutSeconds"`                              // 表示容器必须在多少秒内做出相应反馈给probe，否则视为探测失败 默认值 10
	PeriodSeconds       int    `json:"periodSeconds"`                               // 探测周期，每多少秒探测一次 默认值 30
	SuccessThreshold    int    `json:"successThreshold"`                            // 连续探测几次成功表示成功 默认值 1
	FailureThreshold    int    `json:"failureThreshold"`                            // 连续探测几次失败表示失败 默认值 3
}
