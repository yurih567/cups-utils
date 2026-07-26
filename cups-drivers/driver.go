package drivers

// Driver describes a printer command dialect shared by cups-template and cups-printer.
//
// cups-template typically uses: Init, ClearBuffer, LineEnd, Align*, Raster.
// cups-printer typically uses: Feed, Cut, Beep, OpenDrawer.
//
// Methods return nil when the command is not supported by the model.
type Driver interface {
	ID() string
	Name() string

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
