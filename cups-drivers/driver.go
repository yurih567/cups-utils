package drivers

// Driver describes a printer command dialect shared by cups-template and cups-printer.
//
// cups-template typically uses: Init, ClearBuffer, LineEnd, Align*, Raster, PrintWidthDots.
// cups-printer typically uses: Feed, Cut, Beep, OpenDrawer.
//
// Methods return nil when the command is not supported by the model.
type Driver interface {
	ID() string
	Name() string

	// PrintWidthDots is the printable head width in dots at the model's native DPI
	// (typically 203). This is usually less than the paper roll width (e.g. 576 for
	// 80mm paper). Layout and raster encoding must use this value, not the paper mm.
	PrintWidthDots() int

	Init() []byte
	ClearBuffer() []byte
	LineEnd() []byte
	AlignCenter() []byte
	AlignLeft() []byte

	Raster(width, height int, data []byte) ([]byte, error)

	Feed(lines int) []byte
	Cut(partial bool) []byte
	Beep(times int) []byte
	OpenDrawer(pin int) []byte
}

// Supported reports whether cmd is non-nil and non-empty.
func Supported(cmd []byte) bool {
	return len(cmd) > 0
}
