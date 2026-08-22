// Command build compiles the current Handlebars templates and generates a
// static site. Prefer `hbc build`; this command is retained for compatibility.
package main

import (
	"fmt"
	"os"

	"github.com/andriyg76/go-hbars/internal/buildcmd"
)

func main() {
	if err := buildcmd.RunCLI(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		os.Exit(1)
	}
}
