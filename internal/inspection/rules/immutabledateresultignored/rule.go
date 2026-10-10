// Package immutabledateresultignored implements the ImmutableDateResultIgnored inspection.
package immutabledateresultignored

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ImmutableDateResultIgnored" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }

const message = "Assign the returned immutable date."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	call := n.(*syntax.MethodCall)
	if semanticquery.NativeDateClass(ctx, call.Var) != "DateTimeImmutable" {
		return
	}
	name, ok := call.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	switch strings.ToLower(name.Value) {
	case "modify", "add", "sub", "setdate", "settime", "settimestamp", "settimezone":
	default:
		return
	}
	if _, ok := call.Parent().(*syntax.ExprStmt); !ok {
		return
	}
	ctx.ReportNode(call, message)
}

func (rule) Semantic() {}
