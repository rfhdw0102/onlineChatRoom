package msg

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net"
	"onlineChatRoom/db"
	"onlineChatRoom/utils"
	"time"
)

// broadcast 广播（仅系统消息与群聊）
func (cr *ChatRoom) broadcast(sender, content string) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()

	for username, client := range cr.Clients {
		if username == sender {
			continue
		}
		err := SendJsonMessage(client.Conn, &Message{
			Type:    MessageChat,
			Sender:  sender,
			Content: content,
		})
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println(sender, "已退出聊天室...连接已关闭")
				return
			}
			log.Println("broadcast:", err)
			return
		}
	}
	fmt.Println(content)
}

// PrivateChat 私聊
func (cr *ChatRoom) PrivateChat(msg *Message) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()
	target, ok := cr.Clients[msg.Receiver]
	if !ok {
		_ = SendJsonMessage(msg.Conn, &Message{
			Type:    MessageChat,
			Sender:  "[系统]",
			Content: fmt.Sprintf("用户 %s 不存在或不在线", msg.Receiver),
		})
		return
	}
	err := SendJsonMessage(target.Conn, &Message{
		Type:    MessagePrivate,
		Sender:  msg.Sender,
		Content: msg.Content,
	})
	if err != nil {
		log.Println("PrivateChat:", err)
		return
	}
	fmt.Printf("%s 私聊 %s: %s\n", msg.Sender, msg.Receiver, msg.Content)
}

// ShowClients 查询在线列表
func (cr *ChatRoom) ShowClients(name string, conn net.Conn) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()

	list := "在线用户列表: "
	for username := range cr.Clients {
		list += username + "  "
	}

	err := SendJsonMessage(conn, &Message{
		Type:    MessageList,
		Content: list,
	})
	if err != nil {
		log.Println("ShowClients ", err)
	}
	fmt.Println(name, "请求查看用户列表...")
}

// Register 处理注册信息
func Register(msg *Message) {
	err := db.AddUserDb(msg.Sender, msg.Content)
	if err != nil {
		// 检查是否是唯一约束冲突（用户名已存在）
		if isDuplicateKeyError(err) {
			rr := SendJsonMessage(msg.Conn, &Message{
				Type:    MessageRegister,
				Content: "用户名: " + msg.Sender + " 已被注册",
			})
			if rr != nil {
				log.Println("Register send error:", rr)
			}
		} else {
			log.Println("注册失败:", err)
			rr := SendJsonMessage(msg.Conn, &Message{
				Type:    MessageRegister,
				Content: "注册失败，请稍后重试",
			})
			if rr != nil {
				log.Println("Register send error:", rr)
			}
		}
		return
	}
	// 注册成功
	rr := SendJsonMessage(msg.Conn, &Message{
		Type:    MessageRegister,
		Content: "OK",
	})
	if rr != nil {
		log.Println("Register send error:", rr)
	}
	fmt.Println(msg.Sender, "注册成功...")
}

// isDuplicateKeyError 辅助函数检查是否是唯一约束错误
func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 // 1062是MySQL的重复键问题
	}
	return false
}

// Join 处理登录消息
func (cr *ChatRoom) Join(msg *Message) bool {
	password, err := db.SearchUserDb(msg.Sender)
	// 查询失败的情况
	if err != nil {
		var respContent string
		if errors.Is(err, sql.ErrNoRows) {
			respContent = fmt.Sprintf("%s 不存在，请先注册", msg.Sender)
		} else {
			respContent = "登录失败，数据库异常"
			log.Printf("查询用户 %s 失败: %v", msg.Sender, err)
		}
		// 发送错误响应
		if r := SendJsonMessage(msg.Conn, &Message{
			Type:    MessageChat,
			Content: respContent,
		}); r != nil {
			log.Println("发送登录失败响应错误:", err)
		}
		return false
	}
	// 判断密码
	if password != msg.Content {
		if r := SendJsonMessage(msg.Conn, &Message{
			Type:    MessageChat,
			Content: "密码错误，请重新输入",
		}); r != nil {
			log.Println("发送密码错误响应错误:", err)
		}
		return false
	}

	if _, ok := cr.Clients[msg.Sender]; ok {
		if r := SendJsonMessage(msg.Conn, &Message{
			Type:    MessageChat,
			Content: "该账户已登录",
		}); r != nil {
			log.Println("发送账号已登陆响应错误:", err)
		}
		return false
	}
	// 登录成功
	rr := SendJsonMessage(msg.Conn, &Message{
		Type:    MessageRegister,
		Content: "OK",
	})
	if rr != nil {
		log.Println("Register send error:", rr)
		return false
	}
	client := &Client{Username: msg.Sender, Conn: msg.Conn, LastHeartbeat: time.Now()}
	cr.AddClient(msg.Sender, client)
	//content := fmt.Sprintf("系统广播：%s 加入了聊天室...", msg.Sender)
	//cr.broadcast(msg.Sender, content)
	// 发送历史消息
	historyMsg, rrr := db.ShowHistory()
	if rrr != nil {
		log.Println(rrr)
	}
	message := &Message{Type: MessageChat, Content: historyMsg}
	r := SendJsonMessage(msg.Conn, message)
	if r != nil {
		log.Println("发送历史消息失败:", r)
	}
	// 加入streams流
	_, err = db.AddStreamsData("系统广播", fmt.Sprintf("%s 加入了聊天室...", msg.Sender), msg.Sender)
	if err != nil {
		log.Println("写入 Redis Streams 失败:", err)
	}

	return true
}

// Leave 处理退出消息
func (cr *ChatRoom) Leave(msg *Message) {
	_, err := db.AddStreamsData("系统广播", fmt.Sprintf("%s 离开了聊天室...", msg.Sender), msg.Sender)
	if err != nil {
		log.Println("Leave写入 Redis Streams 失败:", err)
	}
	cr.RemoveClient(msg.Sender)
}

// PongHeart 处理心跳
func (cr *ChatRoom) PongHeart(msg *Message) {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()
	if client, exists := cr.Clients[msg.Sender]; exists {
		client.LastHeartbeat = time.Now()
		err := client.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			log.Printf("PongHeart: %v", err)
		}
	}
}

// StartHeartbeatMonitor 服务端定期检测客户端心跳超时
func (cr *ChatRoom) StartHeartbeatMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		now := time.Now()
		cr.Mutex.Lock()
		for username, client := range cr.Clients {
			if now.Sub(client.LastHeartbeat) > 20*time.Second {
				log.Printf("用户 %s 心跳超时，强制下线\n", username)
				utils.CloseConn(client.Conn, username)
				delete(cr.Clients, username)
				_, err := db.AddStreamsData("系统广播", fmt.Sprintf("%s 离开了聊天室...", username), username)
				if err != nil {
					log.Println("心跳写入 Redis Streams 失败:", err)
				}
			}
		}
		cr.Mutex.Unlock()
	}
}

// SendRank 发送活跃度排行
func SendRank(msg *Message) {
	sprintf, err := db.ShowActivityRank()
	if err != nil {
		log.Println(err)
		return
	}

	rr := SendJsonMessage(msg.Conn, &Message{Type: MessageRank, Content: sprintf})
	if rr != nil {
		log.Println("向客户端发送活跃度排名失败:", rr)
		return
	}
	fmt.Println(msg.Sender, "查看活跃度排行...")
}
