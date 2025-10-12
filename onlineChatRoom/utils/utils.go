package utils

import (
	"bufio"
	"encoding/binary"
	"errors"
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
		//log.Println("消息长度超出限制: ", length)
		return fmt.Errorf("message too long")
	}
	// 写入消息长度
	err := binary.Write(conn, binary.BigEndian, length) //将数据以二进制写入conn
	if err != nil {
		//log.Printf("写入消息长度时出错: %v", err)
		return fmt.Errorf("binary.Write--写入消息长度时出错")
	}
	// 写入消息内容
	_, err = conn.Write(message)
	if err != nil {
		//log.Printf("写入消息内容时出错: %v", err)
		return fmt.Errorf("conn.Write--写入消息时出错")
	}
	return nil
}

// ReadMessage 从连接读取消息
func ReadMessage(reader *bufio.Reader) ([]byte, error) {
	var length int32
	err := binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		// 当连接关闭或长度无法读取时视为 EOF（正常断开）
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		return nil, err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		return nil, err
	}
	return buf, nil
}

func CloseConn(conn net.Conn, name string) {
	err := conn.Close()
	if err != nil {
		log.Printf("%s 关闭连接失败： %s \n", name, err)
	}
}
