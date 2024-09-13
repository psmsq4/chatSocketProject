package clientsession

import (
	"fmt"
	"server/service"
	"sync"
)

type SessionManager struct {
	SessionMap sync.Map
}

var _sessionManager *SessionManager

func Init() {
	_sessionManager = createSessionManager()
}

func createSessionManager() *SessionManager {
	sessionMgr := &SessionManager{
		SessionMap: sync.Map{},
	}

	return sessionMgr
}

/*
Redis에 저장
user_[unique_id]
Sequence :: 클라이언트 최초 접속요청 -> conn := l.Accept -> TcpSession { conn: conn } ->
*/
func AddSession(sessionUniqueId uint64, sessionId int32, userID []byte) bool {
	_, result := FindSession(sessionUniqueId)
	if result {
		return false
	}

	// session := &ClientSession{
	// 	SessionUniqueID: sessionUniqueId,
	// 	SessionID:       sessionId,
	// 	UserID:          userID,
	// 	IsAuth:          true,
	// }

	// _sessionManager.SessionMap.Store(sessionUniqueId, session)
	// -> 이미 network.StartServerBlock()에서 TcpSession 생성 후, appendSession 호출해서 TcpSessionManager에 넣어놓았는 데,
	//    굳이 또 넣어야 할 필요가 있는가?
	fmt.Println(_sessionManager)
	err := service.StoreUserInfo(sessionUniqueId, sessionId, userID, true)
	if err != nil {
		/* 롤백 */
		RemoveSession(sessionUniqueId)
		return false
	}

	return true
}

func FindSession(sessionUniqueId uint64) (*ClientSession, bool) {
	if session, ok := _sessionManager.SessionMap.Load(sessionUniqueId); ok {
		return session.(*ClientSession), true
	}

	return nil, false
}

func RemoveSession(sessionUniqueId uint64) bool {
	err := service.RemoveUserInfo(sessionUniqueId)
	if err != nil {
		return false
	}

	_sessionManager.SessionMap.Delete(sessionUniqueId)
	return true
}
