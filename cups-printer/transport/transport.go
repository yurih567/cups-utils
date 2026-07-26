package transport

import (
	"fmt"
	"net"
	"strings"
)

// Transport delivers raw ESC/POS bytes to a printer.
type Transport interface {
	Send(payload []byte) error
	String() string
}

// Parse builds a Transport from a destination string.
// Callers pass a plain name, IP, or path — the protocol is inferred.
//
// Recognition rules (first match wins):
//   - stdout or -                         → stdout
//   - cups:Name                           → CUPS queue
//   - tcp://host[:port]                   → JetDirect TCP (default port 9100)
//   - file:/path or file:///path          → device/file
//   - absolute path (/dev/..., /tmp/...)  → device/file
//   - IPv4 or IPv4:port                   → TCP (default port 9100)
//   - host:port                           → TCP
//   - anything else (queue name)          → CUPS queue
//
// Examples:
//
//	"TM20"                 → cups:TM20
//	"192.168.0.10"         → tcp://192.168.0.10:9100
//	"192.168.0.10:9100"    → tcp://192.168.0.10:9100
//	"printer.local:9100"   → tcp://printer.local:9100
//	"/dev/usb/lp0"         → file:/dev/usb/lp0
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
	case strings.HasPrefix(d, "/"):
		return File{Path: d}, nil
	case looksLikeHostPort(d):
		return newTCP(d)
	case looksLikeIP(d):
		return newTCP(d)
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
	if strings.Count(s, ":") != 1 {
		return false
	}
	for i := 0; i < len(port); i++ {
		if port[i] < '0' || port[i] > '9' {
			return false
		}
	}
	return true
}

func looksLikeIP(s string) bool {
	host := s
	if strings.Count(s, ":") == 1 {
		h, port, ok := strings.Cut(s, ":")
		if !ok || port == "" {
			return false
		}
		for i := 0; i < len(port); i++ {
			if port[i] < '0' || port[i] > '9' {
				return false
			}
		}
		host = h
	} else if strings.Contains(s, ":") {
		return false
	}
	return net.ParseIP(host) != nil
}
