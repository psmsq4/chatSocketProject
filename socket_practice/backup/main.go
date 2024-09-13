package main

import (
	"fmt"
	"net"
	"server/packet"
)

func main() {
	fmt.Println("Start Aegis Jocker Game Server!")
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := listener.Accept() // listener는 server socket , conn은 client socket
		if err != nil {
			fmt.Println(err)
			continue
		}
		go clientHandler(conn)
	}
}

func clientHandler(conn net.Conn) {
	for {
		buffer := make([]byte, packet.SOCKET_BUFFER)
		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// go는 패키지로 유지됨.
// 패키지에서 패키지를 호출할 수 있음.
// 패키지는 서로 순환참조가 안 됨. (A에서 B를 호출하고 B에서 A를 호출할 수 없음.)
