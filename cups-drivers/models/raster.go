package models

import "fmt"

// maxRasterBandHeight limits each GS v 0 block so thermal printers
// do not overflow the receive buffer on long receipts.
const maxRasterBandHeight = 256

func rasterGSv0(width, height int, data []byte) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid raster size %dx%d", width, height)
	}
	rowBytes := (width + 7) / 8
	expected := rowBytes * height
	if len(data) < expected {
		return nil, fmt.Errorf("raster data too short: got %d want %d", len(data), expected)
	}

	out := make([]byte, 0, expected+(height/maxRasterBandHeight+1)*8)
	for y := 0; y < height; {
		band := maxRasterBandHeight
		if y+band > height {
			band = height - y
		}
		chunk, err := rasterGSv0Band(width, band, data[y*rowBytes:(y+band)*rowBytes])
		if err != nil {
			return nil, err
		}
		out = append(out, chunk...)
		y += band
	}
	return out, nil
}

func rasterGSv0Band(width, height int, data []byte) ([]byte, error) {
	rowBytes := (width + 7) / 8
	expected := rowBytes * height
	if len(data) < expected {
		return nil, fmt.Errorf("raster band too short: got %d want %d", len(data), expected)
	}

	xL := byte(rowBytes & 0xff)
	xH := byte((rowBytes >> 8) & 0xff)
	yL := byte(height & 0xff)
	yH := byte((height >> 8) & 0xff)

	out := make([]byte, 0, 8+expected)
	out = append(out, gs, 'v', '0', 0, xL, xH, yL, yH)
	out = append(out, data[:expected]...)
	return out, nil
}
