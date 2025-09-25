package utils

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

const maxMessageLength = 1 << 20 // 1MB，最大消息长度限制

// SendMessage 向连接发送消息
func SendMessage(conn net.Conn, message []byte) error {
	length := uint32(len(message)) // 消息长度
	if length > maxMessageLength {
		log.Println("消息长度超出限制: ", length)
		return fmt.Errorf("message too long")
	}
	// 写入消息长度
	err := binary.Write(conn, binary.BigEndian, length) //将数据以二进制写入conn
	if err != nil {
		log.Printf("写入消息长度时出错: %v", err)
		return err
	}
	// 写入消息内容
	_, err = conn.Write(message)
	if err != nil {
		log.Printf("写入消息内容时出错: %v", err)
	}
	return err
}

// ReadMessage 从连接读取消息
func ReadMessage(reader *bufio.Reader) ([]byte, error) {
	var length uint32                                     // 消息长度
	err := binary.Read(reader, binary.BigEndian, &length) //从reader中读取二进制数据并解析为结构化数据，自动填充到length中
	if err != nil {
		return nil, err
	}
	if length > maxMessageLength {
		log.Println("消息长度超出限制:", length)
		return nil, fmt.Errorf("message too long")
	}
	// 读取消息内容
	message := make([]byte, length)
	_, err = io.ReadFull(reader, message)
	if err != nil {
		log.Printf("读取消息内容时出错: %v", err)
		return nil, err
	}
	return message, nil
}
