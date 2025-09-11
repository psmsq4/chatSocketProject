package controller

import (
	"bytes"
	"fmt"
	"server/network"
	"server/protocol"
	"server/service"
)

func ProcessPacketTransferMessage(sessionUniqueId uint64, sessionId int32, bodySize int16, bodyData []byte) {
	// 현재 상태 : 임의의 클라이언트가 서버로 채팅방ID와 메세지를 보낸 상황
	// [서버 측 작업]
	//	- 1. 패킷을 구조체에 저장함. 메세지는 bytes.Trim()으로 공백 제거
	//	- 2. 해당 구조체 정보를 DB에 저장함 -> message, timeChat, chatRoomID는 이미 알고 있으므로 userID와 messageID만 StoreMessageToDB에서 반환 받으면 됨.
	//	- 3. 현재 채팅방에 접속해 있는 클라이언트 대상으로 해당 메세지를 전송함. (송신자에게도 전송)
	//	※ 전제조건
	//	: 채팅방에 접속해있는 유저를 파악할 수 있도록 세션 정보(RedisUser)에 ChatRoomID 필드가 추가되어야 함.
	//	: 또한, service/memoryDB.go에는 chatRoomID를 매개변수로 받고, 해당 chatRoomID을 가진 UserSession을 추출할 수 있어야 함.
	//	: RedisUser 구조체에는 Connection은 없으므로, RedisUser 정보 기반으로 TcpSession을 추출해야함.

	var transferMessage protocol.TransferMessageReqPacket
	result := transferMessage.Decoding(bodyData)
	if !result {
		fmt.Println("err")
		return
	}

	// fmt.Println(string(bytes.Trim(transferMessage.Message, "\x00")))

	chatRoomID := transferMessage.ChatRoomID
	timeChat := bytes.Trim(transferMessage.TimeChat, "\x00")
	message := bytes.Trim(transferMessage.Message, "\x00")

	messageID, userName, err := service.StoreMessageToDB(sessionUniqueId, chatRoomID, string(message), string(timeChat))

	listOfLiveUsers := service.RetrieveUsersFromCID(chatRoomID) // chatRoomID를 갖고 있는 세션의 sessionUniqueID들의 배열을 반환한다.
	// fmt.Println(listOfLiveUsers)
	// fmt.Println(messageID, unsafe.Sizeof(messageID))
	// fmt.Println(time, unsafe.Sizeof(time))
	// fmt.Println(user_name, unsafe.Sizeof(user_name))

	BroadcastMessage := protocol.BroadcastMessagePacket{
		Message:  make([]byte, protocol.MAX_CHAT_MESSAGE_BYTE_LENGTH),
		TimeChat: make([]byte, protocol.MAX_CHAT_TIME_BYTE_LENGTH),
		UserName: make([]byte, protocol.MAX_USER_NAME_BYTE_LENGTH),
	}

	BroadcastMessage.MessageSequence = messageID
	copy(BroadcastMessage.Message[:], message)
	copy(BroadcastMessage.TimeChat[:], []byte(timeChat))
	copy(BroadcastMessage.UserName[:], []byte(userName))

	SendTransferMessageResult(sessionUniqueId, sessionId, err)
	BroadcastMessageToLiveUsers(listOfLiveUsers, &BroadcastMessage)
}

func SendTransferMessageResult(sessionUniqueId uint64, sessionId int32, errorcode int16) {
	var transferMessageRes protocol.TransferMessageResPacket

	transferMessageRes.ErrorCode = errorcode
	packet, _ := transferMessageRes.EncodingPacket()

	network.SendToClient(sessionUniqueId, sessionId, packet)
}

func BroadcastMessageToLiveUsers(listOfLiveUsers []int16, broadcastMessage *protocol.BroadcastMessagePacket) {
	packet, _ := broadcastMessage.EncodingPacket()

	// fmt.Println(packet)

	for _, userSessionUniqueId := range listOfLiveUsers {
		// fmt.Println(userSessionUniqueId)
		network.SendToClient(uint64(userSessionUniqueId), 0, packet)
	}
}
