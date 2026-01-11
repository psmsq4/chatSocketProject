package main

import (
	"bytes"
	"client/network"
	"client/protocol"
	"client/send_request"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"database/sql"

	ui "client/chat_ui"

	gc "github.com/gbin/goncurses"
	_ "github.com/mattn/go-sqlite3"
)

var (
	_MessageListener  chan string
	_UIListener       = make(chan string, 1)
	_ChatListListener = make(chan protocol.ViewAvailableChatRoomResPacket)
	_sqlite3Client    *sql.DB
	_chatRoomID       int16
	_chatRoomName     string
	_chatLogBuffer    []string

	isGoncursesInitialized = false
	goncursesMutex         sync.Mutex
	globalStdscr           *gc.Window

	_userID string
)

type LifeGameClient struct {
	PacketChan chan protocol.Packet
}

func initGoncursesOnce() error {
	goncursesMutex.Lock()
	defer goncursesMutex.Unlock()

	if !isGoncursesInitialized {
		gc.CBreak(true)
		gc.Cursor(0)

		stdscr, err := gc.Init()
		if err != nil {
			return err
		}

		stdscr.ScrollOk(true)

		if !gc.HasColors() {
			return fmt.Errorf("This requires a colour capable termial")
		}

		if err := gc.StartColor(); err != nil {
			return err
		}

		globalStdscr = stdscr
		isGoncursesInitialized = true
	}

	return nil
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
	}

	_sqlite3Client = db
}

func DrawGUI() { // goroutine으로 돌아감.
	connectionResult := <-_MessageListener // network.ConnectServer()로부터 message가 올 때까지 대기

	if strings.Compare(connectionResult, "success") == 0 {
		if err := initGoncursesOnce(); err != nil { // goncurses library 초기화
			log.Fatal(err)
		}

		/* 초기화면 (로그인/회원가입)
		   -> 로그인 시, 메인화면(채팅방 신규/채팅방 접속/채팅방 조회)
		      -> 채팅방 신규 등록 -> 신규 채팅방 바로 접속
			  -> 채팅방 접속 -> 채팅방 목록 중 하나 선택 -> 채팅방 접속
		   -> 가입 시, 초기화면(로그인/회원가입)

		* 초기화면 & 로그인 화면 & 회원가입 화면 & 메인화면 & 채팅방 등록 & 채팅방 화면 & 채팅방 접속시도 화면 */
		/* 화면 전환을 어떻게 해야하지?
		   프로시저에 넣어놓고 프로시저 종료 시 그려진 윈도우들은 모두 defer 될 것이다.
		   -> 여기에 stdscr.Clear()를 후속타로 날려주고 다른 프로시저를 호출하면 될 것 같다.
		*/
		for {
			switch <-_UIListener {
			case "Sign In": // 로그인
				id, pw := ui.DrawLogin(globalStdscr, _UIListener)
				send_request.SendLogin(id, pw)
				_userID = id
				globalStdscr.Refresh()
			case "BeforeLogin": // 초기화면
				ui.DrawBeforeLogin(globalStdscr, _UIListener)
				globalStdscr.Refresh()
			case "Create Account": // 회원가입
			case "Exit": // 나가기
				fallthrough
			case "AfterLogin": // 로그인 이후 메인화면
				ui.DrawAfterLogin(globalStdscr, _UIListener, _userID)
				globalStdscr.Refresh()
			case "NewChat": // 새 채팅 등록
				chatname, chatpw := ui.DrawNewChat(globalStdscr, _UIListener)
				_chatRoomName = chatname
				send_request.SendCreateNewChatRoomReq(chatname, chatpw)
				globalStdscr.Refresh()
			case "OldChat": // 채팅방 접속 시도
				send_request.SendConnAvailableChatRoom(_userID, _chatRoomID)
				globalStdscr.Refresh()
			case "ChatList": // 채팅 목록 조회
				/* Requirement
				   - 전체 채팅 개수
				   - 채팅방 별 참여인원
				   - 채팅방 별 제목
				   - 채팅방 별 방장ID
				   -  */
				send_request.SendViewAvailableChatRoom(_userID)
				_chatRoomID, _chatRoomName = ui.DrawChatList(globalStdscr, _ChatListListener, _UIListener)
				globalStdscr.Refresh()
				/* Todo: 서버로부터 채팅 목록을 받아오기 */
				/* Todo: 받아온 채팅 목록을 DrawChatList()에 넘겨주기 */
				/* Todo: DrawChatList()로부터 유저가 선택한 OldChat 받아오기 */
			case "InChat":
				ui.DrawChatRoom(globalStdscr, _chatRoomName, _chatRoomID, _chatLogBuffer, _MessageListener, _UIListener)
				globalStdscr.Refresh()
			}
		}
	} else if strings.Compare(connectionResult, "fail") == 0 {
		return
	}
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
	_MessageListener = make(chan string, 30)
	go client.PacketProcess()
	go DrawGUI()

	network.ConnectServer(snFunctor, _MessageListener, _UIListener)
}

func (client *LifeGameClient) PacketProcess() {
	for packet := range client.PacketChan {
		bodySize := packet.DataSize
		bodyData := packet.Data
		packetId := packet.Id

		switch packetId {
		case protocol.PACKET_ID_LOGIN_RES:
			ProcessPacketLogin(bodySize, bodyData)
		case protocol.PACKET_ID_JOIN_RES:
			ProcessPacketJoin(bodySize, bodyData)
		case protocol.PACKET_CREATE_NEW_CHATROOM_RES:
			ProcessPacketCreateNewChat(bodySize, bodyData)
		case protocol.PACKET_TRANSFER_MESSAGE_RES:
			ProcessPacketTransferMessageRes(bodySize, bodyData)
		case protocol.PACKET_BROADCAST_MESSAGE:
			// go Ticker(_MessageListener)
			ProcessPacketBroadcastMessage(bodySize, bodyData)
		case protocol.PACKET_VIEW_AVAILABLE_CHATROOM_RES:
			ProcessPacketViewAvailableChat(bodySize, bodyData)
		case protocol.PACKET_CONN_AVAILABLE_CHATROOM_RES:
			ProcessPacketConnAvailableChat(bodySize, bodyData)
		case protocol.PACKET_RENEW_CHATLOG_RES:
			ProcessPacketRenewChatLog(bodySize, bodyData)
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

func ProcessPacketRenewChatLog(bodySize int16, bodyData []byte) {
	var renewChatLogRes protocol.RenewChatLogResPacket

	result := renewChatLogRes.Decoding(bodyData)
	if !result {
		fmt.Println("RenewChatLogRes Decoding Fail")
		return
	}

	switch renewChatLogRes.ErrorCode {
	case protocol.ERROR_CODE_RENEW_LAST_MESSAGE:
		/* 서버로 부터 마지막 메세지를 받음. */
		/* Todo: 종료 -> DrawUI 쪽으로 이벤트를 보내 이때까지 수신만 메시지를 그리도록 함. */
		/* └ ProcessConnAvailableChat에서 이미 응답을 받자마자 UIListener <- "InChat"을 그리도록 함. */
	case protocol.ERROR_CODE_NONE:
		/* 서버로 부터 메시지 정상 수신 */
		/* Todo: 받은 메시지를 Sqlite에 저장해야함. */
		message := string(bytes.Trim(renewChatLogRes.Message, "\x00"))
		timestamp := string(bytes.Trim(renewChatLogRes.TimeChat, "\x00"))
		user_name := string(bytes.Trim(renewChatLogRes.UserName, "\x00"))
		result := StoreMessageToDB(
			renewChatLogRes.MessageSequence,
			renewChatLogRes.ChatRoomID,
			message,
			timestamp,
			user_name,
		)
		if result != protocol.ERROR_CODE_NONE {
			fmt.Println("Save Fail")
		}

		RenewChatLogFormatString := fmt.Sprintf(" %s  |  %s\n [%d]: %s\n\n", user_name, timestamp, renewChatLogRes.MessageSequence, message)
		/* 개선사항: _MessageListener를 Buffered(30) Channel로 변경 */
		/* └ 이로 인해, 송신자는 Blocking 없이 서버에 계속 Request를 날릴 수 있음. */
		_MessageListener <- RenewChatLogFormatString

		/* Todo: 다음 메시지를 요청해야함. */
		send_request.SendRenewChatLogReqPacket(renewChatLogRes.MessageSequence, renewChatLogRes.ChatRoomID)
	}
}

func FetchExistChatLogFromDB(chatroom_id int16, MessageListener chan string) {
	stmt, err := _sqlite3Client.Prepare(`
	SELECT USER_NAME, TIME_CHAT, MESSAGE_ID, MESSAGE
	FROM MESSAGE_LOG
	WHERE CHAT_ROOM_ID = ?`)
	if err != nil {
		fmt.Println("FetechExistChatLogFromDB QUERY_PREPARE ERROR")
		return
	}
	defer stmt.Close()

	query_result, err := stmt.Query(chatroom_id)
	if err != nil {
		fmt.Println("FetechExistChatLogFromDB QUERY ERROR")
		return
	}
	defer query_result.Close()

	var fetched_user_name string
	var fetched_time_chat string
	var fetched_message_id int
	var fetched_message string
	for query_result.Next() {
		query_result.Scan(&fetched_user_name, &fetched_time_chat, &fetched_message_id, &fetched_message)

		ExistChatLogFormatString := fmt.Sprintf(" %s  |  %s\n [%d]: %s\n\n", fetched_user_name, fetched_time_chat, fetched_message_id, fetched_message)
		MessageListener <- ExistChatLogFormatString
	}
}

func LastMIDFromCID(chatroom_id int16) int32 {
	stmt, err := _sqlite3Client.Prepare(`
	SELECT MESSAGE_ID 
	FROM MESSAGE_LOG
	WHERE CHAT_ROOM_ID = ?
	ORDER 
		BY MESSAGE_ID DESC
	LIMIT 1
	`)
	if err != nil {
		fmt.Println("LastMIDFromCID QUERY_PREPARE ERROR")
		return -1
	}
	// defer 안 해주면 StoreMessageToDB에서 lock 오류 발생
	defer stmt.Close()

	result, err := stmt.Query(chatroom_id)
	if err != nil {
		fmt.Println("LastMIDFromCID QUERY ERROR")
		return -1
	}
	// defer 안 해주면 StoreMessageToDB에서 lock 오류 발생
	defer result.Close()

	var last_message_id int32
	if result.Next() {
		result.Scan(&last_message_id)
		return last_message_id
	}

	return -1
}

func ProcessPacketConnAvailableChat(bodySize int16, bodyData []byte) {
	var connAvailableChatRes protocol.ConnAvailableChatRoomResPacket

	result := connAvailableChatRes.Decoding(bodyData)
	if !result {
		fmt.Println("ConnAvailableChatRes Decoding Fail")
		return
	}

	// 일단 DrawUI goroutine은 DrawChatRoom에서 (<- MessageListener)에 대기 시켜놓아야 함.
	// _UIListener에 먼저 InChat을 넣고 MessageListener의 수신자를 지정해놓아야 Block이 안 됨.
	switch connAvailableChatRes.ErrorCode {
	case protocol.ERROR_CODE_ALREADY_PART_IN_CONN:
		_UIListener <- "InChat"
		/* [Yet]Todo: sqlite에서 일단 갖고 있는 채팅 내역을 한 개씩 _MessageListener로 쏴주기 */
		FetchExistChatLogFromDB(connAvailableChatRes.ChatRoomID, _MessageListener)

		/* [Done]Todo: sqlite에서 요청한 chatroom_id와 연결된 message_id의 last값을 RenewChatLogReq 형태로 서버에 요청 */
		last_message_id := LastMIDFromCID(connAvailableChatRes.ChatRoomID)
		send_request.SendRenewChatLogReqPacket(last_message_id, connAvailableChatRes.ChatRoomID)
	case protocol.ERROR_CODE_NONE:
		_UIListener <- "InChat"
		send_request.SendRenewChatLogReqPacket(-1, connAvailableChatRes.ChatRoomID)
	case protocol.ERROR_CODE_FAIL_CONN_AVAILABLE_CHATROOM:
		panic("CONNECT Chatting Room Fail..")
	}
}

func ProcessPacketViewAvailableChat(bodySize int16, bodyData []byte) {
	var viewAvailableChatRes protocol.ViewAvailableChatRoomResPacket

	result := (&viewAvailableChatRes).Decoding(bodyData, bodySize)
	if !result {
		fmt.Println("viewAvailableChatRes Decoding Fail")
		return
	}

	_ChatListListener <- viewAvailableChatRes
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
	tx, err := _sqlite3Client.Begin()
	if err != nil {
		fmt.Println("Begin error:", err)
		return protocol.ERROR_CODE_FAIL_STORE_MESSAGE_TO_DB
	}

	stmt, err := tx.Prepare("INSERT INTO MESSAGE_LOG (MESSAGE_ID, CHAT_ROOM_ID, MESSAGE, TIME_CHAT, USER_NAME) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Println("Prepare error:", err)
		tx.Rollback()
		return protocol.ERROR_CODE_FAIL_STORE_MESSAGE_TO_DB
	}
	defer stmt.Close()

	_, err = stmt.Exec(messageID, chatRoomID, message, timeChat, userID)
	if err != nil {
		fmt.Println("Exec error:", err)
		tx.Rollback()
		return protocol.ERROR_CODE_FAIL_STORE_MESSAGE_TO_DB
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("Commit error:", err)
		return protocol.ERROR_CODE_FAIL_STORE_MESSAGE_TO_DB
	}

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
		return
	}

	if joinRes.ErrorCode != protocol.ERROR_CODE_NONE {
		return
	}

	_UIListener <- "BeforeLogin"
}

func ViewUserJoinChatRoom() {

}

func ProcessPacketLogin(bodySize int16, bodyData []byte) {
	var loginRes protocol.LoginResPacket
	result := (&loginRes).Decoding(bodyData)
	if !result {
		return
	}

	if loginRes.ErrorCode != protocol.ERROR_CODE_NONE {
		return
	}

	_UIListener <- "AfterLogin"
	// AfterLoginUserOption()
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

func LoadChatLogBuffer() {
	stmt, err := _sqlite3Client.Prepare("SELECT MESSAGE_ID, MESSAGE, TIME_CHAT, USER_NAME FROM MESSAGE_LOG WHERE CHAT_ROOM_ID = ?")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(_chatRoomID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

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
		globalStdscr.Print("Server: FAIL CREATE NEW CHAT ROOM")
		globalStdscr.Refresh()
	} else {
		// 이게 실행되면 go client.PacketProcess()가 묶여서 패킷 처리가 불가능해짐
		// 따라서 UI는 다른 스레드에서 실행시켜야 함.

		/* 채팅로그 버퍼 초기화 */
		_chatLogBuffer = make([]string, 0)
		LoadChatLogBuffer()

		_UIListener <- "InChat"
	}
}

func (client *LifeGameClient) OnConnect() {
	fmt.Println("Connect Chatting Server!")
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
