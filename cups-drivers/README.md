# cups-drivers

Catálogo compartilhado de comandos por modelo de impressora térmica.

Usado por:

- **cups-template** — `Init`, `ClearBuffer`, `LineEnd`, `Align*`, `Raster`
- **cups-printer** — `Feed`, `Cut`, `Beep`, `OpenDrawer`

## Uso

```go
import (
    drivers "cups-drivers"
    _ "cups-drivers/models" // registra generic, epson, bematech, ...
)

d := drivers.MustGet("generic")

body := append([]byte{}, d.Init()...)
body = append(body, d.ClearBuffer()...)
// ... raster / conteúdo gerado pelo template ...

tail := append(d.Feed(3), d.Cut(false)...)
tail = append(tail, d.Beep(1)...)
```

## Adicionar um novo modelo

1. Crie `models/<marca>.go`:

```go
package models

import drivers "cups-drivers"

func init() {
    drivers.Register(&MinhaMarca{})
}

type MinhaMarca struct{}

func (MinhaMarca) ID() string   { return "minhamarca" }
func (MinhaMarca) Name() string { return "Minha Marca" }

func (MinhaMarca) Init() []byte        { return []byte{0x1b, '@'} }
func (MinhaMarca) ClearBuffer() []byte { return []byte{0x1b, '@'} }
func (MinhaMarca) LineEnd() []byte     { return []byte{0x0a} }
func (MinhaMarca) AlignCenter() []byte { return []byte{0x1b, 'a', 1} }
func (MinhaMarca) AlignLeft() []byte   { return []byte{0x1b, 'a', 0} }

func (MinhaMarca) Raster(width, height int, data []byte) ([]byte, error) {
    return rasterGSv0(width, height, data) // ou envelope próprio
}

func (m MinhaMarca) Feed(lines int) []byte { /* ... */ return nil }
func (MinhaMarca) Cut(partial bool) []byte { return nil }
func (MinhaMarca) Beep(times int) []byte   { return nil }
func (MinhaMarca) OpenDrawer(pin int) []byte { return nil }

var _ drivers.Driver = MinhaMarca{}
```

2. Pronto — o `init()` registra no registry. Consumidores que já fazem `import _ "cups-drivers/models"` passam a enxergar o modelo sem mudança de switch.

## Modelos

| ID | Dialeto |
|----|---------|
| `generic` | ESC/POS padrão (compatível com a maioria das térmicas) |
| `epson` | ESC/POS (mesmo conjunto do `generic`) |
| `bematech` | ESC/Bema |

Liste em runtime: `drivers.List()`.
