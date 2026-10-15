package flowquery

import (
	"container/heap"
	"math"
	"sort"

	"custos/internal/php/syntax"
)

// Reachable reports whether node n (inside scope's body) can be reached
// from the scope entry: no statement on the path from n up to scope is
// preceded, in its statement list, by a statement that Terminates, unless a
// `goto` label written between the two lets a jump land before n. Branches
// guarded by constant conditions are considered reachable. scope may be nil
// (file level); a property hook on the way ends the path as the scope does.
func Reachable(n, scope syntax.Node) bool {
	owner, _ := (&jumps{scope: scope}).blocked(n, n, n.Parent())
	return owner == nil
}

// jumps is one question about a scope.
type jumps struct {
	scope syntax.Node
	// live is what the jumps reach in the variable scope of the list last
	// asked about, read when a label is first asked about.
	live *landings
}

// blocked climbs from child, n or a node around it whose parent is p, to
// the scope, and returns the first statement list on the way, with the index
// in it, that control does not reach n in: a statement before that index
// ends the list and no jump lands after it. The list is nil when every one
// lets control through.
func (j *jumps) blocked(n, child, p syntax.Node) (owner syntax.Node, at int) {
	for ; p != nil && child != j.scope; p = p.Parent() {
		if list, ok := syntax.StmtListOf(p); ok {
			// Statement-list metadata comes before its body: a case's
			// condition or a namespace's name is not one of its statements.
			i := syntax.StmtIndex(list, child)
			if i < 0 {
				i = 0
			}

			if syntax.FirstTerminating(p) < i && !j.landing(p, list, i, n) {
				return p, i
			}
		}
		if _, hook := p.(*syntax.PropertyHook); hook || p == j.scope {
			// A property hook runs when the property is read or written,
			// wherever its class is declared: nothing written around the
			// class stops it, whatever scope was asked about.
			break
		}
		child = p
	}
	return nil, 0
}

// landing reports that a jump can land in or after the first statement that
// ends the list of p and before n, in or below the statement at index at,
// with no statement after the landing ending the list again. A label counts
// when it is written there at any depth, outside the functions and classes
// declared there, when control can go on from it to n (see into and leaves)
// and when a `goto` that can itself be reached names it: one of the
// function, closure or property hook the list is written in, whatever scope
// the question is about. At file level any label counts: the jumps of the
// other namespace blocks of the file cannot be seen from here.
//
// The labels and the places where the list ends are read once per list
// (syntax.TailOf) and what the jumps reach once per function (land):
// a list without a label is answered at once, and a list of a function by
// two lookups, so that asking about every node of a long run of statements
// costs each of them the same, whatever the labels and jumps around it.
func (j *jumps) landing(p syntax.Node, list []syntax.Stmt, at int, n syntax.Node) bool {
	tail := syntax.TailOf(p)
	if len(tail.Labels) == 0 {
		return false
	}
	// The last statement before at that ends the list: nothing written
	// before it goes on to n.
	last := tail.LastTerminating(at - 1)
	known := j.known(p)
	if known == nil {
		if j.passed(p, tail, list[last].Span().Start, list, at) {
			return true
		}
	} else if known.least[last-known.from] < int32(at) {
		return true
	}
	return at < len(list) && j.into(tail, known, list[at], n)
}

// known returns what the jumps reach in the list of p, which has a label
// after a statement that ends it; nil at file level.
func (j *jumps) known(p syntax.Node) *landed {
	if j.live != nil {
		if known := j.live.lists[p]; known != nil {
			return known
		}
	}
	if live := syntax.ScopeMemo(syntax.EnclosingVariableScope(p), land); live != nil {
		j.live = live
		return live.lists[p]
	}
	return nil
}

// passed is the answer at file level: a label is written from the position
// stop on, before the statement at index at, and control can go on from it
// to the end of the statement that holds it. The nearest label is looked at
// first.
func (j *jumps) passed(p syntax.Node, tail *syntax.Tail, stop uint32, list []syntax.Stmt, at int) bool {
	labels := tail.Labels
	if at < len(list) {
		before := list[at].Span().Start
		labels = labels[:sort.Search(len(labels), func(k int) bool { return labels[k].Span().Start >= before })]
	}
	for k := len(labels) - 1; k >= 0 && labels[k].Span().Start >= stop; k-- {
		if leaves(labels[k], p) {
			return true
		}
	}
	return false
}

// labelled reports that a label that counts starts in the positions from
// from up to to, among the labels of a list: a reached one in a function,
// any at file level.
func (j *jumps) labelled(tail *syntax.Tail, known *landed, from, to uint32) bool {
	if known != nil {
		return j.live.between(from, to)
	}
	labels := tail.Labels
	k := sort.Search(len(labels), func(k int) bool { return labels[k].Span().Start >= from })
	return k < len(labels) && labels[k].Span().Start < to
}

// into reports that a label that counts is written below the statement top
// before n, where control can go from it to n without a jump: not in one arm
// of an `if` when n is in another, nor in one `catch` when n is in a later
// one of the same `try`, unless a loop around the two brings it back. The
// labels are asked for by stretches of positions, one more than the `if` and
// `try` statements around n at most.
func (j *jumps) into(tail *syntax.Tail, known *landed, top syntax.Stmt, n syntax.Node) bool {
	from, to := top.Span().Start, n.Span().Start
	if !j.labelled(tail, known, from, to) {
		return false
	}
	// Below the outermost loop around n every label leads to it.
	var loop syntax.Node
	for p := n; p != syntax.Node(top); {
		p = p.Parent()
		if repeats(p) {
			loop = p
		}
	}
	child := n
	for p := n.Parent(); ; p = p.Parent() {
		if loop != nil {
			if p == loop {
				loop = nil
			}
		} else if apart, arms := others(p, child); arms {
			// The labels of the other arms, written before child, lead
			// nowhere; the ones from child on do.
			if j.labelled(tail, known, child.Span().Start, to) {
				return true
			}
			to = apart
		}
		if p == syntax.Node(top) {
			return j.labelled(tail, known, from, to)
		}
		child = p
	}
}

// others reports that p holds, before child, code that control does not go
// on from to child, and returns where that code starts: the arms of an `if`
// before the one that child is, the `catch` blocks of a `try` before the
// `catch` that child is (the body of the `try` does lead to each of them).
func others(p, child syntax.Node) (from uint32, ok bool) {
	switch x := p.(type) {
	case *syntax.If:
		return x.Span().Start, true
	case *syntax.Try:
		if _, caught := child.(*syntax.Catch); caught {
			return x.Catches[0].Span().Start, true
		}
	}
	return 0, false
}

// leaves reports that control can go from a label to the end of the
// statement that holds it in the list of p.
func leaves(label *syntax.Label, p syntax.Node) (out bool) {
	exits(label, p, func(_ syntax.Node, _ int, leaves bool) { out = leaves })
	return out
}

// exits follows a label up to the list of stop and tells visit, for each
// list on the way that a statement ends at or before the label, its owner,
// the index of the statement that holds the label and whether control can
// go from the label to the end of that statement: no statement ends a list
// on the way. A loop, a `switch` or a `try` on the way is left in more ways
// than its lists tell, by `break`, `continue` or an exception; an endless
// loop is left by none of them, whatever it holds.
func exits(label *syntax.Label, stop syntax.Node, visit func(p syntax.Node, i int, leaves bool)) {
	breaks, ended := false, false
	child := syntax.Node(label)
	for p := label.Parent(); ; p = p.Parent() {
		if list, ok := syntax.StmtListOf(p); ok {
			if tail := syntax.TailOf(p); tail != nil {
				i := syntax.StmtIndex(list, child)
				if i >= syntax.FirstTerminating(p) {
					visit(p, i, breaks || !ended)
				}
				ended = ended || tail.LastTerminating(len(list)-1) > i
			}
		}
		if p == stop {
			return
		}
		switch p.(type) {
		case *syntax.Switch, *syntax.Try:
			breaks = true
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			if breaks = !syntax.Endless(p); !breaks {
				ended = true
			}
		}
		child = p
	}
}

// repeats reports a loop statement.
func repeats(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
		return true
	}
	return false
}

// landings is what the jumps of a scope reach: the labels that a `goto`
// which can itself be reached names, and what follows from them for each
// statement list of the scope. It is worked out once per scope (land, kept
// by syntax.ScopeMemo) and only read afterwards.
type landings struct {
	// labels holds the labels of the scope in source order, marks counts
	// the reached ones before each (a Fenwick tree, so that one more is
	// marked without counting again).
	labels []*syntax.Label
	marks  []int32
	// lists holds, by owner, the lists of the scope with a label written
	// from their first terminating statement on.
	lists map[syntax.Node]*landed
}

// landed describes the statements of a list from index from, its first
// terminating one, on. For each of them that ends the list, least is the
// lowest index, from that statement to the next that ends it, of a statement
// holding a reached label that control leaves it after; math.MaxInt32 when
// there is none.
type landed struct {
	from  int
	least []int32
}

// between reports that a reached label starts in the positions from from up
// to to.
func (l *landings) between(from, to uint32) bool {
	first := sort.Search(len(l.labels), func(k int) bool { return l.labels[k].Span().Start >= from })
	end := sort.Search(len(l.labels), func(k int) bool { return l.labels[k].Span().Start >= to })
	return l.marked(end) > l.marked(first)
}

// marked counts the reached labels before index k.
func (l *landings) marked(k int) (count int32) {
	for ; k > 0; k -= k & -k {
		count += l.marks[k]
	}
	return count
}

// mark records the label at index k as reached.
func (l *landings) mark(k int) {
	for k++; k < len(l.marks); k += k & -k {
		l.marks[k]++
	}
}

// building is the search for the landings of a scope: a least fixed point,
// each `goto` followed up from the list that stopped it when a label written
// there becomes reached, so that the work is in proportion to the jumps, the
// labels and the lists around each.
type building struct {
	*landings
	jumps jumps
	// at holds the indexes in labels of the labels of each name, named the
	// names reached.
	at    map[string][]int
	named map[string]bool
	// parked holds the jumps a list stops, by the place that stops them.
	parked map[gap]*parkings
	// fired holds the names of labels reached and not yet followed.
	fired []string
}

// gap is the place of a list where jumps wait: its owner and the index of a
// statement that ends it, for the statements from there to the next one that
// does.
type gap struct {
	owner syntax.Node
	last  int
}

// parking is a jump stopped in a list at the statement of index at.
type parking struct {
	jump  *syntax.Goto
	at    int
	owner syntax.Node
}

// land works out the landings of a scope with a block body.
func land(scope syntax.Node) *landings {
	labels := syntax.ScopeLabels(scope)
	b := &building{
		landings: &landings{labels: labels, marks: make([]int32, len(labels)+1), lists: map[syntax.Node]*landed{}},
		at:       map[string][]int{},
		named:    map[string]bool{},
		parked:   map[gap]*parkings{},
	}
	b.jumps = jumps{scope: scope, live: b.landings}
	body := syntax.VariableScopeBody(scope)
	found := func(p syntax.Node, _ int, _ bool) {
		if b.lists[p] == nil {
			list, _ := syntax.StmtListOf(p)
			known := &landed{from: syntax.FirstTerminating(p)}
			known.least = make([]int32, len(list)-known.from)
			for i := range known.least {
				known.least[i] = math.MaxInt32
			}
			b.lists[p] = known
		}
	}
	for k, label := range labels {
		b.at[label.Name.Value] = append(b.at[label.Name.Value], k)
		exits(label, body, found)
	}
	for _, jump := range syntax.ScopeGotos(scope) {
		b.follow(jump, jump, jump.Parent())
		for len(b.fired) > 0 {
			name := b.fired[len(b.fired)-1]
			b.fired = b.fired[:len(b.fired)-1]
			b.reach(name, body)
		}
	}
	return b.landings
}

// follow climbs with a jump from child, whose parent is p: to the scope,
// and the label it names is reached, or to a list that stops it, where it
// waits for a label of that list to be reached, when the list has one.
func (b *building) follow(jump *syntax.Goto, child, p syntax.Node) {
	owner, at := b.jumps.blocked(jump, child, p)
	if owner == nil {
		b.fired = append(b.fired, jump.Label.Value)
		return
	}
	if b.lists[owner] != nil {
		place := gap{owner, syntax.TailOf(owner).LastTerminating(at - 1)}
		waiting := b.parked[place]
		if waiting == nil {
			waiting = &parkings{byIndex: map[int][]parking{}, queued: map[int]bool{}}
			b.parked[place] = waiting
		}
		if !waiting.queued[at] {
			waiting.queued[at] = true
			heap.Push(&waiting.indexes, at)
		}
		waiting.byIndex[at] = append(waiting.byIndex[at], parking{jump, at, owner})
	}
}

// reach records the labels of a name as reached, in every list around each,
// and follows the jumps that waited there.
func (b *building) reach(name string, body syntax.Node) {
	if b.named[name] {
		return
	}
	b.named[name] = true
	var waking []parking
	wake := func(place gap, at int, through bool) {
		waiting := b.parked[place]
		if waiting == nil {
			return
		}
		take := func(index int) {
			waking = append(waking, waiting.byIndex[index]...)
			delete(waiting.byIndex, index)
		}
		// A label nested in the same statement may reach a parked jump even
		// when the containing statement cannot fall through.
		take(at)
		if through {
			for len(waiting.indexes) > 0 && waiting.indexes[0] > at {
				index := heap.Pop(&waiting.indexes).(int)
				delete(waiting.queued, index)
				take(index)
			}
		}
	}
	for _, k := range b.at[name] {
		b.mark(k)
		exits(b.labels[k], body, func(p syntax.Node, i int, leaves bool) {
			known, tail := b.lists[p], syntax.TailOf(p)
			last := tail.LastTerminating(i)
			improved := leaves && int32(i) < known.least[last-known.from]
			if improved {
				known.least[last-known.from] = int32(i)
			}
			wake(gap{p, last}, i, improved)
			if last == i && i > known.from {
				wake(gap{p, tail.LastTerminating(i - 1)}, i, false)
			}
		})
	}
	for _, w := range waking {
		list, _ := syntax.StmtListOf(w.owner)
		b.follow(w.jump, list[w.at], w.owner)
	}
}

// parkings indexes stopped jumps by their statement. A new earliest landing
// releases only later statements; jumps before it stay parked without work.
type parkings struct {
	byIndex map[int][]parking
	queued  map[int]bool
	indexes parkingIndexes
}

type parkingIndexes []int

func (p parkingIndexes) Len() int           { return len(p) }
func (p parkingIndexes) Less(i, j int) bool { return p[i] > p[j] }
func (p parkingIndexes) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p *parkingIndexes) Push(v any)        { *p = append(*p, v.(int)) }
func (p *parkingIndexes) Pop() any {
	last := len(*p) - 1
	v := (*p)[last]
	*p = (*p)[:last]
	return v
}
