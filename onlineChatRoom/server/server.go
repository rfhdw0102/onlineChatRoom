package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"onlineChatRoom/db"
	"onlineChatRoom/msg"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("server main panic recovered: %v\n", err)
		}
		rr := db.DB.Close()
		if rr != nil {
			log.Println("MySQL连接关闭失败..")
		}
		r := db.RDB.Close()
		if r != nil {
			log.Println("Redis连接关闭失败..")
		}
	}()
	room := msg.NewChatRoom()
	// 连接MySQL
	dbErr := db.ConnectDb()
	if dbErr != nil {
		log.Fatal(dbErr)
	}
	// 连接Redis
	RedisErr := db.InitRedis()
	if RedisErr != nil {
		log.Fatal(RedisErr)
	}
	// 清理Redis数据
	db.ClearRedis()
	go room.HandleStreams()
	go room.HandleMessages()
	go room.StartHeartbeatMonitor()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("server start failed:", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Println("close listener failed:", err)
		}
	}(listener)
	fmt.Println("聊天室已创建，等待客户端连接...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(conn, " connect server failed:", err)
			continue
		}
		//处理客户端
		go handleClientMessage(conn, room)
	}
}

// 处理客户端
func handleClientMessage(conn net.Conn, room *msg.ChatRoom) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("server handleClientMessage panic recovered: %v\n", r)
		}
	}()
	reader := bufio.NewReader(conn)

	username := handleInit(reader, conn, room)
	if username == "" {
		return
	}
	handleMsg(username, reader, conn, room)
}

// handleInit 处理登录注册的消息
func handleInit(reader *bufio.Reader, conn net.Conn, room *msg.ChatRoom) (username string) {
	for {
		initMsg, err := msg.ReadJsonMessage(reader)
		if err != nil {
			log.Printf("%s 在登录注册时失败", conn.RemoteAddr().String())
			return ""
		}
		initMsg.Conn = conn
		var status bool
		switch initMsg.Type {
		case msg.MessageRegister:
			msg.Register(initMsg)
		case msg.MessageJoin:
			status = room.Join(initMsg)
			username = initMsg.Sender
		default:
		}
		if status {
			break
		} else {
			fmt.Println("登录失败..")
		}
	}
	return username
}

// handleMsg 处理登录注册之后的信息
func handleMsg(username string, reader *bufio.Reader, conn net.Conn, room *msg.ChatRoom) {
	for {
		message, err := msg.ReadJsonMessage(reader)
		if message != nil {
			username = message.Sender
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				//log.Printf("正常退出")
			} else {
				log.Printf("%s 被kill或者异常断开...\n", username)
				leaveMsg := &msg.Message{
					Type:   msg.MessageLeave,
					Sender: username,
				}
				room.MsgChan <- leaveMsg
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
