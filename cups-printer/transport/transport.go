package transport

import (
	"fmt"
	"strings"
)

// Transport delivers raw ESC/POS bytes to a printer.
type Transport interface {
	Send(payload []byte) error
	String() string
}

// Parse builds a Transport from a destination string.
//
// Supported forms:
//   - cups:Name or plain Name  → CUPS queue with -o raw
//   - tcp://host:9100 or host:port → JetDirect TCP
//   - file:/path or file:///path or absolute path → device/file write
//   - stdout or - → write to os.Stdout
func Parse(dest string) (Transport, error) {
	d := strings.TrimSpace(dest)
	if d == "" {
		return nil, fmt.Errorf("transport: empty destination")
	}

	lower := strings.ToLower(d)
	switch {
	case d == "-" || lower == "stdout":
		return Stdout{}, nil
	case strings.HasPrefix(lower, "cups:"):
		name := strings.TrimSpace(d[len("cups:"):])
		if name == "" {
			return nil, fmt.Errorf("transport: empty cups queue name")
		}
		return Cups{Queue: name}, nil
	case strings.HasPrefix(lower, "tcp://"):
		return newTCP(d[len("tcp://"):])
	case strings.HasPrefix(lower, "file://"):
		path := d[len("file://"):]
		if path == "" {
			return nil, fmt.Errorf("transport: empty file path")
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return File{Path: path}, nil
	case strings.HasPrefix(lower, "file:"):
		path := strings.TrimSpace(d[len("file:"):])
		if path == "" {
			return nil, fmt.Errorf("transport: empty file path")
		}
		return File{Path: path}, nil
	case looksLikeHostPort(d):
		return newTCP(d)
	case strings.HasPrefix(d, "/"):
		return File{Path: d}, nil
	default:
		return Cups{Queue: d}, nil
	}
}

func looksLikeHostPort(s string) bool {
	host, port, ok := strings.Cut(s, ":")
	if !ok || host == "" || port == "" {
		return false
	}
	if strings.Contains(host, "://") {
		return false
	}
	for i := 0; i < len(port); i++ {
		if port[i] < '0' || port[i] > '9' {
			return false
		}
	}
	return true
}
