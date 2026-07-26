ROOT         := ..
PRINTER_DIR  := $(ROOT)/cups-printer
TEMPLATE_DIR := $(ROOT)/cups-template
BIN          := bin
VERSION      ?= $(shell git -C $(ROOT) describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS      := -s -w

GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
EXT    := $(if $(filter windows,$(GOOS)),.exe,)

# Single self-contained CLI (template engine + drivers + transport).
TOOL := $(BIN)/cups-print$(EXT)

PRINT_BIN  := $(BIN)/print-$(GOOS)-$(GOARCH)$(EXT)
RENDER_BIN := $(BIN)/render-$(GOOS)-$(GOARCH)$(EXT)

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: all tool local build clean list help print render

all: tool

help:
	@echo "Targets:"
	@echo "  make / make tool   - one binary: $(TOOL) (template + print)"
	@echo "  make local         - also build platform-named print/render into $(BIN)/"
	@echo "  make build         - cross-compile cups-print for all platforms"
	@echo "  make print         - build only cups-printer (platform name)"
	@echo "  make render        - build only cups-template (platform name)"
	@echo "  make clean         - remove $(BIN)/"
	@echo "  make list          - list binaries"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo ""
	@echo "Usage:"
	@echo "  $(TOOL) \\"
	@echo "    -template layouts/delivery.xml \\"
	@echo "    -assets ./assets \\"
	@echo "    -model bematech \\"
	@echo "    -dest 192.168.18.133"

# One binary with everything needed to print (XML → ESC/POS → printer).
tool: $(BIN)
	@echo "→ $(TOOL)"
	@CGO_ENABLED=0 go build -C $(PRINTER_DIR) -trimpath -ldflags="$(LDFLAGS)" \
		-o "$(abspath $(TOOL))" ./cmd/print
	@echo "Done: $(abspath $(TOOL))"
	@echo ""
	@echo "Example:"
	@echo "  $(TOOL) -template layouts/delivery.xml -assets ./assets -model bematech -dest 192.168.18.133"

local: tool print render
	@echo "Done. Binaries in $(BIN)/"

print: $(BIN)
	@echo "→ $(PRINT_BIN)"
	@CGO_ENABLED=0 go build -C $(PRINTER_DIR) -trimpath -ldflags="$(LDFLAGS)" -o "$(abspath $(PRINT_BIN))" ./cmd/print

render: $(BIN)
	@echo "→ $(RENDER_BIN)"
	@CGO_ENABLED=0 go build -C $(TEMPLATE_DIR) -trimpath -ldflags="$(LDFLAGS)" -o "$(abspath $(RENDER_BIN))" ./cmd/render

build: $(BIN)
	@set -e; \
	for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		out="$(BIN)/cups-print-$$os-$$arch$$ext"; \
		echo "→ $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -C $(PRINTER_DIR) -trimpath -ldflags="$(LDFLAGS)" -o "$(abspath $$out)" ./cmd/print; \
	done
	@echo "Done. Artifacts in $(BIN)/"

list:
	@ls -la $(BIN) 2>/dev/null || echo "$(BIN)/ is empty — run make tool"

clean:
	rm -rf $(BIN)

$(BIN):
	mkdir -p $(BIN)
