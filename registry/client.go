package registry

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec/frame"
	"github.com/go-netty/go-netty/transport"
	"github.com/go-netty/go-netty/transport/tcp"
)

type Client struct {
	ctx       context.Context
	config    *ClientConfig
	bootstrap netty.Bootstrap
	channel   netty.Channel
}

func NewClient(ctx context.Context, config *ClientConfig) *Client {
	c := &Client{ctx: ctx, config: config}
	if !config.Enabled {
		return c
	}
	clientInitializer := func(channel netty.Channel) {
		pipeline := channel.Pipeline()
		pipeline.
			AddLast(netty.ReadIdleHandler(time.Second), netty.WriteIdleHandler(4*time.Second)).
			AddLast(frame.LengthFieldCodec(binary.BigEndian, 0x7fffffff, 0, 4, 0, 4)).
			AddLast(newMessageCodec(c)).
			AddLast(newTriggerHandler(c))
		if config.Handler != nil {
			pipeline.AddLast(&config.Handler)
		}
	}
	c.bootstrap = netty.NewBootstrap(netty.WithClientInitializer(clientInitializer), netty.WithTransport(tcp.New()))
	return c
}

func (c *Client) Run() error {
	ch, err := c.bootstrap.Connect(c.config.Address, transport.WithContext(c.ctx), transport.WithAttachment(c.config.Name))
	if err != nil {
		return err
	}
	c.channel = ch
	return nil
}

func (c *Client) SendData(data string) error {
	if c.channel.IsActive() {
		message := NewBizMessage(data)
		return c.write(&message)
	}
	return nil
}

func (c *Client) write(data netty.Message) error {
	return c.channel.Write(data)
}
