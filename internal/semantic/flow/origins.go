package flow

import (
	"sort"
	"strings"

	"custos/internal/php/syntax"
)

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
	v := r.values[x]
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
