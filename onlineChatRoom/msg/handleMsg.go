package msg

import (
	"fmt"
	"log"
	"onlineChatRoom/db"
)

// HandleStreams 处理streams流消息
func (cr *ChatRoom) HandleStreams() {
	lastID := "0-0"
	for {
		messages, err := db.ReadStreams(1, lastID)
		if err != nil {
			log.Println("读取 streams 出错:", err)
			continue
		}
		if len(messages) == 0 {
			continue
		}
		for _, m := range messages {
			sender := m.Values["sender"].(string)
			receiver := m.Values["receiver"].(string)
			content := m.Values["content"].(string)
			// 系统广播分支
			if sender == "系统广播" {
				cr.broadcast(receiver, fmt.Sprintf("%s: %s", sender, content))
				lastID = m.ID
				continue
			}

			msg := &Message{
				Sender:   sender,
				Receiver: receiver,
				Content:  content,
				Type:     MessageChat,
			}
			// 如果 sender 在线，再附加 Conn
			if client, ok := cr.Clients[m.Values["sender"].(string)]; ok {
				msg.Conn = client.Conn
			}
			if msg.Receiver != "" {
				cr.PrivateChat(msg)
			} else {
				cr.broadcast(msg.Sender, fmt.Sprintf("%s: %s", msg.Sender, msg.Content))
			}
			_ = db.AddActivity(msg.Sender, 1)
			lastID = m.ID // 更新游标，防止重复读取
		}
	}
}

// HandleChanMessages 普通消息处理
func (cr *ChatRoom) HandleChanMessages() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("server room.HandleMessages panic recovered: %v\n", err)
		}
	}()
	for msg := range cr.MsgChan {
		switch msg.Type {
		case MessageHeart:
			cr.PongHeart(msg.Sender)
		case MessageRegister:
			Register(msg)
		case MessageList:
			cr.ShowClients(msg.Sender, msg.Conn)
		case MessageJoin:
			cr.Join(msg)
		case MessageLeave:
			cr.Leave(msg.Sender)
		case MessageRank:
			SendRank(msg.Sender, msg.Conn)
		default:
		}
	}
}
