package transport

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const defaultTCPPort = "9100"

// TCP sends raw bytes over JetDirect (typically port 9100).
type TCP struct {
	Addr string
}

func newTCP(hostPort string) (TCP, error) {
	hostPort = strings.TrimSpace(hostPort)
	if hostPort == "" {
		return TCP{}, fmt.Errorf("transport: empty tcp address")
	}
	host, port, err := splitHostPort(hostPort)
	if err != nil {
		return TCP{}, err
	}
	if port == "" {
		port = defaultTCPPort
	}
	return TCP{Addr: net.JoinHostPort(host, port)}, nil
}

func splitHostPort(hostPort string) (host, port string, err error) {
	if strings.HasPrefix(hostPort, "[") {
		h, p, e := net.SplitHostPort(hostPort)
		return h, p, e
	}
	if strings.Count(hostPort, ":") == 1 {
		return net.SplitHostPort(hostPort)
	}
	if strings.Contains(hostPort, ":") {
		h, p, e := net.SplitHostPort(hostPort)
		return h, p, e
	}
	return hostPort, "", nil
}

func (t TCP) String() string {
	return "tcp://" + t.Addr
}

func (t TCP) Send(payload []byte) error {
	conn, err := net.DialTimeout("tcp", t.Addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("tcp %s: dial: %w", t.Addr, err)
	}
	defer conn.Close()

	deadline := 30 * time.Second
	if len(payload) > 32*1024 {
		deadline = 120 * time.Second
	}
	_ = conn.SetDeadline(time.Now().Add(deadline))

	for sent := 0; sent < len(payload); {
		n, err := conn.Write(payload[sent:])
		if err != nil {
			return fmt.Errorf("tcp %s: write: %w", t.Addr, err)
		}
		if n == 0 {
			return fmt.Errorf("tcp %s: write: short write", t.Addr)
		}
		sent += n
	}
	return nil
}
