package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	inspections "custos/internal/inspection/catalogue"
	"custos/internal/inspection/meta"
)

const usage = `custos — PHP inspections and fixes

Usage:
  custos analyse [flags] [paths...]   report problems
  custos fix [flags] [paths...]       apply quick-fixes
  custos lsp                          run the language server (stdio)
  custos rules [--json]               list rules
  custos explain <rule>               describe a rule
  custos version                      print version

Run "custos <command> -h" for command flags.
`

// Seams for tests: the rule registry and the catalogue are embedded, so
// their failure paths can only be exercised by swapping them.
var (
	registry  = inspections.All
	catalogue = meta.All
)

// Run executes the command adapter and returns its exit status.
func Run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	code := 0
	switch args[0] {
	case "analyse", "analyze":
		code, err = cmdAnalyse(args[1:], stdout, stderr)
	case "fix":
		code, err = cmdFix(args[1:], stdout, stderr)
	case "rules":
		err = cmdRules(args[1:], stdout, stderr)
	case "explain":
		err = cmdExplain(args[1:], stdout)
	case "lsp":
		err = cmdLSP(version, args[1:], stdin, stdout, stderr)
	case "version", "--version":
		fmt.Fprintln(stdout, "custos", version)
	case "-h", "--help", "help":
		fmt.Fprint(stdout, usage)
	default:
		fmt.Fprintf(stderr, "custos: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(stderr, "custos:", err)
		if code == 0 {
			code = 2
		}
	}
	return code
}
