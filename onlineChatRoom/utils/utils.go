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
		//log.Println("消息长度超出限制: ", length)
		return fmt.Errorf("message too long")
	}
	err := binary.Write(conn, binary.BigEndian, length) //将数据以二进制写入conn
	if err != nil {
		return fmt.Errorf("binary.Write failed")
	}
	_, err = conn.Write(message)
	if err != nil {
		return fmt.Errorf("conn.Write failed")
	}
	return nil
}

// ReadMessage 从连接读取消息
func ReadMessage(reader *bufio.Reader) ([]byte, error) {
	var length int32
	err := binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func CloseConn(conn net.Conn, name string) {
	err := conn.Close()
	if err != nil {
		log.Printf("%s close conn failed:%s \n", name, err)
	}
}
