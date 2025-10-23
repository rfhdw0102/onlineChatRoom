package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io"
	"log"
	"net"
	"onlineChatRoom/db"
	"onlineChatRoom/msg"
	"onlineChatRoom/utils"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("server main panic recovered: %v\n", err)
		}
	}()
	dbErr := db.ConnectDb()
	if dbErr != nil {
		log.Fatal(dbErr)
	}
	defer func(DB *sqlx.DB) {
		err := DB.Close()
		if err != nil {
			log.Println("数据库连接关闭失败..")
		}
	}(db.DB)
	room := msg.NewChatRoom()
	go room.HandleMessages()
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
		if err != nil {
			if errors.Is(err, io.EOF) {
				//log.Printf("正常退出")
			} else {
				log.Printf("%s 非正常退出", username)
				cleanupConnection(conn, room, username)
			}
			return
		}
		username = message.Sender
		message.Conn = conn
		// 发给消息channel，让其处理
		room.MsgChan <- message
	}
}

// 清理连接函数
func cleanupConnection(conn net.Conn, room *msg.ChatRoom, username string) {
	room.Mutex.Lock()
	delete(room.Clients, username)
	room.Mutex.Unlock()
	if username != "" {
		leaveMsg := &msg.Message{Type: msg.MessageChat, Sender: "系统广播", Content: fmt.Sprintf("用户 %s 离开了聊天室", username)}
		room.MsgChan <- leaveMsg
	}
	utils.CloseConn(conn, username)
}
