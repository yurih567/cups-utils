package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	drivers "cups-drivers"
	_ "cups-drivers/models"
	"cups-template/engine"
	"cups-template/engine/font"
	"cups-template/engine/parser"
	escposrenderer "cups-template/renderers/escpos"
	pngrenderer "cups-template/renderers/png"
)

func main() {
	root := findProjectRoot()

	templatePath := flag.String("template", filepath.Join(root, "templates", "receipt.xml"), "path to XML template")
	format := flag.String("format", "cups", "output format: cups | png | escpos")
	outPath := flag.String("out", "", "output path (default depends on format)")
	assetsPath := flag.String("assets", filepath.Join(root, "assets"), "assets base path for images")
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
	if *outPath == "" {
		switch fmtName {
		case "escpos":
			*outPath = filepath.Join(root, "examples", "preview", "receipt.bin")
		default:
			*outPath = filepath.Join(root, "examples", "preview", "preview.png")
		}
	}

	fonts, err := font.NewDefault(*dpi)
	must(err)

	doc, err := parser.ParseFile(*templatePath)
	must(err)

	tree, err := engine.Layout(doc, engine.Options{
		DPI:           *dpi,
		Fonts:         fonts,
		AssetBasePath: *assetsPath,
	})
	must(err)

	commands := engine.BuildDisplayList(tree)

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		must(err)
	}

	switch fmtName {
	case "cups", "png":
		img := pngrenderer.NewRenderer(fonts).Render(commands)
		out, err := os.Create(*outPath)
		must(err)
		defer out.Close()
		must(png.Encode(out, img))
		fmt.Printf("cups/png written to %s (%dx%d)\n", *outPath, img.Bounds().Dx(), img.Bounds().Dy())
		fmt.Println("send to any CUPS printer, for example:")
		fmt.Println("  lp -d YourPrinter " + *outPath)

	case "escpos":
		drv, err := drivers.Get(strings.ToLower(strings.TrimSpace(*model)))
		must(err)
		data, err := escposrenderer.NewRenderer(fonts, escposrenderer.Options{
			Driver:    drv,
			Threshold: 128,
			FeedLines: 3,
		}).Render(commands)
		must(err)
		must(os.WriteFile(*outPath, data, 0o644))
		fmt.Printf("escpos (%s) written to %s (%d bytes)\n", drv.ID(), *outPath, len(data))
		fmt.Println("body only — cut/beep are applied by cups-printer")
		fmt.Println("send via CUPS raw queue, for example:")
		fmt.Println("  lp -d YourPrinter -o raw " + *outPath)

	default:
		must(fmt.Errorf("unsupported format %q (use cups, png or escpos)", *format))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := wd
	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
