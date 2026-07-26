package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cups-printer/job"
	"cups-printer/printer"
)

// Set at build time: -ldflags "-X main.version=v1.0.0"
var version = "dev"

func main() {
	flag.Usage = printUsage

	templatePath := flag.String("template", "-", "XML template path, or - for stdin")
	model := flag.String("model", "generic", "printer model (cups-drivers id)")
	dest := flag.String("dest", "", "printer id (name, IP, path); protocol is auto-detected")
	beep := flag.Int("beep", 0, "beep times (0 = off)")
	drawer := flag.Bool("drawer", false, "open cash drawer")
	drawerPin := flag.Int("drawer-pin", 0, "drawer pin (epson: 0 or 1)")
	cut := flag.Bool("cut", true, "cut paper after print")
	partialCut := flag.Bool("partial-cut", false, "use partial cut when cutting")
	feed := flag.Int("feed", 3, "feed lines before cut")
	assets := flag.String("assets", "", "assets base path for relative image src")
	flag.Parse()

	if len(os.Args) == 1 {
		printUsage()
		os.Exit(0)
	}

	if strings.TrimSpace(*dest) == "" {
		fmt.Fprintln(os.Stderr, "error: -dest is required")
		fmt.Fprintln(os.Stderr)
		printUsage()
		os.Exit(1)
	}

	xml, err := readTemplate(*templatePath)
	must(err)

	err = printer.Print(context.Background(), job.Job{
		XML:        xml,
		Model:      *model,
		Dest:       *dest,
		Beep:       *beep,
		Drawer:     *drawer,
		DrawerPin:  *drawerPin,
		Cut:        *cut,
		PartialCut: *partialCut,
		Feed:       *feed,
		Assets:     *assets,
		DPI:        203,
	})
	must(err)
}

func printUsage() {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "cups-print %s\n\n", version)
	fmt.Fprintf(os.Stderr, "How to use:\n")
	fmt.Fprintf(os.Stderr, "  %s \\\n", name)
	fmt.Fprintf(os.Stderr, "    -template receipt.xml \\\n")
	fmt.Fprintf(os.Stderr, "    -model generic \\\n")
	fmt.Fprintf(os.Stderr, "    -dest 192.168.10.12 \\\n")
	fmt.Fprintf(os.Stderr, "    -assets ./assets \\\n")
	fmt.Fprintf(os.Stderr, "    -beep 1 \\\n")
	fmt.Fprintf(os.Stderr, "    -drawer\n")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  cat receipt.xml | %s -template - -dest MinhaImpressora\n", name)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Destination auto-detect:\n")
	fmt.Fprintf(os.Stderr, "  TM20                 → CUPS queue\n")
	fmt.Fprintf(os.Stderr, "  192.168.0.10         → TCP :9100\n")
	fmt.Fprintf(os.Stderr, "  192.168.0.10:9100    → TCP\n")
	fmt.Fprintf(os.Stderr, "  /dev/usb/lp0         → file/device\n")
	fmt.Fprintf(os.Stderr, "  stdout               → stdout\n")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func readTemplate(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
