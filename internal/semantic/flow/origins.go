package flow

import (
	"sort"
	"strings"

	"custos/internal/php/syntax"
)

// Origin is one expression that may have produced the value read, with the
// failure values the guards on its paths exclude.
type Origin struct {
	Expr     syntax.Expr
	NonFalse bool
	NonNull  bool
}

// MaxOrigins is the ordinary expression retention threshold before compacting
// finite native contracts and non-call ambiguity. Sensitive calls remain exact.
const MaxOrigins = 64

// originSet is never modified once a binding holds it. A path that gave the
// binding a value no expression accounts for adds nothing to it: a set of one
// producer is a binding some other path left unknown. compressed marks finite
// contract representatives; other preserves non-call ambiguity without syntax.
// full is an exhausted host budget, which invalidates the owning scope.
type originSet struct {
	items      []Origin
	full       bool
	compressed bool
	other      bool
}

// origins joins the producers of a binding two paths disagree on. It returns
// the set one of them already holds when the other adds nothing to it, and a
// compact contract witnesses past MaxOrigins, within the owning scope budget.
func origins(a, b Value, owners ...*scopeResult) *originSet {
	var r *scopeResult
	if len(owners) > 0 {
		r = owners[0]
	}
	if a.Expr == b.Expr && (a.Expr != nil || a.origins == nil && b.origins == nil) {
		return nil
	}
	if held := a.origins; a.Expr == nil && held != nil && (held.full || held.holds(&b, r)) {
		return held
	}
	if held := b.origins; b.Expr == nil && held != nil && (held.full || held.holds(&a, r)) {
		return held
	}
	n := &originSet{}
	if a.origins != nil {
		n.compressed = a.origins.compressed
		n.other = a.origins.other
	}
	if b.origins != nil {
		n.compressed = n.compressed || b.origins.compressed
		n.other = n.other || b.origins.other
	}
	size := 0
	for _, v := range [2]*Value{&a, &b} {
		if v.Expr != nil {
			size++
		} else if v.origins != nil {
			size += len(v.origins.items)
		}
	}
	var positions map[syntax.Expr]int
	if size > MaxOrigins {
		positions = make(map[syntax.Expr]int, size)
	}
	add := func(o Origin) {
		if positions == nil {
			n.add(o)
			return
		}
		if i, ok := positions[o.Expr]; ok {
			known := n.items[i]
			n.items[i] = Origin{o.Expr, known.NonFalse && o.NonFalse, known.NonNull && o.NonNull}
			return
		}
		positions[o.Expr] = len(n.items)
		n.items = append(n.items, o)
	}
	for _, v := range [2]*Value{&a, &b} {
		switch {
		case v.Expr != nil:
			add(Origin{v.Expr, v.NonFalse, v.NonNull})
		case v.origins != nil:
			for _, o := range v.origins.items {
				add(o)
			}
		}
	}
	return finishOrigins(n, r, a.origins, b.origins)
}

// add joins one producer into a set under construction.
func (s *originSet) add(o Origin) {
	for i, known := range s.items {
		if known.Expr == o.Expr {
			s.items[i] = Origin{o.Expr, known.NonFalse && o.NonFalse, known.NonNull && o.NonNull}
			return
		}
	}
	s.items = append(s.items, o)
}

// holds reports that joining v adds nothing to the set: each producer of v
// is there with no more guards than v has.
func (s *originSet) holds(v *Value, owners ...*scopeResult) bool {
	var r *scopeResult
	if len(owners) > 0 {
		r = owners[0]
	}
	has := func(o Origin) bool {
		if s.other && !callOrigin(o.Expr) {
			return true
		}
		for _, known := range s.items {
			if s.compressed && r != nil && r.originKey(o.Expr) != "" && r.originKey(o.Expr) == r.originKey(known.Expr) && known.NonFalse == o.NonFalse && known.NonNull == o.NonNull && known.Expr.Span().Start <= o.Expr.Span().Start {
				return true
			}
			if known.Expr == o.Expr {
				return (!known.NonFalse || o.NonFalse) && (!known.NonNull || o.NonNull)
			}
		}
		return false
	}
	switch {
	case v.Expr != nil:
		return has(Origin{v.Expr, v.NonFalse, v.NonNull})
	case v.origins == nil:
		return true
	case v.origins.full:
		return false
	case v.origins.other && !s.other:
		return false
	}
	for _, o := range v.origins.items {
		if !has(o) {
			return false
		}
	}
	return true
}

// excluding returns the set with a failure value excluded for every producer.
func (s *originSet) excluding(nonFalse, nonNull bool, owners ...*scopeResult) *originSet {
	if s == nil {
		return nil
	}
	unchanged := true
	for _, o := range s.items {
		unchanged = unchanged && (!nonFalse || o.NonFalse) && (!nonNull || o.NonNull)
	}
	if unchanged {
		return s
	}
	var r *scopeResult
	if len(owners) > 0 {
		r = owners[0]
	}
	n := &originSet{items: make([]Origin, len(s.items)), full: s.full, compressed: s.compressed, other: s.other}
	for i, o := range s.items {
		n.items[i] = Origin{o.Expr, o.NonFalse || nonFalse, o.NonNull || nonNull}
	}
	return finishOrigins(n, r, s)
}

// sameOrigins compares two sets by content.
func sameOrigins(a, b *originSet) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil || a.full != b.full || a.compressed != b.compressed || a.other != b.other || len(a.items) != len(b.items) {
		return false
	}
	for i, o := range a.items {
		if o != b.items[i] {
			return false
		}
	}
	return true
}

// unguarded returns the set without any exclusion.
func (s *originSet) unguarded(owners ...*scopeResult) *originSet {
	if s == nil {
		return nil
	}
	unchanged := true
	for _, o := range s.items {
		unchanged = unchanged && !o.NonFalse && !o.NonNull
	}
	if unchanged {
		return s
	}
	var r *scopeResult
	if len(owners) > 0 {
		r = owners[0]
	}
	n := &originSet{items: make([]Origin, len(s.items)), full: s.full, compressed: s.compressed, other: s.other}
	for i, o := range s.items {
		n.items[i] = Origin{Expr: o.Expr}
	}
	return finishOrigins(n, r, s)
}

// originKey describes only contracts whose consumers do not distinguish the
// arguments of different calls. Other producers remain individually retained.
func (r *scopeResult) originKey(x syntax.Expr) string {
	if r == nil || r.originTypes == nil {
		return ""
	}
	call, ok := syntax.UnwrapParens(x).(*syntax.FuncCall)
	if !ok {
		return ""
	}
	if key, ok := r.originKeys[x]; ok {
		return key
	}
	key := ""
	if fn := r.originTypes.ResolveFunction(call); fn != nil && fn.Builtin && fn.Avail.In(r.originTypes.PHP) {
		name := strings.ToLower(strings.TrimPrefix(fn.FQN, `\`))
		switch name {
		// Only audited argument-insensitive native failure contracts compact.
		case "file_get_contents", "stream_get_contents", "preg_replace", "preg_replace_callback", "glob", "scandir", "posix_getpwnam", "posix_getpwuid", "normalizer_normalize", "fileperms", "realpath":
			key = name
		}
	}
	if r.originKeys == nil {
		r.originKeys = map[syntax.Expr]string{}
	}
	r.originKeys[x] = key
	if key != "" {
		if r.originContracts == nil {
			r.originContracts = map[string]bool{}
		}
		if !r.originContracts[key] {
			r.originContracts[key] = true
			r.originReserve = 1 + 4*len(r.originContracts)
		}
	}
	return key
}

// finishOrigins publishes a fresh immutable set and charges its retained cells
// once, at allocation. Sharing that set through copied Values costs no cells.
// Temporary builder slices are not retained scope inputs.
func finishOrigins(n *originSet, r *scopeResult, previous ...*originSet) *originSet {
	if r != nil {
		for _, o := range n.items {
			r.originKey(o.Expr)
		}
	}
	compact := len(n.items) > MaxOrigins || n.compressed
	if r != nil && len(n.items)+1 > MaxRetained-r.originReserve-r.retained {
		compact = true
	}
	if compact {
		type contract struct {
			name              string
			nonFalse, nonNull bool
		}
		representatives := map[contract]int{}
		items := make([]Origin, 0, len(n.items))
		for _, o := range n.items {
			if !callOrigin(o.Expr) {
				n.other = true
				n.compressed = true
				continue
			}
			key := r.originKey(o.Expr)
			if key == "" {
				items = append(items, o)
				continue
			}
			k := contract{key, o.NonFalse, o.NonNull}
			if at, ok := representatives[k]; ok {
				if o.Expr.Span().Start < items[at].Expr.Span().Start {
					items[at] = o
				}
			} else {
				representatives[k] = len(items)
				items = append(items, o)
			}
		}
		n.compressed = n.compressed || len(items) < len(n.items)
		n.items = items
		// Source order gives stable real representatives across fixed-point joins.
		sort.Slice(n.items, func(i, j int) bool {
			a, b := n.items[i], n.items[j]
			if a.Expr.Span().Start != b.Expr.Span().Start {
				return a.Expr.Span().Start < b.Expr.Span().Start
			}
			if a.Expr.Span().End != b.Expr.Span().End {
				return a.Expr.Span().End < b.Expr.Span().End
			}
			if a.NonFalse != b.NonFalse {
				return !a.NonFalse
			}
			return !a.NonNull && b.NonNull
		})
	}
	for _, held := range previous {
		if sameOrigins(n, held) {
			return held
		}
	}
	if r != nil && !retainOriginCells(r, len(n.items)+1) {
		return &originSet{full: true}
	}
	// Do not retain an append builder's unused capacity.
	if cap(n.items) != len(n.items) {
		items := make([]Origin, len(n.items))
		copy(items, n.items)
		n.items = items
	}
	return n
}

// CompressedOrigins reports that Reaching supplies contract representatives,
// rather than a complete expression list. Arbitrary predicates cannot use it.
func (e *Env) CompressedOrigins(x syntax.Expr) bool {
	if x == nil {
		return false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v := r.values.at(x)
	return r.complete && v.Expr == nil && v.origins != nil && v.origins.compressed
}

// ProducerBudgetExceeded distinguishes an exhausted host scope from contract
// compression. Neither lexical history nor representative sampling repairs it.
func (e *Env) ProducerBudgetExceeded(x syntax.Expr) bool {
	return x != nil && e.scope(syntax.EnclosingVariableScope(x)).spent
}

// Origin allocations may use the reserved tail for compact representatives;
// ordinary block inputs leave that room available.
func retainOriginCells(r *scopeResult, cells int) bool {
	if cells > MaxRetained-r.retained {
		r.spent = true
		r.complete = false
		return false
	}
	r.retained += cells
	return true
}

// Non-call paths cannot witness a native failure contract. A compact set keeps
// their ambiguity without claiming any expression is the exact value.
func callOrigin(x syntax.Expr) bool {
	for {
		x = syntax.UnwrapParens(x)
		unary, ok := x.(*syntax.Unary)
		if !ok || unary.Op.Kind != syntax.TAt {
			break
		}
		x = unary.Expr
	}
	switch x.(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
		return true
	}
	return false
}
