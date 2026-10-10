// Package base64validationwithoutstrictmode implements the native Base64ValidationWithoutStrictMode inspection.
package base64validationwithoutstrictmode

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Enable strict base64 decoding when validating input."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "Base64ValidationWithoutStrictMode" }
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
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "base64_decode") {
		return
	}
	strict := semanticquery.CallArgument(call.Args, 1, "strict")
	if strict != nil {
		value, known := semanticquery.NativeTruth(ctx, strict)
		if !known || value {
			return
		}
	}
	if rejecting(branch.Body) {
		ctx.ReportNode(call, message)
	}
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
