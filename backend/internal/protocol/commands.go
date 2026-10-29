package protocol

import (
	"encoding/binary"
	"fmt"
)

// Command IDs implemented by the app (little-endian u16 on the wire).
const (
	CmdIDInfo       uint16 = 0x0012 // CMDID_SEND_INFO: read hardware/firmware info (connect handshake)
	CmdIDSendParams uint16 = 0x0029 // CMDID_SEND_PARAMS: read system parameters
	CmdIDSetup      uint16 = 0x0028 // CMDID_SETUP: write system parameters (EEPROM)
	CmdIDOpSetup    uint16 = 0x0020 // CMDID_OP_SETUP: write operation parameters (EEPROM)
)

// EncodeCommand returns the little-endian u16 command id as sent on the wire.
func EncodeCommand(id uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, id)
	return b
}

// VerifyEcho checks that the radar echoed the command id we sent before executing it.
func VerifyEcho(id uint16, echo []byte) error {
	if len(echo) != 2 {
		return fmt.Errorf("protocol: command echo is %d bytes, want 2", len(echo))
	}
	got := binary.LittleEndian.Uint16(echo)
	if got != id {
		return fmt.Errorf("protocol: command echo mismatch: sent 0x%04X, got 0x%04X", id, got)
	}
	return nil
}
