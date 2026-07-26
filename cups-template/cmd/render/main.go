package main

import (
	"flag"
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	drivers "cups-drivers"
	_ "cups-drivers/models"
	"cups-template/engine"
	"cups-template/engine/dom"
	"cups-template/engine/font"
	"cups-template/engine/parser"
	"cups-template/engine/units"
	escposrenderer "cups-template/renderers/escpos"
	pngrenderer "cups-template/renderers/png"
)

func main() {
	templatePath := flag.String("template", "-", "XML template path, or - for stdin")
	format := flag.String("format", "cups", "output format: cups | png | escpos")
	outPath := flag.String("out", "-", "output path, or - for stdout")
	assetsPath := flag.String("assets", "", "optional assets base path for relative image src")
	dpi := flag.Float64("dpi", 0, "render DPI (default: 96 for cups/png, 203 for escpos)")
	model := flag.String("model", "generic", "printer model for escpos: generic | epson | bematech")
	flag.Parse()

	fmtName := strings.ToLower(strings.TrimSpace(*format))
	if *dpi <= 0 {
		if fmtName == "escpos" {
			*dpi = 203
		} else {
			*dpi = 96
		}
	}

	fonts, err := font.NewDefault(*dpi)
	must(err)

	doc, err := parseTemplate(*templatePath)
	must(err)

	var drv drivers.Driver
	if fmtName == "escpos" {
		drv, err = drivers.Get(strings.ToLower(strings.TrimSpace(*model)))
		must(err)
		printDots := drv.PrintWidthDots()
		if printDots <= 0 {
			printDots = escposrenderer.DefaultMaxWidthDots
		}
		doc.Root.Style.Width = units.Mm(float64(printDots) * 25.4 / *dpi)
	}

	tree, err := engine.Layout(doc, engine.Options{
		DPI:           *dpi,
		Fonts:         fonts,
		AssetBasePath: *assetsPath,
	})
	must(err)

	commands := engine.BuildDisplayList(tree)
	out := openOutput(*outPath)
	defer out.Close()

	switch fmtName {
	case "cups", "png":
		img := pngrenderer.NewRenderer(fonts).Render(commands)
		must(png.Encode(out, img))
	case "escpos":
		data, err := escposrenderer.NewRenderer(fonts, escposrenderer.Options{
			Driver:       drv,
			Threshold:    128,
			FeedLines:    3,
			MaxWidthDots: drv.PrintWidthDots(),
		}).Render(commands)
		must(err)
		_, err = out.Write(data)
		must(err)
	default:
		must(fmt.Errorf("unsupported format %q (use cups, png or escpos)", *format))
	}
}

func parseTemplate(path string) (*dom.Document, error) {
	if path == "-" {
		return parser.Parse(os.Stdin)
	}
	return parser.ParseFile(path)
}

func openOutput(path string) io.WriteCloser {
	if path == "-" {
		return nopCloser{os.Stdout}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		must(err)
	}
	f, err := os.Create(path)
	must(err)
	return f
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
