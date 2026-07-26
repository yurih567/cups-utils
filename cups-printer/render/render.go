package render

import (
	"bytes"
	"fmt"

	drivers "cups-drivers"
	"cups-template/engine"
	"cups-template/engine/font"
	"cups-template/engine/parser"
	escposrenderer "cups-template/renderers/escpos"
)

// Options configures XML → ESC/POS body rendering.
type Options struct {
	Driver    drivers.Driver
	AssetsDir string
	DPI       float64
	FeedLines int
}

// Render parses template XML and returns the ESC/POS body (no cut/beep/drawer).
func Render(xml []byte, opts Options) ([]byte, error) {
	if opts.Driver == nil {
		return nil, fmt.Errorf("render: driver is required")
	}
	dpi := opts.DPI
	if dpi <= 0 {
		dpi = 203
	}
	feed := opts.FeedLines
	if feed < 0 {
		feed = 0
	}

	fonts, err := font.NewDefault(dpi)
	if err != nil {
		return nil, fmt.Errorf("render: fonts: %w", err)
	}

	doc, err := parser.Parse(bytes.NewReader(xml))
	if err != nil {
		return nil, fmt.Errorf("render: parse xml: %w", err)
	}

	tree, err := engine.Layout(doc, engine.Options{
		DPI:           dpi,
		Fonts:         fonts,
		AssetBasePath: opts.AssetsDir,
	})
	if err != nil {
		return nil, fmt.Errorf("render: layout: %w", err)
	}

	commands := engine.BuildDisplayList(tree)
	body, err := escposrenderer.NewRenderer(fonts, escposrenderer.Options{
		Driver:    opts.Driver,
		Threshold: 128,
		FeedLines: feed,
	}).Render(commands)
	if err != nil {
		return nil, fmt.Errorf("render: escpos: %w", err)
	}
	return body, nil
}
