package models

import drivers "cups-drivers"

const (
	esc = 0x1b
	gs  = 0x1d
	lf  = 0x0a
)

func init() {
	drivers.Register(&Generic{})
}

// Generic implements the common ESC/POS command set used by most thermal printers.
type Generic struct{}

func (Generic) ID() string   { return "generic" }
func (Generic) Name() string { return "Generic ESC/POS" }

// PrintWidthDots is the common 80mm-paper printable width at 203 DPI (72mm head).
func (Generic) PrintWidthDots() int { return 576 }

func (Generic) Init() []byte        { return []byte{esc, '@'} }
func (Generic) ClearBuffer() []byte { return []byte{esc, '@'} }
func (Generic) LineEnd() []byte     { return []byte{lf} }
func (Generic) AlignCenter() []byte { return []byte{esc, 'a', 1} }
func (Generic) AlignLeft() []byte   { return []byte{esc, 'a', 0} }

func (Generic) Raster(width, height int, data []byte) ([]byte, error) {
	return rasterGSv0(width, height, data)
}

func (g Generic) Feed(lines int) []byte {
	if lines <= 0 {
		return nil
	}
	out := make([]byte, 0, lines)
	for i := 0; i < lines; i++ {
		out = append(out, g.LineEnd()...)
	}
	return out
}

func (Generic) Cut(partial bool) []byte {
	if partial {
		return []byte{gs, 'V', 1}
	}
	return []byte{gs, 'V', 0}
}

// Beep uses ESC B n t (n times, t duration units).
func (Generic) Beep(times int) []byte {
	if times <= 0 {
		return nil
	}
	if times > 9 {
		times = 9
	}
	return []byte{esc, 'B', byte(times), 2}
}

// OpenDrawer uses ESC p m t1 t2. pin 0 → pin 2, pin > 0 → pin 5.
func (Generic) OpenDrawer(pin int) []byte {
	m := byte(0)
	if pin > 0 {
		m = 1
	}
	return []byte{esc, 'p', m, 0x19, 0xfa}
}

var _ drivers.Driver = Generic{}
