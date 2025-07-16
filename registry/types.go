package registry

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"

	"github.com/go-netty/go-netty"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/health"
	"github.com/hosgf/element/logger"
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
	Encoding     = "UTF-8"
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
	msg.SetMessageBodyData(body)
	return msg
}

func ParseFromBytes(data []byte) Message {
	msg := Message{}
	if len(data) < HeadLength {
		logger.Errorf(context.Background(), "数据长度不足: 需要至少%d字节，实际%d字节", HeadLength, len(data))
		return msg
	}

	// 智能处理消息头长度 - 支持17字节和18字节
	var headEndPos int
	if len(data) >= HeadLength+1 {
		// 如果数据足够长，检查是否需要额外处理
		// 先尝试标准17字节
		testHead := data[:HeadLength]
		testMagic := int32(testHead[0])<<24 | int32(testHead[1])<<16 | int32(testHead[2])<<8 | int32(testHead[3])

		if testMagic == MagicNumber {
			// 魔幻数在开头，使用标准17字节
			msg.MessageHead = testHead
			headEndPos = HeadLength
		} else {
			// 尝试18字节格式（可能第一个字节是额外的）
			if len(data) >= HeadLength+1 {
				testHead18 := data[1 : HeadLength+1]
				testMagic18 := int32(testHead18[0])<<24 | int32(testHead18[1])<<16 | int32(testHead18[2])<<8 | int32(testHead18[3])
				if testMagic18 == MagicNumber {
					// 魔幻数在偏移1位置，使用这17字节
					msg.MessageHead = testHead18
					headEndPos = HeadLength + 1
				} else {
					// 都不对，尝试查找魔幻数
					found := false
					for offset := 0; offset <= len(data)-HeadLength; offset++ {
						if offset+HeadLength <= len(data) {
							testData := data[offset : offset+HeadLength]
							magic := int32(testData[0])<<24 | int32(testData[1])<<16 | int32(testData[2])<<8 | int32(testData[3])
							if magic == MagicNumber {
								msg.MessageHead = testData
								headEndPos = offset + HeadLength
								found = true
								break
							}
						}
					}
					if !found {
						// 如果找不到魔幻数，还是使用前17字节，让parseHead处理错误
						msg.MessageHead = data[:HeadLength]
						headEndPos = HeadLength
					}
				}
			} else {
				// 数据不够18字节，使用全部数据作为消息头
				msg.MessageHead = data
				headEndPos = len(data)
			}
		}
	} else {
		// 数据刚好或少于17字节
		msg.MessageHead = data[:HeadLength]
		headEndPos = HeadLength
	}

	// 解析消息头
	if err := msg.parseHead(); err != nil {
		logger.Errorf(context.Background(), "解析消息头失败: %v", err)
		logger.Errorf(context.Background(), "原始数据长度: %d, 十六进制: %X", len(data), data)
	}

	// 设置消息体
	if len(data) > headEndPos {
		msg.MessageBody = data[headEndPos:]
		// 更新长度字段
		if msg.MessageBody != nil {
			msg.Length = int32(len(msg.MessageBody))
		}
	}

	return msg
}

func (m *Message) parseHead() error {
	if m.MessageHead == nil {
		return fmt.Errorf("message head is nil")
	}

	var headData []byte

	// 处理长度兼容性问题 - 支持17字节和18字节
	if len(m.MessageHead) == HeadLength {
		// 标准17字节格式
		headData = m.MessageHead
	} else if len(m.MessageHead) == HeadLength+1 {
		// 18字节格式 - 可能是网络传输或编码导致的额外字节
		// 通常额外字节在末尾，截取前17字节
		headData = m.MessageHead[:HeadLength]
	} else if len(m.MessageHead) > HeadLength {
		// 超过17字节，尝试找到包含魔幻数的17字节段
		found := false
		for offset := 0; offset <= len(m.MessageHead)-HeadLength; offset++ {
			testData := m.MessageHead[offset : offset+HeadLength]
			// 检查魔幻数 (前4字节，大端格式)
			if len(testData) >= 4 {
				testMagic := int32(testData[0])<<24 | int32(testData[1])<<16 | int32(testData[2])<<8 | int32(testData[3])
				if testMagic == MagicNumber {
					headData = testData
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("无法在%d字节数据中找到有效的消息头(魔幻数: %d)", len(m.MessageHead), MagicNumber)
		}
	} else {
		return fmt.Errorf("消息头长度不足: 需要至少%d字节，实际%d字节", HeadLength, len(m.MessageHead))
	}

	// 解析前16字节的4个int32字段 (大端格式)
	buf := bytes.NewReader(headData[:16])
	if err := binary.Read(buf, binary.BigEndian, &m.MagicNumber); err != nil {
		return fmt.Errorf("读取MagicNumber失败: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &m.Length); err != nil {
		return fmt.Errorf("读取Length失败: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &m.MessageType); err != nil {
		return fmt.Errorf("读取MessageType失败: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &m.LogId); err != nil {
		return fmt.Errorf("读取LogId失败: %w", err)
	}

	// 第17字节是flag
	m.Flag = headData[HeadLength-1]

	// 更新MessageHead为标准17字节格式
	m.MessageHead = headData

	// 验证魔幻数
	if m.MagicNumber != MagicNumber {
		return fmt.Errorf("魔幻数不匹配: 期望%d，实际%d", MagicNumber, m.MagicNumber)
	}

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

func (m *Message) GetMessageHead() ([]byte, error) {
	if m.MessageHead == nil || len(m.MessageHead) != HeadLength {
		err := m.composeHead()
		if err != nil {
			return nil, err
		}
	}
	return m.MessageHead, nil
}

func (m *Message) GetMessageBody() []byte {
	return m.MessageBody
}

func (m *Message) SetMessageHead(heads []byte) {
	if heads == nil {
		logger.Errorf(context.Background(), "SetMessageHead: 接收到空的消息头数据")
		return
	}
	m.MessageHead = heads
	if err := m.parseHead(); err != nil {
		logger.Errorf(context.Background(), "SetMessageHead: 解析消息头失败: %v", err)
		logger.Errorf(context.Background(), "SetMessageHead: 原始数据十六进制: %X", heads)
	}
}

func (m *Message) SetMessageHeadData(heads interface{}) {
	if heads == nil {
		return
	}
	m.SetMessageHead(gconv.Bytes(heads))
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
	m.SetMessageBody(gconv.Bytes(body))
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

func (m *Message) ComposeFull() string {
	h, err := m.GetMessageHead()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s%s%s", string(h), Delimiter, m.bodyToString())
}
