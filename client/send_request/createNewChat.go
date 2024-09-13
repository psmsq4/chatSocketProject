package send_request

import (
	"client/network"
	"client/protocol"
)

func SendCreateNewChatRoomReq(chatRoomName string, chatRoomPW string) {
	createNewChatRoomReq := protocol.CreateNewChatRoomReqPacket{
		ChatRoomName: make([]byte, protocol.MAX_CHAT_NAME_BYTE_LENGTH),
		ChatRoomPW:   make([]byte, protocol.MAX_CHAT_PW_BYTE_LENGTH),
	}

	copy(createNewChatRoomReq.ChatRoomName[:], []byte(chatRoomName))
	copy(createNewChatRoomReq.ChatRoomPW[:], []byte(chatRoomPW))

	packet, packetSize := createNewChatRoomReq.EncodingPacket()
	network.SendToServer(packet, packetSize)
}
