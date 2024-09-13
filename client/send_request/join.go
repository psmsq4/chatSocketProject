package send_request

import (
	"client/network"
	"client/protocol"
)

func SendJoin(userID, userPW, userNAME string) {
	joinReq := protocol.JoinReqPacket{
		UserID:   make([]byte, protocol.MAX_USER_ID_BYTE_LENGTH),
		UserPW:   make([]byte, protocol.MAX_USER_PW_BYTE_LENGTH),
		UserName: make([]byte, protocol.MAX_USER_NAME_BYTE_LENGTH),
	}

	copy(joinReq.UserID[:], []byte(userID))
	copy(joinReq.UserPW[:], []byte(userPW))
	copy(joinReq.UserName[:], []byte(userNAME))

	packet, packetSize := joinReq.EncodingPacket()
	network.SendToServer(packet, packetSize)
}
