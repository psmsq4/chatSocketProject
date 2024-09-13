package main

import (
	"fmt"
	clientsession "server/clientSession"
	"server/network"
	"server/protocol"
)

type LifeGameServer struct {
	ServerIndex int
	netConfig   network.NetConfig
	PacketChan  chan protocol.Packet
}

func startLifeGameServer(netConfig network.NetConfig) {
	server := LifeGameServer{
		ServerIndex: 1,
		netConfig:   netConfig,
	}

	protocol.InitPacketHeaderSize()

	/*
		클라이언트 세션매니저 초기화
	*/
	clientsession.Init()

	/*
		채널 버퍼 256으로 설정 :: Packet을 256개 송/수신할 수 있다.
	*/
	server.PacketChan = make(chan protocol.Packet, 256)

	/*
		패킷을 처리하는 고루틴
	*/
	go server.PacketProcessGoroutine()

	/*
		네트워크의 콜백함수를 지정한다.
		네트워크 모듈에서 패킷이 도착하면 지정한 콜백함수를 호출하여 처리한다.
	*/
	snFunctor := network.SessionNetworkFunctor{
		OnConnect:           server.OnConnect,
		OnClose:             server.OnClose,
		OnReceive:           server.OnReceive,
		PacketTotalSizeFunc: network.PacketTotalSize,
		PacketHeaderSize:    protocol.GetPacketHeaderSize(),
	} // server := LifeGameServer { ServerIndex: 1, netConfig: netConfig }

	network.StartServerBlock(netConfig, snFunctor)
	// network 패키지와 main 패키지 간 상호참조 문제를 회피하는 경우?
	// main 패키지는 network 패키지를 참조하고 있다.
	// 원칙적으로 network 패키지는 main 패키지를 참조할 수 없지만,
	// network 패키지의 함수를 호출하여 매개변수에 snFunctor를 전달함으로서
	// main 패키지에 정의된 LifeGameServer 객체와 그에 수반한 매서드들을 호출할 수 있다.
	// network 패키지 로직에서 snFunctor에 저장된 매서드를 호출하면
	// main 패키지에 정의되어 생성된 LifeGameServer 객체에 영향을 미칠 수 있다.
}

/*
클라이언트의 접속이 끊어졌을 때 콜백
*/
func (server *LifeGameServer) OnClose(sessionUniqueId uint64, sessionId int32) {
	fmt.Printf("Client Disconnected:%d - %d\n", sessionUniqueId, sessionId)
	_ = clientsession.RemoveSession(sessionUniqueId)
}

/*
클라이언트가 접속일 하였을 때 콜백
게임 세션 만들어 줘야함
*/
func (server *LifeGameServer) OnConnect(sessionUniqueId uint64, sessionId int32) {
	fmt.Printf("New Client Connected:%d - %d\n", sessionUniqueId, sessionId)
	// _ = clientsession.AddSession(sessionUniqueId, sessionId)
}

/*
클라이언트가 데이터를 주었을 때 콜백
인/디코딩 작업 들어가야함
*/
func (server *LifeGameServer) OnReceive(sessionUniqueId uint64, sessionId int32, packet []byte) {
	server.DistributePacket(sessionUniqueId, sessionId, packet)
}
