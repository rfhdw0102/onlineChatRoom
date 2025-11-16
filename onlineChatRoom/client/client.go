package main

import (
	"fmt"
	"log"
	"net"
	"onlineChatRoom/client/tool"
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
	userMsg := tool.HandleRegOrLog(conn)
	//处理服务端的发来的信息
	go func() {
		err = tool.HandleServerMessage(conn, serverErrFlag)
		if err != nil {
			log.Println(err)
		}
		close(serverErrFlag)
	}()
	//发送心跳
	go func() {
		err = tool.StartHeartbeat(userMsg.Sender, conn)
		if err != nil {
			log.Println(err)
		}
		close(clientQuitFlag)
	}()
	tool.Screen()
	//输入内容的管道
	var msgChan = make(chan string)
	// 键盘并发输入
	go func() {
		for {
			content, inputErr := tool.KeyboardInput()
			if inputErr != nil {
				log.Println(inputErr)
				close(clientQuitFlag)
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
		case <-clientQuitFlag: // 客户端异常
			return
		case content := <-msgChan:
			tool.SendServer(content, conn, userMsg, clientQuitFlag)
		}
	}
}
