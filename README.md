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

Imagens no XML podem ser arquivo relativo (`-assets`), URL ou `data:image/...;base64,...` (inclui SVG).

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
| `-assets` | vazio | Base path de imagens relativas |
| `-cut` | `true` | Cortar ao final |
| `-feed` | `3` | Linhas antes do corte |
| `-beep` | `0` | Apitos |
| `-drawer` | `false` | Abrir gaveta |

## Templates XML

Componentes: `Page`, `Column`, `Row`, `Text`, `Image`, `Divider`, `Spacer`, `QRCode`, `Barcode`.

Exemplos em [`cups-examples/layouts/`](cups-examples/layouts/):

| Layout | Descrição |
|--------|-----------|
| `simple.xml` | Curto, sem imagem (não precisa de `-assets`) |
| `grandchef.xml` | Cupom simples com logo |
| `receipt.xml` | Fechamento de pedido |
| `delivery.xml` | Relatório para entrega |

Referência completa do motor: [`cups-template/README.md`](cups-template/README.md).

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

Baixe o arquivo da sua plataforma na Release e coloque junto do XML (e da pasta `assets`, se o layout usar imagens).

**Linux / macOS**

```bash
chmod +x cups-print-linux-amd64   # ou darwin-arm64, etc.

./cups-print-linux-amd64 \
  -template cupom.xml \
  -assets ./assets \
  -model bematech \
  -dest 192.168.18.133
```

Stdin:

```bash
cat cupom.xml | ./cups-print-linux-amd64 -template - -model bematech -dest 192.168.18.133
```

**Windows (cmd / PowerShell)**

```bat
cups-print-windows-amd64.exe -template cupom.xml -assets .\assets -model bematech -dest 192.168.18.133
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
