package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// relation tells, for a short-circuit operation whose right operand assigns
// a variable and tests it in place, that wherever the left operand, read
// from the values now in place, would skip that operand, the variable is not
// false (relatedFalse) or not null (relatedNull). It holds across the turns
// of a loop that keep the left operand deciding with the assignment: the
// value a turn leaves in place where the next one skips passed the test.
type relation struct {
	by   *syntax.Binary
	bits uint8
}

const (
	relatedFalse uint8 = 1 << iota
	relatedNull
)

// relates returns what is recorded for an operation.
func (f *frame) relates(by *syntax.Binary) uint8 {
	for _, known := range f.related {
		if known.by == by {
			return known.bits
		}
	}
	return 0
}

// relate records what holds for an operation, in place of what did.
func (f *frame) relate(by *syntax.Binary, bits uint8) {
	kept := make([]relation, 0, len(f.related)+1)
	for _, known := range f.related {
		if known.by != by {
			kept = append(kept, known)
		}
	}
	if bits != 0 {
		kept = append(kept, relation{by, bits})
	}
	f.related = kept
}

// settled is what holds for an operation whatever its left operand decides:
// the variable is not false, or not null, on its own. A variable nothing
// assigned is null.
func (f *frame) settled(by *syntax.Binary) uint8 {
	v, tracked := f.vars.lookup(assignedInPlace(by.Right).Var.(*syntax.Variable).Name)
	return excluded(v, !tracked && f.definite)
}

// excluded is what a value settles: it is not false, or not null. null
// reports the value of a variable nothing assigned.
func excluded(v Value, null bool) uint8 {
	bits := uint8(0)
	if v.NonFalse || null {
		bits |= relatedFalse
	}
	if v.NonNull {
		bits |= relatedNull
	}
	return bits
}

// rewritten applies a write to what is recorded: the variable assigned in
// place keeps what the value written has, and a variable the left operand
// reads may now decide otherwise, which leaves what holds on its own.
func (f *frame) rewritten(r *scopeResult, name string, v Value) {
	if _, shared := f.aliased[name]; shared {
		// The names bound with it are written too.
		f.related = nil
		return
	}
	r.operations += len(f.related)
	var kept []relation
	for i, known := range f.related {
		bits := known.bits
		if c, _ := compares(known.by.Left); assignedInPlace(known.by.Right).Var.(*syntax.Variable).Name == name {
			bits &= excluded(v, false)
		} else if plainVariable(c.left) == name || plainVariable(c.right) == name {
			bits &= f.settled(known.by)
		}
		if bits == known.bits && kept == nil {
			continue
		}
		if kept == nil {
			kept = append(make([]relation, 0, len(f.related)), f.related[:i]...)
		}
		if bits != 0 {
			kept = append(kept, relation{known.by, bits})
		}
	}
	if kept != nil {
		f.related = kept
	}
}

// correlated joins what is recorded on two paths: what holds on both, by
// record or on its own. It reports whether the record of the frame changed.
func (f *frame) correlated(b frame) bool {
	var joined []relation
	for _, known := range f.related {
		if bits := (known.bits | f.settled(known.by)) & (b.relates(known.by) | b.settled(known.by)); bits != 0 {
			joined = append(joined, relation{known.by, bits})
		}
	}
	for _, known := range b.related {
		if bits := f.settled(known.by) & (known.bits | b.settled(known.by)); bits != 0 && f.relates(known.by) == 0 {
			joined = append(joined, relation{known.by, bits})
		}
	}
	changed := len(joined) != len(f.related)
	for i := 0; !changed && i < len(joined); i++ {
		changed = joined[i] != f.related[i]
	}
	f.related = joined
	return changed
}

const (
	aliasHeld uint32 = iota
	aliasAny
	aliasGroup // first identifier of a definite group
)

// imported is the reserved binding, under a name no variable has, for
// request data an import (`extract`, one-argument `parse_str`) may have put
// in any variable, tracked or not. It joins and clones like any binding.
const imported = ""

// importing joins request data an import may have spread over the scope.
func (f *frame) importing(v Value) {
	if tainted(v) {
		f.setVar(imported, unsettled(f.vars.value(imported), v, f.owner))
	}
}

// reserved reports a binding that is not a variable: a record kept under a
// name no variable has (see requestElement, received, imported).
func reserved(name string) bool {
	return name == imported || strings.ContainsAny(name, "\x00&")
}

// requestElement returns the reserved binding name, one no variable has, of
// an element of a request table read under one constant key: the binding
// records that a switch just found the element equal to a fixed text. The
// uploads table is left out, its entries are arrays PHP built.
func requestElement(x syntax.Expr) string {
	fetch, ok := syntax.UnwrapParens(x).(*syntax.ArrayDimFetch)
	if !ok {
		return ""
	}
	table, ok := fetch.Var.(*syntax.Variable)
	key, constant := arrayKey(fetch.Dim)
	if !ok || !constant || table.NameExpr != nil || !external(table.Name) || table.Name == "_FILES" {
		return ""
	}
	// An escape sequence may not spell the key it seems to.
	if literal, written := syntax.UnwrapParens(fetch.Dim).(*syntax.Literal); written && strings.Contains(literal.Raw, `\`) {
		return ""
	}
	return table.Name + "\x00" + key
}

// match records that the request table element of that reserved name holds
// a fixed text on this path. Any code can rewrite the table: the record
// does not outlive an expression that may run some (see implicit).
func (f *frame) match(r *scopeResult, element string) {
	f.setVar(element, Value{Complete: true, NonFalse: true})
	f.alias(element, aliasHeld)
	r.matching = true
}

// unmatchAll discards every recorded match: what is called, a PHP function
// included, may run code that rewrites the request tables (a callback, a
// conversion method of an argument, an error handler), and so may a
// destructor when a value is replaced.
func (f *frame) unmatchAll(r *scopeResult) {
	if !r.matching {
		return
	}
	r.operations += len(f.aliased)
	for name := range f.aliased {
		if strings.Contains(name, "\x00") {
			delete(f.aliased, name)
			f.unsetVar(name)
		}
	}
}

// matched reports that the element still holds the text it was matched with
// on every path: no join with a path without the match, no call and no
// write to the table came between, and no reference to the table or one of
// its elements exists.
func (f *frame) matched(element string) bool {
	table, _, _ := strings.Cut(element, "\x00")
	_, shared := f.aliased[table]
	return element != "" && !shared && f.vars.value(element).NonFalse
}

// unmatch discards the matches recorded for the elements of a request table
// that is written.
func (f *frame) unmatch(r *scopeResult, table string) {
	if !external(table) {
		return
	}
	r.operations += len(f.aliased)
	for name := range f.aliased {
		if strings.HasPrefix(name, table+"\x00") {
			delete(f.aliased, name)
			f.unsetVar(name)
		}
	}
}
