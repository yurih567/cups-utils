# cups-printer

Serviço CLI que recebe um XML compatível com o `cups-template`, escolhe o driver (`cups-drivers`), renderiza o cupom e envia para a impressora — com sinais de corte, apito e gaveta.

## Fluxo

```
XML + model + dest + beep/drawer
        ↓
cups-template (lib) → corpo ESC/POS
        ↓
cups-drivers → Feed / Cut / Beep / OpenDrawer
        ↓
transport (CUPS raw | TCP 9100 | file | stdout)
```

## Build

Local (máquina atual):

```bash
go build -o print ./cmd/print
# ou
make local
```

Multiplataforma (100% Go, `CGO_ENABLED=0`):

```bash
make build
```

Gera em `dist/` no padrão de mercado `print-{os}-{arch}`:

```
dist/print-linux-amd64
dist/print-linux-arm64
dist/print-darwin-amd64
dist/print-darwin-arm64
dist/print-windows-amd64.exe
dist/print-windows-arm64.exe
```

PHP/Node/etc. só precisam do binário da plataforma alvo e spawnar o processo.

## Uso

```bash
./print \
  -template ../cups-template/templates/receipt.xml \
  -model epson \
  -dest MinhaImpressora \
  -beep 1 \
  -drawer \
  -assets ../cups-template/assets
```

Stdin:

```bash
cat cupom.xml | ./print -template - -model bematech -dest 192.168.0.10 -beep 1
```

Debug (bytes no terminal):

```bash
./print -template cupom.xml -dest stdout > cupom.bin
```

## Flags

| Flag | Default | Descrição |
|------|---------|-----------|
| `-template` | `-` | Path do XML ou `-` (stdin) |
| `-model` | `generic` | ID do driver (`generic`, `epson`, `bematech`, …) |
| `-dest` | obrigatório | Identificador da impressora (protocolo auto) |
| `-beep` | `0` | Vezes do apito (`0` = off) |
| `-drawer` | `false` | Abrir gaveta |
| `-drawer-pin` | `0` | Pino da gaveta (Epson: 0 ou 1) |
| `-cut` | `true` | Cortar ao final |
| `-partial-cut` | `false` | Corte parcial |
| `-feed` | `3` | Linhas antes do corte |
| `-assets` | vazio | Base path de imagens relativas |

Fonte fixa embutida no `cups-template` (estilo Arial). Não há flag `-fonts`.

## Destinos (`-dest`)

O chamador passa só o identificador — o protocolo é detectado automaticamente:

| Valor | Resolve para |
|-------|----------------|
| `TM20` | fila CUPS |
| `192.168.0.10` | TCP JetDirect `:9100` |
| `192.168.0.10:9100` | TCP |
| `printer.local:9100` | TCP |
| `/dev/usb/lp0` | device/arquivo |
| `stdout` ou `-` | stdout (debug) |

Esquemas explícitos (`tcp://…`, `cups:…`, `file:…`) continuam válidos.

Configuração de ambiente (`.env`) fica no serviço que chama o `cups-printer`, não neste módulo.

## Biblioteca

```go
import (
    "context"
    "cups-printer/job"
    "cups-printer/printer"
)

err := printer.Print(context.Background(), job.Job{
    XML:    xmlBytes,
    Model:  "generic",
    Dest:   "192.168.0.10", // bare IP → TCP :9100
    Beep:   1,
    Drawer: true,
    Cut:    true,
    Feed:   3,
    Assets: "/path/to/assets",
})
```

## Novos modelos

Adicione um arquivo em `cups-drivers/models/` — o printer só precisa receber `-model <id>`.
