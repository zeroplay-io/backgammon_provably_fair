package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gopherd/backgammon_provably_fair/pkg/verifier"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err := verifier.VerifyBytes(data); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}
	fmt.Println("✅ VERIFIED")
}
