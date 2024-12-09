package resource

import (
	"github.com/gogf/gf/v2/container/gset"
	"github.com/hosgf/element/health"
)

// Resource 资源对象
type Resource struct {
	Region string `json:"namespace,omitempty"`
	Type   string `json:"type,omitempty"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
	Time   int64  `json:"time,omitempty"`
	Remark string `json:"remark,omitempty"`
	Nodes  []Node `json:"nodes,omitempty"`
}

func (r *Resource) SetStatus() *Resource {
	if len(r.Nodes) < 1 {
		r.Status = health.UNKNOWN.String()
		return r
	}
	healths := gset.NewStrSet()
	for _, node := range r.Nodes {
		status := node.Status
		if len(status) < 1 {
			// 节点没有设置状态，则认为是当宕机的
			if !healths.Contains(health.DOWN.String()) {
				healths.Add(health.DOWN.String())
			}
			continue
		}
		// 将节点状态添加到缓存中
		if !healths.Contains(status) {
			healths.Add(status)
		}
	}

	if healths.Size() < 1 {
		r.Status = health.UNKNOWN.String()
		return r
	}
	if healths.Size() == 1 {
		r.Status = healths.Pop()
		return r
	}
	// 如果多条，则存在多个状态的进程节点
	up := false
	if healths.Contains(health.UP.String()) {
		up = true
	}
	if up {
		r.Status = health.WARNING.String()
		return r
	}
	if healths.Contains(health.WARNING.String()) {
		r.Status = health.WARNING.String()
		return r
	}
	if healths.Contains(health.DOWN.String()) {
		r.Status = health.DOWN.String()
		return r
	}
	if healths.Contains(health.STOP.String()) {
		r.Status = health.STOP.String()
		return r
	}
	if healths.Contains(health.PENDING.String()) {
		r.Status = health.PENDING.String()
		return r
	}
	if healths.Contains(health.UNKNOWN.String()) {
		r.Status = health.UNKNOWN.String()
		return r
	}
	r.Status = health.UNKNOWN.String()
	return r
}

type Node struct {
	Env        string                 `json:"env"`
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	Time       int64                  `json:"time"`
	Indicators map[string]interface{} `json:"indicators"`
}
