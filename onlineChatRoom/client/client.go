package main

import (
	"fmt"
	"log"
	"net"
	"onlineChatRoom/msg"
	"onlineChatRoom/utils"
	"strings"
)

// 服务端是否关闭
var flag = make(chan struct{})

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("client main panic recovered: %v\n", err)
		}
	}()
	conn, err := net.Dial("tcp", "localhost:8080") //连接服务端
	if err != nil {
		log.Fatal("连接服务器出错...", err)
	}
	defer utils.CloseConn(conn, "客户端")

	fmt.Println("-------------欢迎来到网络聊天室-------------")
	var userMsg *msg.Message
	//注册 or 登录处理
	for {
		fmt.Println("1、注册...")
		fmt.Println("2、登录...")
		n, er := keyboardInput()
		if n != "1" && n != "2" {
			fmt.Println("请输入1 or 2...")
			continue
		}
		if er != nil {
			fmt.Println(er)
		}
		userMsg = registerOrLogin(n, conn)
		if userMsg != nil && n == "2" {
			break
		}
	}
	//处理服务端的发来的信息
	go handleServerMessage(conn)
	//发送心跳
	go startHeartbeat(userMsg.Sender, conn)
	screen()
	//输入内容的管道
	var msgChan = make(chan string)
	// 键盘输入协程
	go func() {
		for {
			content, inputErr := keyboardInput()
			if inputErr != nil {
				log.Println(inputErr)
				continue
			}
			msgChan <- content
		}
	}()
	for {
		select {
		case <-flag:
			return
		case content := <-msgChan:
			if content == "quit" {
				quitErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageLeave, Sender: userMsg.Sender})
				if quitErr != nil {
					log.Println("send msg.MessageLeave failed...", err)
				}
				fmt.Println("退出成功...")
				return
			}
			if content == "list" {
				listErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageList, Sender: userMsg.Sender})
				if listErr != nil {
					log.Println("send msg.MessageList failed...", err)
				}
				continue
			}
			if strings.HasPrefix(content, "To:") {
				parts := strings.Split(content, "-->")
				if len(parts) != 2 {
					fmt.Println("格式错误，应为 To:用户名-->内容")
					continue
				}
				target := strings.TrimPrefix(parts[0], "To:")
				privateErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessagePrivate, Sender: userMsg.Sender, Receiver: target, Content: parts[1]})
				if privateErr != nil {
					log.Println("send msg.MessagePrivate failed...", err)
				}
				fmt.Println("发送成功...")
				continue
			}
			if content == "rank" {
				rankErr := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageRank, Sender: userMsg.Sender})
				if rankErr != nil {
					log.Println("send msg.MessageRank failed...", err)
				}
				continue
			}
			r := msg.SendJsonMessage(conn, &msg.Message{Type: msg.MessageChat, Sender: userMsg.Sender, Content: content})
			if r != nil {
				log.Println("send msg.MessageChat failed...", err)
			}
		}
	}
}
