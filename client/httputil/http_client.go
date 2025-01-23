package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
	"github.com/hosgf/element/util"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	DEFAULT_RETRY_COUNT    = 3
	DEFAULT_RETRY_INTERVAL = 3
	DEFAULT_TIMEOUT        = 3
	DefaultContentType     = "application/json"
)

func DoGetData(ctx context.Context, url string, isDebug bool) (data interface{}, err error) {
	response, err := DoGet(ctx, url, isDebug)
	if err != nil {
		return nil, err
	}
	if response.Code != 200 {
		return nil, errors.New(response.Message)
	}
	return response.Data, err
}

func DoGet(ctx context.Context, url string, isDebug bool) (response result.Response, err error) {
	return doGet(ctx, url, isDebug)
}

func DoPostJson(ctx context.Context, url string, data interface{}) (response result.Response, err error) {
	res := NewJsonHttpClient(ctx, DEFAULT_TIMEOUT, -1).PostContent(ctx, url, data)
	logger.Call(ctx, http.MethodPost, url, DefaultContentType, nil, res, data)
	if len(res) < 1 {
		return response, errors.New(result.REQ_REJECT.Message())
	}
	err = gconv.Struct(res, &response)
	return response, err
}

func DoPost(ctx context.Context, url string, contentType string, data interface{}) (response result.Response, err error) {
	res := NewHttpClient(ctx, DEFAULT_TIMEOUT, -1).ContentType(contentType).PostContent(ctx, url, data)
	logger.Call(ctx, http.MethodPost, url, contentType, nil, res, data)
	if len(res) < 1 {
		return response, errors.New(result.REQ_REJECT.Message())
	}
	err = gconv.Struct(res, &response)
	return response, err
}

func doGet(ctx context.Context, url string, isDebug bool) (response result.Response, err error) {
	res := NewHttpClient(ctx, DEFAULT_TIMEOUT, DEFAULT_RETRY_INTERVAL).GetContent(ctx, url)
	if isDebug {
		logger.Call(ctx, http.MethodGet, url, "application/json", nil, res, nil)
	}
	if len(res) < 1 {
		return response, errors.New(result.REQ_REJECT.Message())
	}
	err = gconv.Struct(res, &response)
	return response, err
}

func DoRequest(ctx context.Context, method, url string, data, resp interface{}, timeout int, isRetry, isJson, isDebug bool) error {
	client := NewHttpClient(ctx, util.AnyInt(timeout < 1, DEFAULT_TIMEOUT, timeout), util.AnyInt(isRetry, DEFAULT_RETRY_INTERVAL, -1))
	if isJson {
		client = client.ContentJson()
	}
	response, err := client.DoRequest(ctx, method, url, data)
	if err != nil {
		glog.Errorf(ctx, "请求客户端失败 %+v,  %+v", response, err.Error())
		return err
	}
	defer func() {
		if err = response.Close(); err != nil {
			glog.Errorf(ctx, "请求客户端失败 %+v, %+v", response, err.Error())
		}
	}()
	r := response.ReadAll()
	if isDebug {
		logger.Call(ctx, method, url, "", nil, isDebug, data)
	}
	if response == nil {
		return gerror.NewCode(result.FAILURE, fmt.Sprintf("【%s】调用失败", url))
	}
	if err := json.Unmarshal(r, resp); err != nil {
		glog.Errorf(ctx, "参数转换异常 \n     %v \n     %s", response, err.Error())
		return gerror.NewCode(result.FAILURE, fmt.Sprintf("【%s】调用失败", url))
	}
	return nil
}

func NewJsonHttpClient(ctx context.Context, timeout int, retryInterval int) (client *gclient.Client) {
	return NewHttpClient(ctx, timeout, retryInterval).ContentJson()
}

func NewHttpClient(ctx context.Context, timeout int, retryInterval int) (client *gclient.Client) {
	client = g.Client().SetTimeout(time.Duration(timeout) * time.Second)
	if headers := request.GetHeader(ctx); headers != nil {
		client.Header(headers)
	}
	if retryInterval > 0 {
		client = client.Retry(DEFAULT_RETRY_COUNT, time.Duration(retryInterval)*time.Second)
	}
	return client
}
