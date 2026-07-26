package transport_test

import (
	"testing"

	"cups-printer/transport"
)

func TestParseDestinations(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"stdout", "stdout"},
		{"-", "stdout"},
		{"cups:TM20", "cups:TM20"},
		{"MinhaFila", "cups:MinhaFila"},
		{"tcp://192.168.0.10:9100", "tcp://192.168.0.10:9100"},
		{"192.168.0.10:9100", "tcp://192.168.0.10:9100"},
		{"192.168.0.10", "tcp://192.168.0.10:9100"},
		{"10.0.0.5", "tcp://10.0.0.5:9100"},
		{"printer.local:9100", "tcp://printer.local:9100"},
		{"file:/dev/usb/lp0", "file:/dev/usb/lp0"},
		{"file:///dev/usb/lp0", "file:/dev/usb/lp0"},
		{"/dev/usb/lp0", "file:/dev/usb/lp0"},
	}

	for _, tt := range tests {
		tr, err := transport.Parse(tt.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.in, err)
		}
		if got := tr.String(); got != tt.want {
			t.Fatalf("Parse(%q).String() = %q want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseEmpty(t *testing.T) {
	if _, err := transport.Parse("  "); err == nil {
		t.Fatal("expected error for empty dest")
	}
}

func TestParseTCPDefaultPort(t *testing.T) {
	tr, err := transport.Parse("tcp://10.0.0.5")
	if err != nil {
		t.Fatal(err)
	}
	if got := tr.String(); got != "tcp://10.0.0.5:9100" {
		t.Fatalf("got %q", got)
	}
}
