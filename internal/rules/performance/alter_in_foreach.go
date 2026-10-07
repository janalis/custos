package performance

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// alterInForeach reports by-reference foreach values left alive after the
// loop, pointless unsets of non-reference values and (opt-in) write-backs
// through the key.
type alterInForeach struct{}

func init() { register(alterInForeach{}) }

func (alterInForeach) ID() string { return "AlterInForeach" }

func (alterInForeach) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach, syntax.KAssign}
}

func (r alterInForeach) Check(ctx *analysis.Context, n syntax.Node) {
	switch n := n.(type) {
	case *syntax.Foreach:
		name, ok := perfPlainVar(n.Value)
		if !ok {
			return // destructuring value (spec Divergences)
		}
		if n.ByRef {
			r.checkByRef(ctx, n, name)
		} else {
			r.checkUnset(ctx, n, name)
		}
	case *syntax.Assign:
		if ctx.Bool("SUGGEST_USING_VALUE_BY_REF") {
			r.checkWriteBack(ctx, n)
		}
	}
}

// checkByRef implements part A (D1-D3).
func (alterInForeach) checkByRef(ctx *analysis.Context, f *syntax.Foreach, name string) {
	next := statementFollower(ctx.File, f) // D2
	for p := f.Parent(); next == nil && p != nil && !util.IsFuncLike(p); p = p.Parent() {
		if _, isBlock := p.(*syntax.Block); !isBlock {
			next = statementFollower(ctx.File, p)
		}
	}
	switch nx := next.(type) { // D3
	case nil, *syntax.Return:
		return
	case *syntax.ExprStmt:
		if _, ok := nx.Expr.(*syntax.Throw); ok {
			return
		}
	case *syntax.Unset:
		for _, v := range nx.Vars {
			if vn, ok := perfPlainVar(v); ok && vn == name {
				return
			}
		}
	}
	ctx.ReportNode(f.Value, "Unset '$"+name+"' right after the loop: it is still a reference to the last element.")
}

// checkUnset implements part B (D4-D6).
func (alterInForeach) checkUnset(ctx *analysis.Context, f *syntax.Foreach, name string) {
	cur := f
	next := followerKeepingDocs(ctx.File, cur) // D5
	for {
		if _, isUnset := next.(*syntax.Unset); isUnset {
			break
		}
		blk, ok := cur.Parent().(*syntax.Block)
		if !ok {
			return
		}
		outer, ok := blk.Parent().(*syntax.Foreach)
		if !ok || outer.Body != syntax.Stmt(blk) {
			return
		}
		cur = outer
		next = followerKeepingDocs(ctx.File, cur)
	}
	u := next.(*syntax.Unset) // D6
	for _, v := range u.Vars {
		if vn, ok := perfPlainVar(v); ok && vn == name {
			ctx.ReportSeverity(v.Span(), meta.SeverityInfo, "'$"+name+"' is not a reference here; unsetting it is unnecessary.")
		}
	}
}

// checkWriteBack implements part C (D7-D8).
func (alterInForeach) checkWriteBack(ctx *analysis.Context, a *syntax.Assign) {
	dim, ok := a.Var.(*syntax.ArrayDimFetch)
	if !ok || dim.Dim == nil {
		return
	}
	key, ok := perfPlainVar(dim.Dim)
	if !ok {
		return
	}
	for p := a.Parent(); p != nil && !util.IsFuncLike(p); p = p.Parent() {
		f, ok := p.(*syntax.Foreach)
		if !ok || f.Key == nil || f.Expr == nil {
			continue
		}
		value, ok := perfPlainVar(f.Value)
		if !ok {
			continue
		}
		if k, ok := perfPlainVar(f.Key); ok && k == key && util.EquivalentFoldNames(ctx.File, f.Expr, dim.Var) {
			if f.ByRef {
				return // already by reference (D8)
			}
			ctx.ReportSeverity(dim.Span(), meta.SeverityInfo, "Iterate '$"+value+"' by reference and assign to it directly instead of writing through the key.")
			return
		}
	}
}

// nextSibling returns the child of n's parent (or the file's top-level
// statement) following n, in source order.
func nextSibling(f *syntax.File, n syntax.Node) syntax.Node {
	parent := n.Parent()
	if parent == nil {
		for i, s := range f.Stmts {
			if syntax.Node(s) == n && i+1 < len(f.Stmts) {
				return f.Stmts[i+1]
			}
		}
		return nil
	}
	var next syntax.Node
	found := false
	syntax.Children(parent, func(c syntax.Node) {
		if next != nil {
			return
		}
		if found {
			next = c
		} else if c == n {
			found = true
		}
	})
	return next
}

// statementFollower is nextSibling for part A, where an elseif/else clause
// following a branch body is not a follower: control continues after the
// whole if statement, so the climb goes on from there (D2).
func statementFollower(f *syntax.File, n syntax.Node) syntax.Node {
	switch next := nextSibling(f, n).(type) {
	case *syntax.ElseIf, *syntax.Else:
		return nil
	default:
		return next
	}
}

// followerKeepingDocs is nextSibling for part B, where a doc comment between
// the loop and the next node is itself the follower; it is then represented
// by n (any non-unset node will do).
func followerKeepingDocs(f *syntax.File, n syntax.Node) syntax.Node {
	next := nextSibling(f, n)
	if next == nil {
		return nil
	}
	gap := syntax.Span{Start: n.Span().End, End: next.Span().Start}
	if _, ok := util.FindToken(f, gap, syntax.TDocComment); ok {
		return n
	}
	return next
}
