package msg

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"onlineChatRoom/utils"
	"strings"
)

type MessageType int //发送类型的种类

const (
	MessageJoin    MessageType = iota //用户加入
	MessageLeave                      //用户离开
	MessageChat                       //公共聊天
	MessagePrivate                    //私聊
	MessageList                       //查看在线用户列表
	MessageHeart                      //心跳检测
)

type Message struct {
	Type     MessageType // 消息类型
	Sender   string      // 发送者
	Receiver string      // 接收者
	Content  string      // 内容
	Conn     net.Conn    // 发送者连接
}

func (msg *Message) JsonMessage() ([]byte, error) {
	return json.Marshal(msg)
}
func UnJsonMessage(msg []byte) (*Message, error) {
	var message Message
	err := json.Unmarshal(msg, &message)
	return &message, err
}
func ReadJsonMessage(reader *bufio.Reader) (*Message, error) {
	message, err := utils.ReadMessage(reader)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		if strings.Contains(err.Error(), "forcibly closed") {
			return nil, err
		}
		log.Printf("ReadJsonMessage失败: %v", err)
		return nil, err
	}
	return UnJsonMessage(message)
}

func SendJsonMessage(conn net.Conn, message *Message) error {
	jsonMessage, err := message.JsonMessage()
	if err != nil {
		log.Println("SendJsonMessage...")
		return err
	}
	return utils.SendMessage(conn, jsonMessage)
}
