package network

import (
	"encoding/binary"
	"errors"
	"reflect"
)

type RawPacketData struct {
	Pos   int
	Data  []byte
	Order binary.ByteOrder
}

/*
패킷 헤더의 구조는 아래와 같음
패킷전체크기(2byte) + 패킷ID(2byte) + 패킷 Type(1byte)
패킷의 2바이트를 꺼내서 패킷의 전체 크기를 구한다.
*/
func PacketTotalSize(data []byte) int16 {
	totalSize := binary.LittleEndian.Uint16(data)
	return int16(totalSize)
	// totalSize는 Header Struct 사이즈(5 Byte)까지 포함한다.
}

func Sizeof(t reflect.Type) int {
	switch t.Kind() {
	case reflect.Array:
		// Array 타입이면 t가 포인터이기 때문에
		// t.Elem()을 통해 실제 Value를 반환받아서 Size를 계산해야한다.
		// 포인터 Size는 운영체제 종속임을 잊지 말자.
		if s := Sizeof(t.Elem()); s >= 0 {
			return s * t.Len()
		}

	case reflect.Struct:
		sum := 0
		for i, n := 0, t.NumField(); i < n; i++ {
			s := Sizeof(t.Field(i).Type)
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return int(t.Size())

	case reflect.Slice:
		return 0

	}

	return -1
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

func (p *RawPacketData) LoadData() []byte {
	return p.Data
}

func (p *RawPacketData) Length() int {
	return len(p.Data)
}

func (p *RawPacketData) ReadByte() (byte, error) {
	var ret byte
	if p.Pos >= len(p.Data) {
		return ret, errors.New("read byte failed")
	}

	ret = p.Data[p.Pos]
	p.Pos++
	return ret, nil
}

func (p *RawPacketData) ReadBytes(readSize int) ([]byte, error) {
	if p.Pos+readSize > len(p.Data) {
		return []byte{}, errors.New("read bytes failed")
	}

	retSlice := p.Data[p.Pos : p.Pos+readSize]
	p.Pos += readSize

	return retSlice, nil
}

func (p *RawPacketData) ReadBool() (bool, error) {
	b, err := p.ReadByte()

	if b != byte(1) {
		return false, err
	}

	return true, err
}

func (p *RawPacketData) ReadS8() (int8, error) {
	_ret, err := p.ReadByte()

	if err != nil {
		return 0, err
	}

	ret := int8(_ret)
	return ret, nil
}

func (p *RawPacketData) ReadU16() (uint16, error) {
	if p.Pos+2 > len(p.Data) {
		return 0, errors.New("read uint16 failed")
	}

	buf := p.Data[p.Pos : p.Pos+2]
	ret := p.Order.Uint16(buf)
	p.Pos += 2
	return ret, nil
}

func (p *RawPacketData) ReadS16() (int16, error) {
	_ret, err := p.ReadU16()
	ret := int16(_ret)
	return ret, err
}

func (p *RawPacketData) ReadU32() (uint32, error) {
	if p.Pos+4 > len(p.Data) {
		return 0, errors.New("read uint32 failed")
	}

	buf := p.Data[p.Pos : p.Pos+4]
	ret := p.Order.Uint32(buf)
	p.Pos += 4
	return ret, nil
}

func (p *RawPacketData) ReadS32() (int32, error) {
	_ret, err := p.ReadU32()
	ret := int32(_ret)
	return ret, err
}

func (p *RawPacketData) ReadU64() (uint64, error) {
	if p.Pos+8 > len(p.Data) {
		return 0, errors.New("read uint64 failed")
	}

	buf := p.Data[p.Pos : p.Pos+8]
	ret := p.Order.Uint64(buf)
	p.Pos += 8
	return ret, nil
}

func (p *RawPacketData) ReadS64() (int64, error) {
	_ret, err := p.ReadU64()
	ret := int64(_ret)
	return ret, err
}

func (p *RawPacketData) ReadString() (string, error) {
	if p.Pos+2 > len(p.Data) {
		return string(""), errors.New("read string header failed")
	}

	size, _ := p.ReadU16()
	if p.Pos+int(size) > len(p.Data) {
		return string(""), errors.New("read string data failed")
	}

	bytes := p.Data[p.Pos : p.Pos+int(size)]
	p.Pos += int(size)
	return string(bytes), nil
}

/* ======== Write ======== */

func (p *RawPacketData) WriteS8(v int8) {
	p.Data[p.Pos] = (byte)(v)
	p.Pos += 1
}

func (p *RawPacketData) WriteU16(v uint16) {
	p.Order.PutUint16(p.Data[p.Pos:], v)
	p.Pos += 2
}

func (p *RawPacketData) WriteS16(v int16) {
	p.WriteU16(uint16(v))
}

func (p *RawPacketData) WriteU32(v uint32) {
	p.Order.PutUint32(p.Data[p.Pos:], v)
	p.Pos += 4
}

func (p *RawPacketData) WriteS32(v int32) {
	p.WriteU32(uint32(v))
}

func (p *RawPacketData) WriteU64(v uint64) {
	p.Order.PutUint64(p.Data[p.Pos:], v)
	p.Pos += 8
}

func (p *RawPacketData) WriteS64(v int64) {
	p.WriteU64(uint64(v))
}

func (p *RawPacketData) WriteBytes(v []byte) {
	copy(p.Data[p.Pos:], v)
	p.Pos += len(v)
}

func (p *RawPacketData) WriteString(v string) {
	copyLen := copy(p.Data[p.Pos:], v)
	p.Pos += copyLen
}
