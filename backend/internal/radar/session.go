package radar

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tfm/backend/internal/protocol"
)

// State is the radar connection state machine.
type State int

const (
	StateDisconnected State = iota
	StateConnected
	StateStreaming
)

// DefaultStreamPort is the fixed radar stream port (47475; 47475-47480 exist).
const DefaultStreamPort = 47475

// DefaultAutoDelayMs is the fixed minimum cycle time written by 0x0020 until the UI
// exposes it as a user setting.
const DefaultAutoDelayMs = 100

func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnected:
		return "connected"
	case StateStreaming:
		return "streaming"
	default:
		return "unknown"
	}
}

// MarshalJSON renders State as its string form in API responses.
func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Status is a snapshot of the connection state.
type Status struct {
	State       State  `json:"state"`
	Host        string `json:"host"`
	CommandPort int    `json:"commandPort"`
	StreamPort  int    `json:"streamPort"`
}

// Session owns the radar state machine and caches device data read at connect time.
type Session struct {
	mu sync.Mutex

	state       State
	host        string
	commandPort int
	streamPort  int

	client *CommandClient
	info   protocol.Info
	params protocol.Params
}

func NewSession() *Session {
	return &Session{state: StateDisconnected}
}

// Connect dials the command port and runs the connect handshake (0x0012 -> 0x0029).
// It changes nothing on the device.
func (s *Session) Connect(ctx context.Context, host string, commandPort, streamPort int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != StateDisconnected {
		return ErrAlreadyConnected
	}

	client, err := Dial(ctx, host, commandPort)
	if err != nil {
		return err
	}

	infoRaw, err := client.Command(ctx, protocol.CmdIDInfo, nil, protocol.InfoSize)
	if err != nil {
		client.Close()
		return fmt.Errorf("radar: connect (0x0012): %w", err)
	}
	info, err := protocol.DecodeInfo(infoRaw)
	if err != nil {
		client.Close()
		return err
	}

	paramsRaw, err := client.Command(ctx, protocol.CmdIDSendParams, nil, protocol.ParamsReplySize)
	if err != nil {
		client.Close()
		return fmt.Errorf("radar: connect (0x0029): %w", err)
	}
	params, err := protocol.DecodeParams(paramsRaw)
	if err != nil {
		client.Close()
		return err
	}

	s.client = client
	s.info = info
	s.params = params
	s.host = host
	s.commandPort = commandPort
	s.streamPort = streamPort
	s.state = StateConnected
	return nil
}

// Disconnect closes the command connection and returns to the disconnected state.
// Stream/recording teardown is added in a later slice.
func (s *Session) Disconnect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateDisconnected {
		return ErrNotConnected
	}

	if s.client != nil {
		s.client.Close()
	}
	s.client = nil
	s.state = StateDisconnected
	return nil
}

// Status returns a snapshot of the current connection state.
func (s *Session) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Status{
		State:       s.state,
		Host:        s.host,
		CommandPort: s.commandPort,
		StreamPort:  s.streamPort,
	}
}

// Info returns the device info read at connect time, and whether it is available.
func (s *Session) Info() (protocol.Info, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == StateDisconnected {
		return protocol.Info{}, false
	}
	return s.info, true
}

// Params returns the system params read at connect time, and whether they are available.
func (s *Session) Params() (protocol.Params, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == StateDisconnected {
		return protocol.Params{}, false
	}
	return s.params, true
}

// ParamsUpdate carries the user-editable subset of the system parameters plus the
// measurement mode. Fields not present here (Band, FftDataType, AdvancedEn,
// FreqPoints) are preserved from the device during apply.
type ParamsUpdate struct {
	TRamp       uint8
	ZeroPad     uint8
	FrontendEn  bool
	PowerSaveEn bool
	Norm        bool
	ActChannels uint16
	MinFreq     uint16
	ManualBW    uint16
	Attenuation uint8
	Mode        protocol.Mode
}

// ApplyParams writes the given settings to the radar using the read-modify-write
// sequence: 0x0029 read -> merge -> 0x0028 write -> 0x0029 read-back -> 0x0020 write.
// The 0x0020 write always re-enables automatic mode with the mode's autoMeasType.
//
// Stream stop/restart around an apply is added when streaming lands (StateStreaming).
func (s *Session) ApplyParams(ctx context.Context, upd ParamsUpdate) (protocol.Params, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateDisconnected || s.client == nil {
		return protocol.Params{}, ErrNotConnected
	}

	timeout := commandCycleTimeout(upd)

	raw, err := s.client.CommandWithTimeout(ctx, protocol.CmdIDSendParams, nil, protocol.ParamsReplySize, timeout)
	if err != nil {
		return protocol.Params{}, fmt.Errorf("radar: apply read params (0x0029): %w", err)
	}
	current, err := protocol.DecodeParams(raw)
	if err != nil {
		return protocol.Params{}, err
	}

	merged := current.SystemParams
	merged.TRamp = upd.TRamp
	merged.ZeroPad = upd.ZeroPad
	merged.FrontendEn = boolToByte(upd.FrontendEn)
	merged.PowerSaveEn = boolToByte(upd.PowerSaveEn)
	merged.Norm = boolToByte(upd.Norm)
	merged.ActChannels = upd.ActChannels
	merged.MinFreq = upd.MinFreq
	merged.ManualBW = upd.ManualBW
	merged.Attenuation = upd.Attenuation

	if _, err := s.client.CommandWithTimeout(ctx, protocol.CmdIDSetup, protocol.EncodeSystemParams(merged), 0, timeout); err != nil {
		return protocol.Params{}, fmt.Errorf("radar: apply write params (0x0028): %w", err)
	}

	readBack, err := s.readParamsRetry(ctx, timeout)
	if err != nil {
		return protocol.Params{}, err
	}

	opSetup := protocol.OpSetup{
		OpMode:       protocol.OpModeAutomatic,
		AutoMeasType: upd.Mode.AutoMeasType(),
		AutoDelay:    DefaultAutoDelayMs,
	}
	if _, err := s.client.CommandWithTimeout(ctx, protocol.CmdIDOpSetup, protocol.EncodeOpSetup(opSetup), 0, timeout); err != nil {
		return protocol.Params{}, fmt.Errorf("radar: apply write op params (0x0020): %w", err)
	}

	s.params = readBack
	return readBack, nil
}

// readParamsRetry reads back 0x0029, retrying once to absorb the config-restart
// window after 0x0028.
func (s *Session) readParamsRetry(ctx context.Context, timeout time.Duration) (protocol.Params, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := s.client.CommandWithTimeout(ctx, protocol.CmdIDSendParams, nil, protocol.ParamsReplySize, timeout)
		if err == nil {
			return protocol.DecodeParams(raw)
		}
		lastErr = err
	}
	return protocol.Params{}, fmt.Errorf("radar: apply read-back (0x0029): %w", lastErr)
}

// commandCycleTimeout exceeds one measurement cycle so commands sent while the radar
// restarts its config are not cut short: max(2 s, 2 x (autoDelay + TRamp)).
func commandCycleTimeout(upd ParamsUpdate) time.Duration {
	cycle := 2 * time.Duration(DefaultAutoDelayMs+int(upd.TRamp)) * time.Millisecond
	if cycle < defaultCommandTimeout {
		cycle = defaultCommandTimeout
	}
	return cycle
}

func boolToByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
