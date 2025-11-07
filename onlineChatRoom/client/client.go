package main

import (
	"fmt"
	"log"
	"net"
	"onlineChatRoom/client/tool"
	"onlineChatRoom/msg"
	"onlineChatRoom/utils"
)

// 服务端是否关闭
var serverErrFlag = make(chan struct{})

// 客户是否退出通道
var clientQuitFlag = make(chan struct{})

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
		n, er := tool.KeyboardInput()
		if n != "1" && n != "2" {
			fmt.Println("请输入1 or 2...")
			continue
		}
		if er != nil {
			fmt.Println(er)
		}
		userMsg = tool.RegisterOrLogin(n, conn)
		if userMsg != nil && n == "2" {
			break
		}
	}
	//处理服务端的发来的信息
	go tool.HandleServerMessage(conn, serverErrFlag)
	//发送心跳
	go tool.StartHeartbeat(userMsg.Sender, conn)
	tool.Screen()
	//输入内容的管道
	var msgChan = make(chan string)
	// 键盘并发输入
	go func() {
		for {
			content, inputErr := tool.KeyboardInput()
			if inputErr != nil {
				log.Println(inputErr)
				continue
			}
			msgChan <- content
		}
	}()
	// 消息发送
	for {
		select {
		case <-serverErrFlag: // 服务端崩溃
			fmt.Println("与服务端断开连接，客户端退出...")
			return
		case <-clientQuitFlag:
			return
		case content := <-msgChan:
			tool.SendServer(content, conn, userMsg, clientQuitFlag)
		}
	}
}
