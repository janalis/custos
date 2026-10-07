package controlflow

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// ifReturnReturnSimplification reports `if (c) { return true; } return false;`
// (and the else / reversed forms), which can return the condition directly.
type ifReturnReturnSimplification struct{}

func init() { register(ifReturnReturnSimplification{}) }

func (ifReturnReturnSimplification) ID() string { return "IfReturnReturnSimplification" }

func (ifReturnReturnSimplification) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KIf}
}

var invertedComparison = map[syntax.TokenKind]string{
	syntax.TIsIdentical:      "!==",
	syntax.TIsNotIdentical:   "===",
	syntax.TIsEqual:          "!=",
	syntax.TIsNotEqual:       "==",
	syntax.TGreater:          "<=",
	syntax.TIsGreaterOrEqual: "<",
	syntax.TLess:             ">=",
	syntax.TIsSmallerOrEqual: ">",
}

// boolValuedOps are the binary operators whose result is always a bool, so
// returning the condition keeps the function's return type (D1).
var boolValuedOps = map[syntax.TokenKind]bool{
	syntax.TIsIdentical: true, syntax.TIsNotIdentical: true,
	syntax.TIsEqual: true, syntax.TIsNotEqual: true,
	syntax.TGreater: true, syntax.TIsGreaterOrEqual: true,
	syntax.TLess: true, syntax.TIsSmallerOrEqual: true,
	syntax.TBooleanAnd: true, syntax.TBooleanOr: true,
	syntax.TAnd: true, syntax.TOr: true, syntax.TXor: true,
}

func (ifReturnReturnSimplification) Check(ctx *analysis.Context, n syntax.Node) {
	s := n.(*syntax.If)
	if s.Cond == nil || len(s.ElseIfs) > 0 { // D2
		return
	}
	c := syntax.UnwrapParens(s.Cond) // D1
	switch c := c.(type) {
	case *syntax.Instanceof:
	case *syntax.Binary:
		if !boolValuedOps[c.Op.Kind] {
			return
		}
	default:
		return
	}
	r1 := soleReturn(s.Body) // D3
	if r1 == nil {
		return
	}
	var r2 *syntax.Return
	if s.Else != nil { // D4a
		r2 = soleReturn(s.Else.Body)
	} else { // D4b
		next, ok := util.NextStmt(ctx.File, s)
		if !ok {
			return
		}
		r2, _ = next.(*syntax.Return)
	}
	if r2 == nil {
		return
	}
	v1, ok1 := returnsBool(r1) // D5
	v2, ok2 := returnsBool(r2)
	if !ok1 || !ok2 || v1 == v2 {
		return
	}
	if s.Else == nil && precededByGuardIf(ctx.File, s) { // D6
		return
	}
	expr := ctx.Text(c)
	if !v1 { // reverse
		if b, ok := c.(*syntax.Binary); ok && invertedComparison[b.Op.Kind] != "" {
			expr = ctx.Text(b.Left) + " " + invertedComparison[b.Op.Kind] + " " + ctx.Text(b.Right)
		} else {
			expr = "!(" + expr + ")"
		}
	}
	repl := "return " + expr
	span := s.Span()
	if s.Else == nil {
		span.End = r2.Span().End
	}
	kw := syntax.Span{Start: s.Span().Start, End: s.Span().Start + 2}
	ctx.Report(kw, "Return the condition directly: '"+repl+"'.", analysis.Fix{
		Title: "Return the condition",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl + ";"}}
		},
	})
}

// soleReturn returns the only statement of a block body when it is a return.
func soleReturn(body syntax.Stmt) *syntax.Return {
	stmts, ok := util.BlockStmts(body)
	if !ok || len(stmts) != 1 {
		return nil
	}
	r, _ := stmts[0].(*syntax.Return)
	return r
}

func returnsBool(r *syntax.Return) (val, ok bool) {
	if r.Expr == nil {
		return false, false
	}
	return util.BoolConst(r.Expr)
}

// precededByGuardIf implements D6.
func precededByGuardIf(f *syntax.File, s *syntax.If) bool {
	prev, ok := util.PrevStmt(f, s)
	if !ok {
		return false
	}
	p, ok := prev.(*syntax.If)
	if !ok || p.Else != nil || len(p.ElseIfs) > 0 {
		return false
	}
	stmts, ok := util.BlockStmts(p.Body)
	if !ok || len(stmts) == 0 {
		return false
	}
	_, ok = stmts[len(stmts)-1].(*syntax.Return)
	return ok
}
