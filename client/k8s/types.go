package k8s

import (
	"fmt"

	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/types"
)

// ProcessGroupConfig 进程组配置对象
type ProcessGroupConfig struct {
	Namespace   string          `json:"namespace,omitempty"`   // 运行进程的资源空间
	GroupName   string          `json:"groupName,omitempty"`   // 进程组名称
	Labels      types.Labels    `json:"labels,omitempty"`      // 进程组标签
	RunningNode string          `json:"runningNode,omitempty"` // 运行节点,可为空
	Replicas    int32           `json:"replicas,omitempty"`    // 节点数，以进程组为纬度 默认为 1
	AllowUpdate bool            `json:"allowUpdate,omitempty"` // 是否允许更新,进程存在则更新
	Secret      string          `json:"secret,omitempty"`      // pull镜像时使用的secret
	Config      []types.Config  `json:"config,omitempty"`      // 配置信息
	Storage     []types.Storage `json:"storage,omitempty"`     // 存储,以进程组的维度来定义,进程通过挂载与存储关联
	Process     []ProcessConfig `json:"process,omitempty"`     // 进程组下的进程信息
}

func (pg *ProcessGroupConfig) toModel() Model {
	return Model{
		Namespace:   pg.Namespace,
		App:         pg.Labels.App,
		Group:       pg.GroupName,
		Owner:       pg.Labels.Owner,
		Scope:       pg.Labels.Scope,
		Labels:      pg.Labels.Labels,
		AllowUpdate: pg.AllowUpdate,
	}
}

func (pg *ProcessGroupConfig) initConfig() {
	if pg.Config == nil {
		pg.Config = make([]types.Config, 0)
	}
	cmap := map[string]string{}
	for _, c := range pg.Config {
		cmap[c.Name] = c.Scope
	}
	if _, ok := cmap["timezone"]; !ok {
		pg.Config = append(pg.Config, types.Config{
			Name: "timezone",
			Type: "config",
			Item: "/usr/share/zoneinfo/Asia/Shanghai",
			Path: "/etc/timezone:ro",
		})
	}
	if _, ok := cmap["localtime"]; !ok {
		pg.Config = append(pg.Config, types.Config{
			Name: "localtime",
			Type: "config",
			Item: "/etc/localtime",
			Path: "/etc/localtime:ro",
		})
	}
}

// ProcessConfig 进程对象
type ProcessConfig struct {
	Name        string              `json:"name,omitempty"`        // 进程名称
	Service     string              `json:"service,omitempty"`     // 服务名
	ServiceType string              `json:"serviceType,omitempty"` // 服务的访问方式
	Source      string              `json:"source,omitempty"`      // 镜像
	PullPolicy  string              `json:"pullPolicy,omitempty"`  // 镜像拉取策略
	Command     []string            `json:"command,omitempty"`     // 运行命令
	Args        []string            `json:"args,omitempty"`        // 运行参数
	Ports       []progress.Port     `json:"ports,omitempty"`       // 服务端口信息
	Resource    []progress.Resource `json:"resource,omitempty"`    // 进程运行所需的资源
	Env         []types.Environment `json:"env,omitempty"`         // 环境变量
	Mounts      []types.Mount       `json:"mounts,omitempty"`      // 卷挂载
	Probe       ProbeConfig         `json:"probe,omitempty"`       // 探针
}

func (p *ProcessConfig) toMounts(pg *ProcessGroupConfig) {
	if p.Mounts == nil {
		p.Mounts = make([]types.Mount, 0)
	}
	cmap := map[string]string{}
	for _, c := range p.Mounts {
		cmap[c.Name] = c.Path
	}
	p.toConfigMounts(cmap, pg)
	p.toStorageMounts(cmap, pg)
}

func (p *ProcessConfig) toConfigMounts(cmap map[string]string, pg *ProcessGroupConfig) {
	if pg.Config == nil || len(pg.Config) == 0 {
		return
	}
	for _, c := range pg.Config {
		if len(c.Name) < 1 || len(c.Path) < 1 {
			continue
		}
		if _, ok := cmap[c.Name]; ok {
			continue
		}
		if !c.IsMatch(c.Name) {
			continue
		}
		cmap[c.Name] = c.Path
		p.Mounts = append(p.Mounts, types.Mount{
			Name: c.Name,
			Path: c.Path,
		})
	}
}

func (p *ProcessConfig) toStorageMounts(cmap map[string]string, pg *ProcessGroupConfig) {
	if pg.Storage == nil || len(pg.Storage) == 0 {
		return
	}
	for _, s := range pg.Storage {
		if len(s.Name) < 1 {
			continue
		}
		if _, ok := cmap[s.Name]; ok {
			continue
		}
		if !s.IsMatch(s.Name) {
			continue
		}
		cmap[s.Name] = s.Path
		p.Mounts = append(p.Mounts, types.Mount{
			Name: s.Name,
			Path: s.Path,
		})
	}
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

func (p *ProcessConfig) toEnvConfig() (map[string]string, []types.Environment) {
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

type Model struct {
	AllowUpdate bool              `json:"allowUpdate,omitempty"` // 是否允许更新,进程存在则更新
	Namespace   string            `json:"namespace,omitempty"`
	Name        string            `json:"name,omitempty"`
	App         string            `json:"app,omitempty"`
	Group       string            `json:"group,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Scope       string            `json:"scope,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	groupLabel  string
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
