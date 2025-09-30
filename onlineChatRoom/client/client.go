package main

import (
	"awesomeProject/onlineChatRoom/utils"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// 处理从服务端接收到的信息
func handleServerMessage(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := utils.ReadMessage(reader)
		if err != nil {
			log.Println("接收服务器消息时出错...")
			return
		}
		msgStr := string(message)

		// 如果是心跳回复 PONG，不作为普通消息打印
		if msgStr == "PONG" {
			//fmt.Println("[心跳] 收到服务器响应: PONG")
			continue
		}

		// 普通消息
		fmt.Println(msgStr)
	}
}

func sendLoginMessage(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("请输入你的网名:")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("读取网名失败...")
			continue
		}
		username := strings.TrimSpace(input)
		if username == "" {
			fmt.Println("网名不能为空...")
			continue
		}
		//发送网名到服务器
		err = utils.SendMessage(conn, []byte(username))
		if err != nil {
			log.Println("发送网名失败...")
			continue
		}
		//等待服务器响应
		response, err := utils.ReadMessage(bufio.NewReader(conn))
		if err != nil {
			log.Println("接收服务器响应失败...")
			continue
		}
		respStr := string(response)
		if respStr == "OK" {
			return
		} else {
			fmt.Println(respStr)
			continue
		}
	}
}

// 发送心跳包
func startHeartbeat(conn net.Conn) {
	ticker := time.NewTicker(10 * time.Second) // 每 10 秒发一次心跳
	defer ticker.Stop()
	for {
		<-ticker.C
		err := utils.SendMessage(conn, []byte("PING"))
		if err != nil {
			log.Println("发送心跳失败:", err)
			return
		}
		//fmt.Println("[心跳] 已发送 PING")
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080") //连接服务端
	if err != nil {
		log.Fatal("连接服务器出错...", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("conn 退出出错...", err)
		}
	}(conn) //延迟关闭

	fmt.Println("-------------欢迎来到网络聊天室-------------")
	sendLoginMessage(conn)       //输入网名
	go handleServerMessage(conn) //开启一个协程来处理服务端的信息
	go startHeartbeat(conn)      // 开启心跳检测

	fmt.Println("成功加入聊天室，可以开始聊天了...")
	fmt.Println("可用命令：")
	fmt.Println("输入：list 查看在线用户...")
	fmt.Println("输入：quit 退出聊天室...")
	fmt.Println("输入：To:+用户名-->+内容 私聊...")

	for {
		reader := bufio.NewReader(os.Stdin)
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("读取用户输入错误...")
		}
		message = strings.TrimSpace(message)
		if message == "" {
			fmt.Println("输入不能为空..")
			continue
		}
		if message == "quit" {
			fmt.Println("退出成功...")
			err := utils.SendMessage(conn, []byte("quit"))
			if err != nil {
				log.Fatal("退出连接时出错...")
			}
			break
		}
		err = utils.SendMessage(conn, []byte(message))
		if err != nil {
			log.Fatal("发送信息失败...")
		}
	}
}
