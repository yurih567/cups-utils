# CUPS Template Engine

Gerador de templates de impressão para o **modelo padrão CUPS**.

Qualquer impressora registrada no CUPS (térmica, laser, jato de tinta, etc.) pode receber o job — o CUPS cuida da conversão via driver/PPD da fila.

Sem CGO, sem Chromium, sem WebKit, sem Playwright e sem Puppeteer.

## Pipeline

```
Template XML
    ↓
Parser
    ↓
DOM
    ↓
Layout
    ↓
Display List
    ↓
Renderer (CUPS/PNG | ESC/POS raw)
```

Parsing, layout e renderização são etapas separadas. O layout **nunca** desenha — ele apenas calcula boxes e emite uma display list.

## Formatos de saída

| Formato | Quando usar | Como enviar |
|---------|-------------|-------------|
| `cups` (padrão) | Qualquer impressora no CUPS | `lp -d NomeDaImpressora cupom.png` |
| `png` | Preview / mesmo payload do `cups` | igual ao `cups` |
| `escpos` | Fila CUPS raw / térmica ESC/POS | `lp -d NomeDaImpressora -o raw cupom.bin` |

O formato `cups` gera PNG, que o CUPS aceita nativamente e filtra para o idioma da impressora.

## Componentes

`Page` · `Column` · `Row` · `Text` · `Image` · `Divider` · `Spacer` · `QRCode` · `Barcode`

---

## Como executar o serviço gerador

O binário de serviço é `cmd/render`. Ele lê um template XML e escreve o resultado renderizado.

### 1. Build

```bash
go build -o render ./cmd/render
```

### 2. Uso básico (stdin → stdout)

O modo padrão é pipe: XML no stdin, bytes no stdout.

```bash
# Job CUPS (qualquer impressora na fila)
./render -format cups < templates/demo.xml > cupom.png
lp -d NomeDaImpressora cupom.png

# Pipe direto para o CUPS
./render -format cups < templates/receipt.xml | lp -d NomeDaImpressora

# Bytes ESC/POS para fila raw (térmicas) — corpo do cupom; cut/beep ficam no cups-printer
./render -format escpos -model epson < templates/receipt.xml | lp -d NomeDaImpressora -o raw
./render -format escpos -model bematech < templates/receipt.xml > cupom.bin
```

Equivalente com `go run`:

```bash
cat templates/demo.xml | go run ./cmd/render -format cups > cupom.png
cat templates/receipt.xml   | go run ./cmd/render -format escpos > cupom.bin
```

### 3. Arquivo de entrada/saída

```bash
./render \
  -template templates/demo.xml \
  -format cups \
  -out cupom.png

./render \
  -template templates/receipt.xml \
  -format escpos \
  -out cupom.bin
```

A largura do cupom vem do template: `<Page width="80mm">` ou `<Page width="58mm">`.

`-template -` e `-out -` significam stdin/stdout (padrão).

### 4. Chamada a partir de outro serviço

```go
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
)

func renderReceipt(svgLogo []byte) ([]byte, error) {
	b64 := base64.StdEncoding.EncodeToString(svgLogo)

	xml := fmt.Sprintf(`
<Page width="80mm" padding="8">
  <Column gap="8" align="center">
    <Image src="data:image/svg+xml;base64,%s" width="180" height="72"/>
    <Text align="center" size="20" weight="bold">Pedido #4521</Text>
    <Divider char="-"/>
    <Row justify="between">
      <Text>Item</Text>
      <Text>10,00</Text>
    </Row>
  </Column>
</Page>
`, b64)

	cmd := exec.Command("./render", "-format", "cups")
	cmd.Stdin = strings.NewReader(xml)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("render failed: %w: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}
```

Fluxo:

```
serviço chamador
    → monta XML (com imagem em data URI, se precisar)
    → spawn ./render -format cups
    → escreve XML no stdin
    → lê PNG no stdout
    → envia para o CUPS (lp / IPP)
```

### 5. Como passar a imagem

| Forma | Exemplo | Quando usar |
|-------|---------|-------------|
| **data URI** | `src="data:image/svg+xml;base64,..."` | Pipe sem FS compartilhado (**recomendado**) |
| **URL** | `src="https://.../logo.svg"` | Imagem já na rede/CDN |
| **path** | `src="logo.svg"` + `-assets ./assets` | Mesmo filesystem entre serviços |

```xml
<Image
    src="data:image/svg+xml;base64,PHN2Zy..."
    width="180"
    height="72"/>
```

### 6. Flags do `render`

| Flag | Default | Descrição |
|------|---------|-----------|
| `-template` | `-` (stdin) | Path do XML, ou `-` para stdin |
| `-out` | `-` (stdout) | Path de saída, ou `-` para stdout |
| `-format` | `cups` | `cups`, `png` ou `escpos` |
| `-model` | `generic` | Dialeto ESC/POS: `generic`, `epson` ou `bematech` (só `escpos`) |
| `-assets` | vazio | Base path para `src` relativo |
| `-dpi` | `96` / `203` | DPI (`203` automático no escpos) |

Fonte fixa embutida (estilo Arial). Não há flag de fonte customizada.

Comandos de corte, bip e gaveta **não** entram no corpo gerado aqui — ficam no `cups-printer`, usando o mesmo pacote [`cups-drivers`](../cups-drivers).

A largura do papel **não** é flag: use `<Page width="80mm">` ou `<Page width="58mm">`.

### 7. Enviar para a impressora via CUPS

```bash
# Qualquer impressora na fila CUPS (formato cups/png)
lp -d NomeDaImpressora cupom.png

# Fila raw ESC/POS (térmicas)
lp -d NomeDaImpressora -o raw cupom.bin

# Listar filas disponíveis
lpstat -p -d
```

---

## Preview local (desenvolvimento)

Helper em `examples/preview` para gerar arquivos rápido durante o desenvolvimento:

```bash
go run ./examples/preview \
  -template templates/demo.xml \
  -out examples/preview/demo.png

go run ./examples/preview \
  -format escpos \
  -model epson \
  -template templates/receipt.xml \
  -out examples/preview/receipt.bin
```

## API em biblioteca

```go
doc, err := parser.Parse(template)

tree, err := engine.Layout(doc, engine.Options{
    DPI:           96,
    Fonts:         fonts, // font.NewDefault(dpi)
    AssetBasePath: "./assets",
})

commands := engine.BuildDisplayList(tree)
img := pngrenderer.NewRenderer(fonts).Render(commands)
png.Encode(file, img)
```

## Estrutura

```
cmd/render/          # CLI de serviço (stdin/stdout)
engine/
  parser/
  layout/
  style/
  dom/
  units/
  font/
  displaylist/
  image/
renderers/
  png/               # payload CUPS padrão (+ preview)
  escpos/            # payload raw via cups-drivers (init/align/raster)
templates/
examples/preview/
assets/
fonts/               # TTF embutida (go:embed) — família Arial
```

Drivers de impressora (generic, epson, bematech, …) vivem em [`../cups-drivers`](../cups-drivers).

## Extensão

Novos renderizadores (PDF, SVG, PostScript) consomem a mesma `[]displaylist.Command` sem alterar o layout.
