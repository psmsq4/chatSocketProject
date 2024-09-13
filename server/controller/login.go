package controller

import (
	"bytes"
	"fmt"
	"server/errorcode"
	"server/network"
	"server/protocol"
	"server/service"
)

func ProcessPacketLogin(sessionUniqueId uint64, sessionId int32, bodySize int16, bodyData []byte) {
	var request protocol.LoginReqPacket
	result := (&request).Decoding(bodyData)
	if !result {
		sendLoginResult(sessionUniqueId, sessionId, errorcode.ERROR_CODE_INVALID_REQUEST)
		return
	}

	userID := bytes.Trim(request.UserID[:], "\x00")
	userPW := bytes.Trim(request.UserPW[:], "\x00")

	if len(userID) <= 0 || len(userPW) <= 0 {
		sendLoginResult(sessionUniqueId, sessionId, errorcode.ERROR_CODE_INVALID_REQUEST)
		return
	}

	err := service.LoginAccount(sessionUniqueId, sessionId, request.UserID, userPW)
	/* 로그인 */
	fmt.Println(err)

	sendLoginResult(sessionUniqueId, sessionId, err)
}

func sendLoginResult(sessionUniqueId uint64, sessionId int32, result int16) {
	loginRes := protocol.LoginResPacket{
		ErrorCode: result,
	}

	packet, _ := loginRes.EncodingPacket()
	network.SendToClient(sessionUniqueId, sessionId, packet)
}
