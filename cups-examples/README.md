# cups-examples

Layouts de exemplo e scripts Node.js para usar o `cups-utils` a partir de JavaScript.

O Node **não** renderiza o cupom — ele chama os binários Go:

| Binário | Módulo | Função |
|---------|--------|--------|
| `render` | `cups-template` | XML → PNG / ESC/POS |
| `print` | `cups-printer` | XML → impressora (TCP / CUPS / arquivo) |

## Estrutura

```
cups-examples/
├── assets/           # imagens usadas pelos layouts
├── layouts/
│   ├── receipt.xml   # fechamento de pedido
│   ├── delivery.xml  # relatório para entrega
│   └── grandchef.xml # cupom simples
├── js/
│   ├── cups.js           # helpers (spawn dos binários)
│   ├── render-png.js     # gera PNG a partir de um layout
│   ├── print-tcp.js      # imprime um layout
│   └── print-dynamic.js  # monta XML em JS e renderiza/imprime
└── package.json
```

## Pré-requisitos

- Node.js 18+
- Go 1.22+

```bash
cd cups-examples
make          # gera bin/cups-print (único binário para imprimir)
```

### Imprimir só com o binário

`cups-print` já embute o motor de template e os drivers — não precisa do `render` separado:

```bash
./bin/cups-print \
  -template layouts/delivery.xml \
  -assets ./assets \
  -model bematech \
  -dest 192.168.18.133
```

| Flag | Descrição |
|------|-----------|
| `-template` | XML do cupom (ou `-` = stdin) |
| `-dest` | IP, nome CUPS ou path (protocolo auto) |
| `-model` | `generic` \| `epson` \| `bematech` |
| `-assets` | pasta das imagens do layout |
| `-cut` / `-feed` / `-beep` / `-drawer` | hardware |

Os scripts JS usam `bin/cups-print` automaticamente (ou `go run` se ainda não existir).

## Layouts

| Arquivo | Uso |
|---------|-----|
| `layouts/receipt.xml` | Fechamento de pedido (mesa / senha) |
| `layouts/delivery.xml` | Relatório para entrega (endereço / telefone) |
| `layouts/grandchef.xml` | Exemplo mínimo |

## JavaScript

### 1. Gerar PNG de preview

```bash
cd cups-examples
node js/render-png.js receipt
node js/render-png.js delivery ./out/delivery.png
```

Saída padrão: `out/<layout>.png`.

### 2. Imprimir (destino no `.env` do examples)

O `.env` é lido pelo **JavaScript** (serviço chamador). O `cups-printer` só recebe `-dest` e detecta o protocolo.

```bash
cp .env.example .env
# edite PRINTER=192.168.18.133   (IP → TCP automático)
#      PRINTER=TM20               (nome → fila CUPS)
#      PRINTER_MODEL=generic
```

```bash
node js/print-tcp.js delivery
node js/print-tcp.js receipt
# override pontual:
node js/print-tcp.js delivery 10.0.0.50
```

### 3. Montar o XML dinamicamente em JS

```bash
# só gera PNG
node js/print-dynamic.js

# imprime usando PRINTER do .env
node js/print-dynamic.js print
```

### Uso como lib no seu app

```js
import { readFileSync } from "node:fs";
import { printXml, renderPng, assetsDir } from "./js/cups.js";

const xml = readFileSync("./layouts/delivery.xml");

const png = await renderPng(xml, { assets: assetsDir });

// só o IP/nome — o cups-printer escolhe o protocolo
await printXml(xml, {
  dest: "192.168.0.10",
  model: "generic",
});

// ou omita dest/model: o JS lê cups-examples/.env e passa -dest
await printXml(xml);
```

## Destinos (`PRINTER` / `-dest`)

Informe só o identificador; o `cups-printer` resolve o protocolo:

| Valor | Resolve para |
|-------|----------------|
| `192.168.0.10` | TCP `192.168.0.10:9100` |
| `192.168.0.10:9100` | TCP |
| `TM20` | fila CUPS |
| `/dev/usb/lp0` | arquivo/device |
| `stdout` | debug |

## Dica

Em produção, empacote o binário da plataforma alvo (`print-linux-amd64`, etc.) junto com o app Node e defina `CUPS_PRINT_BIN` / `CUPS_RENDER_BIN`.
