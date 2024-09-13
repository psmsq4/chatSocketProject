package main

import (
	"fmt"
	"server/network"
	"server/protocol"
)

type LifeGameServer struct {
	ServerIndex int
	netConfig   network.NetConfig
	PacketChan  chan protocol.Packet
}

func startLifeGameServer(netConfig network.NetConfig) { // main.go에서 main 몸체 내에서 실행되는 함수이다.
	server := LifeGameServer{
		ServerIndex: 1, // 서버 스케일 아웃 시 서버별로 번호 부여
		netConfig:   netConfig,
		// main.go에서 ./config/net_config.json [SOURCE]을 netConfig에 디코딩하는 func parseNetConfig() (network.NetConfig, error)으로부터
		// netConfig를 반환받아 여기에 전달하여 초기화.
	}

	protocol.InitPacketHeaderSize()
	// _packetHeaderSize 전역변수를 Header의 Totalsize 맴버변수의 크기로 초기화한다.

	/*
		채널 버퍼 256으로 설정
	*/
	server.PacketChan = make(chan protocol.Packet, 256)
	// DEEP QUEUE는 QUEUE와 다르게, 앞으로 넣고 뒤로 넣고 앞으로 빼고 뒤로 뺄 수 있다.
	// 채널에도 buffer가 있다. 기존에 1이지만, packet의 수용량을 늘려 블로킹을 막기 위해, 256으로 설정

	/*
		패킷을 처리하는 고루틴
		[PATH]: [./distributePacket.go]
	*/
	go server.PacketProcessGoroutin()

	/*
		네트워크의 콜백함수를 지정한다.
		네트워크 모듈에서 패킷이 도착하면 지정한 콜백함수를 호출하여 처리한다.
	*/
	snFunctor := network.SessionNetworkFunctor{
		OnConnect:           server.OnConnect,
		OnClose:             server.OnClose,
		OnReceive:           server.OnReceive,
		PacketTotalSizeFunc: network.PacketTotalSize,
		PacketHeaderSize:    network.PACKET_HEADER_SIZE,
	}
	// 콜백을 만든 이유 : 확장성 / 함수에 프로토타입만 지정해놓고 구현을 네트워크 모듈에 박지 않겠다.
	// 구현은 특정 게임 모듈에 구현.

	network.StartServiceBlock(netConfig, snFunctor)
}

/*
클라이언트의 접속이 끊어졌을 때 콜백
*/
func (server *LifeGameServer) OnClose(sessionUniqueId uint64, sessionId int32) {
	// (server *LifeGameServer): LifeGameServer 구조체에 속한 메서드임을 나타낸다.
	// └ 매개변수도 아니고, 반환형도 아니라는 말이다.
	fmt.Printf("Client Disconnected: %d - %d \n", sessionUniqueId, sessionId)
}

/*
클라이언트가 접속하였을 때 콜백
게임 세션을 만들어줘야 함.
*/
func (server *LifeGameServer) OnConnect(sessionUniqueId uint64, sessionId int32) {
	fmt.Printf("New Client Connected: %d - %d \n", sessionUniqueId, sessionId)
	// 클라이언트 세션을 만들어줘야 한다.
}

/*
클라이언트가 데이터를 주었을 때 콜백
인/디코딩 작업 들어가야함
*/
func (server *LifeGameServer) OnReceive(sessionUniqueId uint64, sessionId int32, packet []byte) {
	fmt.Printf("Client Send Message: %d - %d: %s \n", sessionUniqueId, sessionId, packet)
	server.DistributePacket(sessionUniqueId, sessionId, packet)
}
