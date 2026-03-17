package main

import (
	"os"

	"github.com/robertgumeny/doug-plan/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
