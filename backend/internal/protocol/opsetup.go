package protocol

import (
	"encoding/binary"
	"fmt"
)

// Operation mode values for the 0x0020 opMode field.
const (
	OpModeNormal    uint16 = 0 // host-triggered
	OpModeAutomatic uint16 = 1 // opens the stream sockets
)

// autoMeasType bit masks (Table 2.51). The app always enables system time,
// header and frequency-domain; temperatures/Tx power and time-domain are never set.
const (
	autoMeasCW      uint16 = 1 << 0
	autoMeasDown    uint16 = 1 << 1
	autoMeasSysTime uint16 = 1 << 2
	autoMeasHeader  uint16 = 1 << 4
	autoMeasFreqDom uint16 = 1 << 6
)

// Mode is the measurement mode, mapped onto autoMeasType bits 0-1.
type Mode uint8

const (
	ModeFMCWUp   Mode = iota // bit0=0, bit1=0
	ModeFMCWDown             // bit0=0, bit1=1
	ModeCW                   // bit0=1, bit1=0
)

// ParseMode maps the API's mode string to a Mode.
func ParseMode(s string) (Mode, error) {
	switch s {
	case "fmcw_up":
		return ModeFMCWUp, nil
	case "fmcw_down":
		return ModeFMCWDown, nil
	case "cw":
		return ModeCW, nil
	default:
		return 0, fmt.Errorf("protocol: unknown measurement mode %q", s)
	}
}

// AutoMeasType returns the autoMeasType bit mask for the mode, with the app's
// always-on blocks (system time, header, frequency-domain) included.
func (m Mode) AutoMeasType() uint16 {
	base := autoMeasSysTime | autoMeasHeader | autoMeasFreqDom
	switch m {
	case ModeCW:
		return base | autoMeasCW
	case ModeFMCWDown:
		return base | autoMeasDown
	default:
		return base
	}
}

// OpSetup is the 6-byte 0x0020 payload.
type OpSetup struct {
	OpMode       uint16
	AutoMeasType uint16
	AutoDelay    uint16 // ms
}

// OpSetupSize is the 0x0020 payload size.
const OpSetupSize = 6

// EncodeOpSetup encodes the 6-byte 0x0020 payload (little-endian, unpadded).
func EncodeOpSetup(s OpSetup) []byte {
	b := make([]byte, OpSetupSize)
	binary.LittleEndian.PutUint16(b[0:2], s.OpMode)
	binary.LittleEndian.PutUint16(b[2:4], s.AutoMeasType)
	binary.LittleEndian.PutUint16(b[4:6], s.AutoDelay)
	return b
}
