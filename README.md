# cups-utils

Monorepo para **gerar e imprimir cupons térmicos** a partir de templates XML.

Não usa Chromium, Puppeteer nem WebKit. O cupom é layoutado em Go, renderizado em bitmap e enviado à impressora em ESC/POS (ou PNG para preview/CUPS).

## Visão geral

```
XML do cupom
    ↓
cups-template   → layout + render (PNG / ESC/POS)
    ↓
cups-drivers    → Init, Raster, Cut, Beep, Gaveta (por modelo)
    ↓
cups-printer    → monta o job e envia (TCP / CUPS / arquivo / stdout)
```

Em produção, o binário **`cups-print`** já embute template + drivers + transporte. Quem integra (Node, PHP, etc.) só precisa spawnar esse executável com o XML e o destino da impressora.

## Módulos

| Pasta | Função |
|-------|--------|
| [`cups-template`](cups-template/) | Parser XML, layout, fontes embutidas, render PNG e ESC/POS |
| [`cups-drivers`](cups-drivers/) | Dialetos por marca (`generic`, `epson`, `bematech`) |
| [`cups-printer`](cups-printer/) | CLI/lib que renderiza e envia o job à impressora |
| [`cups-examples`](cups-examples/) | Layouts de exemplo, `.env`, scripts Node e `make` do `cups-print` |

## Como funciona (térmica)

1. O XML descreve o cupom (`Page`, `Row`, `Text`, `Image`, …).
2. O motor calcula o layout e **desenha um bitmap** (usa fontes TrueType embutidas — a impressora recebe pontos, não texto).
3. A largura efetiva é a da **cabeça térmica** (ex.: 576 dots), não a largura nominal da bobina 80 mm.
4. O bitmap vira comandos `GS v 0` (raster) + feed/cut/beep/gaveta do modelo escolhido.
5. O payload vai por TCP `:9100`, fila CUPS raw, arquivo/device ou stdout.

Imagens no XML: arquivo relativo (`-assets`), URL `http(s)`, ou **`data:` embutido no XML** (recomendado em produção — sem pasta assets). Detalhes na seção [Exemplos de layout](#exemplos-de-layout-tipos-de-cupom).

## Requisitos

- **Go** 1.22+ (1.26 usado no desenvolvimento)
- **Node.js** 18+ (opcional, só para os exemplos em `cups-examples`)
- Impressora térmica em modo **ESC/POS** (Bematech: configurar ESC/POS na utilitário da marca)

## Início rápido

### 1. Gerar o binário de impressão

```bash
cd cups-examples
make          # → bin/cups-print
```

Cross-compile (inclui Windows):

```bash
make build    # → bin/cups-print-{os}-{arch}[.exe]
```

### 2. Imprimir um layout

```bash
./bin/cups-print \
  -template layouts/simple.xml \
  -model bematech \
  -dest 192.168.18.133
```

Com assets (logo, etc.):

```bash
./bin/cups-print \
  -template layouts/delivery.xml \
  -assets ./assets \
  -model bematech \
  -dest 192.168.18.133
```

### 3. Preview em PNG (Node)

```bash
cd cups-examples
node js/render-png.js simple
node js/render-png.js delivery
# → out/<layout>.png
```

### 4. Imprimir via Node

Configure `cups-examples/.env`:

```bash
cp .env.example .env
# PRINTER=192.168.18.133
# PRINTER_MODEL=bematech
```

```bash
node js/print-tcp.js simple
node js/print-tcp.js delivery
node js/print-dynamic.js print   # XML montado em JavaScript
```

O Node **não** renderiza o cupom: ele chama `bin/cups-print` (ou `go run` se o binário não existir).

## Destino da impressora (`-dest`)

Informe só o identificador — o protocolo é detectado automaticamente:

| Valor | Resolve para |
|-------|----------------|
| `192.168.0.10` | TCP JetDirect porta **9100** |
| `192.168.0.10:9100` | TCP |
| `TM20` | fila CUPS |
| `/dev/usb/lp0` | arquivo/device |
| `stdout` | debug (bytes no terminal) |

Esquemas explícitos (`tcp://…`, `cups:…`, `file:…`) também são aceitos.

O `.env` com `PRINTER=…` fica no **serviço chamador** (ex.: Node). O `cups-printer` só recebe `-dest`.

## Modelos (`-model`)

| ID | Uso |
|----|-----|
| `generic` | ESC/POS comum (corte `GS V`) |
| `epson` | Epson ESC/POS (largura típica 512 dots) |
| `bematech` | Raster ESC/POS + corte/beep/gaveta Bematech (`ESC i`) |

Na Bematech, use `-model bematech`. Com `generic`, o corte `GS V` pode imprimir a letra **V**.

Novos modelos: veja [`cups-drivers/README.md`](cups-drivers/README.md).

## Flags principais do `cups-print`

| Flag | Default | Descrição |
|------|---------|-----------|
| `-template` | `-` (stdin) | Caminho do XML |
| `-dest` | obrigatório | Identificador da impressora |
| `-model` | `generic` | Driver |
| `-assets` | vazio | Base path **só** para imagens com caminho relativo de arquivo |
| `-cut` | `true` | Cortar ao final |
| `-feed` | `3` | Linhas antes do corte |
| `-beep` | `0` | Apitos |
| `-drawer` | `false` | Abrir gaveta |

## Exemplos de layout (tipos de cupom)

Todos ficam em [`cups-examples/layouts/`](cups-examples/layouts/). Componentes XML: `Page`, `Column`, `Row`, `Text`, `Image`, `Divider`, `Spacer`, `QRCode`, `Barcode`.

### Resumo

| Layout | Precisa de `-assets`? | O que demonstra |
|--------|------------------------|-----------------|
| `simple.xml` | Não | Cupom curto, só texto |
| `datauri-svg-base64.xml` | Não | Logo embutido em `data:…;base64` (SVG) |
| `datauri-svg.xml` | Não | Logo embutido em `data:` SVG URL-encoded |
| `datauri-png-base64.xml` | Não | Logo embutido em `data:…;base64` (PNG) |
| `demo.xml` | **Sim** | Logo como arquivo (`logo.svg` na pasta assets) |
| `receipt.xml` | **Sim** | Fechamento de pedido (mesa, senha, totais) |
| `delivery.xml` | **Sim** | Relatório para entrega (endereço, telefone) |

### 1. Sem imagem — `simple.xml`

Texto puro. Não use `-assets`.

```bash
./cups-examples/bin/cups-print \
  -template ./cups-examples/layouts/simple.xml \
  -model bematech \
  -dest 192.168.18.133 \
  -cut
```

### 2. Com imagem **sem** pasta assets (`data:` URI)

A imagem vai **dentro do XML**, no atributo `src` do `<Image>`.  
Nesse caso **não passe `-assets`** — não há arquivo externo.

Formatos aceitos no `data:`:

| Forma | Exemplo de `src` |
|-------|------------------|
| SVG em Base64 | `data:image/svg+xml;base64,PHN2Zy...` |
| SVG URL-encoded | `data:image/svg+xml,%3Csvg%20xmlns%3D...` |
| PNG em Base64 | `data:image/png;base64,iVBORw0KGgo...` |
| JPEG em Base64 | `data:image/jpeg;base64,/9j/4AAQ...` |

Trecho de XML:

```xml
<Image
  src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciLi4u"
  width="72"
  height="72"
  align="center"/>
```

Gerar o Base64 a partir de um arquivo:

```bash
# macOS
base64 -i logo.svg | tr -d '\n'

# Linux
base64 -w0 logo.svg
```

Depois monte: `data:image/svg+xml;base64,` + o resultado (ou `data:image/png;base64,` para PNG).

Layouts prontos para testar (já com o logo do projeto embutido):

```bash
BIN=./cups-examples/bin/cups-print
DEST=192.168.18.133   # ajuste o IP
MODEL=bematech

# SVG Base64 — sem -assets
$BIN -template ./cups-examples/layouts/datauri-svg-base64.xml \
  -model "$MODEL" -dest "$DEST" -cut

# SVG URL-encoded — sem -assets
$BIN -template ./cups-examples/layouts/datauri-svg.xml \
  -model "$MODEL" -dest "$DEST" -cut

# PNG Base64 — sem -assets
$BIN -template ./cups-examples/layouts/datauri-png-base64.xml \
  -model "$MODEL" -dest "$DEST" -cut
```

Preview PNG:

```bash
cd cups-examples
node js/render-png.js datauri-svg-base64
node js/render-png.js datauri-png-base64
```

**Produção:** o app gera o XML com o `data:` já preenchido (logo do cliente em Base64) e envia o XML inteiro no stdin do `cups-print`. Nenhuma pasta de imagens precisa ir junto do binário.

### 3. Com imagem em arquivo — `-assets`

Use quando o XML aponta para um arquivo relativo, por exemplo `src="logo.svg"`.

```xml
<Image src="logo.svg" width="72" height="72" align="center"/>
```

Aí `-assets` é a pasta onde esse arquivo está:

```bash
./cups-examples/bin/cups-print \
  -template ./cups-examples/layouts/demo.xml \
  -assets ./cups-examples/assets \
  -model bematech \
  -dest 192.168.18.133 \
  -cut
```

Também com assets:

```bash
# Fechamento de pedido
./cups-examples/bin/cups-print \
  -template ./cups-examples/layouts/receipt.xml \
  -assets ./cups-examples/assets \
  -model bematech -dest 192.168.18.133 -cut

# Relatório para entrega
./cups-examples/bin/cups-print \
  -template ./cups-examples/layouts/delivery.xml \
  -assets ./cups-examples/assets \
  -model bematech -dest 192.168.18.133 -cut
```

### 4. URL remota (opcional)

```xml
<Image src="https://exemplo.com/logo.svg" width="72" height="72"/>
```

Precisa de rede no momento da impressão. Em produção, prefira `data:` Base64 para não depender de internet.

### Quando usar cada abordagem

| Situação | Recomendação |
|----------|----------------|
| App gera cupom e já tem o logo em memória/bytes | `data:image/...;base64,...` no XML (**sem** `-assets`) |
| Layouts fixos em disco com pastinha de logos | arquivo relativo + `-assets` |
| Logo hospedado só na web | URL `https://…` (menos ideal offline) |
| Cupom só texto | sem `<Image>`, sem `-assets` |

Referência do motor XML: [`cups-template/README.md`](cups-template/README.md).

## Integração (produção)

1. Empacote o `cups-print` da plataforma alvo (`linux-amd64`, `windows-amd64`, …).
2. No app, faça spawn com XML (arquivo ou stdin) + `-dest` + `-model`.
3. **Não** dispare vários prints em paralelo na **mesma** impressora — os bytes ESC/POS se misturam. Use uma fila por destino.
4. Impressoras diferentes em paralelo: ok.

Exemplo conceitual (Node):

```js
import { spawn } from "node:child_process";
import { readFileSync } from "node:fs";

const xml = readFileSync("cupom.xml");
const child = spawn("./cups-print", [
  "-template", "-",
  "-model", "bematech",
  "-dest", "192.168.0.10",
], { stdio: ["pipe", "inherit", "inherit"] });

child.stdin.end(xml);
```

## Concorrência e performance

- Cada invocação do CLI = **um** cupom.
- Render local de um cupom simples costuma ser dezenas de ms; lentidão costuma vir de `go run`, rede ou buffer da impressora.
- Sempre use o binário compilado em produção (`make` / `make build`), não `go run`.
- Arquitetura atual é **raster** (bitmap). Modo texto nativo da impressora seria outro caminho (sem as mesmas liberdades de layout/SVG).

## Releases (CI)

Ao criar e enviar uma tag `v*`, o GitHub Actions gera os binários e anexa na **Release** da tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Artefatos publicados:

| Arquivo | Plataforma |
|---------|------------|
| `cups-print-linux-amd64` | Linux x86_64 |
| `cups-print-linux-arm64` | Linux ARM64 |
| `cups-print-darwin-amd64` | macOS Intel |
| `cups-print-darwin-arm64` | macOS Apple Silicon |
| `cups-print-windows-amd64.exe` | Windows x64 |
| `cups-print-windows-arm64.exe` | Windows ARM |
| `checksums.txt` | SHA-256 de todos |

Página: [Releases](https://github.com/yurih567/cups-utils/releases)

### Usar o binário da Release

Baixe o arquivo da sua plataforma na Release. Se o XML usar só `data:` ou for só texto, **não precisa** de pasta `assets`.

**Linux / macOS**

```bash
chmod +x cups-print-linux-amd64

# sem assets (texto ou data: URI no XML)
./cups-print-linux-amd64 \
  -template cupom.xml \
  -model bematech \
  -dest 192.168.18.133 \
  -cut

# com arquivos de imagem relativos
./cups-print-linux-amd64 \
  -template cupom.xml \
  -assets ./assets \
  -model bematech \
  -dest 192.168.18.133 \
  -cut
```

Stdin:

```bash
cat cupom.xml | ./cups-print-linux-amd64 -template - -model bematech -dest 192.168.18.133
```

**Windows (cmd / PowerShell)**

```bat
REM sem assets
cups-print-windows-amd64.exe -template cupom.xml -model bematech -dest 192.168.18.133 -cut

REM com pasta de imagens
cups-print-windows-amd64.exe -template cupom.xml -assets .\assets -model bematech -dest 192.168.18.133 -cut
```

**Com Node** (após baixar o binário):

```bash
export CUPS_PRINT_BIN=./cups-print-linux-amd64
node js/print-tcp.js simple
```

No Windows:

```bat
set CUPS_PRINT_BIN=cups-print-windows-amd64.exe
node js\print-tcp.js simple
```

Workflow: [`.github/workflows/release.yml`](.github/workflows/release.yml)

## Desenvolvimento

```bash
# testes
cd cups-drivers && go test ./...
cd ../cups-template && go test ./...
cd ../cups-printer && go test ./...

# printer multiplataforma (alternativa ao make do examples)
cd cups-printer && make build
```

## Documentação por módulo

- [cups-template](cups-template/README.md) — XML, layout, PNG, ESC/POS  
- [cups-drivers](cups-drivers/README.md) — modelos e novos drivers  
- [cups-printer](cups-printer/README.md) — CLI, destinos, biblioteca Go  
- [cups-examples](cups-examples/README.md) — layouts, Node, `.env`, Makefile  
