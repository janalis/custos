// Package unserializefalseambiguity implements the native UnserializeFalseAmbiguity inspection.
package unserializefalseambiguity

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Distinguish valid serialized false from decoding failure."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnserializeFalseAmbiguity" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	branch := n.(*syntax.If)
	b, ok := syntax.UnwrapParens(branch.Cond).(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TIsIdentical {
		return
	}
	other := b.Left
	if truth, known := astquery.BoolConst(b.Right); !known || truth {
		if truth, known := astquery.BoolConst(b.Left); !known || truth {
			return
		}
		other = b.Right
	}
	call, ok := semanticquery.NativeValue(ctx, other).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "unserialize") {
		return
	}

	if rejecting(branch.Body) && !payloadCompared(ctx, branch, call) {
		ctx.ReportNode(call, message)
	}
}

// payloadCompared reports whether the decoded variable is compared with the
// serialized form of false in the same variable scope, no later than the end
// of the rejecting branch: such code already tells the two outcomes apart.
func payloadCompared(ctx *analysis.Context, branch *syntax.If, call *syntax.FuncCall) bool {
	name, ok := astquery.PlainVariableName(syntax.UnwrapParens(semanticquery.CallArgument(call.Args, 0, "data")))
	if !ok {
		return false
	}
	scope, limit, found := syntax.EnclosingVariableScope(branch), branch.Body.Span().End, false
	visit := func(n syntax.Node) bool {
		if found || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		b, ok := n.(*syntax.Binary)
		if !ok || b.Span().Start >= limit {
			return true
		}
		switch b.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
			left, right := syntax.UnwrapParens(b.Left), syntax.UnwrapParens(b.Right)
			found = payloadMarker(ctx, name, left, right) || payloadMarker(ctx, name, right, left)
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}

// payloadMarker reports whether payload is the named variable and marker is
// the serialized false text, written literally or as serialize(false).
func payloadMarker(ctx *analysis.Context, name string, payload, marker syntax.Expr) bool {
	if written, ok := astquery.PlainVariableName(payload); !ok || written != name {
		return false
	}
	if text, ok := astquery.QuotedStringValue(marker); ok {
		return text == "b:0;"
	}
	c, ok := marker.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "serialize") {
		return false
	}
	truth, known := astquery.BoolConst(semanticquery.CallArgument(c.Args, 0, "value"))
	return known && !truth
}

func rejecting(body syntax.Stmt) bool {
	statements, ok := astquery.BlockStmts(body)
	if !ok || len(statements) == 0 {
		return false
	}
	last := statements[len(statements)-1]
	if ret, ok := last.(*syntax.Return); ok {
		v, known := astquery.BoolConst(ret.Expr)
		return ret.Expr == nil || (known && !v)
	}
	s, ok := last.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	_, ok = s.Expr.(*syntax.Throw)
	return ok
}
