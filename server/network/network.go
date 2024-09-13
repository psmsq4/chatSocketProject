package network

import (
	"fmt"
	"net"
	"server/service"
	"sync"
	"sync/atomic"
)

type NetConfig struct {
	BindAdress string `json:"bind_address"`
	Port       int    `json:"port"`
}

/*
* 	RING BUFFER?
 */
type SessionNetworkFunctor struct {
	OnConnect           func(sessionUniqueId uint64, sessionId int32)
	OnClose             func(sessionUniqueId uint64, sessionId int32)
	OnReceive           func(sessionUniqueId uint64, sessionId int32, packet []byte)
	PacketTotalSizeFunc func([]byte) int16
	PacketHeaderSize    int16
}

type TcpSession struct {
	SeqIndex uint64 /* 모든 서버 통틀어서 유니크한 값 */
	Index    int32  /* 특정 서버에서 유니크한 값 */
	Conn     net.Conn

	NetworkFunctor SessionNetworkFunctor
}

/*
NetworkFunctor 클라이언트에 대한 콜백 함수 모음
sync.Map의 경우 여러 고루틴이 접근 보장
SessionIndexPool 클라이언트 접속시 세션에 대한 ID - DeepQueue로 관리 - Lock & Unlock
*/
type TcpSessionManager struct {
	NetworkFunctor SessionNetworkFunctor
	SessionMap     sync.Map
	/* Queue */
	SessionIndexPool *Deque
}

func StartServerBlock(netConfig NetConfig, networkFunctor SessionNetworkFunctor) {
	fmt.Printf("TCP Server Start on %s:%d port\n", netConfig.BindAdress, netConfig.Port)
	/* TcpSession 매니저 생성 */
	_tcpSessionManager = createSessionManager(networkFunctor)
	address := fmt.Sprintf(":%d", netConfig.Port)
	fmt.Println(address)
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		/*
			Create TCP Session
			Index : Allocate from Session Index Pool (Deque)
		*/
		///var seqNumber uint64 = 0
		seqNumber := service.GetUniqueSessionId()
		if seqNumber == 0 {
			fmt.Println("Failed Alloc SessionUniqueId")
			continue
		}
		newTcpSession := &TcpSession{
			SeqIndex:       seqNumber,
			Conn:           conn,
			NetworkFunctor: _tcpSessionManager.NetworkFunctor,
		}

		/*
			Append TCP Session at TCP Session Manager
		*/
		ok := _tcpSessionManager.appendSession(newTcpSession)
		fmt.Println(_tcpSessionManager.SessionMap)
		if !ok {
			fmt.Println("Not Enough Index Pool")
			continue
		}

		/*
			Handle Tcp Read
		*/
		go newTcpSession.handleTcpRead()
	}
}

/*
링버퍼 구조 유지
PACKET HEADER STRUCTURE
*/
func (session *TcpSession) handleTcpRead() {
	/* 클라이언트의 접속 처리 */
	session.NetworkFunctor.OnConnect(session.SeqIndex, session.Index)

	for {
		var startRecvPos int16 = 0
		var result int
		recvBuff := make([]byte, MAX_BUFFER)

		for {
			recvBytes, err := session.Conn.Read(recvBuff[startRecvPos:])
			/*
				클라이언트 연결 끊어짐
			*/
			if err != nil {
				session.closeProcess()
				return
			}

			/*
				이 문장은 필요없을 듯
			*/
			/*
				if recvBytes < PACKET_HEADER_SIZE {
					session.closeProcess()r
					return
				}
			*/

			/*
				링버퍼 구조를 이루고 있음 startRecvPos는 패킷을 만들기 위한 길이까지만 설정하고
				남는 패킷은 나음에 startRecvPos부터 시작
			*/
			readAbleByte := int16(startRecvPos) + int16(recvBytes)
			startRecvPos, result = session.makePacket(readAbleByte, recvBuff)
			if result != NET_ERROR_NONE {
				session.closeProcess()
				return
			}
		}
	}
}

/*
makePacket 함수를 호출하여 아이디와 패킷길이, 패킷길이만큼의 데이터 바이트를 넘겨줌
그럼 아이디를 기반으로 데이터바이트를 디코딩함

makePacket함수는 처리할 수 있는 패킷만 처리하고 요구 처리 바이트보다 적게 남을 경우 버퍼의 시작으로 옮긴다음
버퍼의 시작 위치를 설정하여 다음패킷 수신시에 받아서 같이 처리한다.

예를들어 33바이트가 왔다고 가정한다.
패킷의 헤더는 5바이트고 바디는 10바이트이다. 총 15바이트이다.

15 / 15 처리하고 남는 바이트는 3바이트이다.
3바이트는 패킷의 최소 요구바이트 15바이트 이하이기에 처리할 수 없다.

그렇기에 30번째위치부터 3바이트 크기만큼 recvBuff의 시작지점으로 옮긴다.
그리고 다음 패킷의 수신 지점 (startRecvPos)의 값을 0이 아닌 3(readAbleByte)으로 설정한다.
그럼 다음 패킷 수신시 3바이트를 포함해서 파싱이 가능해진다!

makePacket은 수신받은 바이트 배열을 패킷 구조체로 디코딩하는 것을 의미
*/
func (session *TcpSession) makePacket(readAbleByte int16, recvBuff []byte) (int16, int) {
	sessionUniqueId := session.SeqIndex
	sessionId := session.Index

	var startRecvPos int16 = 0
	var readPos int16 = 0

	packetHeaderSize := session.NetworkFunctor.PacketHeaderSize // protocol.GetPacketHeaderSize()
	PacketTotalSizeFunc := session.NetworkFunctor.PacketTotalSizeFunc
	// session.NetworkFunctor 필드는 game.go에서 snFunctor로서 최초로 초기화가 완료된 상태로
	// network.StartServerBlock() 함수에서 매개변수로 전달받고 이를 newTcpSession을 생성할 때 한 번더 초기화한다.

	for {
		if readAbleByte < packetHeaderSize {
			break
		}

		requireDataSize := PacketTotalSizeFunc(recvBuff[readPos:])

		/*
			현재 읽을 수 있는 바이트의 수가 요구바이트보다 작다면 다음번에 읽는다.
		*/
		if readAbleByte < requireDataSize {
			break
		}

		/*
			현재 읽을 수 있는 바이트의 수가 최대 읽을 수 있는 바이트보다 크다면
		*/
		if readAbleByte > MAX_PACKET_SIZE {
			return startRecvPos, NET_ERROR_TOO_LARGE_PACKET
		}

		ltvPacket := recvBuff[readPos:(readPos + requireDataSize)]
		readPos += requireDataSize
		readAbleByte -= requireDataSize

		/*
			Receive 콜백함수 호출 및 패킷 인/디코딩
		*/
		session.NetworkFunctor.OnReceive(sessionUniqueId, sessionId, ltvPacket)
	}

	/*
		아직 읽을 수 있는 바이트의 수가 있는데 패킷의 요구크기보다 작아서 처리하지 못하는 경우라면
		버퍼의 시작지점에 값을 쓰고 startRecvPos를 설정하여 다음 패킷받는 위치를 지정한다.
	*/
	if readAbleByte > 0 {
		copy(recvBuff, recvBuff[readPos:(readPos+readAbleByte)])
	}
	/*
		다음 패킷을 수신할 때 어디서부터 수신할 것인지 설정
	*/
	startRecvPos = readAbleByte
	return startRecvPos, NET_ERROR_NONE
}

func (session *TcpSession) closeProcess() {
	session.Conn.Close()
	session.NetworkFunctor.OnClose(session.SeqIndex, session.Index)
	_tcpSessionManager.removeSession(session.SeqIndex, session.Index)
}

/*
만약 scale out시에 redis로 유니크값 관리하면 좋을듯
*/
func SeqNumIncrement() uint64 {
	newValue := atomic.AddUint64(&_seqNumber, 1)
	return newValue
}

func SendToClient(sessionUniqueId uint64, sessionId int32, data []byte) {
	_tcpSessionManager.sendPacket(sessionUniqueId, sessionId, data)
}

func SendToAllClient(data []byte) {
	_tcpSessionManager.sendPacketAllClient(data)
}

/* 전역변수 */
var _seqNumber uint64
var _tcpSessionManager *TcpSessionManager
