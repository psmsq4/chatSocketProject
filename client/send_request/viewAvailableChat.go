package send_request

import (
	"client/network"
	"client/protocol"
)

func SendViewAvailableChatRoom(userID string) {
	viewAvailableChat := protocol.ViewAvailableChatRoomReqPacket{
		UserID: make([]byte, protocol.MAX_USER_ID_BYTE_LENGTH),
	}

	copy(viewAvailableChat.UserID[:], []byte(userID))

	packet, totalSize := viewAvailableChat.EncodingPacket()
	network.SendToServer(packet, totalSize)
}
