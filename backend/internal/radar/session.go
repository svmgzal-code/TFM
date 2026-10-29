package radar

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"tfm/backend/internal/protocol"
)

// State is the radar connection state machine.
type State int

const (
	StateDisconnected State = iota
	StateConnected
	StateStreaming
)

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
