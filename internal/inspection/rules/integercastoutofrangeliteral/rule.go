// Package integercastoutofrangeliteral implements the native IntegerCastOutOfRangeLiteral inspection.
package integercastoutofrangeliteral

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve integers that exceed the supported integer range."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntegerCastOutOfRangeLiteral" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	u := n.(*syntax.Unary)
	if u.Op.Kind != syntax.TIntCast {
		return
	}
	s, ok := semanticquery.NativeString(ctx, u.Expr)
	if !ok || s == "" {
		return
	}
	limit := "9223372036854775807"
	switch s[0] {
	case '-':
		limit = "9223372036854775808"
		s = s[1:]
	case '+':
		s = s[1:]
	}
	if s == "" {
		return
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return
		}
	}
	s = strings.TrimLeft(s, "0")
	if len(s) > len(limit) || (len(s) == len(limit) && s > limit) {
		ctx.ReportNode(u, message)
	}
}
