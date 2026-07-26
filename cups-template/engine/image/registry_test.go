package imagecodec

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeDataURIBase64SVG(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"><rect width="10" height="10" fill="black"/></svg>`)
	src := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(svg)

	img, err := DefaultRegistry().DecodeSource(src, "")
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("unexpected bounds: %v", img.Bounds())
	}
}

func TestDecodeDataURIBase64PNG(t *testing.T) {
	// 1x1 black PNG
	raw, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
	)
	if err != nil {
		t.Fatal(err)
	}
	src := "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
	img, err := DefaultRegistry().DecodeSource(src, "")
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 1 || img.Bounds().Dy() != 1 {
		t.Fatalf("unexpected bounds: %v", img.Bounds())
	}
}

	root := findModuleRoot(t)
	path := filepath.Join(root, "assets", "logo.svg")
	if _, err := os.Stat(path); err != nil {
		t.Skip(err)
	}
	img, err := DefaultRegistry().DecodeSource("logo.svg", filepath.Join(root, "assets"))
	if err != nil {
		t.Fatal(err)
	}
	if img == nil {
		t.Fatal("nil image")
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
