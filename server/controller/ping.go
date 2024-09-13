package controller

import (
	"fmt"
	"server/network"
	"server/protocol"
)

func ProcessPacketPing(sessionUniqueId uint64, sessionId int32, bodySize int16, bodyData []byte) {
	var request protocol.PingReqPacket
	result := (&request).Decoding(bodyData)
	if !result {
		return
	}

	fmt.Printf("[%d]-[%d] Client Send PING\n", sessionUniqueId, sessionId)
	sendPingResult(sessionUniqueId, sessionId, protocol.PING)
}

func sendPingResult(sessionUniqueId uint64, sessionId int32, result int8) {
	response := protocol.PingResPacket{
		Pong: result,
	}

	packet, _ := response.EncodingPacket()
	network.SendToClient(sessionUniqueId, sessionId, packet)
}
