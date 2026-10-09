package types

import "strings"

type relativeBinding struct {
	self, parent, receiver string
	visited                int
}

// HasRelative reports relative atoms anywhere in a contract's metadata.
// Oversized metadata is treated conservatively as requiring binding.
func (t Type) HasRelative() bool {
	r := relativeBinding{}
	return r.has(t, 0)
}

func (r *relativeBinding) has(t Type, depth int) bool {
	if depth >= 64 || r.visited >= 4096 {
		return true
	}
	r.visited++
	for _, a := range t.atoms {
		switch strings.TrimRight(a, "[]") {
		case "self", "parent", "static":
			return true
		}
	}
	for _, g := range t.gen {
		for _, arg := range g.args {
			if r.has(arg, depth+1) {
				return true
			}
		}
	}
	return r.hasArray(t.arr, depth+1)
}

func (r *relativeBinding) hasArray(a *arrayInfo, depth int) bool {
	if a == nil {
		return false
	}
	if depth >= 64 || r.visited >= 4096 {
		return true
	}
	r.visited++
	for _, k := range a.keys {
		if r.has(k.Type, depth+1) {
			return true
		}
	}
	return r.hasArray(a.elem, depth+1)
}

// BindRelative resolves self/parent at their effective declaration and static
// at the receiver. Empty names make the corresponding contract unknown.
// Array, generic, callable and intersection metadata are retained immutably.
func (t Type) BindRelative(self, parent, receiver string) Type {
	r := relativeBinding{self: self, parent: parent, receiver: receiver}
	out, _ := r.bind(t, 0)
	return out
}

func (r *relativeBinding) atom(a string) string {
	base := strings.TrimRight(a, "[]")
	var name string
	switch base {
	case "self":
		name = r.self
	case "parent":
		name = r.parent
	case "static":
		name = r.receiver
	default:
		return a
	}
	if name == "" {
		return ""
	}
	return `\` + strings.TrimPrefix(name, `\`) + a[len(base):]
}

func (r *relativeBinding) bind(t Type, depth int) (Type, bool) {
	if depth >= 64 || r.visited >= 4096 {
		return Unknown, true
	}
	r.visited++
	var atoms []string
	for i, a := range t.atoms {
		b := r.atom(a)
		if b == "" {
			return Unknown, true
		}
		if b != a && atoms == nil {
			atoms = append([]string(nil), t.atoms...)
		}
		if atoms != nil {
			atoms[i] = b
		}
	}
	out := t
	changed := atoms != nil
	if changed {
		out = Of(atoms...)
		out.arr, out.gen, out.inter = t.arr, t.gen, t.inter
	}
	var gens []genEntry
	for i, g := range t.gen {
		var args []Type
		for j, arg := range g.args {
			if r.visited >= 4096 {
				return Unknown, true
			}
			bound, change := r.bind(arg, depth+1)
			if change && args == nil {
				args = append([]Type(nil), g.args...)
			}
			if args != nil {
				args[j] = bound
			}
		}
		atom := r.atom(g.atom)
		if (args != nil || atom != g.atom) && gens == nil {
			gens = append([]genEntry(nil), t.gen...)
		}
		if gens != nil {
			if args == nil {
				args = g.args
			}
			gens[i] = genEntry{atom: atom, args: args}
		}
	}
	if gens != nil {
		out = out.withGen(gens)
		changed = true
	}
	if arr, change := r.array(t.arr, depth+1); change {
		out.arr = arr
		changed = true
	}
	return out, changed
}

func (r *relativeBinding) array(a *arrayInfo, depth int) (*arrayInfo, bool) {
	if a == nil {
		return nil, false
	}
	if depth >= 64 || r.visited >= 4096 {
		return nil, true
	}
	r.visited++
	var keys []ShapeKey
	for i, k := range a.keys {
		if r.visited >= 4096 {
			return nil, true
		}
		bound, change := r.bind(k.Type, depth+1)
		if change && keys == nil {
			keys = append([]ShapeKey(nil), a.keys...)
		}
		if keys != nil {
			keys[i].Type = bound
		}
	}
	elem, change := r.array(a.elem, depth+1)
	if keys == nil && !change {
		return a, false
	}
	copy := *a
	if keys != nil {
		copy.keys = keys
	}
	copy.elem = elem
	return &copy, true
}
