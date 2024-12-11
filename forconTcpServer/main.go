package main

import (
	"fmt"
	"net"
)

func main() {
	// 监听TCP端口 8080
	listener, err := net.Listen("tcp", "localhost:6217")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()
	fmt.Println("Listening on localhost:6217...")

	for {
		// 等待客户端连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

// 处理单个连接
func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 512)

	for {
		// 读取客户端发送的数据
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		} else {
			//fmt.Println("msg: ", string(buffer[:n]), "len=",n)
		}



		if n == 4 {
			if buffer[0] == 'M' && buffer[1] == '0' && buffer[2] == '\r' && buffer[3] == '\n'{
				str := "M0,-000001167,+000004966" + "\r"
				byteArray := []byte(str)
				_, err = conn.Write(byteArray)
				if err != nil {
					fmt.Println("Error writing:", err.Error())
					break
				} else {
					fmt.Println("out:" + str)
				}
			}
		}
		if n == 5 {
			if buffer[0] == 'M' && buffer[1] == '0' && buffer[2] == ',' && buffer[3] == '0' && buffer[4] == '\r'{
				//"M0,+53.93920,+53.93920"
				str := "M0,+53.93920,+53.93920" + "\r"
				byteArray := []byte(str)
				_, err = conn.Write(byteArray)
				if err != nil {
					fmt.Println("Error writing:", err.Error())
					break
				} else {
					fmt.Println("out:" + str)
				}
			}
		}
	}
}