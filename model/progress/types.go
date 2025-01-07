package progress

import (
	"strings"

	"github.com/hosgf/element/health"
	"github.com/hosgf/element/types"
)

type ProgressGroup struct {
	Region    string          `json:"region,omitempty"`
	Namespace string          `json:"namespace,omitempty"`
	GroupID   string          `json:"groupId,omitempty"`
	Labels    *ProgressLabels `json:"labels,omitempty"`
	Status    health.Health   `json:"status,omitempty"`
	Time      int64           `json:"time,omitempty"`
	Details   []Progress      `json:"details,omitempty"`
}

func (p *ProgressGroup) MatchNamespace(namespace string) bool {
	return strings.EqualFold(p.Namespace, namespace)
}

type Progress struct {
	Region     string                 `json:"region,omitempty"`
	Namespace  string                 `json:"namespace,omitempty"`
	PID        string                 `json:"pid,omitempty"`
	Service    string                 `json:"service,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Labels     *ProgressLabels        `json:"labels,omitempty"`
	Status     health.Health          `json:"status,omitempty"`
	Time       int64                  `json:"time,omitempty"`
	Indicators map[string]interface{} `json:"indicators,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (p *Progress) ToGroupProgress() Progress {
	return Progress{
		PID:        p.PID,
		Service:    p.Service,
		Name:       p.Name,
		Status:     p.Status,
		Indicators: p.Indicators,
		Details:    p.Details,
	}
}

func (p *Progress) ToGroup() ProgressGroup {
	return ProgressGroup{
		Namespace: p.Namespace,
		Region:    p.Region,
		Labels:    p.Labels,
		GroupID:   p.PID,
		Time:      p.Time,
		Status:    health.UNKNOWN,
		Details:   []Progress{p.ToGroupProgress()},
	}
}

func (p *Progress) ToHealth() Health {
	return Health{
		Namespace: p.Namespace,
		Region:    p.Region,
		PID:       p.PID,
		Name:      p.Name,
		Service:   p.Service,
		Status:    p.Status,
		Time:      p.Time,
		Group:     p.GetGroup(),
		Address:   p.GetAddress(),
		Ports:     p.GetPorts(),
	}
}

func (p *Progress) MatchGroup(group string) bool {
	return strings.EqualFold(p.GetGroup(), group)
}

func (p *Progress) MatchNamespace(namespace string) bool {
	return strings.EqualFold(p.Namespace, namespace)
}

func (p *Progress) GetGroup() string {
	if p.Labels == nil {
		return ""
	}
	return p.Labels.Group
}

func (p *Progress) GetAddress() string {
	if p.Indicators == nil {
		return ""
	}
	return p.Indicators["address"].(string)
}

func (p *Progress) GetPorts() []*ProgressPort {
	if p.Details == nil {
		return nil
	}
	return p.Details["ports"].([]*ProgressPort)
}

func (p *Progress) SetAddress(address string) {
	if p.Indicators == nil {
		p.Indicators = make(map[string]interface{})
	}
	p.Indicators["address"] = address
}

func (p *Progress) SetPorts(ports []*ProgressPort) {
	if p.Details == nil {
		p.Details = make(map[string]interface{})
	}
	p.Details["ports"] = ports
}

type ProgressLabels struct {
	App    string            `json:"app,omitempty"`    // 所属应用
	Group  string            `json:"group,omitempty"`  // 所属进程组
	Owner  string            `json:"owner,omitempty"`  // 所属人
	Scope  string            `json:"scope,omitempty"`  // 作用范围
	Labels map[string]string `json:"labels,omitempty"` // 标签
}

type Db struct {
	Status  health.Health `json:"status"`
	Details Database      `json:"details"`
}
type Details struct {
	Status  health.Health     `json:"status"`
	Details map[string]string `json:"details"`
}

type Ping struct {
	Status health.Health `json:"status"`
}

type RefreshScope struct {
	Status health.Health `json:"status"`
}

type Database struct {
	Database string `json:"database"`
	Select   string `json:"select * "`
}

type ProgressPort struct {
	Name     string             `json:"name,omitempty"`     // 名称
	Protocol types.ProtocolType `json:"protocol,omitempty"` // 协议
	Port     int32              `json:"port,omitempty"`     // 对外的端口号,外部可访问的
}

// Port 端口号
type Port struct {
	Name       string             `json:"name,omitempty"`       // 名称
	Protocol   types.ProtocolType `json:"protocol,omitempty"`   // 协议
	Port       int32              `json:"port,omitempty"`       // 对外的端口号,外部可访问的
	TargetPort int32              `json:"targetPort,omitempty"` // 被代理的端口号,应用服务端口号
	NodePort   int32              `json:"nodePort,omitempty"`   // 代理端口号
}

// Resource 进程资源
type Resource struct {
	Type      types.ResourceType `json:"type,omitempty"`      // 资源类型(RAM OR CPU)
	Unit      string             `json:"unit,omitempty"`      // 单位
	Minimum   int64              `json:"minimum,omitempty"`   // 最小
	Maximum   int64              `json:"maximum,omitempty"`   // 最大
	Threshold int64              `json:"threshold,omitempty"` // 阈值
}

func (r *Resource) Update(res Resource) {
	if len(res.Type) > 0 {
		res.Type = r.Type
	}
	if res.Minimum > 0 {
		r.Minimum = res.Minimum
	}
	if res.Maximum > 0 {
		r.Maximum = res.Maximum
	}
	if res.Threshold > 0 {
		r.Threshold = res.Threshold
	}
}

func (r *Resource) SetMinimum(data string) {
	if len(data) < 1 {
		return
	}
	value, unit := types.Parse(data)
	if len(r.Unit) < 1 {
		r.Unit = unit
	}
	if value > 0 {
		r.Minimum = value
	}
}

func (r *Resource) SetMaximum(data string) {
	if len(data) < 1 {
		return
	}
	value, unit := types.Parse(data)
	if len(r.Unit) < 1 {
		r.Unit = unit
	}
	if value > 0 {
		r.Maximum = value
	}
}
