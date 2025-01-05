package progress

import (
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/types"
)

type ProgressGroup struct {
	Namespace string        `json:"namespace,omitempty"`
	Group     string        `json:"group,omitempty"`
	Status    health.Health `json:"status,omitempty"`
	Details   []Progress    `json:"details,omitempty"`
}

type Progress struct {
	Namespace  string                 `json:"namespace,omitempty"`
	PID        string                 `json:"pid,omitempty"`
	Service    string                 `json:"service,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Group      string                 `json:"group,omitempty"`
	Status     health.Health          `json:"status,omitempty"`
	Time       int64                  `json:"time,omitempty"`
	Indicators map[string]interface{} `json:"indicators,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

type Db struct {
	Status  health.Health `json:"status"`
	Details Database      `json:"details"`
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

// GroupHealth  健康检查
type GroupHealth struct {
	Namespace string        `json:"namespace,omitempty"`
	Group     string        `json:"group,omitempty"`
	Status    health.Health `json:"status,omitempty"`
	Time      int64         `json:"time,omitempty"`
	Details   []Health      `json:"details,omitempty"`
}

// Health 健康检查
type Health struct {
	Namespace string                 `json:"namespace,omitempty"`
	PID       string                 `json:"pid,omitempty"`
	Service   string                 `json:"service,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Group     string                 `json:"group,omitempty"`
	Status    health.Health          `json:"status,omitempty"`
	Time      int64                  `json:"time,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
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
