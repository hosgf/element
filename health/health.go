package health

import (
	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/text/gstr"
)

type Health int

// Health
const (
	// UNKNOWN 未知的，不存在的
	UNKNOWN Health = iota
	// PENDING 启动中,未准备就绪
	PENDING
	// READ_ONLY 只读，不能进行调度
	READ_ONLY
	// DOWN 宕机,下线，有问题的
	DOWN
	// STOP 停止的，不可用的
	STOP
	// WARNING 告警的，基本可用的
	WARNING
	// UP 正常的，健康的
	UP
)

func (h Health) String() string {
	return [...]string{"UNKNOWN", "PENDING", "READ_ONLY", "DOWN", "STOP", "WARNING", "UP"}[h]
}

func IsUnknown(status string) bool {
	if len(status) < 1 {
		return true
	}
	return gstr.Equal(status, UNKNOWN.String())
}

func IsDown(status string) bool {
	if len(status) < 1 {
		return true
	}
	return gstr.Equal(status, DOWN.String())
}

func GetHealth(states []string) Health {
	if nil == states || len(states) < 1 {
		return UNKNOWN
	}
	healths := gset.NewSet()
	for _, p := range states {
		healths.Add(p)
	}
	if healths.Size() < 1 {
		return UNKNOWN
	}
	if healths.Size() == 1 {
		return healths.Pop().(Health)
	}
	up := false
	if healths.Contains(UP) {
		up = true
	}
	if up {
		return WARNING
	}
	if healths.Contains(WARNING) {
		return WARNING
	}
	if healths.Contains(DOWN) {
		return DOWN
	}
	if healths.Contains(READ_ONLY) {
		return READ_ONLY
	}
	if healths.Contains(PENDING) {
		return PENDING
	}
	if healths.Contains(UNKNOWN) {
		return UNKNOWN
	}
	return UNKNOWN
}
