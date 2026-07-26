package models

import (
	"bytes"
	"testing"

	drivers "cups-drivers"
)

func TestRegistryListsModels(t *testing.T) {
	ids := drivers.List()
	if len(ids) < 3 {
		t.Fatalf("expected at least generic, epson and bematech, got %v", ids)
	}
	for _, id := range []string{"generic", "epson", "bematech"} {
		if _, err := drivers.Get(id); err != nil {
			t.Fatalf("Get(%q): %v", id, err)
		}
	}
}

func TestUnknownModel(t *testing.T) {
	if _, err := drivers.Get("no-such-model"); err == nil {
		t.Fatal("expected error for unknown model")
	}
}

func TestGenericCommands(t *testing.T) {
	d := drivers.MustGet("generic")
	if d.PrintWidthDots() != 576 {
		t.Fatalf("PrintWidthDots = %d want 576", d.PrintWidthDots())
	}
	if !bytes.Equal(d.Init(), []byte{0x1b, '@'}) {
		t.Fatalf("Init = %v", d.Init())
	}
	if !bytes.Equal(d.Cut(false), []byte{0x1d, 'V', 0}) {
		t.Fatalf("Cut(false) = %v", d.Cut(false))
	}
	if !bytes.Equal(d.Cut(true), []byte{0x1d, 'V', 1}) {
		t.Fatalf("Cut(true) = %v", d.Cut(true))
	}
	if !bytes.Equal(d.LineEnd(), []byte{0x0a}) {
		t.Fatalf("LineEnd = %v", d.LineEnd())
	}
	if !bytes.Equal(d.Beep(1), []byte{0x1b, 'B', 1, 2}) {
		t.Fatalf("Beep(1) = %v", d.Beep(1))
	}
	if !bytes.Equal(d.OpenDrawer(0), []byte{0x1b, 'p', 0, 0x19, 0xfa}) {
		t.Fatalf("OpenDrawer(0) = %v", d.OpenDrawer(0))
	}
}

func TestEpsonMatchesGenericCommands(t *testing.T) {
	g := drivers.MustGet("generic")
	e := drivers.MustGet("epson")
	if !bytes.Equal(g.Init(), e.Init()) || !bytes.Equal(g.Cut(false), e.Cut(false)) {
		t.Fatal("epson should share the generic ESC/POS command set")
	}
	if g.ID() == e.ID() {
		t.Fatal("generic and epson must have distinct IDs")
	}
}

func TestEpsonCommands(t *testing.T) {
	d := drivers.MustGet("epson")
	if d.PrintWidthDots() != 512 {
		t.Fatalf("PrintWidthDots = %d want 512", d.PrintWidthDots())
	}
	if !bytes.Equal(d.Init(), []byte{0x1b, '@'}) {
		t.Fatalf("Init = %v", d.Init())
	}
	if !bytes.Equal(d.Cut(false), []byte{0x1d, 'V', 0}) {
		t.Fatalf("Cut(false) = %v", d.Cut(false))
	}
	if !bytes.Equal(d.Cut(true), []byte{0x1d, 'V', 1}) {
		t.Fatalf("Cut(true) = %v", d.Cut(true))
	}
	if !bytes.Equal(d.LineEnd(), []byte{0x0a}) {
		t.Fatalf("LineEnd = %v", d.LineEnd())
	}
}

func TestBematechCommands(t *testing.T) {
	d := drivers.MustGet("bematech")
	if d.ID() == drivers.MustGet("epson").ID() {
		t.Fatal("bematech and epson must have distinct IDs")
	}
	if !bytes.Equal(d.Init(), []byte{0x1b, '@'}) {
		t.Fatalf("Init = %v", d.Init())
	}
	if !bytes.Equal(d.Cut(false), []byte{0x1b, 'i'}) {
		t.Fatalf("Cut(false) = %v", d.Cut(false))
	}
	if !bytes.Equal(d.Beep(2), []byte{0x07, 0x07}) {
		t.Fatalf("Beep(2) = %v", d.Beep(2))
	}
}

func TestRasterEnvelope(t *testing.T) {
	d := drivers.MustGet("generic")
	data := make([]byte, 8) // 8px wide, 1 row
	out, err := d.Raster(8, 1, data)
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := []byte{0x1d, 'v', '0', 0, 1, 0, 1, 0}
	if !bytes.HasPrefix(out, wantPrefix) {
		t.Fatalf("raster prefix = %v want %v", out[:8], wantPrefix)
	}
}

func TestRasterBandsLongImage(t *testing.T) {
	d := drivers.MustGet("generic")
	width := 8
	height := 300
	rowBytes := 1
	data := make([]byte, rowBytes*height)
	out, err := d.Raster(width, height, data)
	if err != nil {
		t.Fatal(err)
	}
	// two bands: 256 + 44
	bands := 0
	i := 0
	for i+8 <= len(out) {
		if out[i] == 0x1d && out[i+1] == 'v' && out[i+2] == '0' {
			bands++
			rowBytes := int(out[i+4]) | int(out[i+5])<<8
			h := int(out[i+6]) | int(out[i+7])<<8
			i += 8 + rowBytes*h
			continue
		}
		t.Fatalf("unexpected bytes at %d: %v", i, out[i:min(i+8, len(out))])
	}
	if bands != 2 {
		t.Fatalf("expected 2 bands, got %d (len=%d)", bands, len(out))
	}
}
