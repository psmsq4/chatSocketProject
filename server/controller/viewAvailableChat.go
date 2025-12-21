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

	var chatroom_infos []protocol.ChatRoom
	var chatroom_id int16
	var chatroom_create_date string
	var chatroom_creator string
	var chatroom_name string
	var num_attendance int16
	idx := 0

	for results.Next() {
		results.Scan(&chatroom_id, &chatroom_create_date, &chatroom_creator, &chatroom_name, &num_attendance)

		chatroom_infos = append(chatroom_infos, protocol.ChatRoom{})
		chatroom_infos[idx].ID = chatroom_id
		chatroom_infos[idx].CREATE_TIME = make([]byte, protocol.MAX_CHAT_TIME_BYTE_LENGTH)
		copy(chatroom_infos[idx].CREATE_TIME[:], []byte(chatroom_create_date))
		chatroom_infos[idx].CREATOR_NAME = make([]byte, protocol.MAX_USER_NAME_BYTE_LENGTH)
		copy(chatroom_infos[idx].CREATOR_NAME[:], []byte(chatroom_creator))
		chatroom_infos[idx].CHATROOM_NAME = make([]byte, protocol.MAX_CHAT_NAME_BYTE_LENGTH)
		copy(chatroom_infos[idx].CHATROOM_NAME[:], []byte(chatroom_name))
		chatroom_infos[idx].NUM_ATTENDANCE = num_attendance

		// if !results.Next() {
		// 	break
		// }
		idx++
	}
	if idx == 0 {
		fmt.Println("접속 가능한 채팅방 없음.")
	}
	SendViewAvailableChatRoomResult(sessionUniqueId, sessionId, chatroom_infos, 0)
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
