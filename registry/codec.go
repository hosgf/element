package registry

import (
	"strings"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/utils"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/logger"
)

func newMessageCodec(client *Client) messageCodec {
	return messageCodec{mh: client.config.MessageHandler}
}

type messageCodec struct {
	mh MessageHandler
}

func (m messageCodec) CodecName() string {
	return "message-codec"
}

func (m messageCodec) HandleRead(ctx netty.InboundContext, message netty.Message) {
	textBytes, err := utils.ToBytes(message)
	if err != nil {
		logger.Errorf(ctx.Channel().Context(), "Reader Message error: %s", err)
		return
	}
	sb := strings.Builder{}
	sb.Write(textBytes)
	data := sb.String()
	strs := strings.Split(data, "@@@")
	if len(strs) != 2 {
		logger.Errorf(ctx.Channel().Context(), "Reader Message failure")
		return
	}
	var obj Message
	obj.SetMessageHead(gconv.Bytes(strs[0]))
	obj.SetMessageBody(gconv.Bytes(strs[1]))
	if m.mh == nil {
		ctx.HandleRead(obj)
		return
	}
	messageType := MessageType(obj.MessageType)
	switch messageType {
	case MessageTypeHB:
		m.mh.HandleReplyPingData(obj.bodyToString())
	case MessageTypeBIZ:
		m.mh.HandleReplyData(obj.bodyToString())
	}
}

func (m messageCodec) HandleWrite(ctx netty.OutboundContext, message netty.Message) {
	switch r := message.(type) {
	case Message:
		ctx.HandleWrite(r.ComposeFull())
	case *Message:
		ctx.HandleWrite(r.ComposeFull())
	case string:
		ctx.HandleWrite(r)
	default:
		ctx.HandleWrite(gjson.MustEncodeString(r))
	}
}
