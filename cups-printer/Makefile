APP       := print
CMD       := ./cmd/print
DIST      := dist
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)

# Market-standard naming: print-{os}-{arch}[.exe]
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: all build local clean list help

all: build

help:
	@echo "Targets:"
	@echo "  make build   - cross-compile $(APP) for all platforms (pure Go, CGO=0)"
	@echo "  make local   - build for current OS/arch only"
	@echo "  make clean   - remove $(DIST)/"
	@echo "  make list    - show output binaries"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Output:  $(DIST)/$(APP)-{os}-{arch}"

build: $(DIST)
	@set -e; \
	for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		out="$(DIST)/$(APP)-$$os-$$arch$$ext"; \
		echo "→ $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -trimpath -ldflags="$(LDFLAGS)" -o "$$out" $(CMD); \
	done
	@echo "Done. Artifacts in $(DIST)/"

local: $(DIST)
	@os=$$(go env GOOS); arch=$$(go env GOARCH); \
	ext=""; if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
	out="$(DIST)/$(APP)-$$os-$$arch$$ext"; \
	echo "→ $$out"; \
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o "$$out" $(CMD)

list:
	@ls -la $(DIST) 2>/dev/null || echo "$(DIST)/ is empty — run make build"

clean:
	rm -rf $(DIST)

$(DIST):
	mkdir -p $(DIST)
