package service

import (
	"bytes"
	"database/sql"
	"fmt"
	"server/errorcode"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type DbConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Protocol string `json:"protocol"`
}

/*
데이터베이스 users 테이블 칼럼
*/
type Account struct {
	ID       int64
	UserID   string
	UserNAME string
	UserPW   string
}

var _mysqlClient *sql.DB
var _mysqlConfig DbConfig

/*
Mysql 초기화 부분
쿠버네티스 업로드시 환경변수로 진행
*/
func InitMysql(dbConfig DbConfig) error {
	_mysqlConfig = dbConfig

	user := dbConfig.User
	password := dbConfig.Password
	protocol := dbConfig.Protocol
	host := dbConfig.Host
	port := dbConfig.Port
	database := dbConfig.Database

	var err error
	addr := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", user, password, protocol, host, port, database)
	_mysqlClient, err = sql.Open("mysql", addr)
	if err != nil {
		return err
	}

	return nil
}

/*
user_id를 기반으로 users 테이블에서 user를 가지고옴
*/
func LoadAccount(strUserId string) (Account, error) {
	stmt, err := _mysqlClient.Prepare("select id, user_id, user_name, user_pw from USERS where user_id = ?")
	if err != nil {
		return Account{}, err
	}

	defer stmt.Close()

	row := stmt.QueryRow(strUserId)
	var Id int64
	var userId, userName, userPw string
	err = row.Scan(&Id, &userId, &userName, &userPw)
	if err != nil && err == sql.ErrNoRows {
		return Account{}, err
	}

	return Account{
		ID:       Id,
		UserID:   userId,
		UserNAME: userName,
		UserPW:   userPw,
	}, nil
}

/*
사용자 로그인
*/
func LoginAccount(sessionUniqueId uint64, sessionId int32, userID, userPW []byte) int16 {
	account, err := LoadAccount(string(bytes.Trim(userID, "\x00")))
	if err != nil {
		return errorcode.ERROR_CODE_MYSQL_ERROR
	}
	fmt.Println(account)

	err = StoreUserInfo(sessionUniqueId, sessionId, userID, true) // Redis에 유저 정보를 추가하는 로직 추가되어야 함.
	if err != nil {
		/* 롤백 */
		RemoveUserInfo(sessionUniqueId)

		// 원래 clientSession 패키지에서 관리하던 작업이지만, 별도 sessionMap에 저장할 필요 없이 Redis에만 저장하면 되므로,
		// memoryDB 함수를 다이렉트로 콜한다. 이렇게 함으로서 상호참조 문제를 해결했다. (2024.08.06)
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.UserPW), userPW)
	if err != nil {
		return errorcode.ERROR_CODE_MYSQL_ERROR
	}
	return errorcode.ERROR_CODE_NONE
}

/*
사용자 추가
*/
func JoinAccount(userID, userPW, userNAME []byte) int16 {
	_, err := LoadAccount(string(userID))
	if err == nil {
		fmt.Println(err)
		return errorcode.ERROR_CODE_MYSQL_ERROR
	}

	hashPW, err := bcrypt.GenerateFromPassword(userPW, bcrypt.DefaultCost)
	fmt.Println(string(hashPW))
	if err != nil {
		fmt.Println(err)
		return errorcode.ERROR_CODE_MYSQL_ERROR
	}

	stmt, err := _mysqlClient.Prepare("insert into USERS (user_id, user_name, user_pw) values(?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return errorcode.ERROR_CODE_MYSQL_ERROR
	}

	defer stmt.Close()

	result, err := stmt.Exec(string(userID), string(userNAME), string(hashPW))
	if err != nil {
		fmt.Println(err)
		return errorcode.ERROR_CODE_MYSQL_ERROR
	}

	/* auto increasement된 값 */
	_, _ = result.LastInsertId()
	return errorcode.ERROR_CODE_NONE
}

/*
채팅방 추가
*/
func CreateNewChatRoom(userID string, chatRoomName []byte, chatRoomPW []byte) (int16, int16) {

	account, err := LoadAccount(string(userID))
	if err != nil {
		fmt.Println("b")
	}
	SurrogateKeyFromUserID := account.ID

	stmt, err := _mysqlClient.Prepare("insert into CHAT_ROOM (CREATOR_ID, CHAT_ROOM_NAME, CHAT_ROOM_PW) values(?, ?, ?)")
	if err != nil {
		fmt.Println("c")
		return -1, errorcode.ERROR_CODE_FAIL_CREATE_NEW_CHATROOM
	}

	defer stmt.Close()

	fmt.Println(SurrogateKeyFromUserID, string(chatRoomName), string(chatRoomPW))

	result, err := stmt.Exec(SurrogateKeyFromUserID, string(chatRoomName), string(chatRoomPW))
	if err != nil {
		fmt.Println("d")
		return -1, errorcode.ERROR_CODE_FAIL_CREATE_NEW_CHATROOM
	}

	ChatRoomID, _ := result.LastInsertId()
	return int16(ChatRoomID), errorcode.ERROR_CODE_NONE
}

/* 메세지 저장 */
func StoreMessageToDB(sessionUniqueID uint64, chatRoomID int16, message string) (int32, string, string, int16) {
	stmt, err := _mysqlClient.Prepare("INSERT INTO MESSAGE_TRANSACTION (CHAT_ROOM_ID, MESSAGE, SENDER) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Println("StoreMessageToDB INSERT QUERY_PREPARE ERROR")
		return -1, "", "", errorcode.ERROR_CODE_FAIL_TRANSFER_MESSAGE
	}
	defer stmt.Close()

	userID := bytes.Trim(LoadUserInfo(sessionUniqueID, 0).UserID, "\x00")
	userAccount, _ := LoadAccount(string(userID))

	result, err := stmt.Exec(strconv.Itoa(int(chatRoomID)), message, strconv.Itoa(int(userAccount.ID)))
	fmt.Println(strconv.Itoa(int(chatRoomID)), " ", message, " ", strconv.Itoa(int(userAccount.ID)))
	if err != nil && result != nil {
		fmt.Println("StoreMessageToDB INSERT EXEC ERROR", err)
		return -1, "", "", errorcode.ERROR_CODE_FAIL_TRANSFER_MESSAGE
	}

	var message_sequence int32
	var time string
	var user_name string

	stmt, err = _mysqlClient.Prepare(`
		SELECT T1.MESSAGE_ID, T1.TIME_CHAT, T2.USER_NAME 
		FROM MESSAGE_TRANSACTION T1
		INNER JOIN USERS T2
			ON T1.SENDER = T2.ID
		WHERE CHAT_ROOM_ID = ? AND MESSAGE = ? AND SENDER = ?`)
	if err != nil {
		fmt.Println("StoreMessageToDB SELECT QUERY_PREPARE ERROR (MESSAGE_ID)")
	}
	row := stmt.QueryRow(strconv.Itoa(int(chatRoomID)), message, strconv.Itoa(int(userAccount.ID)))

	row.Scan(&message_sequence, &time, &user_name)
	// fmt.Println(unsafe.Sizeof(time))

	return message_sequence, time, user_name, errorcode.ERROR_CODE_NONE
}

func InsertAttendanceInformation(chatRoomID int16, userID string, auth string) {
	stmt, err := _mysqlClient.Prepare("INSERT INTO CHAT_ATTENDANCE (CHAT_ROOM_ID, ID, AUTHORITY_CODE) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Println("InsertAttendanceInformation INSERT QUREY_PREPARE ERROR")
	}
	defer stmt.Close()

	stmt.Exec(strconv.Itoa(int(chatRoomID)), userID, auth)
}

func SelectChatRoomInfo() (*sql.Rows, int) {
	stmt, err := _mysqlClient.Prepare(
		`SELECT T1.CHAT_ROOM_ID, T1.CREATE_DATE, T2.USER_NAME, T1.CHAT_ROOM_NAME 
		 FROM CHAT_ROOM T1
		 INNER JOIN USERS T2
		 	ON T1.CREATOR_ID = T2.ID`)
	if err != nil {
		fmt.Println("SelectChatRoomInfo SELECT QUERY_PREPARE ERROR")
		return nil, errorcode.ERROR_CODE_FAIL_VIEW_AVAILABLE_CHATROOM
	}

	results, err := stmt.Query()
	if err != nil {
		fmt.Println("SelectChatRoomInfo SELECT EXEC ERROR")
		return nil, errorcode.ERROR_CODE_FAIL_VIEW_AVAILABLE_CHATROOM
	}

	return results, errorcode.ERROR_CODE_NONE
}
