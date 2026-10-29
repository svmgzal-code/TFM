package protocol

import (
	"encoding/binary"
	"fmt"
)

// Info is the decoded 0x0012 CMDID_SEND_INFO reply (38 bytes, unpadded Ethernet).
type Info struct {
	FirmwareVersion      uint32
	FirmwareRevision     uint32
	SntBoardVersion      uint32
	BasebandBoardVersion uint32
	FrontendBoardVersion uint32
	AvailableChannels    uint16
	AvailableAlgorithms  uint16
	RadarHardwareID      uint16
	RadarNumber          uint32
	FlashingDate         uint32
	PhaseOffset          int32
}

// InfoSize is the unpadded reply size of 0x0012.
const InfoSize = 38

// DecodeInfo decodes the 38-byte 0x0012 reply.
func DecodeInfo(b []byte) (Info, error) {
	if len(b) != InfoSize {
		return Info{}, fmt.Errorf("protocol: info reply is %d bytes, want %d", len(b), InfoSize)
	}
	return Info{
		FirmwareVersion:      binary.LittleEndian.Uint32(b[0:4]),
		FirmwareRevision:     binary.LittleEndian.Uint32(b[4:8]),
		SntBoardVersion:      binary.LittleEndian.Uint32(b[8:12]),
		BasebandBoardVersion: binary.LittleEndian.Uint32(b[12:16]),
		FrontendBoardVersion: binary.LittleEndian.Uint32(b[16:20]),
		AvailableChannels:    binary.LittleEndian.Uint16(b[20:22]),
		AvailableAlgorithms:  binary.LittleEndian.Uint16(b[22:24]),
		RadarHardwareID:      binary.LittleEndian.Uint16(b[24:26]),
		RadarNumber:          binary.LittleEndian.Uint32(b[26:30]),
		FlashingDate:         binary.LittleEndian.Uint32(b[30:34]),
		PhaseOffset:          int32(binary.LittleEndian.Uint32(b[34:38])),
	}, nil
}
