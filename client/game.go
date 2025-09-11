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

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var _MessageListener chan string
var _sqlite3Client *sql.DB
var _chatRoomID int16
var _chatLogBuffer []string

type LifeGameClient struct {
	PacketChan chan protocol.Packet
}

func InitSqlite3() {
	// 네, sqlite3는 MySQL과 달리 별도의 서버 데몬이 필요하지 않고, 데이터베이스가 단일 파일로 관리됩니다.
	// 아래 코드는 "log.db"라는 파일을 데이터베이스로 사용하도록 엽니다.
	db, err := sql.Open("sqlite3", "log.db")
	if err != nil {
		fmt.Println(err)
	}

	createTable := `
		CREATE TABLE IF NOT EXISTS MESSAGE_LOG (
			MESSAGE_ID INTEGER NOT NULL,
			MESSAGE TEXT NOT NULL,
			TIME_CHAT TEXT NOT NULL,
			USER_NAME TEXT NOT NULL,
			CHAT_ROOM_ID INTEGER NOT NULL,
			PRIMARY KEY (MESSAGE_ID, CHAT_ROOM_ID)
		)
	`

	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Println("테이블 생성 오류:", err)
		return
	} else {
		fmt.Println("MESSAGE_LOG 테이블이 성공적으로 생성되었습니다.")
	}

	_sqlite3Client = db
}

func ConnectLifeGameServer() {

	client := LifeGameClient{}
	/* 패킷헤더 사이즈 정의 */
	protocol.InitPacketHeaderSize()

	/* 패킷 채널 생성 */
	client.PacketChan = make(chan protocol.Packet, 256)

	/* SQLite3 초기화 */
	InitSqlite3()

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

func StoreMessageToDB(messageID int32, chatRoomID int16, message string, timeChat string, userID string) int {
	if _sqlite3Client == nil {
		fmt.Println("sqlite3Client is not initialized")
		return protocol.ERROR_CODE_FAIL_STORE_MESSAGE_TO_DB
	}

	// In Go's database/sql package, calling Begin() is necessary to start a transaction explicitly.
	// This is required if you want to group multiple operations into a single atomic unit,
	// or if you want to control commit/rollback behavior.
	// For a single insert, you could use Exec() directly on the DB handle without Begin(),
	// but using Begin() here ensures that the insert is part of a transaction,
	// which can be important for consistency or future extensibility.
	tx, _ := _sqlite3Client.Begin()
	stmt, err := tx.Prepare("INSERT INTO MESSAGE_LOG (MESSAGE_ID, CHAT_ROOM_ID, MESSAGE, TIME_CHAT, USER_NAME) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return protocol.ERROR_CODE_FAIL_STORE_MESSAGE_TO_DB
	}
	defer stmt.Close()

	_, err = stmt.Exec(messageID, chatRoomID, message, timeChat, userID)
	if err != nil {
		fmt.Println(err)
	}
	tx.Commit()
	return protocol.ERROR_CODE_NONE
}

func ProcessPacketBroadcastMessage(bodySize int16, bodyData []byte) {
	var broadcastMessage protocol.BroadcastMessagePacket

	// time.Sleep(2 * time.Second)

	result := (&broadcastMessage).Decoding(bodyData)
	if !result {
		fmt.Println("Broadcast Chat Message Decoding Failed")
		return
	}

	userName := bytes.Trim(broadcastMessage.UserName, "\x00")
	message := bytes.Trim(broadcastMessage.Message, "\x00")
	time_chat := bytes.Trim(broadcastMessage.TimeChat, "\x00")

	/* 서버로부터 메세지를 받았고.. sqlite3에 이 메세지 정보들을 저장하고.. */
	StoreMessageToDB(broadcastMessage.MessageSequence, _chatRoomID, string(message), string(time_chat), string(userName))
	broadcastMessageFormatString := fmt.Sprintf(" %s  |  %s\n [%d]: %s\n\n", userName, time_chat, broadcastMessage.MessageSequence, message)
	_MessageListener <- broadcastMessageFormatString
}

/*
새로운 채팅방 개설 -> 채팅방 진입 -> 메세지 전송 -> 메세지 Reply from server -> 메세지 저장 to SQLite3 -> 메세지 출력
기존 채팅방 진입
-> 클라이언트 SQLite3에서 ChatRoomID로 MESSAGE_ID 최대값 출력 -> server에 MAX(MESSAGE_ID) + 1 전송
-> 만약 서버에서 MAX(MESSAGE_ID) + 1 이상의 메세지가 존재한다면 클라이언트로 모두 전송
-> 진입 전 chatRoomBuffer 초기화: SQLite3에서 ChatRoomID로 모든 메세지 이력을 가져와 포맷에 맞게 변경 후 chatRoomBuffer에 저장
-> 메세지 Reply from server -> 메세지 저장 to SQLite3 -> 메세지 출력
*/

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

func LoadChatLogBuffer() {
	stmt, err := _sqlite3Client.Prepare("SELECT MESSAGE_ID, MESSAGE, TIME_CHAT, USER_NAME FROM MESSAGE_LOG WHERE CHAT_ROOM_ID = ?")
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(_chatRoomID)
	for rows.Next() {
		var messageID int32
		var message string
		var timeChat string
		var userID string
		err = rows.Scan(&messageID, &message, &timeChat, &userID)
		if err != nil {
			fmt.Println(err)
		}

		/* 채팅 로그는 총 3줄이며, 3줄은 하나의 인덱스에 모두 저장됨. */
		_chatLogBuffer = append(_chatLogBuffer, fmt.Sprintf(" %s  |  %s\n [%d]: %s\n\n", userID, timeChat, messageID, message))
	}
}

func ProcessPacketCreateNewChat(bodySize int16, bodyData []byte) {
	var createNewChatRoomRes protocol.CreateNewChatRoomResPacket
	createNewChatRoomRes.Decoding(bodyData)

	errResp := createNewChatRoomRes.ErrorCode
	_chatRoomID = createNewChatRoomRes.ChatRoomID

	if errResp == protocol.ERROR_CODE_FAIL_CREATE_NEW_CHATROOM {
		fmt.Println("Server: FAIL CREATE NEW CHAT ROOM")
	} else {
		fmt.Println("Success!")
		// 이게 실행되면 go client.PacketProcess()가 묶여서 패킷 처리가 불가능해짐
		// 따라서 UI는 다른 스레드에서 실행시켜야 함.

		/* 채팅로그 버퍼 초기화 */
		_chatLogBuffer = make([]string, 0)
		LoadChatLogBuffer()

		go chatui.DrawFrame(_chatRoomID, _MessageListener, _chatLogBuffer)
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
