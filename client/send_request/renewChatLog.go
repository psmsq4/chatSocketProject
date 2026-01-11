package send_request

import (
	"client/network"
	"client/protocol"
)

func SendRenewChatLogReqPacket(last_message_id int32, chatroom_id int16) {
	var renewChatLogReq protocol.RenewChatLogReqPacket

	renewChatLogReq.ChatLogEndSequence = int32(last_message_id)
	renewChatLogReq.ChatRoomID = chatroom_id
	sendBuf, totalSize := renewChatLogReq.EncodingPacket()

	network.SendToServer(sendBuf, totalSize)
}
