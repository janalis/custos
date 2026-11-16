package flow

import (
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
	if a.Expr == b.Expr && (a.Expr != nil || a.origins() == nil && b.origins() == nil) {
		return nil
	}
	if held := a.origins(); a.Expr == nil && held != nil && (held.full || held.holds(&b, r)) {
		return held
	}
	if held := b.origins(); b.Expr == nil && held != nil && (held.full || held.holds(&a, r)) {
		return held
	}
	n := &originSet{}
	if a.origins() != nil {
		n.compressed = a.origins().compressed
		n.other = a.origins().other
	}
	if b.origins() != nil {
		n.compressed = n.compressed || b.origins().compressed
		n.other = n.other || b.origins().other
	}
	size := 0
	for _, v := range [2]*Value{&a, &b} {
		if v.Expr != nil {
			size++
		} else if v.origins() != nil {
			size += len(v.origins().items)
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
		case v.origins() != nil:
			for _, o := range v.origins().items {
				add(o)
			}
		}
	}
	return finishOrigins(n, r, a.origins(), b.origins())
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
	case v.origins() == nil:
		return true
	case v.origins().full:
		return false
	case v.origins().other && !s.other:
		return false
	}
	for _, o := range v.origins().items {
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
