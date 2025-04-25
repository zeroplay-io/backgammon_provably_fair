//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/gopherd/backgammon_provably_fair/pkg/verifier"
)

func verify(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return "Error: need JSON string"
	}
	jsonStr := args[0].String()

	if err := verifier.VerifyBytes([]byte(jsonStr)); err != nil {
		return "Error: " + err.Error()
	}
	return "OK"
}

func main() {
	js.Global().Set("verify", js.FuncOf(verify))
	<-make(chan struct{}) // keep alive
}
