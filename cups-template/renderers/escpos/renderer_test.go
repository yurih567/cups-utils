package escpos

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	_ "cups-drivers/models"
)

func TestEncodeImageFitsPrintableWidth(t *testing.T) {
	// Simulate an 80mm@203dpi page (~640px) that exceeds the 576-dot head.
	src := image.NewRGBA(image.Rect(0, 0, 640, 40))
	draw.Draw(src, src.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	// Black pixels on the far right — would be clipped without fitWidth.
	for y := 0; y < 40; y++ {
		src.Set(639, y, color.Black)
		src.Set(0, y, color.Black)
	}

	data, err := EncodeImage(src, Options{Threshold: 128, MaxWidthDots: DefaultMaxWidthDots})
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 16 {
		t.Fatalf("payload too short: %d", len(data))
	}

	fitted := fitWidth(src, DefaultMaxWidthDots)
	if fitted.Bounds().Dx() != DefaultMaxWidthDots {
		t.Fatalf("fit width = %d want %d", fitted.Bounds().Dx(), DefaultMaxWidthDots)
	}
	aligned := alignWidth(fitted)
	if aligned.Bounds().Dx() > DefaultMaxWidthDots {
		t.Fatalf("aligned width %d exceeds max %d", aligned.Bounds().Dx(), DefaultMaxWidthDots)
	}
}

func TestEncodeImageUsesAlignLeft(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 64, 8))
	draw.Draw(src, src.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	data, err := EncodeImage(src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	// ESC a 0 = align left
	found := false
	for i := 0; i+2 < len(data); i++ {
		if data[i] == 0x1b && data[i+1] == 'a' && data[i+2] == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected ESC a 0 (align left) in payload")
	}
	for i := 0; i+2 < len(data); i++ {
		if data[i] == 0x1b && data[i+1] == 'a' && data[i+2] == 1 {
			t.Fatal("should not send align center for full-width raster")
		}
	}
}
