package protocol

import (
	"bytes"
	"testing"
)

func TestEncodeSystemParamsGolden(t *testing.T) {
	// Golden fixture from docs/protocol.md section 8 (17 bytes, unpadded).
	want := []byte{
		0x04,       // Band = 4
		0x02,       // TRamp = 2
		0x02,       // ZeroPad = 2
		0x00,       // FftDataType = 0
		0x01,       // FrontendEn = 1
		0x00,       // PowerSaveEn = 0
		0x00,       // Norm = 0
		0x0F, 0x00, // ActChannels = 0x000F
		0x00,       // AdvancedEn = 0
		0x00, 0x01, // FreqPoints = 256
		0x00, 0x00, // MinFreq = 0
		0x00, 0x00, // ManualBW = 0
		0x00, // Attenuation = 0
	}

	in := SystemParams{
		Band:        4,
		TRamp:       2,
		ZeroPad:     2,
		FftDataType: 0,
		FrontendEn:  1,
		PowerSaveEn: 0,
		Norm:        0,
		ActChannels: 0x000F,
		AdvancedEn:  0,
		FreqPoints:  256,
		MinFreq:     0,
		ManualBW:    0,
		Attenuation: 0,
	}

	got := EncodeSystemParams(in)
	if !bytes.Equal(got, want) {
		t.Errorf("EncodeSystemParams mismatch:\n got % X\nwant % X", got, want)
	}
}

func TestEncodeSystemParamsRoundTrip(t *testing.T) {
	in := SystemParams{
		Band:        3,
		TRamp:       5,
		ZeroPad:     4,
		FftDataType: 0,
		FrontendEn:  1,
		PowerSaveEn: 1,
		Norm:        1,
		ActChannels: 0x0005,
		AdvancedEn:  0x42,
		FreqPoints:  256,
		MinFreq:     24000,
		ManualBW:    250,
		Attenuation: 3,
	}

	raw := EncodeSystemParams(in)
	got := decodeSystemParams(raw)
	if got != in {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, in)
	}
}

func TestEncodeOpSetupGolden(t *testing.T) {
	// Golden fixture from docs/protocol.md section 8: opMode 1, autoMeasType 0x0054, autoDelay 100.
	want := []byte{0x01, 0x00, 0x54, 0x00, 0x64, 0x00}
	got := EncodeOpSetup(OpSetup{OpMode: OpModeAutomatic, AutoMeasType: 0x0054, AutoDelay: 100})
	if !bytes.Equal(got, want) {
		t.Errorf("EncodeOpSetup mismatch:\n got % X\nwant % X", got, want)
	}
}

func TestAutoMeasType(t *testing.T) {
	tests := []struct {
		mode Mode
		want uint16
	}{
		{ModeFMCWUp, 0x0054},
		{ModeFMCWDown, 0x0056},
		{ModeCW, 0x0055},
	}
	for _, tt := range tests {
		if got := tt.mode.AutoMeasType(); got != tt.want {
			t.Errorf("AutoMeasType(%v) = 0x%04X, want 0x%04X", tt.mode, got, tt.want)
		}
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		s    string
		want Mode
		ok   bool
	}{
		{"fmcw_up", ModeFMCWUp, true},
		{"fmcw_down", ModeFMCWDown, true},
		{"cw", ModeCW, true},
		{"bogus", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		got, err := ParseMode(tt.s)
		if (err == nil) != tt.ok {
			t.Fatalf("ParseMode(%q): got err %v, want ok %v", tt.s, err, tt.ok)
		}
		if err == nil && got != tt.want {
			t.Errorf("ParseMode(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}
