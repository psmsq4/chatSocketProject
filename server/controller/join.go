package controller

import (
	"bytes"
	"server/errorcode"
	"server/network"
	"server/protocol"
	"server/service"
)

func ProcessPacketJoin(sessionUniqueId uint64, sessionId int32, bodySize int16, bodyData []byte) {
	var request protocol.JoinReqPacket
	result := (&request).Decoding(bodyData)
	if !result {
		sendJoinResult(sessionUniqueId, sessionId, errorcode.ERROR_CODE_INVALID_REQUEST)
		return
	}

	userID := bytes.Trim(request.UserID[:], "\x00")
	userPW := bytes.Trim(request.UserPW[:], "\x00")
	userNAME := bytes.Trim(request.UserName[:], "\x00")

	if len(userID) <= 0 || len(userPW) <= 0 {
		sendJoinResult(sessionUniqueId, sessionId, errorcode.ERROR_CODE_INVALID_REQUEST)
		return
	}

	/* 회원가입 */
	err := service.JoinAccount(userID, userPW, userNAME)
	sendJoinResult(sessionUniqueId, sessionId, err)
}

func sendJoinResult(sessionUniqueId uint64, sessionId int32, result int16) {
	joinRes := protocol.JoinResPacket{
		ErrorCode: result,
	}

	packet, _ := joinRes.EncodingPacket()
	network.SendToClient(sessionUniqueId, sessionId, packet)
}
