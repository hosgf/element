package resource

import (
	"github.com/gogf/gf/v2/container/gset"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/types"
)

// Resource 资源对象
type Resource struct {
	Env       string        `json:"env,omitempty"`
	Region    string        `json:"region,omitempty"`
	Namespace string        `json:"namespace,omitempty"`
	Type      string        `json:"type,omitempty"`
	Name      string        `json:"name,omitempty"`
	Status    health.Health `json:"status,omitempty"`
	Time      int64         `json:"time,omitempty"`
	Remark    string        `json:"remark,omitempty"`
	Nodes     []Node        `json:"nodes,omitempty"`
}

func (r *Resource) ToResourceItem() Resource {
	return Resource{
		Name:   r.Name,
		Remark: r.Remark,
		Type:   r.Type,
		Status: r.Status,
		Time:   r.Time,
		Nodes:  r.Nodes,
	}
}

func (r *Resource) SetStatus() *Resource {
	if len(r.Nodes) < 1 {
		r.Status = health.UNKNOWN
		return r
	}
	healths := gset.NewSet()
	for _, node := range r.Nodes {
		status := node.Status
		if len(status.String()) < 1 {
			// 节点没有设置状态，则认为是当宕机的
			if !healths.Contains(health.DOWN) {
				healths.Add(health.DOWN)
			}
			continue
		}
		// 将节点状态添加到缓存中
		if !healths.Contains(status) {
			healths.Add(status)
		}
	}

	if healths.Size() < 1 {
		r.Status = health.UNKNOWN
		return r
	}
	if healths.Size() == 1 {
		r.Status = healths.Pop().(health.Health)
		return r
	}
	// 如果多条，则存在多个状态的进程节点
	up := false
	if healths.Contains(health.UP) {
		up = true
	}
	if up {
		r.Status = health.WARNING
		return r
	}
	if healths.Contains(health.WARNING) {
		r.Status = health.WARNING
		return r
	}
	if healths.Contains(health.DOWN) {
		r.Status = health.DOWN
		return r
	}
	if healths.Contains(health.STOP) {
		r.Status = health.STOP
		return r
	}
	if healths.Contains(health.PENDING) {
		r.Status = health.PENDING
		return r
	}
	if healths.Contains(health.UNKNOWN) {
		r.Status = health.UNKNOWN
		return r
	}
	r.Status = health.UNKNOWN
	return r
}

type Node struct {
	Name       string                 `json:"name"`
	Roles      string                 `json:"roles"`
	Status     health.Health          `json:"status"`
	Time       int64                  `json:"time"`
	Indicators map[string]interface{} `json:"indicators"`
	Details    map[string]interface{} `json:"details"`
}

type Details struct {
	Unit  string `json:"unit,omitempty"` // 单位
	Total int64  `json:"total,omitempty"`
	Free  int64  `json:"free,omitempty"`
	Usage int64  `json:"usage,omitempty"`
}

func (d *Details) SetTotal(data string) {
	if len(data) < 1 {
		return
	}
	value, unit := types.Parse(data)
	if len(d.Unit) < 1 {
		d.Unit = unit
	}
	if value > 0 {
		d.Total = value
	}
}

func (d *Details) SetFree(data string) {
	if len(data) < 1 {
		return
	}
	value, unit := types.Parse(data)
	if len(d.Unit) < 1 {
		d.Unit = unit
	}
	if value > 0 {
		d.Free = value
	}
}

func (d *Details) SetUsage(data string) {
	if len(data) < 1 {
		return
	}
	value, unit := types.Parse(data)
	if len(d.Unit) < 1 {
		d.Unit = unit
	}
	if value > 0 {
		d.Usage = value
	}
}

func (d *Details) SetTotalValue(data int64) {
	d.Total = data
}

func (d *Details) SetFreeValue(data int64) {
	d.Free = data
}

func (d *Details) SetUsageValue(data int64) {
	d.Usage = data
}

func (d *Details) SetUnit(data string) {
	d.Unit = data
}

func (d *Details) ThroughUsageConstruction(data int64) {
	d.Usage = data
	d.Free = d.Total - data
}
