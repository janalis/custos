package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

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
