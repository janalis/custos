package flow

import (
	"maps"
	"math/bits"
	"slices"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const (
	uploadEntry uint8 = iota + 1
	uploadTable
)

// elementSet is never modified once a binding holds it: a write makes a new
// one, which lets every copy of the array value share it.
type elementSet struct {
	hasResources bool
	whole        bool // every runtime entry is tracked; false after losing an entry
	optional     bool // tracked keys may be absent, so membership/count is not exact
	items        map[string]Value
	order        []string // nil unless insertion order is proven for the complete set
	depth        int      // bounded structural cost, including branching
	width        int      // upper bound on the number of entries during construction
}

// Empty element sets contain no mutable evaluation metadata and are shared.
var emptyElements = &elementSet{whole: true}

func cloneElements(s *elementSet) *elementSet {
	out := *s
	out.items = maps.Clone(s.items)
	return &out
}

// put is only used while constructing an unpublished set. A loop that wraps
// its previous array cannot accumulate an unbounded chain of nested facts.
func (s *elementSet) put(key string, value Value) {
	height := bits.Len(uint(max(1, s.width)))
	if value.elements() != nil {
		if value.elements().depth+height > MaxDepth {
			value.setElements(nil)
		} else {
			height += value.elements().depth
		}
	}
	s.hasResources = s.hasResources || value.storedIdentity() != 0 || value.record() != nil || value.elements() != nil && value.elements().hasResources
	s.depth = max(s.depth, height)
	s.items[key] = value
}

// without returns the set less one element, which a callee holding a
// reference to it may have rewritten; nothing when the element is not known.
func (s *elementSet) without(key string, known bool) *elementSet {
	if s == nil || !known {
		return nil
	}
	n := &elementSet{items: map[string]Value{}, width: len(s.items)}
	for k, v := range s.items {
		if k != key {
			n.put(k, v)
		}
	}
	if len(n.items) == 0 {
		return nil
	}
	return n
}

// with returns the set with one more element written.
func (s *elementSet) with(key string, v Value) *elementSet {
	width := 1
	if s != nil {
		width += len(s.items)
	}
	n := &elementSet{items: map[string]Value{}, width: width}
	if s != nil {
		n.whole = s.whole
		n.optional = s.optional
	}
	n.put(key, v)
	if s != nil {
		for k, old := range s.items {
			if k != key {
				n.put(k, old)
			}
		}
	}
	return n
}

// common keeps the elements tracked in both sets, each holding what either
// may. It returns a itself when b adds nothing to it.
func common(a, b *elementSet) *elementSet {
	if a == b {
		return a
	}
	if a == nil || b == nil {
		return nil
	}
	kept := slices.Equal(a.order, b.order) && (!a.whole || b.whole && len(a.items) == len(b.items)) && (!b.optional || a.optional)
	for k, v := range a.items {
		w, ok := b.items[k]
		if !ok || !equal(merge(v, w), v) {
			kept = false
			break
		}
	}
	if kept {
		return a
	}
	n := &elementSet{items: map[string]Value{}, width: min(len(a.items), len(b.items)), whole: a.whole && b.whole && len(a.items) == len(b.items), optional: a.optional || b.optional}
	if slices.Equal(a.order, b.order) {
		n.order = a.order
	}
	for k, v := range a.items {
		if w, ok := b.items[k]; ok {
			n.put(k, merge(v, w))
		}
	}
	if len(n.items) == 0 {
		return nil
	}
	if len(n.items) != len(a.items) {
		n.whole = false
	}
	return n
}

// lookup returns a tracked element.
func (s *elementSet) lookup(key string) (Value, bool) {
	if s == nil {
		return Value{}, false
	}
	v, ok := s.items[key]
	return v, ok
}

// all returns the tracked elements; none when nothing is tracked.
func (s *elementSet) all() map[string]Value {
	if s == nil {
		return nil
	}
	return s.items
}

// list reports an array a PHP function built and numbered itself: none of
// its keys comes from the data.
func (e *Env) list(array syntax.Expr) bool {
	call, ok := syntax.UnwrapParens(array).(*syntax.FuncCall)
	if !ok {
		return false
	}
	switch name, builtin := e.callName(call); {
	case !builtin:
	case name == "explode", name == "preg_split", name == "str_split", name == "mb_str_split", name == "array_values", name == "array_keys", name == "str_getcsv", name == "range":
		return true
	}
	return false
}

// edits reports a PHP function whose result is its text arguments, changed
// in case, trimmed, cut, padded, replaced, formatted or escaped.
func edits(name string) bool {
	switch name {
	case "strtolower", "strtoupper", "ucfirst", "ucwords", "lcfirst", "trim", "ltrim", "rtrim", "substr", "str_pad", "str_replace", "sprintf", "strval",
		"htmlspecialchars", "htmlentities":
		return true
	}
	return false
}

// yieldsElements reports a PHP function whose result is made of the values
// of the arrays it is given, without their keys.
func yieldsElements(name string) bool {
	switch name {
	case "implode", "join", "array_values", "array_sum", "array_product", "max", "min", "count", "sizeof", "in_array", "reset", "end", "current", "array_pop", "array_shift":
		return true
	}
	return false
}

// keepsElements reports a PHP function whose result is its first argument
// with some elements left out or reordered, under the same keys.
func keepsElements(name string) bool {
	switch name {
	case "array_unique", "array_filter", "array_slice", "array_reverse", "array_diff", "array_diff_key", "array_intersect", "array_intersect_key":
		return true
	}
	return false
}

// passed returns the value given to the parameter of that name and position.
func passed(list *syntax.ArgList, values []Value, name string, position int) (Value, bool) {
	at := 0
	for i, v := range values {
		if a, ok := list.Args[i].(*syntax.Arg); ok && a.Name != nil {
			if a.Name.Value == name {
				return v, true
			}
			continue
		}
		if at == position {
			return v, true
		}
		at++
	}
	return Value{}, false
}

// completeSpread accepts only an evaluated literal-array producer. Completeness
// already accounts for nested spreads, references and unknown element values.
func (e *Env) completeSpread(v Value) bool {
	if !v.Complete || v.Expr == nil || e.types.PHP.Below(phpversion.PHP74) {
		return false
	}
	array, ok := syntax.UnwrapParens(v.Expr).(*syntax.Array)
	if !ok {
		return false
	}
	if e.types.PHP.AtLeast(phpversion.PHP81) {
		return true
	}
	// Before 8.1 string keys cannot be unpacked. Restrict this proof to
	// implicit numeric keys; explicit and nested keys remain conservative.
	for _, item := range array.Items {
		if item != nil && (item.Key != nil || item.Unpack) {
			return false
		}
	}
	return true
}
