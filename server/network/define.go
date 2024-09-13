package network

const (
	POOL_SIZE       = 512
	MAX_BUFFER      = 4096
	MAX_PACKET_SIZE = 1024
)

const (
	BYTE_OF_CHATROOM_ID      = 2
	BYTE_OF_MESSAGE_SEQUENCE = 4
)

const (
	NET_ERROR_NONE             = 10001
	NET_ERROR_TOO_LARGE_PACKET = 10002
)
