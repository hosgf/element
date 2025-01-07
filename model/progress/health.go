package progress

import (
	"strings"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/hosgf/element/health"
)

// GroupHealth  健康检查
type GroupHealth struct {
	Region    string        `json:"region,omitempty"`
	Namespace string        `json:"namespace,omitempty"`
	Group     string        `json:"group,omitempty"`
	Status    health.Health `json:"status,omitempty"`
	Time      int64         `json:"time,omitempty"`
	Details   []Health      `json:"details,omitempty"`
}

// Health 健康检查
type Health struct {
	Region    string          `json:"region,omitempty"`
	Namespace string          `json:"namespace,omitempty"`
	PID       string          `json:"pid,omitempty"`
	Service   string          `json:"service,omitempty"`
	Name      string          `json:"name,omitempty"`
	Group     string          `json:"group,omitempty"`
	Address   string          `json:"address,omitempty"`
	Ports     []*ProgressPort `json:"ports,omitempty"`
	Status    health.Health   `json:"status,omitempty"`
	Time      int64           `json:"time,omitempty"`
}

func (p *Health) MatchGroup(group string) bool {
	return strings.EqualFold(p.Group, group)
}

func GetHealth(ps []Progress) health.Health {
	if nil == ps || len(ps) < 1 {
		return health.UNKNOWN
	}
	healths := gset.NewSet()
	for _, p := range ps {
		healths.Add(p.Status)
	}
	if healths.Size() < 1 {
		return health.UNKNOWN
	}
	if healths.Size() == 1 {
		return healths.Pop().(health.Health)
	}
	up := false
	if healths.Contains(health.UP) {
		up = true
	}
	if up {
		return health.WARNING
	}
	if healths.Contains(health.WARNING.String()) {
		return health.WARNING
	}
	if healths.Contains(health.DOWN) {
		return health.DOWN
	}
	if healths.Contains(health.READ_ONLY) {
		return health.READ_ONLY
	}
	if healths.Contains(health.PENDING) {
		return health.PENDING
	}
	if healths.Contains(health.UNKNOWN) {
		return health.UNKNOWN
	}
	return health.UNKNOWN
}
