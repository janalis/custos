package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime"

	"github.com/janalis/custos/internal/inspection/analysis"
	inspections "github.com/janalis/custos/internal/inspection/catalogue"
	"github.com/janalis/custos/internal/inspection/meta"
)

const usage = `custos — PHP inspections and fixes

Usage:
  custos analyse [flags] [paths...]   report problems
  custos fix [flags] [paths...]       apply quick-fixes
  custos mcp [flags]                  run the MCP server (stdio or HTTP)
  custos lsp                          run the language server (stdio)
  custos rules [--json]               list rules
  custos explain <rule>               describe a rule
  custos test-rules [dir]             run rules over their fixtures
  custos self-update [--check] [--version X.Y.Z]
                                      update custos to the latest release
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
	return run(version, args, stdin, stdout, stderr, nil, false)
}

// releaseVersion is the custos release a binary is, for the update notice:
// none for a program embedding custos.
func releaseVersion(version string, embedded bool) string {
	if embedded {
		return ""
	}
	return version
}

// extra holds the rules a program embedding custos adds to the catalogue
// (see RunWith).
var extra []analysis.Rule

// rulesWithExtra returns the registered rules followed by the extra ones.
func rulesWithExtra() []analysis.Rule {
	base := registry()
	return append(base[:len(base):len(base)], extra...)
}

// RunWith is Run with rules added to the catalogue, for programs that build
// their own custos binary with rules written in Go. It is not safe to call
// concurrently with different rules.
//
// Such a program is not a custos release: it neither checks for updates
// nor replaces itself with one (self-update).
func RunWith(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, rules []analysis.Rule) int {
	return run(version, args, stdin, stdout, stderr, rules, true)
}

func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, rules []analysis.Rule, embedded bool) int {
	extra = rules
	defer func() { extra = nil }()
	// Allocation sampling is off unless --memprofile turns it on: it would
	// only cost the deep recursive stacks of the analysis a few percent.
	runtime.MemProfileRate = 0
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	code := 0
	switch args[0] {
	case "analyse", "analyze":
		tuneGC()
		notice := startUpdateNotice(releaseVersion(version, embedded), stderr)
		code, err = cmdAnalyse(args[1:], stdout, stderr)
		notice.finish(stderr)
	case "fix":
		tuneGC()
		notice := startUpdateNotice(releaseVersion(version, embedded), stderr)
		code, err = cmdFix(args[1:], stdout, stderr)
		notice.finish(stderr)
	case "self-update":
		if embedded {
			code, err = 2, errors.New("self-update installs the stock custos release; rebuild this program to update it")
			break
		}
		code, err = cmdSelfUpdate(version, args[1:], stdout, stderr)
	case "rules":
		err = cmdRules(args[1:], stdout, stderr)
	case "test-rules":
		code, err = cmdTestRules(args[1:], stdout, stderr)
	case "explain":
		err = cmdExplain(args[1:], stdout)
	case "mcp":
		err = cmdMCP(version, args[1:], stdin, stdout, stderr)
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
