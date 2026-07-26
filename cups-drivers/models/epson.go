package models

import drivers "cups-drivers"

func init() {
	drivers.Register(&Epson{})
}

// Epson is the Epson-branded ESC/POS dialect (same command set as Generic).
type Epson struct {
	Generic
}

func (Epson) ID() string   { return "epson" }
func (Epson) Name() string { return "Epson ESC/POS" }

// PrintWidthDots matches many Epson TM 80mm models (512 dots / ~64mm printable).
func (Epson) PrintWidthDots() int { return 512 }

var _ drivers.Driver = Epson{}
