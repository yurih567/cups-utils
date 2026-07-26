package font

import (
	"fmt"
	"os"
	"sync"

	"cups-template/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type FaceKey struct {
	Family string
	Weight int
	Size   float64
}

type Manager struct {
	mu    sync.RWMutex
	fonts map[string]map[int]*opentype.Font
	faces map[FaceKey]font.Face
	dpi   float64
}

func NewManager(dpi float64) *Manager {
	if dpi <= 0 {
		dpi = 96
	}
	return &Manager{
		fonts: make(map[string]map[int]*opentype.Font),
		faces: make(map[FaceKey]font.Face),
		dpi:   dpi,
	}
}

// NewDefault returns a manager with the fixed built-in Arial-like face (regular + bold).
func NewDefault(dpi float64) (*Manager, error) {
	m := NewManager(dpi)
	if err := m.Register(fonts.Family, 400, fonts.Regular); err != nil {
		return nil, err
	}
	if err := m.Register(fonts.Family, 700, fonts.Bold); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) RegisterFile(family string, weight int, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read font %s: %w", path, err)
	}
	return m.Register(family, weight, data)
}

func (m *Manager) Register(family string, weight int, data []byte) error {
	f, err := opentype.Parse(data)
	if err != nil {
		return fmt.Errorf("parse font %s/%d: %w", family, weight, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.fonts[family] == nil {
		m.fonts[family] = make(map[int]*opentype.Font)
	}
	m.fonts[family][weight] = f
	return nil
}

func (m *Manager) Face(family string, weight int, size float64) (font.Face, error) {
	if size <= 0 {
		size = 14
	}
	key := FaceKey{Family: family, Weight: weight, Size: size}

	m.mu.RLock()
	if face, ok := m.faces[key]; ok {
		m.mu.RUnlock()
		return face, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if face, ok := m.faces[key]; ok {
		return face, nil
	}

	otFont, err := m.lookupLocked(family, weight)
	if err != nil {
		return nil, err
	}

	// size is a design pixel at 96 DPI (CSS-like). Physical size stays stable
	// across preview (96) and thermal print (203) DPIs.
	const designDPI = 96.0
	face, err := opentype.NewFace(otFont, &opentype.FaceOptions{
		Size:    size * 72 / designDPI,
		DPI:     m.dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	m.faces[key] = face
	return face, nil
}

func (m *Manager) Measure(family string, weight int, size float64, text string) (width, height float64, err error) {
	face, err := m.Face(family, weight, size)
	if err != nil {
		return 0, 0, err
	}
	metrics := face.Metrics()
	height = float64(metrics.Ascent+metrics.Descent) / 64
	width = float64(font.MeasureString(face, text)) / 64
	return width, height, nil
}

func (m *Manager) lookupLocked(family string, weight int) (*opentype.Font, error) {
	byWeight, ok := m.fonts[family]
	if !ok {
		for _, candidate := range m.fonts {
			byWeight = candidate
			ok = true
			break
		}
		if !ok {
			return nil, fmt.Errorf("no fonts registered")
		}
	}

	if f, ok := byWeight[weight]; ok {
		return f, nil
	}
	if weight >= 600 {
		if f, ok := byWeight[700]; ok {
			return f, nil
		}
	}
	if f, ok := byWeight[400]; ok {
		return f, nil
	}
	for _, f := range byWeight {
		return f, nil
	}
	return nil, fmt.Errorf("font family %q has no faces", family)
}
