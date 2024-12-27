package logger

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

func Call(ctx context.Context, method string, url string, contentType string, headers interface{}, response interface{}, param interface{}) {
	var str = fmt.Sprintf("\n------> %s  %s\n", method, url)
	if headers != nil {
		str += fmt.Sprintf("Headers: %+v \n", headers)
	}
	if len(contentType) > 0 {
		str += fmt.Sprintf("ContentType: %s \n", contentType)
	}
	if param != nil {
		str += fmt.Sprintf("Params: %+v \n", param)
	}
	str += fmt.Sprintf("Response: %s \n", response)
	str += "------> END HTTP\n"
	Log().Debug(ctx, str)
}

func Debug(ctx context.Context, isDebug bool, method string, url string, contentType string, headers interface{}, response interface{}, param interface{}, err error) {
	if !isDebug {
		return
	}
	var str = fmt.Sprintf("\n------> %s  %s\n", method, url)
	if headers != nil {
		str += fmt.Sprintf("Headers: %+v \n", headers)
	}
	if len(contentType) > 0 {
		str += fmt.Sprintf("ContentType: %s \n", contentType)
	}
	if param != nil {
		str += fmt.Sprintf("Params: %+v \n", param)
	}
	str += fmt.Sprintf("Response: %s \n", response)
	if err != nil {
		str += fmt.Sprintf("Error: %+v \n", err)
	}
	str += "------> END HTTP\n"
	Log().Debug(ctx, str)
}

func Info(ctx context.Context, v ...interface{}) {
	Log().Info(ctx, v...)
}

func Error(ctx context.Context, message string, err error) {
	Log().Errorf(ctx, "\n -- %s ,err: %s", message, err.Error())
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	Log().Errorf(ctx, format, v...)
}

func Fatalf(ctx context.Context, format string, v ...interface{}) {
	Log().Fatalf(ctx, format, v...)
}

func Warningf(ctx context.Context, format string, v ...interface{}) {
	Log().Warningf(ctx, format, v...)
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	Log().Debugf(ctx, format, v...)
}

func SetLevelStr(levelStr string) error {
	return Log().SetLevelStr(levelStr)
}

func Log(name ...string) *glog.Logger {
	return g.Log(name...)
}
