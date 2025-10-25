package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"io"
	"log"
	"net"
	"onlineChatRoom/db"
	"onlineChatRoom/msg"
)

func main() {
	// panic 捕获
	defer func() {
		if err := recover(); err != nil {
			log.Printf("server main panic recovered: %v\n", err)
		}
	}()
	room := msg.NewChatRoom()
	// 连接MySQL
	dbErr := db.ConnectDb()
	if dbErr != nil {
		log.Fatal(dbErr)
	}
	// 关闭MySQL连接
	defer func(DB *sqlx.DB) {
		err := DB.Close()
		if err != nil {
			log.Println("MySQL连接关闭失败..")
		}
	}(db.DB)
	// 连接Redis
	RedisErr := db.InitRedis()
	if RedisErr != nil {
		log.Fatal(RedisErr)
	}

	// 关闭Redis连接
	defer func(RDB *redis.Client) {
		err := RDB.Close()
		if err != nil {
			log.Println("Redis连接关闭失败..")
		}
	}(db.RDB)

	// 处理各种消息
	go room.HandleMessages()
	// 心跳定时检测
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
	var username string
	reader := bufio.NewReader(conn)
	for {
		message, err := msg.ReadJsonMessage(reader)
		if message != nil {
			username = message.Sender
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				//log.Printf("正常退出")
			} else {
				log.Printf("%s 非正常退出", username)
			}
			return
		}
		message.Conn = conn
		// 发给消息channel，让其处理
		room.MsgChan <- message
	}
}
