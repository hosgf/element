package registry

import (
	"math"
	_ "sync/atomic"
	"time"

	"github.com/go-netty/go-netty"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/hosgf/element/logger"
)

type (
	triggerHandler struct {
		sh        SendDataHandler
		mh        MessageHandler
		client    *Client
		closeChan chan struct{}
	}

	SendDataHandler interface {
		SendPingData() Message
		SendData(data string) error
	}

	MessageHandler interface {
		HandleReplyPingData(data string)
		HandleReplyData(data string)
	}
)

func newTriggerHandler(client *Client) *triggerHandler {
	config := client.config
	return &triggerHandler{
		sh:        config.SendDataHandler,
		mh:        config.MessageHandler,
		client:    client,
		closeChan: make(chan struct{}),
	}
}

func (h *triggerHandler) SendPingData() Message {
	return h.sh.SendPingData()
}

func (h *triggerHandler) SendData(data string) error {
	return h.client.SendData(data)
}

func (h *triggerHandler) HandleActive(ctx netty.ActiveContext) {
	go h.ping(ctx)
	ctx.HandleActive()
}

func (h *triggerHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	h.stop()
	go h.retries(ctx)
}

func (h *triggerHandler) HandleException(ctx netty.ExceptionContext, ex netty.Exception) {
	logger.Warningf(ctx.Channel().Context(), "Lost the TCP connection with the server.")
	ctx.Channel().Close(ex)
}

func (h *triggerHandler) ping(ctx netty.ActiveContext) {
	for {
		ticker := time.NewTicker(h.nextTime())
		defer ticker.Stop()
		select {
		case <-h.closeChan:
			logger.Debugf(ctx.Channel().Context(), "The channel is closed.！！！！！")
			return
		case <-ticker.C:
			if ctx.Channel().IsActive() {
				logger.Debugf(ctx.Channel().Context(), "Send heartbeat request to start execution")
				ctx.Write(h.SendPingData())
			}
		}
	}
}

func (h *triggerHandler) stop() {
	close(h.closeChan)
}

func (h *triggerHandler) retries(ctx netty.InactiveContext) {
	for {
		nextTime := h.nextTime()
		ticker := time.NewTicker(nextTime)
		logger.Info(ctx.Channel().Context(), "正在尝试重新连接...")
		defer ticker.Stop()
		select {
		case <-ticker.C:
			if err := h.client.Run(); err != nil {
				logger.Warningf(ctx.Channel().Context(), "注册服务重连失败: %v，等待 %v 秒后重试...\n", err, nextTime.Seconds())
			} else {
				logger.Info(ctx.Channel().Context(), "注册服务重连成功")
				return
			}
		}
	}
}

func (h *triggerHandler) nextTime() time.Duration {
	second := math.Max(5, float64(grand.Intn(BaseRandom)))
	return time.Duration(second) * time.Second
}
