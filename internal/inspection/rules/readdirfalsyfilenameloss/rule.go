// Package readdirfalsyfilenameloss implements the native ReaddirFalsyFilenameLoss inspection.
package readdirfalsyfilenameloss

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare the directory entry with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReaddirFalsyFilenameLoss" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "readdir") || firstClassCallable(c) {
		return
	}
	use, entry, truthy := conditionUse(c)
	if use == nil {
		return
	}
	// Everything outside the entry expression is dropped by the rewrite, so a
	// comment there withholds the fix; comments inside it are copied verbatim.
	src := ctx.File.Src
	dropped := string(src[use.Span().Start:entry.Span().Start]) + " " + string(src[entry.Span().End:use.Span().End])
	if strings.Contains(dropped, "/*") || strings.Contains(dropped, "//") || strings.Contains(dropped, "#") {
		ctx.ReportNode(use, message)
		return
	}
	op := " === false"
	if truthy {
		op = " !== false"
	}
	fix := diagnostic.Fix{Title: "Compare the entry with false", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: use.Span(), NewText: "(" + ctx.Text(entry) + ")" + op}}
	}}
	ctx.ReportNode(use, message, fix)
}

func firstClassCallable(c *syntax.FuncCall) bool {
	for _, a := range c.Args.Args {
		if _, ok := a.(*syntax.VariadicPlaceholder); ok {
			return true
		}
	}
	return false
}

// conditionUse climbs from a readdir call to the if/loop condition testing its
// result for truthiness. It returns the whole condition, the widest expression
// still holding the raw entry (the call, parenthesized or plainly assigned) and
// whether the condition holds for a truthy entry. A nil condition means the
// result is not tested that way, or is already compared strictly.
func conditionUse(call *syntax.FuncCall) (use, entry syntax.Expr, truthy bool) {
	var e syntax.Expr = call
	entry, truthy = call, true
	raw := true // e still evaluates to the entry rather than to a boolean test of it
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			// A compound or by-reference assignment yields another value.
			if !raw || p.Op.Kind != syntax.TEqual || p.ByRef || p.Value != e {
				return nil, nil, false
			}
			e = p
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return nil, nil, false
			}
			e, raw, truthy = p, false, !truthy
		case *syntax.Binary:
			strict := p.Op.Kind == syntax.TIsIdentical || p.Op.Kind == syntax.TIsNotIdentical
			if !strict && p.Op.Kind != syntax.TIsEqual && p.Op.Kind != syntax.TIsNotEqual {
				return nil, nil, false
			}
			if raw && strict {
				return nil, nil, false
			}
			other := p.Right
			if other == e {
				other = p.Left
			}
			v, ok := astquery.BoolConst(other)
			if !ok {
				return nil, nil, false
			}
			if (p.Op.Kind == syntax.TIsEqual || p.Op.Kind == syntax.TIsIdentical) != v {
				truthy = !truthy
			}
			e, raw = p, false
		case *syntax.If:
			return e, entry, truthy
		case *syntax.ElseIf:
			return e, entry, truthy
		case *syntax.While:
			return e, entry, truthy
		case *syntax.DoWhile:
			return e, entry, truthy
		case *syntax.For:
			// Only the last condition expression decides whether the loop continues.
			if len(p.Cond) == 0 || p.Cond[len(p.Cond)-1] != e {
				return nil, nil, false
			}
			return e, entry, truthy
		default:
			return nil, nil, false
		}
		if raw {
			entry = e
		}
	}
}
