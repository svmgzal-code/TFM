package protocol

import (
	"testing"
)

func TestDecodeInfo(t *testing.T) {
	// Golden fixture: a 38-byte 0x0012 reply with known field values.
	raw := []byte{
		0x02, 0x01, 0x04, 0x00, // FirmwareVersion = 0x00040102
		0x68, 0x00, 0x00, 0x00, // FirmwareRevision = 104
		0x00, 0x01, 0x02, 0x00, // SntBoardVersion = 0x00020100
		0x01, 0x00, 0x03, 0x00, // BasebandBoardVersion = 0x00030001
		0x02, 0x00, 0x01, 0x00, // FrontendBoardVersion = 0x00010002
		0x0F, 0x00, // AvailableChannels = 0x000F
		0x03, 0x00, // AvailableAlgorithms = 0x0003
		0x01, 0x00, // RadarHardwareID = 0x0001
		0x3D, 0x2C, 0x1B, 0x0A, // RadarNumber = 0x0A1B2C3D
		0x1D, 0x78, 0x33, 0x01, // FlashingDate = 20150301 (decimal)
		0x00, 0x00, 0x00, 0x00, // PhaseOffset = 0
	}

	got, err := DecodeInfo(raw)
	if err != nil {
		t.Fatalf("DecodeInfo: %v", err)
	}

	want := Info{
		FirmwareVersion:      0x00040102,
		FirmwareRevision:     104,
		SntBoardVersion:      0x00020100,
		BasebandBoardVersion: 0x00030001,
		FrontendBoardVersion: 0x00010002,
		AvailableChannels:    0x000F,
		AvailableAlgorithms:  0x0003,
		RadarHardwareID:      0x0001,
		RadarNumber:          0x0A1B2C3D,
		FlashingDate:         20150301,
		PhaseOffset:          0,
	}
	if got != want {
		t.Errorf("DecodeInfo mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestDecodeInfoRejectsWrongSize(t *testing.T) {
	if _, err := DecodeInfo(make([]byte, InfoSize-1)); err == nil {
		t.Fatal("expected error for short info reply")
	}
	if _, err := DecodeInfo(make([]byte, InfoSize+1)); err == nil {
		t.Fatal("expected error for long info reply")
	}
}

func TestDecodeParams(t *testing.T) {
	// Golden fixture: a 29-byte 0x0029 reply matching the app's default system params.
	raw := []byte{
		0x04,       // Band = 4
		0x02,       // TRamp = 2
		0x02,       // ZeroPad = 2
		0x00,       // FftDataType = 0 (magnitude)
		0x01,       // FrontendEn = 1
		0x00,       // PowerSaveEn = 0
		0x01,       // Norm = 1
		0x0F, 0x00, // ActChannels = 0x000F
		0x00,       // AdvancedEn = 0
		0x00, 0x01, // FreqPoints = 256
		0x00, 0x00, // MinFreq = 0
		0x00, 0x00, // ManualBW = 0
		0x00,                   // Attenuation = 0
		0x30, 0x26, 0x09, 0x00, // RangeBin = 599600 um
		0x00, 0x00, 0x00, 0x00, // DopplerBin = 0
		0x00, 0x00, 0x00, 0x00, // FrequencyBin = 0
	}

	got, err := DecodeParams(raw)
	if err != nil {
		t.Fatalf("DecodeParams: %v", err)
	}

	want := Params{
		SystemParams: SystemParams{
			Band:        4,
			TRamp:       2,
			ZeroPad:     2,
			FftDataType: 0,
			FrontendEn:  1,
			PowerSaveEn: 0,
			Norm:        1,
			ActChannels: 0x000F,
			AdvancedEn:  0,
			FreqPoints:  256,
			MinFreq:     0,
			ManualBW:    0,
			Attenuation: 0,
		},
		RangeBin:     599600,
		DopplerBin:   0,
		FrequencyBin: 0,
	}
	if got != want {
		t.Errorf("DecodeParams mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestDecodeParamsRejectsWrongSize(t *testing.T) {
	if _, err := DecodeParams(make([]byte, ParamsReplySize-1)); err == nil {
		t.Fatal("expected error for short params reply")
	}
}

func TestEncodeCommand(t *testing.T) {
	got := EncodeCommand(CmdIDInfo)
	if len(got) != 2 || got[0] != 0x12 || got[1] != 0x00 {
		t.Errorf("EncodeCommand(0x0012) = % X, want 12 00", got)
	}
}

func TestVerifyEcho(t *testing.T) {
	tests := []struct {
		name    string
		id      uint16
		echo    []byte
		wantErr bool
	}{
		{"ok", CmdIDInfo, []byte{0x12, 0x00}, false},
		{"mismatch", CmdIDInfo, []byte{0x29, 0x00}, true},
		{"short", CmdIDInfo, []byte{0x12}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyEcho(tt.id, tt.echo)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyEcho: got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
