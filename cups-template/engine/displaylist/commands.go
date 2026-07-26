package displaylist

import (
	"image"
	"image/color"
)

type Command interface {
	command()
}

type DrawRect struct {
	X, Y, Width, Height float64
	Fill                color.Color
	Stroke              color.Color
	StrokeWidth         float64
	Radius              float64
	Opacity             float64
	Rotate              float64
}

func (DrawRect) command() {}

type DrawText struct {
	X, Y, Width, Height float64
	Text                string
	FontFamily          string
	FontSize            float64
	FontWeight          int
	Color               color.Color
	Align               int
	Opacity             float64
	Rotate              float64
}

func (DrawText) command() {}

type DrawImage struct {
	X, Y, Width, Height float64
	Image               image.Image
	Src                 string
	Opacity             float64
	Rotate              float64
}

func (DrawImage) command() {}

type DrawLine struct {
	X1, Y1, X2, Y2 float64
	Color          color.Color
	StrokeWidth    float64
	Opacity        float64
}

func (DrawLine) command() {}

type DrawQRCode struct {
	X, Y, Width, Height float64
	Data                string
	Opacity             float64
}

func (DrawQRCode) command() {}

type DrawBarcode struct {
	X, Y, Width, Height float64
	Data                string
	Format              string
	Opacity             float64
}

func (DrawBarcode) command() {}
