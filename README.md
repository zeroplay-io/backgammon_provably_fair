# Provably-Fair Backgammon Verifier

This repository provides **all the client-side tools players need to independently
verify every dice roll in our Backgammon game**.  
The C++ game server (private repo) publishes a JSON report at the end of each
match; everything here is used to confirm that report.

## Backgammon - Fair Board Game

<p>
  <a href="https://apps.apple.com/app/id6739750400" target="_blank">
    <img alt="Download Backgammon from AppStore" src="img/stores/app-store.svg" height="40" style="margin-right:20px;">
  </a>
  <a href="https://play.google.com/store/apps/details?id=io.zeroplay.backgammon2" target="_blank">
    <img alt="Download Backgammon from Google Play" src="img/stores/google-play.png" height="40" style="margin-right:20px;">
  </a>
  <a href="https://play.zeroplay.io/backgammon/" target="_blank">
    <img alt="Play Backgammon on Web" src="img/stores/h5.svg" height="40">
  </a>
</p>



## What’s inside

| Path | Description |
|------|-------------|
| `pkg/verifier` | Pure-Go library that reproduces the dice stream from a JSON report and returns `true/false`. |
| `cmd/verifier` | Tiny command-line wrapper around the library. |
| `cmd/wasm` | Same code compiled to WebAssembly and exported as a single JS function `verify(json)`. |
| `web/` | Minimal HTML page that loads `verifier.wasm` and lets a player paste a report to get an immediate ✅ / ❌ result. |


## Quick start

### 1. Build the CLI

```bash
go run ./cmd/verifier < report.json   # prints VERIFIED ✅ or error
```

### 2. Build the browser demo

```bash
# Compile to WASM
GOOS=js GOARCH=wasm go build -o web/verifier.wasm ./cmd/wasm

# Copy the Go runtime
cp $(go env GOROOT)/misc/wasm/wasm_exec.js web/

# Open web/index.html (or drop the folder into your game's WebView)
```

The verifier reproduces every roll with HMAC-SHA256(serverSeed, combinedSeed || nonce_be) and checks it against the rolls array.
