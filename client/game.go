package main

import (
	"bytes"
	chatui "client/chat_ui"
	"client/network"
	"client/protocol"
	"client/send_request"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var _MessageListener chan string
var _ChatLogFileDescriptor *os.File

type LifeGameClient struct {
	PacketChan chan protocol.Packet
}

func ConnectLifeGameServer() {

	client := LifeGameClient{}
	/* 패킷헤더 사이즈 정의 */
	protocol.InitPacketHeaderSize()

	/* 패킷 채널 생성 */
	client.PacketChan = make(chan protocol.Packet, 256)

	snFunctor := network.SessionNetworkFunctor{
		OnConnect:           client.OnConnect,
		OnReceive:           client.OnReceive,
		PacketTotalSizeFunc: network.PacketTotalSize,
		PacketHeaderSize:    protocol.GetPacketHeaderSize(),
	}
	_MessageListener = make(chan string)
	go client.PacketProcess()

	network.ConnectServer(snFunctor)
}

func (client *LifeGameClient) PacketProcess() {
	for {
		select {
		case packet := <-client.PacketChan:
			{
				bodySize := packet.DataSize
				bodyData := packet.Data
				packetId := packet.Id

				if packetId == protocol.PACKET_ID_LOGIN_RES {
					fmt.Println("Login Response")
					ProcessPacketLogin(bodySize, bodyData)
				} else if packetId == protocol.PACKET_ID_JOIN_RES {
					fmt.Println("Join Response")
					ProcessPacketJoin(bodySize, bodyData)
				} else if packetId == protocol.PACKET_CREATE_NEW_CHATROOM_RES {
					fmt.Println("Create Response")
					ProcessPacketCreateNewChat(bodySize, bodyData)
				} else if packetId == protocol.PACKET_TRANSFER_MESSAGE_RES {
					ProcessPacketTransferMessageRes(bodySize, bodyData)
				} else if packetId == protocol.PACKET_BROADCAST_MESSAGE {
					// go Ticker(_MessageListener)
					ProcessPacketBroadcastMessage(bodySize, bodyData)
				} else if packetId == protocol.PACKET_VIEW_AVAILABLE_CHATROOM_RES {
					ProcessPacketViewAvailableChat(bodySize, bodyData)
				} else if packetId == protocol.PACKET_RENEW_CHATLOG_RES {

				}
			}
		}
	}
}

// func Ticker(messageListener chan string) {
// 	var i int
// 	i = 0
// 	for {
// 		time.Sleep(1 * time.Second)
// 		messageListener <- string(i + 97)
// 		i++
// 	}
// }

func ProcessPacketViewAvailableChat(bodySize int16, bodyData []byte) {
	var viewAvailableChatRes protocol.ViewAvailableChatRoomResPacket

	result := (&viewAvailableChatRes).Decoding(bodyData, bodySize)
	if !result {
		fmt.Println("Can't Bring Available Chatting Rooms From Server!")
		return
	}
	var i int16
	for i = 0; i < viewAvailableChatRes.Len; i++ {
		fmt.Printf("%-5d|%s|%s|%s\n", viewAvailableChatRes.ChatRooms[i].ID, string(viewAvailableChatRes.ChatRooms[i].CHATROOM_NAME), string(viewAvailableChatRes.ChatRooms[i].CREATE_TIME), string(viewAvailableChatRes.ChatRooms[i].CREATOR_NAME))
	}
}

func ProcessPacketBroadcastMessage(bodySize int16, bodyData []byte) {
	var broadcastMessage protocol.BroadcastMessagePacket

	// time.Sleep(2 * time.Second)

	result := (&broadcastMessage).Decoding(bodyData)
	if !result {
		fmt.Println("Broadcast Chat Message Decoding Failed")
		return
	}

	sender := bytes.Trim(broadcastMessage.Sender, "\x00")
	message := bytes.Trim(broadcastMessage.Message, "\x00")
	time_chat := bytes.Trim(broadcastMessage.TimeChat, "\x00")

	broadcastMessageFormatString := fmt.Sprintf(" %s  |  %s\n [%d]: %s\n\n", sender, time_chat, broadcastMessage.MessageSequence, message)
	_MessageListener <- broadcastMessageFormatString
}

func ProcessPacketTransferMessageRes(bodySize int16, bodyData []byte) {
	var transferMsgRes protocol.TransferMessageResPacket
	result := (&transferMsgRes).Decoding(bodyData)
	if !result {
		fmt.Println("Transfer Chat Message Decoding Failed")
		return
	}

	if transferMsgRes.ErrorCode == protocol.ERROR_CODE_FAIL_TRANSFER_MESSAGE {
		// _MessageListener <- "Message Transmission Failed."
	} else if transferMsgRes.ErrorCode == protocol.ERROR_CODE_NONE {
		// _MessageListener <- "Success."
	}
}

func ProcessPacketJoin(bodySize int16, bodyData []byte) {
	var joinRes protocol.LoginResPacket
	result := (&joinRes).Decoding(bodyData)
	if !result {
		fmt.Println("Join Failed")
		return
	}

	if joinRes.ErrorCode != protocol.ERROR_CODE_NONE {
		fmt.Println("Join Failed")
		return
	}

	fmt.Println("Join!")
	StartMenuProcess()
}

func ViewAvailableChatRoom() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Println("접속 가능한 채팅방 목록")
	send_request.SendViewAvailableChatRoom(_userID)
}

func ViewUserJoinChatRoom() {

}

func CreateNewChatRoom() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	var chatRoomName string
	var chatRoomPW string

	fmt.Print("채팅방 제목: ")
	fmt.Scanf("%s", &chatRoomName)
	fmt.Print("채팅방 비밀번호(없으면 Enter): ")
	fmt.Scanf("%s", &chatRoomPW)

	send_request.SendCreateNewChatRoomReq(chatRoomName, chatRoomPW)
}

func AfterLoginUserOption() {
	var option int8
	fmt.Printf("***** 환영합니다. %s님. *****\n", _userID)
	fmt.Println("1. 채팅방 조회")
	fmt.Println("2. 기존 채팅방 접속")
	fmt.Println("3. 신규 채팅방 생성")
	fmt.Print("Option Select: ")

	fmt.Scanf("%d", &option)
	switch option {
	case 1:
		ViewAvailableChatRoom()
	case 2:
		ViewUserJoinChatRoom()
	case 3:
		CreateNewChatRoom()
	}
}

func ProcessPacketLogin(bodySize int16, bodyData []byte) {
	var loginRes protocol.LoginResPacket
	result := (&loginRes).Decoding(bodyData)
	if !result {
		fmt.Println("Login Failed")
		return
	}

	if loginRes.ErrorCode != protocol.ERROR_CODE_NONE {
		fmt.Println("Login Failed")
		return
	}

	fmt.Println("Login!")
	AfterLoginUserOption()
}

func SendPing() {
	for {
		return
		time.Sleep(1 * time.Millisecond)
		pingReq := protocol.PingReqPacket{
			Ping: protocol.PING,
		}

		packet, packetSize := pingReq.EncodingPacket()
		network.SendToServer(packet, packetSize)
	}
}

var _userID string

func StartMenuProcess() {
	var option int8
	fmt.Println("***** LOGIN MENU *****")
	fmt.Println("1. 로그인(LOGIN)")
	fmt.Println("2. 가입(JOIN)")
	fmt.Print("Option Select: ")
	fmt.Scanf("%d", &option)

	var userID string
	var userPW string
	var userNAME string
	if option == 1 {
		fmt.Print("USER ID: ")
		fmt.Scanf("%s", &userID)
		fmt.Print("USER PW: ")
		fmt.Scanf("%s", &userPW)

		fmt.Println("Try Logining...")
		send_request.SendLogin(userID, userPW) // -> Server로 전송 -> Response 수신 -> packetChan -> ProcessPakcetLogin 실행 -> AfterLoginUserOption 실행
		_userID = userID
	} else {
		fmt.Print("NEW USER ID: ")
		fmt.Scanf("%s", &userID)
		fmt.Print("NEW USER PW: ")
		fmt.Scanf("%s", &userPW)
		fmt.Print("NEW USER NAME: ")
		fmt.Scanf("%s", &userNAME)
		fmt.Println("Try Joining...")

		send_request.SendJoin(userID, userPW, userNAME)
	}
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ProcessPacketCreateNewChat(bodySize int16, bodyData []byte) {
	var createNewChatRoomRes protocol.CreateNewChatRoomResPacket
	createNewChatRoomRes.Decoding(bodyData)

	errResp := createNewChatRoomRes.ErrorCode
	chatRoomID := createNewChatRoomRes.ChatRoomID

	if errResp == protocol.ERROR_CODE_FAIL_CREATE_NEW_CHATROOM {
		fmt.Println("Server: FAIL CREATE NEW CHAT ROOM")
	} else {
		fmt.Println("Success!")
		// 이게 실행되면 go client.PacketProcess()가 묶여서 패킷 처리가 불가능해짐
		// 따라서 UI는 다른 스레드에서 실행시켜야 함.
		filename := fmt.Sprintf("chatlog_%d", chatRoomID)
		var err error
		_ChatLogFileDescriptor, err = os.Create(filename)
		if err != nil {
			panic(err)
		}
		go chatui.DrawFrame(chatRoomID, _MessageListener, _ChatLogFileDescriptor)
	}
}

func (client *LifeGameClient) OnConnect() {
	fmt.Println("Connect Chatting Server!")

	StartMenuProcess()
}

func (client *LifeGameClient) OnReceive(packetData []byte) {
	packetID := protocol.PeekPacketID(packetData)
	bodySize, packetBody := protocol.PeekPacketBody(packetData)

	packet := protocol.Packet{
		Id:       packetID,
		DataSize: bodySize,
		Data:     make([]byte, bodySize),
	}

	copy(packet.Data, packetBody)
	client.PacketChan <- packet
}
