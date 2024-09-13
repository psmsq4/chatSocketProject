package send_request

import (
	"client/network"
	"client/protocol"
)

func SendTransferMessage(chatRoomID int16, message string) {
	transferMessageReq := protocol.TransferMessageReqPacket{
		Message: make([]byte, protocol.MAX_CHAT_MESSAGE_BYTE_LENGTH),
	}

	transferMessageReq.ChatRoomID = chatRoomID
	copy(transferMessageReq.Message[:], []byte(message))

	packet, packetSize := transferMessageReq.EncodingPacket()
	network.SendToServer(packet, packetSize)
}
