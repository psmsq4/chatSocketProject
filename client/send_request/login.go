package send_request

import (
	"client/network"
	"client/protocol"
)

func SendLogin(userID, userPW string) {
	loginReq := protocol.LoginReqPacket{
		UserID: make([]byte, protocol.MAX_USER_ID_BYTE_LENGTH),
		UserPW: make([]byte, protocol.MAX_USER_PW_BYTE_LENGTH),
	}
	copy(loginReq.UserID[:], []byte(userID))
	copy(loginReq.UserPW[:], []byte(userPW))
	packet, packetSize := loginReq.EncodingPacket()

	network.SendToServer(packet, packetSize)
}
