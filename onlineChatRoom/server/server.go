package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"onlineChatRoom/msg"
	"onlineChatRoom/utils"
	"strings"
	"time"
)

type Client struct {
	Username      string
	Conn          net.Conn
	LastHeartbeat time.Time
}

func main() {
	room := msg.NewChatRoom()
	go room.HandleMessages()
	//go startHeartbeatCheck(room)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("服务端启动失败:", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("监听关闭失败...")
			return
		}
	}(listener)

	fmt.Println("聊天室已创建，等待客户端连接...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("客户端连接失败:", err)
			continue
		}
		go handleClient(conn, room)
	}
}

// 处理客户端
func handleClient(conn net.Conn, room *msg.ChatRoom) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] handleClient 出现异常: %v", r)
		}
	}()

	reader := bufio.NewReader(conn)
	username := checkLogin(reader, conn, room)
	client := &Client{Username: username, Conn: conn, LastHeartbeat: time.Now()}

	room.AddClient(username, conn)
	room.MsgChan <- &msg.Message{Type: msg.MessageJoin, Sender: username}

	defer func() {
		room.RemoveClient(username)
		room.MsgChan <- &msg.Message{Type: msg.MessageLeave, Sender: username}
		utils.CloseConn(conn, username)
	}()
	for {
		message, err := msg.ReadJsonMessage(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				//log.Printf("%s 正常退出聊天室", username)
			} else if strings.Contains(err.Error(), "forcibly closed") {
				log.Printf("%s 连接被远程关闭", username)
			} else {
				log.Printf("%s 连接断开: %v", username, err)
			}
			return
		}
		message.Conn = conn
		handleMessage(message, client, room)
	}
}

// 验证用户名唯一
func checkLogin(reader *bufio.Reader, conn net.Conn, room *msg.ChatRoom) string {
	for {
		loginMes, err := msg.ReadJsonMessage(reader)
		if err != nil {
			log.Println("读取用户名失败:", err)
			continue
		}
		username := strings.TrimSpace(loginMes.Sender)
		if username == "" {
			_ = msg.SendJsonMessage(conn, &msg.Message{
				Type:    msg.MessageChat,
				Sender:  "[系统]",
				Content: "用户名不能为空",
			})
			continue
		}
		room.Mutex.Lock()
		_, exists := room.Clients[username]
		room.Mutex.Unlock()
		if exists {
			_ = msg.SendJsonMessage(conn, &msg.Message{
				Type:    msg.MessageChat,
				Sender:  "[系统]",
				Content: "用户名已存在，请重新输入",
			})
			continue
		}
		_ = msg.SendJsonMessage(conn, &msg.Message{Content: "OK"})
		return username
	}
}

// 处理消息
func handleMessage(m *msg.Message, client *Client, room *msg.ChatRoom) {
	switch m.Type {
	case msg.MessageHeart:
		client.LastHeartbeat = time.Now()
		err := client.Conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		if err != nil {
			log.Printf("handleMessage设置超时时间失败: %v", err)
		}
		_ = msg.SendJsonMessage(m.Conn, &msg.Message{
			Type:    msg.MessageHeart,
			Content: "PONG",
		})
	case msg.MessageChat:
		room.MsgChan <- m
	case msg.MessagePrivate:
		room.PrivateChat(m)
	case msg.MessageList:
		room.ShowClients(m.Sender, m.Conn)
	default:
	}
}

//// 心跳检测
//func startHeartbeatCheck(room *msg.ChatRoom) {
//	ticker := time.NewTicker(15 * time.Second)
//	defer ticker.Stop()
//	for range ticker.C {
//		now := time.Now()
//		room.Mutex.Lock()
//		for username, conn := range room.Clients {
//			err := conn.SetReadDeadline(now.Add(10 * time.Second))
//			if err != nil {
//				log.Printf("%s 设置超时时间失败: %v", username, err)
//			}
//		}
//		room.Mutex.Unlock()
//	}
//}
