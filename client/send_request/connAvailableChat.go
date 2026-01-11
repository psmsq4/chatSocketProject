package send_request

import (
	"client/network"
	"client/protocol"
)

func SendConnAvailableChatRoom(userID string, chatroomID int16) {
	connAvailableChat := protocol.ConnAvailableChatRoomReqPacket{
		UserID: make([]byte, protocol.MAX_USER_ID_BYTE_LENGTH),
	}

	copy(connAvailableChat.UserID[:], []byte(userID))
	connAvailableChat.ChatRoomID = chatroomID

	packet, totalSize := connAvailableChat.EncodingPacket()
	network.SendToServer(packet, totalSize)
}
