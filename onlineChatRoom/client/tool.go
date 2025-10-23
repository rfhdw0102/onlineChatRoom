package main

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

func screen() {
	fmt.Println("成功加入聊天室，可以开始聊天了...")
	fmt.Println("可用便捷命令：")
	fmt.Println("输入：list 查看在线用户...")
	fmt.Println("输入：quit 退出聊天室...")
	fmt.Println("输入：To:+用户名-->+内容 私聊...")
}

// 键盘输入处理
func keyboardInput() (string, error) {
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

// 注册 and 登录处理
func registerOrLogin(n string, conn net.Conn) *msg.Message {
	fmt.Println("请输入用户名和密码，用一个空格隔开...(quit退出)")
	reader := bufio.NewReader(conn)
	for {
		input, err := keyboardInput()
		if err != nil {
			log.Println(err)
		}
		if input == "quit" {
			fmt.Println("退出成功...")
			return nil
		}
		split := strings.Split(input, " ")
		if len(split) != 2 {
			fmt.Println("输入格式错误，请重新输入")
		}
		var loginMes *msg.Message
		if n == "1" {
			loginMes = &msg.Message{Type: msg.MessageRegister, Sender: split[0], Content: split[1]}
		} else {
			loginMes = &msg.Message{Type: msg.MessageJoin, Sender: split[0], Content: split[1]}
		}
		err = msg.SendJsonMessage(conn, loginMes)
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
			return loginMes
		} else {
			fmt.Println(response.Content)
			continue
		}
	}
}

// 处理服务端发来的信息
func handleServerMessage(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("client handleServerMessage panic recovered: %v\n", err)
		}
	}()
	reader := bufio.NewReader(conn)
	for {
		message, err := msg.ReadJsonMessage(reader)
		if err != nil {
			fmt.Println("server is end...")
			close(flag)
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
