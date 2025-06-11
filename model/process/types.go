package process

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/types"
)

type ProcessGroup struct {
	Region    string        `json:"region,omitempty"`
	Namespace string        `json:"namespace,omitempty"`
	GroupID   string        `json:"groupId,omitempty"`
	Labels    *types.Labels `json:"labels,omitempty"`
	Status    health.Health `json:"status,omitempty"`
	Time      int64         `json:"time,omitempty"`
	Details   []Process     `json:"details,omitempty"`
}

func (p *ProcessGroup) MatchNamespace(namespace string) bool {
	return strings.EqualFold(p.Namespace, namespace)
}

type Process struct {
	Region     string                 `json:"region,omitempty"`
	Namespace  string                 `json:"namespace,omitempty"`
	PID        string                 `json:"pid,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Service    string                 `json:"service,omitempty"`
	Labels     *types.Labels          `json:"labels,omitempty"`
	Status     health.Health          `json:"status,omitempty"`
	Time       int64                  `json:"time,omitempty"`
	Indicators map[string]interface{} `json:"indicators,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (p *Process) GetServiceType() string {
	if p.Details == nil || len(p.Details) < 1 {
		return ""
	}
	if svc, ok := p.Details["service"]; ok {
		var details Details
		if err := gconv.Struct(svc, &details); err != nil {
			return ""
		}
		if details.Details == nil || len(details.Details) < 1 {
			return ""
		}
		return details.Details["serviceType"]
	}
	return ""
}

func (p *Process) ToGroupProcess() Process {
	return Process{
		PID:        p.PID,
		Service:    p.Service,
		Name:       p.Name,
		Status:     p.Status,
		Indicators: p.Indicators,
		Details:    p.Details,
	}
}

func (p *Process) ToGroup() ProcessGroup {
	return ProcessGroup{
		Namespace: p.Namespace,
		Region:    p.Region,
		Labels:    p.Labels,
		GroupID:   p.PID,
		Time:      p.Time,
		Status:    health.UNKNOWN,
		Details:   []Process{p.ToGroupProcess()},
	}
}

func (p *Process) ToHealth() Health {
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

func (p *Process) MatchScope(scope string) bool {
	return strings.EqualFold(p.GetScope(), scope)
}

func (p *Process) MatchGroup(group string) bool {
	return strings.EqualFold(p.GetGroup(), group)
}

func (p *Process) MatchNamespace(namespace string) bool {
	return strings.EqualFold(p.Namespace, namespace)
}

func (p *Process) GetScope() string {
	if p.Labels == nil {
		return ""
	}
	return p.Labels.Scope
}

func (p *Process) GetGroup() string {
	if p.Labels == nil {
		return ""
	}
	return p.Labels.Group
}

func (p *Process) GetRunningNode() string {
	if p.Details == nil {
		return ""
	}
	return p.Details["runningNode"].(string)
}

func (p *Process) GetAddress() string {
	if p.Details == nil {
		return ""
	}
	return p.Details["address"].(string)
}

func (p *Process) GetPorts() []ProcessPort {
	if p.Details == nil {
		return nil
	}
	return p.Details["ports"].([]ProcessPort)
}

func (p *Process) SetRunningNode(runningNode string) {
	if p.Details == nil {
		p.Details = make(map[string]interface{})
	}
	p.Details["runningNode"] = runningNode
}

func (p *Process) SetAddress(address string) {
	if p.Details == nil {
		p.Details = make(map[string]interface{})
	}
	p.Details["address"] = address
}

func (p *Process) SetPorts(ports []ProcessPort) {
	if p.Details == nil {
		p.Details = make(map[string]interface{})
	}
	p.Details["ports"] = ports
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

type ProcessPort struct {
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

func (p *Port) GetProtocol() types.ProtocolType {
	if p.Protocol == "" {
		p.Protocol = types.ProtocolTcp
	}
	return p.Protocol
}

func (p *Port) GetName() string {
	p.Format()
	if p.Name == "" {
		return string(p.Port)
	}
	return fmt.Sprintf("%s-%d", p.Name, p.Port)
}

func (p *Port) Format() *Port {
	if p.Protocol == "" {
		p.Protocol = types.ProtocolTcp
	}
	if p.NodePort == 0 {
		p.NodePort = -1
	}
	if p.Name == "" {
		p.Name = p.Protocol.String()
	}
	port := p.Port
	targetPort := p.TargetPort
	if targetPort > 0 && port > 0 {
		return p
	}
	if targetPort < 1 {
		targetPort = port
	}
	if port < 1 {
		port = targetPort
	}
	p.Port = port
	p.TargetPort = targetPort
	return p
}

func (p *Port) Copy() Port {
	p.Format()
	return Port{
		Name:       p.Name,
		Protocol:   p.Protocol,
		Port:       p.Port,
		TargetPort: p.TargetPort,
		NodePort:   p.NodePort,
	}
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
