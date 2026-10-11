// Package unvalidatedredirecttarget implements the native UnvalidatedRedirectTarget inspection.
package unvalidatedredirecttarget

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

const message = "Validate the redirect target against the permitted destination policy."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnvalidatedRedirectTarget" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltinName(ctx, c) != "header" {
		// Project wrappers around header() are proven by their sink summaries.
		semanticquery.CheckUnvalidatedRedirectTarget(ctx, n, message)
		return
	}
	value := semanticquery.CallArgument(c.Args, 0, "header")
	if value == nil || !ctx.Flow().Tainted(value, flow.URL) {
		return
	}
	pieces := appendPieces(nil, value)
	prefix, literals := literalPrefix(ctx, pieces)
	if !strings.HasPrefix(strings.ToLower(prefix), "location:") {
		return
	}
	fixed := fixedLocalQuery(prefix[len("location:"):])
	for _, piece := range pieces[literals:] {
		if !pieceTainted(ctx, piece) || fixed && encodedQueryValue(ctx, piece) {
			continue
		}
		ctx.ReportNode(c, message)
		return
	}
}

// appendPieces flattens a concatenation chain and interpolated strings into
// their operands, in evaluation order.
func appendPieces(pieces []syntax.Expr, x syntax.Expr) []syntax.Expr {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Binary:
		if n.Op.Kind == syntax.TDot {
			return appendPieces(appendPieces(pieces, n.Left), n.Right)
		}
	case *syntax.InterpolatedString:
		// A backtick command yields shell output, not written header text.
		if !n.Backtick {
			for _, part := range n.Parts {
				pieces = appendPieces(pieces, part)
			}
			return pieces
		}
	}
	return append(pieces, syntax.UnwrapParens(x))
}

// literalPrefix joins the leading operands whose text is statically known and
// returns how many operands it consumed.
func literalPrefix(ctx *analysis.Context, pieces []syntax.Expr) (string, int) {
	var prefix strings.Builder
	consumed := 0
	for ; consumed < len(pieces); consumed++ {
		if part, ok := pieces[consumed].(*syntax.StringPart); ok {
			prefix.WriteString(part.Raw)
			continue
		}
		text, known := semanticquery.NativeString(ctx, pieces[consumed])
		if !known {
			break
		}
		prefix.WriteString(text)
	}
	return prefix.String(), consumed
}

// fixedLocalQuery reports whether the written target already pins a local
// path and has entered its query string, so later operands are query data.
func fixedLocalQuery(target string) bool {
	target = strings.TrimLeft(target, " \t")
	return len(target) > 1 && target[0] == '/' && target[1] != '/' && target[1] != '\\' && strings.Contains(target, "?")
}

// encodedQueryValue proves the operand is the result of a builtin URL encoder.
func encodedQueryValue(ctx *analysis.Context, piece syntax.Expr) bool {
	if v, variable := piece.(*syntax.Variable); variable && aliased(ctx, v) {
		return false
	}
	call, ok := semanticquery.NativeValue(ctx, piece).(*syntax.FuncCall)
	if !ok {
		return false
	}
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "urlencode", "rawurlencode", "http_build_query":
		return true
	}
	return false
}

// pieceTainted reports request data in one operand. The flow analysis already
// gives a match expression the taint of its arm results only: its subject
// selects, but is not part of, the produced value.
func pieceTainted(ctx *analysis.Context, piece syntax.Expr) bool {
	return ctx.Flow().Tainted(piece, flow.URL)
}
