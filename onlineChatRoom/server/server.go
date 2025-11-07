package main

import (
	"fmt"
	"log"
	"net"
	"onlineChatRoom/db"
	"onlineChatRoom/msg"
	"onlineChatRoom/server/tool"
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
	go room.HandleChanMessages()
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
		go tool.HandleClientMessage(conn, room)
	}
}
