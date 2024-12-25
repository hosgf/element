package util

import (
	"github.com/gogf/gf/v2/os/genv"
	"github.com/hosgf/element/config"
)

func GetHomePath() string {
	key := InitAppHomeKey("")
	homePath := genv.Get(key)
	if !homePath.IsEmpty() {
		return homePath.String()
	}
	return ""
}

func SetHome(data string) {
	key := InitAppHomeKey("")
	genv.Set(key, data)
}

func InitAppHomeKey(key string) string {
	return config.InitAppHomeKey(key)
}
