package main

import (
	"awesomeProject/onlineChatRoom/utils"
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Client 表示一个客户端连接
type Client struct {
	Username      string   // 客户端的网名
	Conn          net.Conn // 客户端的连接
	LastHeartbeat time.Time // 最近一次心跳时间
}

type Message struct {
	Content string
	conn    net.Conn
}

// ChatRoom 聊天室结构体
type ChatRoom struct {
	Client  map[*Client]bool // 在线的客户端列表
	JoinCh  chan *Client     // 加入聊天室通道
	LeaveCh chan *Client     // 离开聊天室通道
	Message chan *Message    // 正常聊天通道
	mutex   sync.Mutex       // 互斥锁
}

// NewChatRoom 初始化聊天室
func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		Client:  make(map[*Client]bool),
		JoinCh:  make(chan *Client),
		LeaveCh: make(chan *Client),
		Message: make(chan *Message),
	}
}

// 广播
func (cr *ChatRoom) broadcast(conn net.Conn, message []byte) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	for client := range cr.Client {
		if client.Conn != conn {
			err := utils.SendMessage(client.Conn, message)
			if err != nil {
				log.Printf("广播给：%s 失败...", client.Username)
				continue
			}
		}
	}
}

// 显示在线用户列表
func (cr *ChatRoom) showClient(conn net.Conn) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	message := "在线用户列表: "
	for client := range cr.Client {
		message = message + client.Username + "   "
	}
	for client := range cr.Client {
		if client.Conn == conn {
			err := utils.SendMessage(client.Conn, []byte(message))
			if err != nil {
				log.Printf("发送在线用户列表给：%s 失败...", client.Username)
				continue
			}
		}
	}
}

// 处理客户端事务
func (cr *ChatRoom) handleEvent() {
	for {
		select {
		case client := <-cr.JoinCh: // 用户加入聊天室
			cr.mutex.Lock()
			cr.Client[client] = true
			cr.mutex.Unlock()
			message := fmt.Sprintf("系统广播：%s 加入了聊天室...", client.Username)
			fmt.Printf("系统广播：%s 加入了聊天室...\n", client.Username)
			cr.broadcast(client.Conn, []byte(message))
		case client := <-cr.LeaveCh: // 用户离开聊天室
			cr.mutex.Lock()
			delete(cr.Client, client)
			cr.mutex.Unlock()
			message := fmt.Sprintf("系统广播：%s 离开了聊天室...", client.Username)
			fmt.Printf("系统广播：%s 离开了聊天室...\n", client.Username)
			cr.broadcast(client.Conn, []byte(message))
		case message := <-cr.Message: // 聊天室公共信息
			fmt.Println(message.Content)
			cr.broadcast(message.conn, []byte(message.Content))
		}
	}
}

// 处理客户端连接
func (cr *ChatRoom) handleClient(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("关闭连接出错:", err)
		}
	}(conn)

	reader := bufio.NewReader(conn)
	var username string
	for {
		// 获取用户名
		message, err := utils.ReadMessage(reader)
		if err != nil {
			log.Println("接收客户网名失败:", err)
			return
		}
		username = string(message)
		// 检查用户名是否有效
		if username == "" {
			_ = utils.SendMessage(conn, []byte("Error : 网名不能为空"))
			continue
		}
		// 检查用户名是否已存在
		cr.mutex.Lock()
		usernameExists := false
		for client := range cr.Client {
			if username == client.Username {
				usernameExists = true
				break
			}
		}
		cr.mutex.Unlock()
		if usernameExists {
			_ = utils.SendMessage(conn, []byte("Error : 网名已存在"))
			continue
		}
		break
	}

	// 用户名有效，发送确认
	_ = utils.SendMessage(conn, []byte("OK"))

	// 创建客户端并加入聊天室
	client := &Client{
		Username:      username,
		Conn:          conn,
		LastHeartbeat: time.Now(),
	}
	cr.JoinCh <- client
	defer func() {
		cr.LeaveCh <- client
	}()

	// 处理客户端消息
	for {
		message, err := utils.ReadMessage(reader)
		if err != nil {
			if strings.Contains(err.Error(), "forcibly closed by the remote host") {
				fmt.Println(username + " 已断开连接...")
				break
			} else {
				log.Printf("读取 %s 消息出错: %v", username, err)
				return
			}
		}
		msgStr := string(message)

		// 心跳包处理
		if msgStr == "PING" {
			client.LastHeartbeat = time.Now()
			log.Printf("[心跳] 收到 %s 的 PING", client.Username)
			_ = utils.SendMessage(conn, []byte("PONG"))
			log.Printf("[心跳] 已向 %s 回复 PONG", client.Username)
			continue
		}

		if msgStr == "quit" {
			return
		}
		if msgStr == "list" {
			fmt.Println(username + " 请求查看在线用户列表...")
			cr.showClient(conn)
			continue
		}
		if strings.HasPrefix(msgStr, "To:") {
			strs := strings.Split(msgStr, "-->")
			usernameStr := strs[0][3:]
			if usernameStr == "" {
				_ = utils.SendMessage(conn, []byte("私聊用户名为空..."))
				continue
			}
			flag := false
			var ci *Client
			for client := range cr.Client {
				if client.Username == usernameStr {
					ci = client
					flag = true
					break
				}
			}
			if flag {
				msg := username + " 私聊你: " + strs[1]
				err := utils.SendMessage(ci.Conn, []byte(msg))
				fmt.Println(username + " 私聊 " + usernameStr + " : " + strs[1])
				if err != nil {
					log.Println("私聊发送消息失败...")
				}
			} else {
				_ = utils.SendMessage(conn, []byte("私聊用户名不存在..."))
			}
			continue
		}
		msg := &Message{
			Content: username + ":" + msgStr,
			conn:    conn,
		}
		cr.Message <- msg
	}
}

func main() {
	chatRoom := NewChatRoom() // 初始化聊天室
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("服务端启动失败...")
	}
	defer listen.Close()

	fmt.Println("聊天室已创建...")
	// 处理客户端事务
	go safeHandleEvent(chatRoom) //安全包装
	go heart(chatRoom)           // 启动心跳检测

	// 接收客户端连接
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("连接客户端时出错...")
			continue
		}
		go chatRoom.safeHandleClient(conn) //安全保护
	}
}

// 心跳检测
func heart(cr *ChatRoom) {
	ticker := time.NewTicker(15 * time.Second) // 每 15 秒检查一次
	defer ticker.Stop()
	for {
		<-ticker.C
		now := time.Now()
		cr.mutex.Lock()
		for client := range cr.Client {
			if now.Sub(client.LastHeartbeat) > 30*time.Second { // 超过 30 秒没心跳
				log.Printf("[心跳] 客户端 %s 心跳超时，断开连接", client.Username)
				client.Conn.Close()
				delete(cr.Client, client)
				cr.LeaveCh <- client
			}
		}
		cr.mutex.Unlock()
	}
}
func  safeHandleEvent(cr *ChatRoom) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[panic] handleEvent 出现异常: %v", err)
		}
	}()
	cr.handleEvent()
}
func (cr *ChatRoom) safeHandleClient(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[panic] handleClient 出现异常: %v", err)
		}
	}()
	cr.handleClient(conn)
}
