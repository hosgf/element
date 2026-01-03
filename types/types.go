package types

import (
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/util"
)

const (
	UserIdKey    string = "user_id"
	RequestIdKey string = "request_id"
)

const (
	DefaultServiceType   string = "ClusterIP"
	DefaultGroupLabel    string = "app"
	DefaultMinimumCpu    int64  = 100
	DefaultMinimumMemory int64  = 30
	DefaultMaximumCpu    int64  = -1
	DefaultMaximumMemory int64  = 2048
	DefaultCpuUnit       string = "m"
	DefaultMemoryUnit    string = "Mi"
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
	return strings.ToUpper(string(t))
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

// AccessMode 存储访问模式
type AccessMode string

const (
	ReadWriteOnce AccessMode = "ReadWriteOnce"
	ReadOnlyMany  AccessMode = "ReadOnlyMany"
	ReadWriteMany AccessMode = "ReadWriteMany"
)

func (a AccessMode) String() string {
	return string(a)
}

// StorageType 存储类型
type StorageType string

const (
	StoragePVC    StorageType = "pvc"
	StorageConfig StorageType = "config"
)

func (a StorageType) String() string {
	if string(a) == "" {
		return string(StoragePVC)
	}
	return gstr.ToLower(string(a))
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

// Mount 持久化挂载,存储关系映射,最终持久化目录为:{Path}/{SubPath}
type Mount struct {
	Name    string `json:"name,omitempty"`    // 挂载设备名称,关联存储的名字
	Path    string `json:"path,omitempty"`    // 挂载目录,应用要使用的目录, 容器内部工作目录
	SubPath string `json:"subPath,omitempty"` // 挂载子目录,可以是个目录或者文件,
}

func (m *Mount) GetPath() string {
	return util.GetOrDefault(m.Path, filepath.Join("data", m.Name))
}

// Config 配置
type Config struct {
	Name  string `json:"name,omitempty"`  // 配置名称
	Type  string `json:"type,omitempty"`  // 配置类型,config , volume
	Item  string `json:"item,omitempty"`  // 配置详情条目
	Path  string `json:"path,omitempty"`  // 配置路径
	Scope string `json:"scope,omitempty"` // 作用域,默认全部
}

func (c *Config) IsMatch(name string) bool {
	if len(c.Scope) < 1 {
		return true
	}
	if gstr.Contains(c.Scope, "*") {
		return true
	}
	for _, s := range gstr.Split(c.Scope, ",") {
		if gstr.Equal(s, name) {
			return true
		}
	}
	return false
}

type StorageResource struct {
	Type string `json:"type,omitempty"` // 存储名称
	Item string `json:"item,omitempty"` // 存储详情
}

// Storage 存储 对象
type Storage struct {
	Name       string          `json:"name,omitempty"`       // 存储名称
	Type       StorageType     `json:"type,omitempty"`       // 存储类型,config , volume , pvc
	AccessMode AccessMode      `json:"accessMode,omitempty"` // 访问模式
	Size       string          `json:"size,omitempty"`       // 存储大小
	Path       string          `json:"path,omitempty"`       // 存储路径
	Item       interface{}     `json:"item,omitempty"`       // 存储详情,
	Resource   StorageResource `json:"resource,omitempty"`   // 存储资源，可为空
	Scope      string          `json:"scope,omitempty"`      // 作用域,默认全部
}

func (s *Storage) IsMatch(name string) bool {
	if len(s.Scope) < 1 {
		return true
	}
	if gstr.Contains(s.Scope, "*") {
		return true
	}
	for _, s := range gstr.Split(s.Scope, ",") {
		if gstr.Equal(s, name) {
			return true
		}
	}
	return false
}

func (s *Storage) ReadOnly() bool {
	switch s.ToAccessMode() {
	case ReadOnlyMany:
		return true
	default:
		return false
	}
}

func (s *Storage) ToStorageType() StorageType {
	if s.Type == "" {
		return StoragePVC
	}
	return StorageType(s.Type.String())
}

func (s *Storage) ToAccessMode() AccessMode {
	if s.AccessMode == "" {
		return ReadWriteOnce
	}
	return AccessMode(gstr.CaseCamel(s.AccessMode.String()))
}

func (s *Storage) GetPath() string {
	return util.GetOrDefault(s.Path, filepath.Join(util.GetOrDefault(util.GetHomePath(), "/"), "data", s.Name))
}
