package registry

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/go-netty/go-netty"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/health"
	"golang.org/x/text/encoding/charmap"
)

type MessageType int32

const (
	MessageTypeBIZ MessageType = 1 // 业务报文
	MessageTypeHB  MessageType = 2 // 心跳报文
)

func (m MessageType) ToInt32() int32 {
	return int32(m)
}

const (
	HeadLength   = 17
	MagicNumber  = 16787777
	DefaultLogId = 10000000
	Encoding     = "GB18030"
	BaseRandom   = 20
)

type ClientConfig struct {
	Name            string          `json:"name"`
	Enabled         bool            `json:"enabled"`
	Address         string          `json:"address"`
	Retry           bool            `json:"retry"`
	MaxRetries      int             `json:"maxRetries"`
	Handler         netty.Handler   `json:"handler"`
	SendDataHandler SendDataHandler `json:"sendDataHandler"`
	MessageHandler  MessageHandler  `json:"messageHandler"`
}

type Message struct {
	MagicNumber int32  `json:"magicNumber" `
	Length      int32  `json:"length"`
	LogId       int32  `json:"logId"`
	Flag        byte   `json:"flag"`
	MessageType int32  `json:"messageType"`
	MessageHead []byte `json:"messageHead"`
	MessageBody []byte `json:"messageBody"`
}

func NewHeartBeatMessage() Message {
	body := []byte(health.UP)
	return Message{
		MagicNumber: MagicNumber,
		MessageType: MessageTypeHB.ToInt32(),
		LogId:       DefaultLogId,
		MessageBody: body,
		Length:      int32(len(body)),
	}
}

func NewMessage() Message {
	return Message{
		MagicNumber: MagicNumber,
		LogId:       DefaultLogId,
	}
}

func NewBizMessage(body interface{}) Message {
	msg := Message{
		MagicNumber: MagicNumber,
		MessageType: MessageTypeBIZ.ToInt32(),
		LogId:       DefaultLogId,
	}
	msg.SetMessageBody(gconv.Bytes(body))
	return msg
}

func ParseFromBytes(data []byte) Message {
	msg := Message{}
	if len(data) < HeadLength {
		return msg
	}
	msg.MessageHead = data[:HeadLength]
	msg.parseHead()
	if len(data) > HeadLength {
		msg.MessageBody = data[HeadLength:]
	}
	return msg
}

func (m *Message) parseHead() error {
	if (m.MessageHead == nil) || len(m.MessageHead) != HeadLength {
		return nil
	}
	buf := bytes.NewReader(m.MessageHead[:HeadLength-1]) // Exclude flag
	if err := binary.Read(buf, binary.BigEndian, &m.MagicNumber); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &m.Length); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &m.MessageType); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &m.LogId); err != nil {
		return err
	}
	m.Flag = m.MessageHead[HeadLength-1]
	return nil
}

func (m *Message) composeHead() error {
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, m.MagicNumber); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, m.Length); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, m.MessageType); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, m.LogId); err != nil {
		return err
	}
	m.MessageHead = buf.Bytes()
	m.MessageHead = append(m.MessageHead, m.Flag)
	return nil
}

func (m *Message) ComposeFull() ([]byte, error) {
	if m.MessageBody != nil {
		m.Length = int32(len(m.MessageBody))
	}
	err := m.composeHead()
	if err != nil {
		return nil, err
	}
	data := append(m.MessageHead, m.MessageBody...)
	return data, nil
}

func (m *Message) GetMessageHead() ([]byte, error) {
	if m.MessageHead == nil || len(m.MessageHead) != HeadLength {
		err := m.composeHead()
		if err != nil {
			return nil, err
		}
	}
	return m.MessageHead, nil
}

func (m *Message) SetMessageBody(body []byte) {
	m.MessageBody = body
	if body != nil {
		m.Length = int32(len(body))
	}
}

func (m *Message) SetMessageBodyData(body interface{}) {
	if body == nil {
		return
	}
	m.MessageBody = gconv.Bytes(body)
	m.Length = int32(len(m.MessageBody))
}

func (m *Message) bodyToString() string {
	if m.MessageBody == nil || len(m.MessageBody) == 0 {
		return ""
	}
	body := string(m.MessageBody)
	if Encoding != "UTF-8" {
		enc, _ := charmap.ISO8859_1.NewDecoder().Bytes(m.MessageBody)
		body = string(enc)
	}
	return body
}

func (m *Message) ToBytes() []byte {
	d, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return d
}

func (m *Message) ToString() string {
	m.composeHead()
	return gjson.MustEncodeString(m)
}
