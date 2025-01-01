package k8s

import (
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/health"
)

// pod状态
const (
	// Pending 等待中
	Pending = "pending"
	// Running 运行中
	Running = "running"
	Active  = "active"
	// Completed 运行中
	Completed = "completed"
	// Succeeded 正常终止
	Succeeded   = "succeeded"
	Terminating = "terminating"
	// Failed 异常停止
	Failed = "failed"
	Evicte = "evicte"
	// Unkonwn 未知状态
	Unkonwn = "unkonwn"
	// CrashLoopBackOff 容器退出，kubelet正在将它重启
	CrashLoopBackOff = "crashloopbackoff"
	// InvalidImageName 无法解析镜像名称
	InvalidImageName = "invalidimagename"
	// ImageInspectError 无法校验镜像
	ImageInspectError = "imageinspecterror"
	// ErrImageNeverPull 策略禁止拉取镜像
	ErrImageNeverPull = "errimageneverpull"
	// ImagePullBackOff 正在重试拉取
	ImagePullBackOff = "imagepullbackoff"
	// RegistryUnavailable 连接不到镜像中心
	RegistryUnavailable = "registryunavailable"
	// ErrImagePull 通用的拉取镜像出错
	ErrImagePull = "errimagepull"
	// CreateContainerConfigError 不能创建kubelet使用的容器配置
	CreateContainerConfigError = "createcontainerconfigerror"
	// CreateContainerError 创建容器失败
	CreateContainerError = "createcontainererror"
	// RunContainerError 启动容器失败
	RunContainerError = "runcontainererror"
	// PostStartHookError 执行hook报错
	PostStartHookError = "poststarthookerror"
	// ContainersNotInitialized 容器没有初始化完毕
	ContainersNotInitialized = "containersnotinitialized"
	// ContainersNotReady 容器没有准备完毕
	ContainersNotReady = "containersnotready"
	// ContainerCreating 容器创建中
	ContainerCreating = "containercreating"
	// PodInitializing pod 初始化中
	PodInitializing = "podinitializing"
	// DockerDaemonNotReady docker还没有完全启动
	DockerDaemonNotReady = "dockerdaemonnotready"
	// NetworkPluginNotReady 网络插件还没有完全启动
	NetworkPluginNotReady = "networkpluginnotready"
)

func Status(status string) health.Health {
	switch gstr.ToLower(status) {
	case
		Running,
		Active,
		Completed:
		return health.UP
	case
		Pending,
		PodInitializing,
		ContainerCreating,
		NetworkPluginNotReady,
		ContainersNotReady,
		ContainersNotInitialized,
		DockerDaemonNotReady:
		return health.PENDING
	case
		Succeeded,
		Terminating,
		CrashLoopBackOff,
		Evicte:
		return health.STOP
	case Unkonwn:
		return health.UNKNOWN
	default:
		return health.DOWN
	}
}

func NodeStatus(status string) health.Health {
	switch gstr.ToLower(status) {
	case
		"true":
		return health.UP
	case
		"false":
		return health.DOWN
	case Unkonwn:
		return health.UNKNOWN
	default:
		return health.DOWN
	}
}
