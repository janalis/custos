package flowquery

import "custos/internal/php/syntax"

// SteppedBefore reports whether use may observe an increment, a decrement or
// a write through a by-reference closure capture of its variable. Local
// flow facts keep an assigned literal across those constructs, so the value
// at use is no longer the proven one. A later change still counts when a
// loop repeats the use.
func SteppedBefore(f *syntax.File, use *syntax.Variable) bool {
	for _, access := range VarAccesses(f, syntax.EnclosingFuncLike(use), use.Name) {
		switch by := access.By.(type) {
		case *syntax.IncDec:
		case *syntax.ClosureUse:
			if !by.ByRef {
				continue
			}
		default:
			continue
		}
		if access.Var.Span().Start < use.Span().Start || SharesLoop(use, access.Var) {
			return true
		}
	}
	return false
}

// SharesLoop reports whether a loop around use, within its function, also
// contains other.
func SharesLoop(use, other syntax.Node) bool {
	for p := use.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if span := p.Span(); span.Start <= other.Span().Start && other.Span().End <= span.End {
				return true
			}
		}
	}
	return false
}
