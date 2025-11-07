package tool

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"onlineChatRoom/db"
	"onlineChatRoom/msg"
)

// HandleClientMessage 处理客户端
func HandleClientMessage(conn net.Conn, room *msg.ChatRoom) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("server handleClientMessage panic recovered: %v\n", r)
		}
	}()
	reader := bufio.NewReader(conn)

	username := handleRegisterOrLogin(reader, conn, room)
	if username == "" {
		return
	}
	handleCommonMsg(username, reader, conn, room)
}

// handleRegisterOrLogin 处理登录注册的消息
func handleRegisterOrLogin(reader *bufio.Reader, conn net.Conn, room *msg.ChatRoom) (username string) {
	for {
		initMsg, err := msg.ReadJsonMessage(reader)
		if err != nil {
			log.Printf("%s 在登录注册时失败", conn.RemoteAddr().String())
			return ""
		}
		initMsg.Conn = conn
		switch initMsg.Type {
		case msg.MessageRegister:
			msg.Register(initMsg)
			continue
		case msg.MessageJoin:
			status := room.Join(initMsg)
			if status {
				username = initMsg.Sender
				return username
			}
		default:
		}
	}

}

// handleCommonMsg 处理登录注册之后的信息
func handleCommonMsg(username string, reader *bufio.Reader, conn net.Conn, room *msg.ChatRoom) {
	for {
		message, err := msg.ReadJsonMessage(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				//log.Printf("正常退出")
			} else {
				log.Printf("%s 被kill或者异常断开...\n", username)
				room.Leave(username)
			}
			return
		}
		message.Conn = conn
		switch message.Type {
		case msg.MessageLeave, msg.MessageList, msg.MessageRank, msg.MessageHeart:
			room.MsgChan <- message
		default:
			// 聊天消息才异步入 Redis Streams
			_, err = db.AddStreamsData(message.Sender, message.Content, message.Receiver)
			if err != nil {
				log.Println("写入 Redis Streams 失败:", err)
			}
		}
	}
}
