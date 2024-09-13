package main

import (
	"fmt"
	"server/controller"
	"server/protocol"
	"time"
)

/*
실제 패킷에서 아이디, 바디를 가지고옴
*/
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
DistributePacket함수에서 채널형식으로 넘겨줌
실질적인 패킷 처리 함수
*/
func (server *LifeGameServer) PacketProcessGoroutine() {

	roomUpdateTimerTicker := time.NewTicker(time.Second)
	defer roomUpdateTimerTicker.Stop()

	for {
		select {
		case packet := <-server.PacketChan:
			{
				sessionId := packet.UserSessionIndex
				sessionUniqueId := packet.UserSessionUniqueIndex
				bodySize := packet.DataSize
				bodyData := packet.Data

				if packet.Id == protocol.PACKET_ID_LOGIN_REQ {
					controller.ProcessPacketLogin(sessionUniqueId, sessionId, bodySize, bodyData)
				} else if packet.Id == protocol.PACKET_ID_JOIN_REQ {
					controller.ProcessPacketJoin(sessionUniqueId, sessionId, bodySize, bodyData)
				} else if packet.Id == protocol.PACKET_ID_PING_REQ {
					controller.ProcessPacketPing(sessionUniqueId, sessionId, bodySize, bodyData)
				} else if packet.Id == protocol.PACKET_CREATE_NEW_CHATROOM_REQ {
					controller.ProcessPacketCreateNewChat(sessionUniqueId, sessionId, bodySize, bodyData)
				} else if packet.Id == protocol.PACKET_TRANSFER_MESSAGE_REQ {
					controller.ProcessPacketTransferMessage(sessionUniqueId, sessionId, bodySize, bodyData)
				} else if packet.Id == protocol.PACKET_VIEW_AVAILABLE_CHATROOM_REQ {
					controller.ProcessPacketViewAvailableChatRoom(sessionUniqueId, sessionId, bodySize, bodyData)
				} else {
					fmt.Println("Invalid Packet ID")
				}
			}

		/*
			초당 호출되는 로직
			이때 방을 업데이트를 진행한다.
		*/
		case curTime := <-roomUpdateTimerTicker.C:
			{
				_ = curTime
				//fmt.Println("Update Room")
				//fmt.Println(curTime)
			}
		}
	}
}
