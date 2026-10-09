package uselessreturn

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// uselessReturn reports a trailing bare `return;` and `return $local = x;`.
type uselessReturn struct{}

func (uselessReturn) ID() string               { return "UselessReturn" }
func (uselessReturn) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KReturn} }
func (r uselessReturn) Check(ctx *analysis.Context, n syntax.Node) {
	ret := n.(*syntax.Return)
	if ret.Expr == nil {
		r.checkTrailing(ctx, ret)
		return
	}
	r.checkAssign(ctx, ret)
}

// checkTrailing implements D1/D2.
func (uselessReturn) checkTrailing(ctx *analysis.Context, ret *syntax.Return) {
	b, ok := ret.Parent().(*syntax.Block)
	if !ok || b.Alt {
		return
	}
	switch b.Parent().(type) {
	case *syntax.Function, *syntax.Method, *syntax.Closure:
	default:
		return
	}
	if syntax.FuncLikeBody(b.Parent()) != b || len(b.Stmts) == 0 || b.Stmts[len(b.Stmts)-1] != syntax.Stmt(ret) {
		return
	}
	ctx.ReportSeverity(ret.Span(), diagnostic.SeverityInfo, "Redundant 'return;' at the end of the body; remove it.")
}

// checkAssign implements D3–D5.
func (uselessReturn) checkAssign(ctx *analysis.Context, ret *syntax.Return) {
	a, ok := ret.Expr.(*syntax.Assign)
	if !ok || a.Op.Kind != syntax.TEqual || a.Value == nil { // D3 (plain `=` only, see spec Divergences)
		return
	}
	v, ok := a.Var.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name == "" {
		return
	}
	scope := syntax.EnclosingFuncLike(ret) // D4
	body := syntax.FuncLikeBody(scope)
	if body == nil {
		return
	}
	name := v.Name
	for _, p := range syntax.FuncLikeParams(scope) { // D5a
		if p.ByRef && p.Var != nil && p.Var.NameExpr == nil && p.Var.Name == name {
			return
		}
	}
	if c, ok := scope.(*syntax.Closure); ok && byRefUse(c, name) { // D5b
		return
	}
	observable := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if observable {
			return false
		}
		switch s := x.(type) {
		case *syntax.Closure: // D5b
			observable = byRefUse(s, name)
		case *syntax.StaticStmt: // D5c
			for _, sv := range s.Vars {
				if sv.Var != nil && sv.Var.NameExpr == nil && sv.Var.Name == name {
					observable = true
				}
			}
		case *syntax.Global: // global $v (spec Divergences)
			for _, g := range s.Vars {
				if gv, ok := g.(*syntax.Variable); ok && gv.NameExpr == nil && gv.Name == name {
					observable = true
				}
			}
		}
		return !observable
	})
	if observable {
		return
	}
	for p := ret.Parent(); p != nil && p != scope; p = p.Parent() { // D5d
		if t, ok := p.(*syntax.Try); ok {
			if t.Finally != nil && t.Finally.Body != nil && astquery.MentionsVariable(t.Finally.Body, name) {
				return
			}
			break
		}
	}
	span := ret.Span()
	value := a.Value.Span()
	src := ctx.Src
	ctx.Report(span, "The assigned variable is never used after returning; return the value directly.", diagnostic.Fix{
		Title: "Return the value directly",
		Edits: func() []diagnostic.TextEdit {
			text := "return " + string(src[value.Start:value.End])
			if src[span.End-1] == ';' {
				text += ";"
			} else {
				return []diagnostic.TextEdit{{Span: syntax.Span{Start: span.Start, End: value.End}, NewText: text}}
			}
			return []diagnostic.TextEdit{{Span: span, NewText: text}}
		},
	})
}

func byRefUse(c *syntax.Closure, name string) bool {
	for _, u := range c.Uses {
		if u.ByRef && u.Var != nil && u.Var.Name == name {
			return true
		}
	}
	return false
}
