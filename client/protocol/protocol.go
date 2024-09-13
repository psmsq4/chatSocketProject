package protocol

type Header struct {
	TotalSize  int16
	ID         int16
	PacketType int8
}

type Packet struct {
	UserSessionIndex       int32
	UserSessionUniqueIndex uint64
	Id                     int16
	DataSize               int16
	Data                   []byte
}

type LoginReqPacket struct {
	UserID []byte
	UserPW []byte
}

/*
로그인 응답
세션ID를 쿠키로써 갖고 있도록 서버에서 생성한 SessionId를 보낸다
*/
type LoginResPacket struct {
	ErrorCode int16
}

type JoinReqPacket struct {
	UserID   []byte
	UserPW   []byte
	UserName []byte
}

type JoinResPacket struct {
	ErrorCode int16
}

type PingReqPacket struct {
	Ping int8
}

type PingResPacket struct {
	Pong int8
}

type CreateNewChatRoomReqPacket struct {
	ChatRoomName []byte
	ChatRoomPW   []byte
}

// 서버가 인코딩하여 클라이언트로 송신하고
// 클라이언트가 수신받아 디코딩하는 구조체이다.
type CreateNewChatRoomResPacket struct {
	ErrorCode  int16
	ChatRoomID int16
}

type ViewAvailableChatRoomReqPacket struct {
	UserID []byte
}

type ChatRoom struct {
	ID            int16
	CREATE_TIME   []byte
	CREATOR_NAME  []byte
	CHATROOM_NAME []byte
} // ViewAvailableChatRoomResPacket에 종속되는 구조체

type ViewAvailableChatRoomResPacket struct {
	ErrorCode int16
	Len       int16
	ChatRooms []ChatRoom
}

/*

 */

type ViewUserJoinChatRoomReqPacket struct {
	UserID []byte
}

type ViewUserJoinChatRoomResPacket struct {
	ErrorCode int16
}

type TransferMessageReqPacket struct {
	ChatRoomID int16
	Message    []byte
}

type TransferMessageResPacket struct {
	ErrorCode int16
} // 송신자한테만 발송하는 패킷

type BroadcastMessagePacket struct {
	MessageSequence int32
	Message         []byte
	TimeChat        []byte // 16 byte (static)
	Sender          []byte
} // req, res 없이 udp처럼 일방적인 패킷으로 구현

type RenewChatLogReqPacket struct {
	ChatLogEndSequence int32
} // 재접속 또는 신규접속 시 채팅로그를 갱신하기 위한 요청 패킷

type RenewChatLogResPacket struct {
	ErrorCode       int16
	MessageSequence int32
	Message         []byte
	TimeChat        []byte
	Sender          []byte
} // 상기 패킷에 대한 응답 패킷
