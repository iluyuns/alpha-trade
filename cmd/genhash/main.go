package main

import (
	"fmt"
	"os"

	"github.com/iluyuns/alpha-trade/internal/pkg/crypto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: genhash <password>")
		os.Exit(1)
	}
	hash, err := crypto.HashPassword(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(hash)
}
