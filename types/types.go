package types

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
	ProtocolTcp ProtocolType = "tcp"
	ProtocolUdp ProtocolType = "udp"
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

//func (c Container) toVolumes() corev1.Container {
//	corev1.Container{
//		Name:    c.Name,
//		Image:   c.Image,
//		Command: c.Command,
//		Args:    c.Args,
//		Args:    c.Args,
//	}
//}
