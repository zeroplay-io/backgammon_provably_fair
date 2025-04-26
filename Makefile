# Backgammon Provably-Fair Verifier – Master Makefile
# ---------------------------------------------------
# build : compile native CLI + WebAssembly module
# dist  : publish web release (runs user-supplied dist.local.mk if present)

.PHONY: build build/verifier build/wasm dist

# complete build (native + WASM)
build: build/verifier build/wasm

# native command-line verifier
build/verifier:
	@mkdir -p build
	go build -o build/ ./cmd/verifier

# WebAssembly verifier for browser use
build/wasm:
	GOOS=js GOARCH=wasm go build -o web/verifier.wasm ./cmd/wasm

# publish web distribution (delegates to optional local script)
dist:
	@if [ -f dist.local.mk ]; then \
		echo "==> Publishing via dist.local.mk"; \
		$(MAKE) -f dist.local.mk; \
	else \
		echo "Skip: dist.local.mk not found"; \
	fi

