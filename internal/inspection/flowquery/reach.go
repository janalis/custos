package flowquery

import (
	"slices"
	"sort"

	"custos/internal/php/syntax"
)

// Reachable reports whether node n (inside scope's body) can be reached
// from the scope entry: no statement on the path from n up to scope is
// preceded, in its statement list, by a statement that Terminates, unless a
// `goto` label written between the two lets a jump land before n. Branches
// guarded by constant conditions are considered reachable. scope may be nil
// (file level).
func Reachable(n, scope syntax.Node) bool {
	return (&jumps{scope: scope, budget: jumpChecks}).reach(n, nil)
}

// jumpChecks bounds the `goto` statements whose own reachability is asked
// for one node; past it a jump counts as taken.
const jumpChecks = 32

// jumps is one question about a scope and the jumps it may still ask about.
type jumps struct {
	scope  syntax.Node
	budget int
}

// reach is Reachable with the labels in via left out: the ones the question
// already goes through, which a jump cannot be reached by again.
func (j *jumps) reach(n syntax.Node, via []string) bool {
	child := n
	for p := n.Parent(); p != nil && child != j.scope; p = p.Parent() {
		if list, ok := syntax.StmtListOf(p); ok {
			// Statements before child (all of them when child is not in the
			// list, e.g. a case condition); FirstTerminating is cached so
			// long lists stay linear overall.
			i := syntax.StmtIndex(list, child)
			if i < 0 {
				i = len(list)
			}
			if syntax.FirstTerminating(p) < i && !j.landing(p, list, i, n, via) {
				return false
			}
		}
		if p == j.scope {
			break
		}
		child = p
	}
	return true
}

// landing reports that a jump can land in or after the first statement that
// ends the list of p and before n, in or below the statement at index at,
// with no statement after the landing ending the list again. A label counts
// when it is written there at any depth, outside the functions and classes
// declared there, when control can go on from it to n (see leads and leaves)
// and when a `goto` of the scope that can itself be reached names it. At
// file level any label counts: the jumps of the other namespace blocks of
// the file cannot be seen from here.
//
// The labels and the places where the list ends are read once per list
// (syntax.TailOf): a list without a label between the two is answered at
// once, and only the statements that hold one are looked at, the nearest
// first, so that a long run of statements nothing reaches costs each of its
// nodes a search and not a walk.
func (j *jumps) landing(p syntax.Node, list []syntax.Stmt, at int, n syntax.Node, via []string) bool {
	tail := syntax.TailOf(p)
	labels := tail.Labels
	if at < len(list) {
		before := n.Span().Start
		labels = labels[:sort.Search(len(labels), func(k int) bool { return labels[k].Span().Start >= before })]
	}
	// The last statement before at that ends the list: nothing written
	// before it goes on to n.
	stop := list[tail.LastTerminating(at-1)].Span().Start
	var names []string
	for k := len(labels) - 1; k >= 0 && labels[k].Span().Start >= stop; {
		start := labels[k].Span().Start
		i := sort.Search(len(list), func(i int) bool { return list[i].Span().Start > start }) - 1
		names = names[:0]
		for ; k >= 0 && labels[k].Span().Start >= list[i].Span().Start; k-- {
			label := labels[k]
			if slices.Contains(via, label.Name.Value) {
				continue
			}
			if i == at && leads(label, n, list[i]) || i < at && leaves(label, list[i]) {
				names = append(names, label.Name.Value)
			}
		}
		if len(names) > 0 && (j.scope == nil || j.taken(names, via)) {
			return true
		}
	}
	return false
}

// taken reports that a `goto` of the scope naming one of the labels can be
// reached without going through that label or one of via.
func (j *jumps) taken(names, via []string) bool {
	for _, jump := range syntax.ScopeGotos(j.scope) {
		if !slices.Contains(names, jump.Label.Value) {
			continue
		}
		if j.budget == 0 {
			return true
		}
		j.budget--
		if j.reach(jump, append(slices.Clip(via), jump.Label.Value)) {
			return true
		}
	}
	return false
}

// leads reports that control can go from a label to n, both below the
// statement top, without a jump: not from one arm of an `if` to another,
// unless a loop around the two brings it back.
func leads(label, n syntax.Node, top syntax.Stmt) bool {
	shared := label.Parent()
	for shared.Span().Start > n.Span().Start || shared.Span().End < n.Span().End {
		shared = shared.Parent()
	}
	if _, arms := shared.(*syntax.If); !arms {
		return true
	}
	for p := shared; ; p = p.Parent() {
		if repeats(p) {
			return true
		}
		if p == syntax.Node(top) {
			return false
		}
	}
}

// leaves reports that control can go from a label to the end of the
// statement top that holds it: no statement ends a list on the way. A loop,
// a `switch` or a `try` on the way is left in more ways than its lists
// tell, by `break`, `continue` or an exception.
func leaves(label syntax.Node, top syntax.Stmt) bool {
	for p := label; p != syntax.Node(top); {
		p = p.Parent()
		switch p.(type) {
		case *syntax.Switch, *syntax.Try:
			return true
		}
		if repeats(p) {
			return true
		}
	}
	for child := label; child != syntax.Node(top); child = child.Parent() {
		if list, ok := syntax.StmtListOf(child.Parent()); ok {
			if i := syntax.StmtIndex(list, child); i >= 0 && slices.ContainsFunc(list[i+1:], syntax.Terminates) {
				return false
			}
		}
	}
	return true
}

// repeats reports a loop statement.
func repeats(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
		return true
	}
	return false
}
