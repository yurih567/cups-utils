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
  -dest 'cups:MinhaImpressora' \
  -beep 1 \
  -drawer \
  -assets ../cups-template/assets
```

Stdin:

```bash
cat cupom.xml | ./print -template - -model bematech -dest 'tcp://192.168.0.10:9100' -beep 1
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
| `-dest` | obrigatório | Destino da impressora |
| `-beep` | `0` | Vezes do apito (`0` = off) |
| `-drawer` | `false` | Abrir gaveta |
| `-drawer-pin` | `0` | Pino da gaveta (Epson: 0 ou 1) |
| `-cut` | `true` | Cortar ao final |
| `-partial-cut` | `false` | Corte parcial |
| `-feed` | `3` | Linhas antes do corte |
| `-assets` | vazio | Base path de imagens relativas |

Fonte fixa embutida no `cups-template` (estilo Arial). Não há flag `-fonts`.

## Destinos (`-dest`)

| Forma | Exemplo |
|-------|---------|
| Fila CUPS | `cups:TM20` ou `TM20` |
| Rede JetDirect | `tcp://192.168.0.10:9100` ou `192.168.0.10:9100` |
| Device/arquivo | `file:/dev/usb/lp0` ou `/dev/usb/lp0` |
| Debug | `stdout` ou `-` |

Filas CUPS sempre usam `lp -d Nome -o raw` (payload ESC/POS completo).

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
    Dest:   "cups:MinhaImpressora",
    Beep:   1,
    Drawer: true,
    Cut:    true,
    Feed:   3,
    Assets: "/path/to/assets",
})
```

## Novos modelos

Adicione um arquivo em `cups-drivers/models/` — o printer só precisa receber `-model <id>`.
