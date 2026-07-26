package assemble

import drivers "cups-drivers"

// Options controls hardware commands appended after the receipt body.
type Options struct {
	Feed       int
	Cut        bool
	PartialCut bool
	Beep       int
	Drawer     bool
	DrawerPin  int
}

// Tail builds feed/cut/beep/drawer commands in that order.
// Unsupported (nil/empty) commands from the driver are skipped.
func Tail(d drivers.Driver, opts Options) []byte {
	var out []byte

	if feed := d.Feed(opts.Feed); drivers.Supported(feed) {
		out = append(out, feed...)
	}
	if opts.Cut {
		if cut := d.Cut(opts.PartialCut); drivers.Supported(cut) {
			out = append(out, cut...)
		}
	}
	if opts.Beep > 0 {
		if beep := d.Beep(opts.Beep); drivers.Supported(beep) {
			out = append(out, beep...)
		}
	}
	if opts.Drawer {
		if drawer := d.OpenDrawer(opts.DrawerPin); drivers.Supported(drawer) {
			out = append(out, drawer...)
		}
	}
	return out
}

// Payload concatenates body and tail.
func Payload(body []byte, d drivers.Driver, opts Options) []byte {
	tail := Tail(d, opts)
	if len(tail) == 0 {
		return body
	}
	out := make([]byte, 0, len(body)+len(tail))
	out = append(out, body...)
	out = append(out, tail...)
	return out
}
