package flow

import (
	"maps"
	"math/bits"
	"slices"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
	if value.elements != nil {
		if value.elements.depth+height > MaxDepth {
			value.elements = nil
		} else {
			height += value.elements.depth
		}
	}
	s.hasResources = s.hasResources || value.storedIdentity != 0 || value.record != nil || value.elements != nil && value.elements.hasResources
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

// elements returns the tracked elements of the array a plain assignment
// stores: those of the variable, array literal or choice of either it reads.
// A long literal, or one with a key the analysis cannot name, is read as a
// whole.
func (e *Env) elements(value syntax.Expr, f *frame, r *scopeResult) *elementSet {
	switch n := syntax.UnwrapParens(value).(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.ArrayDimFetch:
		return r.values.at(n).elements
	case *syntax.Variable:
		if n.NameExpr == nil && !external(n.Name) {
			return f.vars.value(n.Name).elements
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TArrayCast {
			return e.elements(n.Expr, f, r)
		}
	case *syntax.Ternary:
		if truth, known := constantTruth(n.Cond); known {
			chosen := n.Else
			if truth {
				chosen = n.Then
				if chosen == nil {
					chosen = n.Cond
				}
			}
			return e.elements(chosen, f, r)
		}
		if n.Then != nil {
			return common(e.elements(n.Then, f, r), e.elements(n.Else, f, r))
		}
	case *syntax.Array:
		return r.values.at(n).elements
	}
	return nil
}

func (e *Env) literalElements(n *syntax.Array, r *scopeResult) *elementSet {
	if len(n.Items) == 0 {
		return emptyElements
	}
	if len(n.Items) > MaxSources {
		return nil
	}
	next, counting := int64(0), true
	set := &elementSet{items: make(map[string]Value, len(n.Items)), width: len(n.Items), whole: true}
	for _, item := range n.Items {
		if item == nil || item.Unpack || item.ByRef {
			return nil
		}
		key, known := strconv.FormatInt(next, 10), counting
		if item.Key != nil {
			key, known = arrayKey(item.Key)
		}
		if !known {
			return nil
		}
		// PHP numbers an item without a key after the largest integer
		// key so far; what follows a negative one depends on its version.
		if index, err := strconv.ParseInt(key, 10, 64); err == nil && index >= next {
			next = index + 1
		} else if err == nil && index < 0 {
			counting = false
		}
		own := r.values.at(item.Value)
		if _, existed := set.items[key]; !existed {
			set.order = append(set.order, key)
		}
		entry := elementValue(own)
		if own.Complete && !own.replaced && own.Invalidated == 0 {
			entry.storedIdentity = own.Identity
		} else {
			entry.storedIdentity = 0
		}
		set.put(key, entry)
	}
	r.operations += len(set.items)
	return set
}

// arrayKey returns the key PHP uses for a constant array offset: a decimal
// integer for an integer in any notation, a float (truncated), a boolean and
// a string that spells an integer; the empty string for null.
func arrayKey(x syntax.Expr) (string, bool) {
	negative := false
	x = syntax.UnwrapParens(x)
	if sign, ok := x.(*syntax.Unary); ok && sign.Op.Kind == syntax.TMinus {
		negative, x = true, syntax.UnwrapParens(sign.Expr)
	}
	switch n := x.(type) {
	case *syntax.ConstFetch:
		if n.Name == nil || negative {
			break
		}
		switch strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`)) {
		case "true":
			return "1", true
		case "false":
			return "0", true
		case "null":
			return "", true
		}
	case *syntax.Literal:
		raw := strings.ReplaceAll(n.Raw, "_", "")
		switch {
		case n.LitKind == syntax.LitInt:
			if index, err := strconv.ParseInt(raw, 0, 64); err == nil {
				if negative {
					index = -index
				}
				return strconv.FormatInt(index, 10), true
			}
		case n.LitKind == syntax.LitFloat:
			if number, err := strconv.ParseFloat(raw, 64); err == nil && number < 1<<53 {
				if negative {
					number = -number
				}
				return strconv.FormatInt(int64(number), 10), true
			}
		case !negative:
			text, known := literalText(n)
			if index, err := strconv.ParseInt(text, 10, 64); err == nil && strconv.FormatInt(index, 10) == text {
				return text, true
			}
			return text, known
		}
	}
	return "", false
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
