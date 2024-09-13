package protocol

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

const (
	PACKET_TYPE_NORMAL   = 0
	PACKET_TYPE_COMPRESS = 1
	PACKET_TYPE_SECURE   = 2
)

const (
	MAX_USER_ID_BYTE_LENGTH      = 16
	MAX_USER_PW_BYTE_LENGTH      = 16
	MAX_CHAT_MESSAGE_BYTE_LENGTH = 126
)

type Header struct {
	TotalSize  int16 // 2바이트
	ID         int16 // 2바이트
	PacketType int8  // 1바이트
} // 5바이트

type Packet struct {
	UserSessionIndex       int32  // 서버마다 sessionIndex가 같을 수 있다.
	UserSessionUniqueIndex uint64 // 그래서 구별을 위해 UniqueIndex를 생성한다. (Redis_Single Thread를 통해 uniqueIndex를 부여받는다.)
	// redis는 원자적 처리 HOWEVER, 모든 thread를 원자적으로 처리하는 것이라, 명령을 원자적으로 처리함.
	// 그래서 redis에 로아 스크립트로 날리면, 원자성이 깨지는 것을 막을 수 있다.
	Id       int16
	DataSize int16
	Data     []byte
}

var _packetHeaderSize int16 // 전역변수

func InitPacketHeaderSize() {
	_packetHeaderSize = PacketHeaderSize() // 전역변수
} // 패킷헤더 사이즈는 고정되어 있지 않기 때문에 이렇게 함수를 정의한다. ()

/*
전체 패킷에서 총 크기를 제외한 다음 2바이트를 꺼내옴
*/
func PeekPacketID(rawData []byte) int16 {
	packetID := binary.LittleEndian.Uint16(rawData[2:])
	return int16(packetID) // 형변환함으로써 2바이트 만큼 알아서 짤리나?
}

/*
전체 패킷에서 헤더를 뺀 만큼 바디로 지정
*/
func PeekPacketBody(rawData []byte) (int16, []byte) {
	headerSize := _packetHeaderSize
	totalSize := int16(binary.LittleEndian.Uint16(rawData))
	// 뭔소리고
	bodySize := totalSize - headerSize

	if bodySize > 0 {
		return bodySize, rawData[headerSize:]
	}

	return bodySize, []byte{}
}

/*
패킷헤더의 크기를 사전에 구함
*/
func PacketHeaderSize() int16 {
	var header Header // 5바이트 짜리 구조체 변수

	hSize := unsafe.Sizeof(reflect.TypeOf(header))
	// unsafe package는 뭐하는 패키지?
	// reflect package는 뭐하는 패키지?

	return (int16)(hSize)
	// 굳이 형변환을 하는 이유는?
}

type LoginReqPacket struct {
	UserID []byte
	UserPW []byte
}

func (loginReq LoginReqPacket) EncodingPacket() {

}

func (loginReq *LoginReqPacket) Decoding() {

}

type LoginResPacket struct {
}

func (loginRes LoginResPacket) EncodingPacket() {

}

func (loginRes *LoginResPacket) Decoding() {

}

type JoinReqPacket struct {
}

func (joinReq JoinReqPacket) EncodingPacket() {

}

func (JoinReq *JoinReqPacket) Decoding() {

}

type PingReqPacket struct {
}

func (pingReq PingReqPacket) EncodingPacket() {

}

func (pingReq *PingReqPacket) Decoding() {

}

type PingResPacket struct {
	Pong int8
}

func (pingRes PingResPacket) EncodingPacket() {

}

func (pingRes *PingResPacket) Decoding() {

}
