package abi

import "encoding/binary"

type EventDescriptor struct {
	Version    uint16
	EventID    uint16
	Flags      uint32
	PlayerID   uint64
	RequestKey uint64
}

func EncodeDescriptor(dst []byte, d EventDescriptor) {
	_ = dst[EventDescriptorSize-1]
	binary.LittleEndian.PutUint16(dst[0:], d.Version)
	binary.LittleEndian.PutUint16(dst[2:], d.EventID)
	binary.LittleEndian.PutUint32(dst[4:], d.Flags)
	binary.LittleEndian.PutUint64(dst[8:], d.PlayerID)
	binary.LittleEndian.PutUint64(dst[16:], d.RequestKey)
}

func DecodeDescriptor(src []byte) EventDescriptor {
	_ = src[EventDescriptorSize-1]
	return EventDescriptor{
		Version:    binary.LittleEndian.Uint16(src[0:]),
		EventID:    binary.LittleEndian.Uint16(src[2:]),
		Flags:      binary.LittleEndian.Uint32(src[4:]),
		PlayerID:   binary.LittleEndian.Uint64(src[8:]),
		RequestKey: binary.LittleEndian.Uint64(src[16:]),
	}
}
