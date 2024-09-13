package controller

import (
	"bytes"
	"fmt"
	"server/errorcode"
	"server/network"
	"server/protocol"
	"server/service"
)

func ProcessPacketCreateNewChat(sessionUniqueId uint64, sessionId int32, bodySize int16, bodyData []byte) {
	var request protocol.CreateNewChatRoomReqPacket

	result := request.Decoding(bodyData)
	if !result {
		fmt.Println("a")
		sendCreateNewChatResult(sessionUniqueId, sessionId, -1, errorcode.ERROR_CODE_FAIL_CREATE_NEW_CHATROOM)
		return
	}

	chatRoomName := bytes.Trim(request.ChatRoomName, "\x00")
	chatRoomPW := bytes.Trim(request.ChatRoomPW, "\x00")

	session := service.LoadUserInfo(sessionUniqueId, 0)
	UserID := bytes.Trim(session.UserID, "\x00")

	chatRoomID, err := service.CreateNewChatRoom(string(UserID), chatRoomName, chatRoomPW)
	sendCreateNewChatResult(sessionUniqueId, sessionId, chatRoomID, err)

	service.InsertCIDToUser(sessionUniqueId, sessionId, chatRoomID)
	service.InsertAttendanceInformation(chatRoomID, string(UserID), "AUTH")
}

func sendCreateNewChatResult(sessionUniqueId uint64, sessionId int32, chatRoomID int16, err int16) {
	createNewChatRoomResPacket := protocol.CreateNewChatRoomResPacket{
		ErrorCode:  err,
		ChatRoomID: chatRoomID,
	}

	sendBuf, _ := createNewChatRoomResPacket.EncodingPacket()

	network.SendToClient(sessionUniqueId, sessionId, sendBuf)
}
