package network

import "encoding/binary"

type RawPacketData struct {
	Pos   int
	Data  []byte
	Order binary.ByteOrder
}

/*
패킷 헤더의 구조는 아래와 같음
패킷전체크기(2Byte) + 패킷ID(2Byte) + 패킷 Type(1Byte)
패킷의 2바이트를 꺼내서 패킷의 전체 크기를 구한다.
*/
func PacketTotalSize(data []byte) int16 {
	totalSize := binary.LittleEndian.Uint16(data)
	return int16(totalSize)
}

func MakeReader(buffer []byte, isLittleEndian bool) RawPacketData {
	if isLittleEndian {
		return RawPacketData{
			Data:  buffer,
			Pos:   0,
			Order: binary.LittleEndian,
		}
	}

	return RawPacketData{
		Data:  buffer,
		Pos:   0,
		Order: binary.BigEndian,
	}
}

func MakeWrite(buffer []byte, isLittleEndian bool) RawPacketData {
	if isLittleEndian {
		return RawPacketData{
			Data:  buffer,
			Pos:   0,
			Order: binary.LittleEndian,
		}
	}

	return RawPacketData{
		Data:  buffer,
		Pos:   0,
		Order: binary.BigEndian,
	}
}
