// Package hardcodedcredentialatknownsink implements the native HardcodedCredentialAtKnownSink inspection.
package hardcodedcredentialatknownsink

import (
	"path"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Load credentials from external configuration."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HardcodedCredentialAtKnownSink" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew, syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	for _, pattern := range ctx.List("excludedPaths") {
		if glob(strings.Split(pattern, "/"), strings.Split(strings.ReplaceAll(ctx.File.Path, `\`, "/"), "/")) {
			return
		}
	}
	var password syntax.Expr
	switch c := n.(type) {
	case *syntax.New:
		if !semanticquery.NamesGlobalClass(ctx, c.Class, "PDO") && !semanticquery.NamesGlobalClass(ctx, c.Class, "mysqli") {
			return
		}
		name := c.Class.(*syntax.Name)
		constructor := ctx.Index().FindMethod(ctx.Names().Class(name.Value, name.Span().Start), "__construct", ctx.PHP)
		if constructor != nil && constructor.Builtin {
			password = semanticquery.CallArgument(c.Args, 2, "password")
		}
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "mysqli_connect":
			password = semanticquery.CallArgument(c.Args, 2, "password")
		case "mysqli_real_connect":
			password = semanticquery.CallArgument(c.Args, 3, "password")
		default:
			return
		}
	}
	s, known := semanticquery.NativeString(ctx, password)
	if known && s != "" {
		ctx.ReportNode(password, message)
	}
}

func glob(pattern, parts []string) bool {
	if len(pattern) > 128 || len(parts) > 256 {
		return false
	}
	row := make([]bool, len(parts)+1)
	next := make([]bool, len(row))
	row[0] = true
	for _, segment := range pattern {
		clear(next)
		for j := range next {
			if segment == "**" {
				next[j] = row[j] || (j > 0 && next[j-1])
			} else if j > 0 && row[j-1] {
				matched, err := path.Match(segment, parts[j-1])
				next[j] = err == nil && matched
			}
		}
		row, next = next, row
	}
	return row[len(parts)]
}
