package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"onlineChatRoom/msg"
	"onlineChatRoom/utils"
	"os"
	"strings"
	"time"
)

func screen() {
	fmt.Println("成功加入聊天室，可以开始聊天了...")
	fmt.Println("可用命令：")
	fmt.Println("输入：list 查看在线用户...")
	fmt.Println("输入：quit 退出聊天室...")
	fmt.Println("输入：To:+用户名-->+内容 私聊...")
}

// 处理服务端发来的信息
func handleServerMessage(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := msg.ReadJsonMessage(reader)
		if err != nil {
			log.Println("服务端已关闭...")
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

// 发送心跳包
func startHeartbeat(username string, conn net.Conn) {
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
			log.Println("startHeartbeat发送心跳结构体失败:", err)
			return
		}
	}
}

// 键盘输入处理
func ReadInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	message := strings.TrimSpace(input)
	if message == "" {
		return "", fmt.Errorf("输入格式错误，输入不能为空")
	}
	return message, nil
}

// 验证用户名能否登录
func sendLoginMessage(conn net.Conn) *msg.Message {
	reader := bufio.NewReader(conn)
	for {
		fmt.Println("请输入你的网名:")
		username, err := ReadInput()
		if err != nil {
			log.Println("sendLoginMessage读取网名失败...")
		}
		loginMes := &msg.Message{Type: msg.MessageJoin, Sender: username}
		err = msg.SendJsonMessage(conn, loginMes)
		if err != nil {
			log.Println("sendLoginMessage发送网名失败...")
			continue
		}
		response, err := msg.ReadJsonMessage(reader)
		if err != nil {
			log.Println("sendLoginMessage接收服务器响应失败...")
			continue
		}
		if response.Content == "OK" {
			return loginMes
		} else {
			fmt.Println(response.Content)
			continue
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080") //连接服务端
	if err != nil {
		log.Fatal("连接服务器出错...", err)
	}
	defer utils.CloseConn(conn, "客户端")

	fmt.Println("-------------欢迎来到网络聊天室-------------")
	loginMes := sendLoginMessage(conn)       //输入网名
	go handleServerMessage(conn)             //开启一个协程来处理服务端的信息
	go startHeartbeat(loginMes.Sender, conn) //开启心跳检测
	screen()
	for {
		content, err := ReadInput()
		if err != nil {
			log.Println("main输入失败...")
			continue
		}
		if content == "quit" {
			fmt.Println("退出成功...")
			break
		}
		if content == "list" {
			err := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageList, Sender: loginMes.Sender, Content: "quit"})
			if err != nil {
				log.Println("main发送list失败...")
				continue
			}
			continue
		}
		if strings.HasPrefix(content, "To:") {
			parts := strings.SplitN(content, "-->", 2)
			if len(parts) < 2 {
				fmt.Println("格式错误，应为 To:用户名-->内容")
				continue
			}
			target := strings.TrimPrefix(parts[0], "To:")
			err := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessagePrivate, Sender: loginMes.Sender, Receiver: target, Content: parts[1]})
			if err != nil {
				log.Println("main发送私聊失败...")
				continue
			}
			continue
		}
		r := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageChat, Sender: loginMes.Sender, Content: content})
		if r != nil {
			log.Println("main发送普通消息失败...")
		}
	}
}
