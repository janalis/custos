package architecture

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// badExceptionsProcessing reports oversized try blocks and caught
// exceptions that are never used.
type badExceptionsProcessing struct{}

func init() { register(badExceptionsProcessing{}) }

func (badExceptionsProcessing) ID() string { return "BadExceptionsProcessing" }

func (badExceptionsProcessing) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KTry, syntax.KCatch}
}

func (badExceptionsProcessing) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Try:
		if x.Body == nil || countStmts(x.Body.Stmts) <= 3 { // D1/E1
			return
		}
		s := x.Span()
		ctx.Report(syntax.Span{Start: s.Start, End: s.Start + 3},
			"Too many statements in this try block; extract some of them so the failing call is obvious.")
	case *syntax.Catch:
		if x.Var == nil || x.Var.NameExpr != nil || x.Var.Name == "" || x.Body == nil { // D3/E3
			return
		}
		if util.MentionsVariable(x.Body, x.Var.Name) { // D4/E2
			return
		}
		if usedAfterCatch(ctx, x) { // D4b/E2b
			return
		}
		if countStmts(x.Body.Stmts) == 0 {
			ctx.ReportNode(x.Var, "Caught exception is silently discarded; at least log it.")
		} else {
			ctx.ReportNode(x.Var, "Caught exception is dropped; log it or pass it on as the previous exception.")
		}
	}
}

// countStmts counts the statements of a list, ignoring empty statements and
// zero-width recovery nodes.
func countStmts(stmts []syntax.Stmt) int {
	c := 0
	for _, s := range stmts {
		if _, ok := s.(*syntax.Nop); ok || s.Span().Len() == 0 {
			continue
		}
		c++
	}
	return c
}

// usedAfterCatch reports whether the caught variable occurs after the catch
// clause in the enclosing scope, outside other catch clauses rebinding the
// same name (D4b).
func usedAfterCatch(ctx *analysis.Context, c *syntax.Catch) bool {
	last := lastFreeAccesses(ctx, syntax.EnclosingFuncLike(c))[c.Var.Name]
	return last > 0 && last-1 >= c.Span().End
}

// lastFreeAccesses returns, for scope, name -> 1 + the start offset of the
// last access to the variable that is not inside a catch clause rebinding
// it (0: none). Computed once per scope and file: scanning the accesses per
// catch clause was quadratic on long try/catch sequences.
func lastFreeAccesses(ctx *analysis.Context, scope syntax.Node) map[string]uint32 {
	cache := ctx.Memo("lastFreeAccesses", func() any { return map[syntax.Node]map[string]uint32{} }).(map[syntax.Node]map[string]uint32)
	if m, ok := cache[scope]; ok {
		return m
	}
	m := map[string]uint32{}
	for name, accs := range util.VarAccessesByName(ctx.File, scope) {
		for _, a := range accs {
			if !rebindingCatch(a.Var, scope, name) {
				m[name] = max(m[name], a.Var.Span().Start+1)
			}
		}
	}
	cache[scope] = m
	return m
}

// rebindingCatch reports whether v sits inside a catch clause of scope
// that binds name.
func rebindingCatch(v, scope syntax.Node, name string) bool {
	for p := v.Parent(); p != nil && p != scope; p = p.Parent() {
		if c, ok := p.(*syntax.Catch); ok && c.Var != nil && c.Var.NameExpr == nil && c.Var.Name == name {
			return true
		}
	}
	return false
}
