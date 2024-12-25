package config

import (
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"
	"strings"
)

var (
	Version    = "1.0.0"
	AppCode    = ""
	AppHomeKey = ""
	AesKey     = "youedata12345678"
)

func Init(appCode, version, aesKey string) {
	if len(appCode) > 0 {
		AppCode = appCode
		AppHomeKey = InitAppHomeKey(appCode)
	}
	if len(version) > 0 {
		AppCode = version
	}
	if len(aesKey) > 0 {
		AesKey = aesKey
	}
}

func InitAppHomeKey(key string) string {
	if len(key) > 0 {
		newkey := strings.ToUpper(key)
		if gstr.HasSuffix(newkey, "_HOME") {
			AppHomeKey = newkey
			return AppHomeKey
		}
		AppHomeKey = newkey + "_HOME"
		return AppHomeKey
	}
	if len(AppHomeKey) > 0 {
		return AppHomeKey
	}
	if len(AppCode) > 0 {
		AppHomeKey = strings.ToUpper(AppCode) + "_HOME"
		return AppHomeKey
	}
	AppHomeKey = grand.Letters(10) + "_HOME"
	return AppHomeKey
}
