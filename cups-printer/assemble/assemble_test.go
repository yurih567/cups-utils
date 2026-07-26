package assemble_test

import (
	"bytes"
	"testing"

	"cups-printer/assemble"

	drivers "cups-drivers"
	_ "cups-drivers/models"
)

func TestTailEpsonFull(t *testing.T) {
	d := drivers.MustGet("epson")
	got := assemble.Tail(d, assemble.Options{
		Feed:       3,
		Cut:        true,
		PartialCut: false,
		Beep:       2,
		Drawer:     true,
		DrawerPin:  0,
	})

	want := append([]byte{}, d.Feed(3)...)
	want = append(want, d.Cut(false)...)
	want = append(want, d.Beep(2)...)
	want = append(want, d.OpenDrawer(0)...)

	if !bytes.Equal(got, want) {
		t.Fatalf("Tail = %v want %v", got, want)
	}
}

func TestTailSkipsBeepAndDrawerWhenOff(t *testing.T) {
	d := drivers.MustGet("epson")
	got := assemble.Tail(d, assemble.Options{
		Feed: 1,
		Cut:  true,
		Beep: 0,
	})
	want := append(d.Feed(1), d.Cut(false)...)
	if !bytes.Equal(got, want) {
		t.Fatalf("Tail = %v want %v", got, want)
	}
	if bytes.Contains(got, d.Beep(1)) {
		t.Fatal("beep should not be present")
	}
	if bytes.Contains(got, d.OpenDrawer(0)) {
		t.Fatal("drawer should not be present")
	}
}

func TestPayloadAppendsTail(t *testing.T) {
	d := drivers.MustGet("bematech")
	body := []byte{0x01, 0x02}
	got := assemble.Payload(body, d, assemble.Options{Feed: 2, Cut: true, Beep: 1})
	want := append(append([]byte{0x01, 0x02}, d.Feed(2)...), append(d.Cut(false), d.Beep(1)...)...)
	if !bytes.Equal(got, want) {
		t.Fatalf("Payload = %v want %v", got, want)
	}
}
