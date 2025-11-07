package tool

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"onlineChatRoom/msg"
	"os"
	"strings"
	"time"
)

func Screen() {
	fmt.Println("成功加入聊天室，可以开始聊天了...")
	fmt.Println("可用便捷命令：")
	fmt.Println("1、输入：list 查看在线用户...")
	fmt.Println("2、输入：quit 退出...")
	fmt.Println("3、输入：To:+用户名-->+内容 私聊...")
	fmt.Println("4、输入：rank 查看聊天室所有用户活跃度排名...")
}

// KeyboardInput 键盘输入处理
func KeyboardInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	message := strings.TrimSpace(input)
	if message == "" {
		return "", fmt.Errorf("keyboardInput failed: input is empty")
	}
	return message, nil
}

// RegisterOrLogin 注册 and 登录处理
func RegisterOrLogin(n string, conn net.Conn) *msg.Message {
	reader := bufio.NewReader(conn)
	for {
		fmt.Println("请输入账户:")
		username, _ := KeyboardInput()
		if username == "quit" {
			fmt.Println("退出成功...")
			return nil
		}
		fmt.Println("请输入密码:")
		password, _ := KeyboardInput()
		var loginMes *msg.Message
		if n == "1" {
			loginMes = &msg.Message{Type: msg.MessageRegister, Sender: username, Content: password}
		} else {
			loginMes = &msg.Message{Type: msg.MessageJoin, Sender: username, Content: password}
		}
		err := msg.SendJsonMessage(conn, loginMes)
		if err != nil {
			log.Println("register send Message failed...")
			continue
		}
		response, err := msg.ReadJsonMessage(reader)
		if err != nil {
			log.Println("register read Message failed...")
			continue
		}
		if response.Content == "OK" {
			if n == "1" {
				fmt.Println("注册成功...")
			} else {
				fmt.Println("登录成功...")
			}
			return loginMes
		} else {
			fmt.Println(response.Content)
			continue
		}
	}
}

// HandleServerMessage 处理服务端发来的信息
func HandleServerMessage(conn net.Conn, serverErrFlag chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("client handleServerMessage panic recovered: %v\n", err)
		}
	}()
	reader := bufio.NewReader(conn)
	for {
		message, err := msg.ReadJsonMessage(reader)
		if err != nil {
			close(serverErrFlag)
			return
		}
		switch message.Type {
		case msg.MessageHeart:
			//fmt.Println("接收到pong...")
			continue
		case msg.MessagePrivate:
			fmt.Println(message.Sender, "私聊你:", message.Content)
		default:
			fmt.Println(message.Content)
		}
	}
}

// StartHeartbeat 发送心跳包
func StartHeartbeat(username string, conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("client startHeartbeat panic recovered: %v\n", err)
		}
	}()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		message := &msg.Message{
			Type:    msg.MessageHeart,
			Sender:  username,
			Content: "PING",
		}
		err := msg.SendJsonMessage(conn, message)
		if err != nil {
			log.Println("send heartBeat failed:", err)
			return
		}
	}
}

// SendServer 向服务端发送消息
func SendServer(content string, conn net.Conn, userMsg *msg.Message, clientQuitFlag chan struct{}) {
	if content == "quit" {
		quitErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageLeave, Sender: userMsg.Sender})
		if quitErr != nil {
			log.Println("send msg.MessageLeave failed...", quitErr)
		}
		fmt.Println("退出成功...")
		close(clientQuitFlag)
		return
	}
	if content == "list" {
		listErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageList, Sender: userMsg.Sender})
		if listErr != nil {
			log.Println("send msg.MessageList failed...", listErr)
		}
		return
	}
	if strings.HasPrefix(content, "To:") {
		parts := strings.Split(content, "-->")
		if len(parts) != 2 {
			fmt.Println("格式错误，应为 To:用户名-->内容")
			return
		}
		target := strings.TrimPrefix(parts[0], "To:")
		privateErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessagePrivate, Sender: userMsg.Sender, Receiver: target, Content: parts[1]})
		if privateErr != nil {
			log.Println("send msg.MessagePrivate failed...", privateErr)
		}
		fmt.Println("发送成功...")
		return
	}
	if content == "rank" {
		rankErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageRank, Sender: userMsg.Sender})
		if rankErr != nil {
			log.Println("send msg.MessageRank failed...", rankErr)
		}
		return
	}
	r := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageChat, Sender: userMsg.Sender, Content: content})
	if r != nil {
		log.Println("send msg.MessageChat failed...", r)
	}
}
