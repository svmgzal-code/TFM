package protocol

import (
	"encoding/binary"
	"fmt"
)

// SystemParams is the read/write 17-byte system parameter block (0x0028 payload / 0x0029 head).
type SystemParams struct {
	Band        uint8
	TRamp       uint8
	ZeroPad     uint8
	FftDataType uint8
	FrontendEn  uint8
	PowerSaveEn uint8
	Norm        uint8
	ActChannels uint16
	AdvancedEn  uint8
	FreqPoints  uint16
	MinFreq     uint16
	ManualBW    uint16
	Attenuation uint8
}

// SystemParamsSize is the read/write system parameter block size.
const SystemParamsSize = 17

// ParamsReplySize is the unpadded 0x0029 reply size (17-byte block + 12-byte read-only tail).
const ParamsReplySize = 29

// Params is the decoded 0x0029 reply: system params plus the read-only bins.
type Params struct {
	SystemParams
	RangeBin     uint32 // um
	DopplerBin   uint32 // um/s
	FrequencyBin uint32 // Hz
}

// DecodeParams decodes the 29-byte 0x0029 reply.
func DecodeParams(b []byte) (Params, error) {
	if len(b) != ParamsReplySize {
		return Params{}, fmt.Errorf("protocol: params reply is %d bytes, want %d", len(b), ParamsReplySize)
	}
	return Params{
		SystemParams: decodeSystemParams(b[0:SystemParamsSize]),
		RangeBin:     binary.LittleEndian.Uint32(b[17:21]),
		DopplerBin:   binary.LittleEndian.Uint32(b[21:25]),
		FrequencyBin: binary.LittleEndian.Uint32(b[25:29]),
	}, nil
}

func decodeSystemParams(b []byte) SystemParams {
	return SystemParams{
		Band:        b[0],
		TRamp:       b[1],
		ZeroPad:     b[2],
		FftDataType: b[3],
		FrontendEn:  b[4],
		PowerSaveEn: b[5],
		Norm:        b[6],
		ActChannels: binary.LittleEndian.Uint16(b[7:9]),
		AdvancedEn:  b[9],
		FreqPoints:  binary.LittleEndian.Uint16(b[10:12]),
		MinFreq:     binary.LittleEndian.Uint16(b[12:14]),
		ManualBW:    binary.LittleEndian.Uint16(b[14:16]),
		Attenuation: b[16],
	}
}

// EncodeSystemParams encodes the 17-byte 0x0028 payload (little-endian, unpadded).
func EncodeSystemParams(p SystemParams) []byte {
	b := make([]byte, SystemParamsSize)
	b[0] = p.Band
	b[1] = p.TRamp
	b[2] = p.ZeroPad
	b[3] = p.FftDataType
	b[4] = p.FrontendEn
	b[5] = p.PowerSaveEn
	b[6] = p.Norm
	binary.LittleEndian.PutUint16(b[7:9], p.ActChannels)
	b[9] = p.AdvancedEn
	binary.LittleEndian.PutUint16(b[10:12], p.FreqPoints)
	binary.LittleEndian.PutUint16(b[12:14], p.MinFreq)
	binary.LittleEndian.PutUint16(b[14:16], p.ManualBW)
	b[16] = p.Attenuation
	return b
}
