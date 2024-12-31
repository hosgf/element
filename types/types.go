package types

type Label string

const (
	LabelGroupName Label = "app"              // 所属进程组名称
	LabelApp       Label = "x-platform-app"   // 应用名称
	LabelOwner     Label = "x-platform-owner" // 所属人
	LabelScope     Label = "x-platform-scope" // 所属作用域
)

func (l Label) String() string {
	return string(l)
}

type ResourceType string

const (
	ResourceCPU     ResourceType = "CPU"
	ResourceRAM     ResourceType = "RAM"
	ResourceStorage ResourceType = "STORAGE"
	//	v1.ResourceCPU:    "500m",  // 请求 500m CPU
	//
	// v1.ResourceMemory: "256Mi", // 请求 256Mi 内存
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
