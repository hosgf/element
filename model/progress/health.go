package progress

import (
	"github.com/gogf/gf/v2/container/gset"
	"github.com/hosgf/element/health"
)

func GetHealth(progresss []*Progress) health.Health {
	if nil == progresss || len(progresss) < 1 {
		return health.UNKNOWN
	}
	healths := gset.NewSet()
	for _, p := range progresss {
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
