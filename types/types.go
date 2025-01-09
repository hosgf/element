package types

import (
	"github.com/hosgf/element/health"
)

var (
	DefaultServiceType string = "ClusterIP"
)

// Label 标签类型
type Label string

const (
	LabelApp   Label = "x-platform-app"   // 应用名称
	LabelGroup Label = "x-platform-group" // 所属进程组名称
	LabelOwner Label = "x-platform-owner" // 所属人
	LabelScope Label = "x-platform-scope" // 所属作用域
)

func (l Label) String() string {
	return string(l)
}

// ProtocolType 协议类型
type ProtocolType string

const (
	ProtocolTcp  ProtocolType = "TCP"
	ProtocolUdp  ProtocolType = "UDP"
	ProtocolSctp ProtocolType = "SCTP"
)

func (t ProtocolType) String() string {
	return string(t)
}

// ResourceType 资源类型
type ResourceType string

const (
	ResourceCPU     ResourceType = "cpu"
	ResourceMemory  ResourceType = "memory"
	ResourceStorage ResourceType = "storage"
)

func (r ResourceType) String() string {
	return string(r)
}

type Namespace struct {
	Region string        `json:"region,omitempty"`
	Name   string        `json:"name,omitempty"`
	Label  string        `json:"label,omitempty"`
	Remark string        `json:"remark,omitempty"`
	Status health.Health `json:"status,omitempty"`
}

// Environment 环境变量
type Environment struct {
	Name  string            `json:"name,omitempty"`  // 环境变量名称
	Path  string            `json:"path,omitempty"`  // 映射地址
	Items map[string]string `json:"items,omitempty"` // 变量信息
}

// Labels 标签信息
type Labels struct {
	App    string            `json:"app,omitempty"`    // 所属应用
	Group  string            `json:"group,omitempty"`  // 所属进程组
	Owner  string            `json:"owner,omitempty"`  // 所属人
	Scope  string            `json:"scope,omitempty"`  // 作用范围
	Labels map[string]string `json:"labels,omitempty"` // 标签
}

//func (c Container) toVolumes() corev1.Container {
//	corev1.Container{
//		Name:    c.Name,
//		Image:   c.Image,
//		Command: c.Command,
//		Args:    c.Args,
//		Args:    c.Args,
//	}
//}
