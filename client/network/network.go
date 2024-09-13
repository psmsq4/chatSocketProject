package network

import (
	"fmt"
	"net"
)

type SessionNetworkFunctor struct {
	OnConnect           func()
	OnReceive           func(packet []byte)
	PacketTotalSizeFunc func([]byte) int16
	PacketHeaderSize    int16
}

type TcpSession struct {
	Conn       net.Conn
	NetFunctor SessionNetworkFunctor
}

var _clientSession *TcpSession

func ConnectServer(functor SessionNetworkFunctor) {
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println(err)
		return
	}

	_clientSession = &TcpSession{
		Conn:       conn,
		NetFunctor: functor,
	}

	_clientSession.handleToRead()
}

func (session *TcpSession) handleToRead() {
	session.NetFunctor.OnConnect()
	recvBuff := make([]byte, MAX_BUFFER)
	var startRecvPos int16 = 0
	var result int
	for {
		recvBytes, err := session.Conn.Read(recvBuff[startRecvPos:])
		if err != nil {
			fmt.Println(err)
			return
		}

		readAbleByte := startRecvPos + int16(recvBytes)
		startRecvPos, result = session.makePacket(readAbleByte, recvBuff)
		if result != NET_ERROR_NONE {
			fmt.Println("Network Error")
			return
		}
	}
}

func (session *TcpSession) makePacket(readAbleByte int16, recvBuff []byte) (int16, int) {
	var startRecvPos int16 = 0
	var readPos int16 = 0

	packetHeaderSize := session.NetFunctor.PacketHeaderSize
	packetTotalSizeFunc := session.NetFunctor.PacketTotalSizeFunc

	for {

		/* 최소한의 크보다 작다면(패킷 헤더) */
		if readAbleByte < packetHeaderSize {
			break
		}

		requireBytes := packetTotalSizeFunc(recvBuff[readPos:])

		/* 읽어야 하는 크기보다 작다면 다음에 다시 받음 */
		if readAbleByte < requireBytes {
			break
		}

		/* 읽을 수 있는 최대 패킷보다 크다면 */
		if readAbleByte > MAX_PACKET_SIZE {
			return startRecvPos, NET_ERROR_TOO_LARGE_PACKET
		}

		ltvPacket := recvBuff[readPos:(readPos + requireBytes)]
		readPos += requireBytes
		readAbleByte -= requireBytes

		/* 패킷을 읽음을 알림 */
		session.NetFunctor.OnReceive(ltvPacket)
	}

	if readAbleByte > 0 {
		copy(recvBuff, recvBuff[readPos:(readPos+readAbleByte)])
	}

	startRecvPos = readAbleByte
	return startRecvPos, NET_ERROR_NONE
}

func (session *TcpSession) sendToServer(data []byte, size int16) error {
	_, err := session.Conn.Write(data)
	return err
}

func SendToServer(data []byte, size int16) error {
	err := _clientSession.sendToServer(data, size)
	return err
}
