package main

import (
	"fmt"
	"os"

	"github.com/tomocy/smoothie/cmd/smoothie/runner"
)

func main() {
	r := runner.New()
	if err := r.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run: %s\n", err)
		os.Exit(1)
	}
}
