package job

// Job describes a print request: XML body, printer model, destination, and hardware signals.
type Job struct {
	XML        []byte
	Model      string
	Dest       string
	Beep       int
	Drawer     bool
	DrawerPin  int
	Cut        bool
	PartialCut bool
	Feed       int
	Assets     string
	DPI        float64
}
