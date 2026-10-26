package usinginclusionreturnvalue

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// usingInclusionReturnValue reports include/require expressions whose value
// is used, i.e. that are not a statement of their own.
type usingInclusionReturnValue struct{}

const usingInclusionReturnValueMsg = "Avoid relying on the value returned by an included file."

func (usingInclusionReturnValue) ID() string { return "UsingInclusionReturnValue" }
func (usingInclusionReturnValue) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KInclude}
}

func (usingInclusionReturnValue) Check(ctx *analysis.Context, n syntax.Node) {
	cur := n
	for guarded := false; ; { // D2: look through `@`, parentheses and (E2) a guarding operator
		parent := cur.Parent()
		if u, ok := parent.(*syntax.Unary); ok && u.Op.Kind == syntax.TAt {
			cur = u
			continue
		}
		if _, ok := parent.(*syntax.Paren); ok {
			cur = parent
			continue
		}
		if r, ok := parent.(*syntax.Return); ok && !guarded && syntax.EnclosingFuncLike(r) == nil { // E3
			return
		}
		if b, ok := parent.(*syntax.Binary); ok && b.Right == cur && guardOp(b.Op.Kind) {
			cur, guarded = b, true
			continue
		}
		break
	}
	if _, ok := cur.Parent().(*syntax.ExprStmt); ok { // E1, E2
		return
	}
	ctx.ReportNode(n, usingInclusionReturnValueMsg)
}

// guardOp reports the operators whose right operand only runs depending on
// the left one: `class_exists($c, false) || require $file;`.
func guardOp(k syntax.TokenKind) bool {
	switch k {
	case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
		return true
	}
	return false
}
