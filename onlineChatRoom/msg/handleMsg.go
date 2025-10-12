package msg

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// 聊天室
type ChatRoom struct {
	Clients map[string]net.Conn
	MsgChan chan *Message
	Mutex   sync.Mutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		Clients: make(map[string]net.Conn),
		MsgChan: make(chan *Message, 100),
	}
}

func (cr *ChatRoom) AddClient(username string, conn net.Conn) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()
	cr.Clients[username] = conn
}

func (cr *ChatRoom) RemoveClient(username string) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()
	delete(cr.Clients, username)
}

// -------------------- 核心消息处理 --------------------

func (cr *ChatRoom) HandleMessages() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[panic] HandleMessages 出现异常: %v", err)
		}
	}()
	for msg := range cr.MsgChan {
		switch msg.Type {
		case MessageJoin:
			content := fmt.Sprintf("系统广播：%s 加入了聊天室...", msg.Sender)
			cr.broadcast("[系统]", content)
		case MessageLeave:
			content := fmt.Sprintf("系统广播：%s 离开了聊天室...", msg.Sender)
			cr.broadcast("[系统]", content)
		case MessageChat:
			cr.broadcast(msg.Sender, fmt.Sprintf("%s: %s", msg.Sender, msg.Content))
		default:
		}
	}
}

// 广播（仅系统消息与群聊）
func (cr *ChatRoom) broadcast(sender, content string) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()

	for username, conn := range cr.Clients {
		if username == sender {
			continue
		}
		_ = SendJsonMessage(conn, &Message{
			Type:    MessageChat,
			Sender:  sender,
			Content: content,
		})
	}
	fmt.Println(content)
}

// 私聊
func (cr *ChatRoom) PrivateChat(msg *Message) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()

	target, ok := cr.Clients[msg.Receiver]
	if !ok {
		_ = SendJsonMessage(msg.Conn, &Message{
			Type:    MessagePrivate,
			Sender:  "[系统]",
			Content: fmt.Sprintf("用户 %s 不存在或不在线", msg.Receiver),
		})
		return
	}
	_ = SendJsonMessage(target, &Message{
		Type:     MessagePrivate,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
		Content:  msg.Content,
	})
	fmt.Printf("%s 私聊 %s: %s\n", msg.Sender, msg.Receiver, msg.Content)
}

// 查询在线列表
func (cr *ChatRoom) ShowClients(name string, conn net.Conn) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()

	list := "在线用户列表: "
	for username := range cr.Clients {
		list += username + "  "
	}

	_ = SendJsonMessage(conn, &Message{
		Type:    MessageList,
		Content: list,
	})
	fmt.Println(name, "请求查看用户列表...")
}
