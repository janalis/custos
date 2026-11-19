// Command custom-binary is custos with two rules of its own: one written in
// Go, one declared as a pattern.
package main

import (
	"strings"

	"github.com/janalis/custos/pkg/custos"
	"github.com/janalis/custos/pkg/custos/ast"
)

var version = "dev"

// noShortTempName reports tempnam() called with a prefix shorter than three
// characters, which some systems truncate silently.
var noShortTempName = custos.MustRule(custos.Info{
	ID:      "NoShortTempName",
	Summary: "tempnam() prefixes shorter than three characters are truncated on some systems.",
}, []ast.NodeKind{ast.KFuncCall}, func(ctx *custos.Context, n ast.Node) {
	call := n.(*ast.FuncCall)
	if ctx.GlobalFunctionName(call) != "tempnam" || call.Args == nil || len(call.Args.Args) < 2 {
		return
	}
	arg, ok := call.Args.Args[1].(*ast.Arg)
	if !ok {
		return
	}
	lit, ok := ast.UnwrapParens(arg.Value).(*ast.Literal)
	if !ok {
		return
	}
	if prefix, ok := ast.LiteralStringValue(lit); ok && len(strings.TrimSpace(prefix)) < 3 {
		ctx.ReportNode(arg, "Use a prefix of at least three characters")
	}
})

func main() {
	noDump, err := custos.Pattern(custos.PatternDefinition{
		ID:      "NoDump",
		Match:   custos.Patterns{`\dump(...)`, `\dd(...)`},
		Message: "Remove the debug call",
	})
	if err != nil {
		panic(err)
	}
	custos.Main(version, noShortTempName, noDump)
}
