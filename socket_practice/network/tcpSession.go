package network

import "sync"

/*
네트워크에 접속한 클라이언트에 대해 새로운 세션을 추가함
세션 유니크 인덱스에 대해 기존에 존재하는 세션인지 확인하고 아니라면 append함
*/
func (sessionMgr *TcpSessionManager) appendSession(session *TcpSession) bool {
	sessionUniqueId := session.SeqIndex
	sessionIndex := sessionMgr.allocSession()

	if sessionIndex == -1 {
		return false
	}

	_, result := sessionMgr.findSession(sessionUniqueId, sessionIndex)
	/* 이미 세션이 존재한다면 추가로 만들지 않음 */
	if result {
		return false
	}

	/* 기존에 존재하지 않는 새로운 세션이라면 */
	session.Index = sessionIndex
	sessionMgr.SessionMap.Store(sessionUniqueId, session)
	return true
}

func (sessionMgr *TcpSessionManager) removeSession(sessionUniqueId uint64, sessionId int32) {
	sessionMgr.freeSession(sessionId)
	sessionMgr.SessionMap.Delete(sessionUniqueId)
}

/*
딥큐로부터 인덱스 하나를 꺼내옴
*/
func (sessionMgr *TcpSessionManager) allocSession() int32 {
	index := sessionMgr.SessionIndexPool.Shift()

	if index == nil {
		return -1
	}

	return index.(int32)
}

func (sessionMgr *TcpSessionManager) freeSession(sessionIndex int32) {
	sessionMgr.SessionIndexPool.Append(sessionIndex)
}

/*
클라언트 TcpSession을 생성하기 위해서는 딥큐로 관리되는 인덱스를 필요로 함
딥큐 인덱스 풀을 미리 생성한다.
*/
func (sessionMgr *TcpSessionManager) createSessionPool(poolSize int) {
	sessionMgr.SessionIndexPool = NewCappedDeque(poolSize)
	// SessionIndexPool은 *Deque 자료형을 정의되어 있다.

	for index := 0; index < poolSize; index++ {
		sessionMgr.SessionIndexPool.Append(int32(index))
	}
}

func (sessionMgr *TcpSessionManager) findSession(sessionUniqueId uint64, sessionId int32) (*TcpSession, bool) {
	if session, ok := sessionMgr.SessionMap.Load(sessionUniqueId); ok {
		return session.(*TcpSession), true
	}

	return nil, false
}

/*
접속된 클라이언트는 네트워크단에서 TcpSession을 생성한다.
해당 세션을 관리하는 매니저 생성
*/
func createSessionManager(networkFunctor SessionNetworkFunctor) *TcpSessionManager {
	// TcpSessionManager(struct) [PATH]: [./network.go]
	sessionMgr := &TcpSessionManager{
		SessionMap: sync.Map{},
		// sync.Map을 쓰면 알아서 mutex처리
	}

	sessionMgr.NetworkFunctor = networkFunctor
	sessionMgr.createSessionPool(POOL_SIZE)
	// POOL_SIZE :: 요청받을 최대 클라이언트 수

	return sessionMgr
}
