package main

import (
	"os"

	"github.com/kenta-tanaka/appc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
