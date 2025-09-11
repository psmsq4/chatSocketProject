package send_request

import (
	"client/network"
	"client/protocol"
	"time"
)

func SendTransferMessage(chatRoomID int16, message string) {
	transferMessageReq := protocol.TransferMessageReqPacket{
		TimeChat: make([]byte, protocol.MAX_CHAT_TIME_BYTE_LENGTH),
		Message:  make([]byte, protocol.MAX_CHAT_MESSAGE_BYTE_LENGTH),
	}

	transferMessageReq.ChatRoomID = chatRoomID
	copy(transferMessageReq.TimeChat[:], []byte(time.Now().Format("2006-01-02 15:04:05.000")))
	copy(transferMessageReq.Message[:], []byte(message))

	packet, packetSize := transferMessageReq.EncodingPacket()
	network.SendToServer(packet, packetSize)
}
