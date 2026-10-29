package protocol

import "fmt"

// FirmwareVersionString formats a firmware version u32 (e.g. 0x00040102) as "4.1.2".
// Major = bits 16-23, minor = bits 8-15, patch = bits 0-7.
func FirmwareVersionString(v uint32) string {
	return fmt.Sprintf("%d.%d.%d", (v>>16)&0xFF, (v>>8)&0xFF, v&0xFF)
}

// FlashingDateString formats a decimal YYYYMMDD u32 as "2015-03-01".
func FlashingDateString(d uint32) string {
	return fmt.Sprintf("%04d-%02d-%02d", d/10000, (d/100)%100, d%100)
}

// Hex32 formats a u32 as an 8-digit uppercase hex string, e.g. "0x00040102".
func Hex32(v uint32) string {
	return fmt.Sprintf("0x%08X", v)
}
