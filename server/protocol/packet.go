package protocol

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"server/errorcode"
	"server/network"
)

const (
	PACKET_TYPE_NORMAL   = 0
	PACKET_TYPE_COMPRESS = 1
	PACKET_TYPE_SECURE   = 2
)

const (
	MAX_USER_ID_BYTE_LENGTH      = 16
	MAX_USER_PW_BYTE_LENGTH      = 16
	MAX_USER_NAME_BYTE_LENGTH    = 16
	MAX_CHAT_MESSAGE_BYTE_LENGTH = 512
	MAX_CHAT_NAME_BYTE_LENGTH    = 20
	MAX_CHAT_PW_BYTE_LENGTH      = 20
	MAX_CHAT_TIME_BYTE_LENGTH    = 20
)

var _packetHeaderSize int16

func InitPacketHeaderSize() {
	_packetHeaderSize = PacketHeaderSize()
}

func GetPacketHeaderSize() int16 {
	return _packetHeaderSize
}

/*
전체 패킷에서 총 크기를 제외한 다음 2바이트를 꺼내옴
*/
func PeekPacketID(rawData []byte) int16 {
	packetID := binary.LittleEndian.Uint16(rawData[2:])
	return int16(packetID)
}

/*
전체 패킷에서 헤더를 뺸 만큼 바디로 지정
*/
func PeekPacketBody(rawData []byte) (int16, []byte) {
	headerSize := _packetHeaderSize
	totalSize := int16(binary.LittleEndian.Uint16(rawData))
	bodySize := totalSize - headerSize

	if bodySize > 0 {
		return bodySize, rawData[headerSize:]
	}

	return bodySize, []byte{}
}

/*
패킷 헤더를 추가한다.
*/
func EncodingPacketHeader(writer *network.RawPacketData, totalSize int16, pktId int16, pktType int8) {
	writer.WriteS16(totalSize)
	writer.WriteS16(pktId)
	writer.WriteS8(pktType)
}

/*
패킷 헤더를 분석한다.
*/
func DecodingPacketHeader(header *Header, data []byte) {
	reader := network.MakeReader(data, true)
	header.TotalSize, _ = reader.ReadS16()
	header.ID, _ = reader.ReadS16()
	header.PacketType, _ = reader.ReadS8()
}

/*
패킷헤더의 크기를 사전에 구함
*/
func PacketHeaderSize() int16 {
	var header Header
	hSize := network.Sizeof(reflect.TypeOf(header))
	return (int16)(hSize)
}

/* 로그인 요청 */
func (loginReq LoginReqPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + MAX_USER_ID_BYTE_LENGTH + MAX_USER_PW_BYTE_LENGTH
	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_LOGIN_REQ, 0)
	writer.WriteBytes(loginReq.UserID[:])
	writer.WriteBytes(loginReq.UserPW[:])
	return sendBuf, totalSize
}

func (loginReq *LoginReqPacket) Decoding(bodyData []byte) bool {
	bodySize := MAX_USER_ID_BYTE_LENGTH + MAX_USER_PW_BYTE_LENGTH
	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)

	var err error
	loginReq.UserID, err = reader.ReadBytes(MAX_USER_ID_BYTE_LENGTH)
	if err != nil {
		return false
	}

	loginReq.UserPW, err = reader.ReadBytes(MAX_USER_PW_BYTE_LENGTH)
	return err == nil
}

func (loginRes LoginResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + 2
	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_LOGIN_RES, 0)
	writer.WriteS16(loginRes.ErrorCode)
	return sendBuf, totalSize
}

/* 로그인 응답 */
func (loginRes *LoginResPacket) Decoding(bodyData []byte) bool {
	bodySize := 2
	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)

	var err error
	loginRes.ErrorCode, err = reader.ReadS16()
	return err == nil
}

func (joinReq JoinReqPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + MAX_USER_ID_BYTE_LENGTH + MAX_USER_PW_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH
	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_LOGIN_REQ, 0)
	writer.WriteBytes(joinReq.UserID[:])
	writer.WriteBytes(joinReq.UserPW[:])
	writer.WriteBytes(joinReq.UserName[:])
	return sendBuf, totalSize
}

/* 회원가입 요청 */
func (joinReq *JoinReqPacket) Decoding(bodyData []byte) bool {
	bodySize := MAX_USER_ID_BYTE_LENGTH + MAX_USER_PW_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH
	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)

	var err error
	joinReq.UserID, err = reader.ReadBytes(MAX_USER_ID_BYTE_LENGTH)
	if err != nil {
		return false
	}

	joinReq.UserPW, err = reader.ReadBytes(MAX_USER_PW_BYTE_LENGTH)
	if err != nil {
		return false
	}

	joinReq.UserName, err = reader.ReadBytes(MAX_USER_NAME_BYTE_LENGTH)
	return err == nil
}

/* 회원가입 응답 */
func (joinRes JoinResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + 2
	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_JOIN_RES, 0)
	writer.WriteS16(joinRes.ErrorCode)
	return sendBuf, totalSize
}

func (joinRes *JoinResPacket) Decoding(bodyData []byte) bool {
	bodySize := 2
	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)

	var err error
	joinRes.ErrorCode, err = reader.ReadS16()
	return err == nil
}

/* 핑 요청 */
func (pingReq PingReqPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + int16(network.Sizeof(reflect.TypeOf(int8(0))))
	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_PING_REQ, 0)
	writer.WriteS8(pingReq.Ping)
	return sendBuf, totalSize
}

func (pingReq *PingReqPacket) Decoding(bodyData []byte) bool {
	bodySize := network.Sizeof(reflect.TypeOf(int8(0)))
	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	pingReq.Ping, err = reader.ReadS8()
	return err == nil
}

/* 핑 응답 */
func (pingRes PingResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + int16(network.Sizeof(reflect.TypeOf(int8(0))))
	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_PING_RES, 0)
	writer.WriteS8(pingRes.Pong)
	return sendBuf, totalSize
}

func (pingRes *PingResPacket) Decoding(bodyData []byte) bool {
	bodySize := network.Sizeof(reflect.TypeOf(int8(0)))
	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)

	var err error
	pingRes.Pong, err = reader.ReadS8()

	return err == nil
}

func (createNewChatReq *CreateNewChatRoomReqPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + MAX_CHAT_NAME_BYTE_LENGTH + MAX_CHAT_PW_BYTE_LENGTH
	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true) // &sendBuf 하지 않아도 배열이므로 이미 그 자체로 포인터
	EncodingPacketHeader(&writer, totalSize, PACKET_CREATE_NEW_CHATROOM_REQ, 0)
	writer.WriteBytes(createNewChatReq.ChatRoomName[:])
	writer.WriteBytes(createNewChatReq.ChatRoomPW[:])
	return sendBuf, totalSize
}

func (CreateNewChatReq *CreateNewChatRoomReqPacket) Decoding(bodyData []byte) bool {
	BodySize := MAX_CHAT_NAME_BYTE_LENGTH + MAX_CHAT_PW_BYTE_LENGTH
	if len(bodyData) != BodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	CreateNewChatReq.ChatRoomName, err = reader.ReadBytes(MAX_CHAT_NAME_BYTE_LENGTH)
	CreateNewChatReq.ChatRoomPW, err = reader.ReadBytes(MAX_CHAT_PW_BYTE_LENGTH)

	return err == nil
}

func (createNewChatRes *CreateNewChatRoomResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + errorcode.BYTE_OF_ERROR_CODE + network.BYTE_OF_CHATROOM_ID

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)

	EncodingPacketHeader(&writer, totalSize, PACKET_CREATE_NEW_CHATROOM_RES, 0)
	writer.WriteS16(createNewChatRes.ErrorCode)
	writer.WriteS16(createNewChatRes.ChatRoomID)

	return sendBuf, totalSize
}

func (createNewChatRes *CreateNewChatRoomResPacket) Decoding(bodyData []byte) bool {
	BodySize := errorcode.BYTE_OF_ERROR_CODE + network.BYTE_OF_CHATROOM_ID
	if len(bodyData) != int(BodySize) {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	createNewChatRes.ErrorCode, err = reader.ReadS16()
	if err != nil {
		fmt.Println(err)
	}
	createNewChatRes.ChatRoomID, err = reader.ReadS16()
	if err != nil {
		fmt.Println(err)
	}

	return err == nil
}

func (transferMessageReq *TransferMessageReqPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + network.BYTE_OF_CHATROOM_ID + MAX_CHAT_MESSAGE_BYTE_LENGTH

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)

	EncodingPacketHeader(&writer, totalSize, PACKET_TRANSFER_MESSAGE_REQ, 0)
	writer.WriteS16(transferMessageReq.ChatRoomID)
	writer.WriteBytes(transferMessageReq.Message)

	return sendBuf, totalSize
}

func (TransferMessageReq *TransferMessageReqPacket) Decoding(bodyData []byte) bool {
	BodySize := network.BYTE_OF_CHATROOM_ID + MAX_CHAT_MESSAGE_BYTE_LENGTH

	if len(bodyData) != int(BodySize) {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	TransferMessageReq.ChatRoomID, err = reader.ReadS16()
	if err != nil {
		fmt.Println(err)
	}
	TransferMessageReq.Message, err = reader.ReadBytes(MAX_CHAT_MESSAGE_BYTE_LENGTH)
	if err != nil {
		fmt.Println(err)
	}

	return err == nil
}

func (TransferMessageRes *TransferMessageResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + errorcode.BYTE_OF_ERROR_CODE

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)

	EncodingPacketHeader(&writer, totalSize, PACKET_TRANSFER_MESSAGE_RES, 0)
	writer.WriteS16(TransferMessageRes.ErrorCode)

	return sendBuf, totalSize
}

func (TransferMessageRes *TransferMessageResPacket) Decoding(bodyData []byte) bool {
	BodySize := errorcode.BYTE_OF_ERROR_CODE

	if len(bodyData) != int(BodySize) {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	TransferMessageRes.ErrorCode, err = reader.ReadS16()
	if err != nil {
		fmt.Println(err)
	}

	return err == nil
}

func (BroadCastMessage *BroadcastMessagePacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + network.BYTE_OF_MESSAGE_SEQUENCE + MAX_CHAT_MESSAGE_BYTE_LENGTH + MAX_CHAT_TIME_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)

	EncodingPacketHeader(&writer, totalSize, PACKET_BROADCAST_MESSAGE, 0)
	writer.WriteS32(BroadCastMessage.MessageSequence)
	writer.WriteBytes(BroadCastMessage.Message)
	writer.WriteBytes(BroadCastMessage.TimeChat)
	writer.WriteBytes(BroadCastMessage.Sender)

	fmt.Println(sendBuf)
	return sendBuf, totalSize
}

func (BroadCastMessage *BroadcastMessagePacket) Decoding(bodyData []byte) bool {
	bodySize := network.BYTE_OF_MESSAGE_SEQUENCE + MAX_CHAT_MESSAGE_BYTE_LENGTH + MAX_CHAT_TIME_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH

	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	BroadCastMessage.MessageSequence, err = reader.ReadS32()
	if err != nil {
		fmt.Println("BroadCastMessagePacket Reading Err", err)
		return false
	}
	BroadCastMessage.Message, err = reader.ReadBytes(MAX_CHAT_MESSAGE_BYTE_LENGTH)
	if err != nil {
		fmt.Println("BroadCastMessagePacket Reading Err", err)
		return false
	}
	BroadCastMessage.TimeChat, err = reader.ReadBytes(MAX_CHAT_TIME_BYTE_LENGTH)
	if err != nil {
		fmt.Println("BroadCastMessagePacket Reading Err", err)
		return false
	}
	BroadCastMessage.Sender, err = reader.ReadBytes(MAX_USER_NAME_BYTE_LENGTH)
	if err != nil {
		fmt.Println("BroadCastMessagePacket Reading Err", err)
		return false
	}

	return true
}

func (viewAvailableChatRoomReq *ViewAvailableChatRoomReqPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + MAX_USER_ID_BYTE_LENGTH

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)

	EncodingPacketHeader(&writer, totalSize, PACKET_VIEW_AVAILABLE_CHATROOM_REQ, 0)
	writer.WriteBytes(viewAvailableChatRoomReq.UserID)

	return sendBuf, totalSize
}

func (viewAvailableChatRoomReq *ViewAvailableChatRoomReqPacket) Decoding(bodyData []byte) bool {
	bodySize := MAX_USER_ID_BYTE_LENGTH

	if bodySize != len(bodyData) {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	viewAvailableChatRoomReq.UserID, err = reader.ReadBytes(MAX_USER_ID_BYTE_LENGTH)
	if err != nil {
		fmt.Println("ViewAvailableChatRoomReq Reading Err")
		return false
	}

	return true
}

func (chatroominfo *ChatRoom) Encoding() []byte {
	totalSize := network.BYTE_OF_CHATROOM_ID + MAX_CHAT_TIME_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH + MAX_CHAT_NAME_BYTE_LENGTH

	tmpBuf := make([]byte, totalSize)
	writer := network.MakeWrite(tmpBuf, true)

	writer.WriteS16(chatroominfo.ID)
	writer.WriteBytes(chatroominfo.CREATE_TIME)
	writer.WriteBytes(chatroominfo.CREATOR_NAME)
	writer.WriteBytes(chatroominfo.CHATROOM_NAME)

	return tmpBuf
}

func (chatroominfo *ChatRoom) Decoding(bodyData []byte) {
	reader := network.MakeReader(bodyData, true)

	chatroominfo.ID, _ = reader.ReadS16()
	chatroominfo.CREATE_TIME, _ = reader.ReadBytes(MAX_CHAT_TIME_BYTE_LENGTH)
	chatroominfo.CREATOR_NAME, _ = reader.ReadBytes(MAX_USER_NAME_BYTE_LENGTH)
	chatroominfo.CHATROOM_NAME, _ = reader.ReadBytes(MAX_CHAT_NAME_BYTE_LENGTH)
}

func (viewAvailableChatRoomRes *ViewAvailableChatRoomResPacket) Encoding() ([]byte, int16) {
	totalSize := _packetHeaderSize + errorcode.BYTE_OF_ERROR_CODE + 2 + viewAvailableChatRoomRes.Len*(network.BYTE_OF_CHATROOM_ID+MAX_CHAT_TIME_BYTE_LENGTH+MAX_USER_NAME_BYTE_LENGTH+MAX_CHAT_NAME_BYTE_LENGTH)

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)

	EncodingPacketHeader(&writer, totalSize, PACKET_VIEW_AVAILABLE_CHATROOM_RES, 0)
	writer.WriteS16(viewAvailableChatRoomRes.ErrorCode)
	writer.WriteS16(viewAvailableChatRoomRes.Len)
	var i int16
	for i = 0; i < viewAvailableChatRoomRes.Len; i++ {
		writer.WriteBytes(viewAvailableChatRoomRes.ChatRooms[i].Encoding())
	}

	return sendBuf, totalSize
}

func (viewAvailableChatRoomRes *ViewAvailableChatRoomResPacket) Decoding(bodyData []byte, bodySize int16) bool {
	if bodySize != int16(len(bodyData)) {
		return false
	}

	reader := network.MakeReader(bodyData, true)

	var err error
	viewAvailableChatRoomRes.ErrorCode, err = reader.ReadS16()
	if err != nil {
		return false
	}
	viewAvailableChatRoomRes.Len, err = reader.ReadS16()
	if err != nil {
		return false
	}

	var i int16

	viewAvailableChatRoomRes.ChatRooms = make([]ChatRoom, 0)
	for i = 0; i < viewAvailableChatRoomRes.Len; i++ {
		viewAvailableChatRoomRes.ChatRooms = append(viewAvailableChatRoomRes.ChatRooms, ChatRoom{})

		rawData, _ := reader.ReadBytes(2 + MAX_CHAT_TIME_BYTE_LENGTH + MAX_CHAT_NAME_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH)
		viewAvailableChatRoomRes.ChatRooms[i].Decoding(rawData)
	}

	return true
}

func (RenewChatLogReq *RenewChatLogReqPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + network.BYTE_OF_MESSAGE_SEQUENCE

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)

	EncodingPacketHeader(&writer, totalSize, PACKET_RENEW_CHATLOG_REQ, 0)
	writer.WriteS32(RenewChatLogReq.ChatLogEndSequence)

	return sendBuf, totalSize
}

func (RenewChatLogReq *RenewChatLogReqPacket) Decoding(bodyData []byte) bool {
	bodySize := network.BYTE_OF_MESSAGE_SEQUENCE

	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)

	var err error
	RenewChatLogReq.ChatLogEndSequence, err = reader.ReadS32()
	if err != nil {
		fmt.Println("RenewChatLogReq Reading Err")
		return false
	}

	return true
}

func (RenewChatLogRes *RenewChatLogResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _packetHeaderSize + errorcode.BYTE_OF_ERROR_CODE + network.BYTE_OF_MESSAGE_SEQUENCE + MAX_CHAT_MESSAGE_BYTE_LENGTH + MAX_CHAT_TIME_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH

	sendBuf := make([]byte, totalSize)
	writer := network.MakeWrite(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_RENEW_CHATLOG_RES, 0)

	writer.WriteS32(RenewChatLogRes.MessageSequence)
	writer.WriteBytes(RenewChatLogRes.Message)
	writer.WriteBytes(RenewChatLogRes.TimeChat)
	writer.WriteBytes(RenewChatLogRes.Sender)

	return sendBuf, totalSize
}

func (RenewChatLogRes *RenewChatLogResPacket) Decoding(bodyData []byte) bool {
	bodySize := errorcode.BYTE_OF_ERROR_CODE + network.BYTE_OF_MESSAGE_SEQUENCE + MAX_CHAT_MESSAGE_BYTE_LENGTH + MAX_CHAT_TIME_BYTE_LENGTH + MAX_USER_NAME_BYTE_LENGTH

	if len(bodyData) != bodySize {
		return false
	}

	reader := network.MakeReader(bodyData, true)
	var err error
	RenewChatLogRes.ErrorCode, err = reader.ReadS16()
	if err != nil {
		fmt.Println("RenewChatLogRes Reading Err", err)
		return false
	}
	RenewChatLogRes.MessageSequence, err = reader.ReadS32()
	if err != nil {
		fmt.Println("RenewChatLogRes Reading Err", err)
		return false
	}
	RenewChatLogRes.Message, err = reader.ReadBytes(MAX_CHAT_MESSAGE_BYTE_LENGTH)
	if err != nil {
		fmt.Println("RenewChatLogRes Reading Err", err)
		return false
	}
	RenewChatLogRes.TimeChat, err = reader.ReadBytes(MAX_CHAT_TIME_BYTE_LENGTH)
	if err != nil {
		fmt.Println("RenewChatLogRes Reading Err", err)
		return false
	}
	RenewChatLogRes.Sender, err = reader.ReadBytes(MAX_USER_NAME_BYTE_LENGTH)
	if err != nil {
		fmt.Println("RenewChatLogRes Reading Err", err)
		return false
	}

	return true
}
