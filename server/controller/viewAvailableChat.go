package controller

import (
	"fmt"
	"server/errorcode"
	"server/network"
	"server/protocol"
	"server/service"
)

func ProcessPacketViewAvailableChatRoom(sessionUniqueId uint64, sessionId int32, bodySize int16, bodyData []byte) {
	var viewAvailableChatRoomReq protocol.ViewAvailableChatRoomReqPacket

	result := viewAvailableChatRoomReq.Decoding(bodyData)

	if !result {
		fmt.Println("ViewAvailableChatRoom Decoding Fail")
	}

	// userID := bytes.Trim(viewAvailableChatRoomReq.UserID, "\x00")
	results, err := service.SelectChatRoomInfo()
	if err != errorcode.ERROR_CODE_NONE {
		fmt.Println("Fetch Error")
	}

	var chatRoomInfos []protocol.ChatRoom
	var chatRoomID int16
	var chatRoomCreateDate string
	var chatRoomCreatorName string
	var chatRoomName string
	idx := 0

	for results.Next() {
		results.Scan(&chatRoomID, &chatRoomCreateDate, &chatRoomCreatorName, &chatRoomName)

		chatRoomInfos = append(chatRoomInfos, protocol.ChatRoom{})
		chatRoomInfos[idx].ID = chatRoomID
		chatRoomInfos[idx].CREATE_TIME = make([]byte, protocol.MAX_CHAT_TIME_BYTE_LENGTH)
		copy(chatRoomInfos[idx].CREATE_TIME[:], []byte(chatRoomCreateDate))
		chatRoomInfos[idx].CREATOR_NAME = make([]byte, protocol.MAX_USER_NAME_BYTE_LENGTH)
		copy(chatRoomInfos[idx].CREATOR_NAME[:], []byte(chatRoomCreatorName))
		chatRoomInfos[idx].CHATROOM_NAME = make([]byte, protocol.MAX_CHAT_NAME_BYTE_LENGTH)
		copy(chatRoomInfos[idx].CHATROOM_NAME[:], []byte(chatRoomName))

		// if !results.Next() {
		// 	break
		// }
		idx++
	}
	if idx == 0 {
		fmt.Println("접속 가능한 채팅방 없음.")
	}
	SendViewAvailableChatRoomResult(sessionUniqueId, sessionId, chatRoomInfos, 0)
	results.Close()
}

func SendViewAvailableChatRoomResult(sessionUniqueId uint64, sessionId int32, chatRoomInfos []protocol.ChatRoom, err int) {
	var viewAvailableChatRoomRes protocol.ViewAvailableChatRoomResPacket

	viewAvailableChatRoomRes.ErrorCode = int16(err)
	viewAvailableChatRoomRes.Len = int16(len(chatRoomInfos))
	viewAvailableChatRoomRes.ChatRooms = chatRoomInfos

	fmt.Println(viewAvailableChatRoomRes.ChatRooms)

	packet, _ := viewAvailableChatRoomRes.Encoding()

	network.SendToClient(sessionUniqueId, sessionId, packet)
}
