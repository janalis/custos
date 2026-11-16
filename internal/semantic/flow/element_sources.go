package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
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
