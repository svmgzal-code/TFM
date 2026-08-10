package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const (
	cmdPing     uint16 = 0x0000
	defaultAddr        = "192.168.1.2:1024"
)

func main() {
	addr := flag.String("addr", defaultAddr, "radar TCP address")
	timeout := flag.Duration("timeout", 3*time.Second, "connection and read/write timeout")
	dryRun := flag.Bool("dry-run", false, "print the command bytes without opening a TCP connection")
	flag.Parse()

	request := make([]byte, 2)
	binary.LittleEndian.PutUint16(request, cmdPing)

	if *dryRun {
		fmt.Printf("CMDID_PING request: % X\n", request)
		return
	}

	if err := ping(*addr, *timeout, request); err != nil {
		fmt.Fprintf(os.Stderr, "ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ping ok: radar acknowledged command 0x%04X\n", cmdPing)
}

func ping(addr string, timeout time.Duration, request []byte) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	if _, err := conn.Write(request); err != nil {
		return fmt.Errorf("write CMDID_PING: %w", err)
	}

	response := make([]byte, 2)
	if _, err := io.ReadFull(conn, response); err != nil {
		return fmt.Errorf("read acknowledgement: %w", err)
	}

	ack := binary.LittleEndian.Uint16(response)
	if ack != cmdPing {
		return fmt.Errorf("unexpected acknowledgement 0x%04X, want 0x%04X", ack, cmdPing)
	}

	return nil
}
