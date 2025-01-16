package registry

import (
	"encoding/json"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/utils"
	"github.com/gogf/gf/v2/encoding/gjson"
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
	var obj Message
	jsonDecoder := json.NewDecoder(utils.MustToReader(message))
	if err := jsonDecoder.Decode(&obj); err != nil {
		logger.Errorf(ctx.Channel().Context(), "decode error: %s", err)
		return
	}
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
		ctx.HandleWrite(r.ToString())
	case *Message:
		ctx.HandleWrite(r.ToString())
	case string:
		ctx.HandleWrite(r)
	default:
		ctx.HandleWrite(gjson.MustEncodeString(r))
	}
}
