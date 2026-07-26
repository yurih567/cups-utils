package models

import drivers "cups-drivers"

func init() {
	drivers.Register(&Bematech{})
}

// Bematech uses ESC/POS graphics (GS v 0) — set the printer to ESC/POS mode
// in the Bematech utility, otherwise binary raster is printed as garbage text.
// Cut / beep / drawer keep classic ESC/Bema opcodes.
type Bematech struct {
	Generic
}

func (Bematech) ID() string   { return "bematech" }
func (Bematech) Name() string { return "Bematech (ESC/POS graphics)" }

// Cut: ESC i (full), ESC m (partial) — classic ESC/Bema.
func (Bematech) Cut(partial bool) []byte {
	if partial {
		return []byte{esc, 'm'}
	}
	return []byte{esc, 'i'}
}

// Beep uses BEL (0x07), repeated.
func (Bematech) Beep(times int) []byte {
	if times <= 0 {
		return nil
	}
	if times > 9 {
		times = 9
	}
	out := make([]byte, times)
	for i := range out {
		out[i] = 0x07
	}
	return out
}

// OpenDrawer uses ESC v (Bematech drawer pulse). pin is ignored on most MP models.
func (Bematech) OpenDrawer(pin int) []byte {
	_ = pin
	return []byte{esc, 'v', 0}
}

var _ drivers.Driver = Bematech{}
