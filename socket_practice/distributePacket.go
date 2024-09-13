package main

import (
	"fmt"
	"server/protocol"
	"time"
)

func (server *LifeGameServer) DistributePacket(sessionUniqueId uint64, sessionId int32, packetData []byte) {
	packetID := protocol.PeekPacketID(packetData)
	bodySize, packetBody := protocol.PeekPacketBody(packetData)

	packet := protocol.Packet{
		UserSessionIndex:       sessionId,
		UserSessionUniqueIndex: sessionUniqueId,
		Id:                     packetID,
		DataSize:               bodySize,
		Data:                   make([]byte, bodySize),
	}

	copy(packet.Data, packetBody)

	/*
		수신한 패킷을 처리하는 채널로 보냄
	*/
	server.PacketChan <- packet
}

/*
DistributePacket 함수에서 채널형식으로 넘겨줌
실질적인 패킷 처리 함수
*/
func (server *LifeGameServer) PacketProcessGoroutin() {
	roomUpdateTimerTicker := time.NewTicker(time.Second)
	defer roomUpdateTimerTicker.Stop()
	// defer : 지연실행
	// └ 특정 문장/함수를 나중에(defer를 호출하는 함수가 리턴하기 직전) 실행하게 한다.
	// 일반적으로 defer는 C#, java 같은 언어에서 finally 블럭처럼 마지막에 Clean-Up 작업을 위해 사용한다.

	for {
		select {
		// select는 switch와 비슷하지만, case에 채널이 사용됨.
		// 덕분에 동기화 코딩을 위해 화려한 코딩을 할 수 있음.
		case packet := <-server.PacketChan: // 다시 한 번 말하지만, server는 LifeGameServer 구조체 객체고 game.go에서 초기화 되었다.
			// case문에서 채널로부터 이벤트 수신
			// case문의 채널에 값이 들어올 때까지 select문에서 블로킹됨.
			{
				sessionId := packet.UserSessionIndex
				sessionUniqueId := packet.UserSessionUniqueIndex
				bodySize := packet.DataSize
				bodyData := packet.Data

				_ = sessionId
				_ = sessionUniqueId
				_ = bodySize
				_ = bodyData

				if packet.Id == protocol.PACKET_ID_LOGIN_REQ {

				} else if packet.Id == protocol.PACKET_ID_JOIN_REQ {

				} else if packet.Id == protocol.PACKET_ID_PING_REQ {

				} else {
					fmt.Println("Invalid Packet ID")
				}
			}

		/*
			초당 호출되는 로직
			이때 방을 업데이트 한다.
		*/
		case curTime := <-roomUpdateTimerTicker.C:
			{
				fmt.Println("Update Room")
				fmt.Println(curTime)
			}
		} // 몬스터의 위치 포지션
	}
}
