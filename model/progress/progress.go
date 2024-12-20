package progress

type ProgressGroup struct {
	Namespace string     `json:"namespace,omitempty"`
	Group     string     `json:"group,omitempty"`
	Status    string     `json:"status,omitempty"`
	Details   []Progress `json:"details,omitempty"`
}

type Progress struct {
	Namespace  string                 `json:"namespace,omitempty"`
	PID        string                 `json:"pid,omitempty"`
	Svc        string                 `json:"svc,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Group      string                 `json:"group,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Time       int64                  `json:"time,omitempty"`
	Indicators map[string]interface{} `json:"indicators,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

type Db struct {
	Status  string   `json:"status"`
	Details Database `json:"details"`
}
type Ping struct {
	Status string `json:"status"`
}

type RefreshScope struct {
	Status string `json:"status"`
}

type Database struct {
	Database string `json:"database"`
	Select   string `json:"select * "`
}

// GroupHealth  健康检查
type GroupHealth struct {
	Namespace string   `json:"namespace,omitempty"`
	Group     string   `json:"group,omitempty"`
	Status    string   `json:"status,omitempty"`
	Time      int64    `json:"time,omitempty"`
	Details   []Health `json:"details,omitempty"`
}

// Health 健康检查
type Health struct {
	Namespace string                 `json:"namespace,omitempty"`
	PID       string                 `json:"pid,omitempty"`
	Svc       string                 `json:"svc,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Group     string                 `json:"group,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Time      int64                  `json:"time,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Port 端口号
type Port struct {
	Name       string `json:"name,omitempty"`       // 名称
	Protocol   string `json:"protocol,omitempty"`   // 协议
	Port       int32  `json:"port,omitempty"`       // 对外的端口号,外部可访问的
	TargetPort int32  `json:"targetPort,omitempty"` // 被代理的端口号,应用服务端口号
	NodePort   int32  `json:"nodePort,omitempty"`   // 代理端口号
}

// Resource 进程资源
type Resource struct {
	Type      string `json:"type,omitempty"`      // 资源类型(RAM OR CPU)
	Unit      string `json:"unit,omitempty"`      // 单位
	Minimum   int64  `json:"minimum,omitempty"`   // 最小
	Maximum   int64  `json:"maximum,omitempty"`   // 最大
	Threshold int64  `json:"threshold,omitempty"` // 阈值
}
