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

var _ drivers.Driver = Epson{}
