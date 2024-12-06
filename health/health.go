package health

import (
	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/progress"
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

func GetProgressHealth(list []progress.Progress) string {
	if nil == list || len(list) < 1 {
		return UNKNOWN.String()
	}
	healths := gset.NewStrSet()
	for _, p := range list {
		healths.Add(p.Status)
	}
	if healths.Size() < 1 {
		return UNKNOWN.String()
	}
	if healths.Size() == 1 {
		return gconv.String(healths.Pop())
	}
	up := false
	if healths.ContainsI(UP.String()) {
		up = true
	}
	if up {
		return WARNING.String()
	}
	if healths.ContainsI(WARNING.String()) {
		return WARNING.String()
	}
	if healths.ContainsI(DOWN.String()) {
		return DOWN.String()
	}
	if healths.ContainsI(READ_ONLY.String()) {
		return READ_ONLY.String()
	}
	if healths.ContainsI(PENDING.String()) {
		return PENDING.String()
	}
	if healths.ContainsI(UNKNOWN.String()) {
		return UNKNOWN.String()
	}
	return UNKNOWN.String()
}

func GetHealth(states []string) string {
	if nil == states || len(states) < 1 {
		return UNKNOWN.String()
	}
	healths := gset.NewStrSet()
	for _, p := range states {
		healths.Add(p)
	}
	if healths.Size() < 1 {
		return UNKNOWN.String()
	}
	if healths.Size() == 1 {
		return gconv.String(healths.Pop())
	}
	up := false
	if healths.ContainsI(UP.String()) {
		up = true
	}
	if up {
		return WARNING.String()
	}
	if healths.ContainsI(WARNING.String()) {
		return WARNING.String()
	}
	if healths.ContainsI(DOWN.String()) {
		return DOWN.String()
	}
	if healths.ContainsI(READ_ONLY.String()) {
		return READ_ONLY.String()
	}
	if healths.ContainsI(PENDING.String()) {
		return PENDING.String()
	}
	if healths.ContainsI(UNKNOWN.String()) {
		return UNKNOWN.String()
	}
	return UNKNOWN.String()
}
