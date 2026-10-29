package radar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"tfm/backend/internal/protocol"
)

// defaultCommandTimeout is used before the app knows the radar's measurement timing
// (i.e. during connect). Once streaming, timeouts must exceed one measurement cycle:
// max(2 s, 2 x (autoDelay + TRamp)).
const defaultCommandTimeout = 2 * time.Second

// CommandClient owns the radar's command TCP connection and serializes requests so
// that only one command is in flight at a time.
type CommandClient struct {
	conn net.Conn
	mu   sync.Mutex

	timeout time.Duration
}

// Dial opens the command connection to host:port.
func Dial(ctx context.Context, host string, port int) (*CommandClient, error) {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("radar: dial command port %s: %w", addr, err)
	}
	return &CommandClient{conn: conn, timeout: defaultCommandTimeout}, nil
}

// Close closes the command connection.
func (c *CommandClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Command performs one command transaction:
//
//	write CommandID (u16) -> read+verify echo -> write payload (if any) ->
//	read reply payload (if wantReply > 0).
//
// It is safe for concurrent use; requests are serialized and each has a deadline.
func (c *CommandClient) Command(ctx context.Context, id uint16, payload []byte, wantReply int) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	deadline := time.Now().Add(c.timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := c.conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("radar: set deadline: %w", err)
	}

	if _, err := c.conn.Write(protocol.EncodeCommand(id)); err != nil {
		return nil, wrapTimeout(fmt.Errorf("write command 0x%04X: %w", id, err))
	}

	echo := make([]byte, 2)
	if _, err := io.ReadFull(c.conn, echo); err != nil {
		return nil, wrapTimeout(fmt.Errorf("read echo 0x%04X: %w", id, err))
	}
	if err := protocol.VerifyEcho(id, echo); err != nil {
		return nil, err
	}

	if len(payload) > 0 {
		if _, err := c.conn.Write(payload); err != nil {
			return nil, wrapTimeout(fmt.Errorf("write payload 0x%04X: %w", id, err))
		}
	}

	if wantReply > 0 {
		reply := make([]byte, wantReply)
		if _, err := io.ReadFull(c.conn, reply); err != nil {
			return nil, wrapTimeout(fmt.Errorf("read reply 0x%04X: %w", id, err))
		}
		return reply, nil
	}

	return nil, nil
}

func wrapTimeout(err error) error {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return fmt.Errorf("%w: %v", ErrTimeout, err)
	}
	return err
}
