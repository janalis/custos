// Command custos inspects and fixes PHP code.
package main

import (
	"os"

	"custos/internal/cli"
)

var version = "dev"

func main() { os.Exit(cli.Run(version, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
