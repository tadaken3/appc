package main

import (
	"os"

	"github.com/tadaken3/appc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
