package infer

import (
	"slices"
	"strings"

	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// Generic templates: a receiver typed `Collection<int, Foo>` (a class atom
// carrying generic arguments, see types.TypeArgs) binds the class-level
// templates of Collection; @extends / @implements / @use arguments carry
// the bindings to the ancestors. Methods whose documented return type
// mentions class templates (index.Method.GenReturn) return the bound types,
// and iterating a Traversable yields the bindings of Traversable's TKey and
// TValue.

// tplBindings maps a template name to its bound type.
type tplBindings map[string]types.Type

// maxGenAncestors caps the classes genBindings visits for one receiver:
// each @extends level substitutes the bindings into the next arguments, so
// a long hostile chain must not cost one doc parse per class without end.
// Ancestors beyond it get no bindings (their templates read as before:
// bound or mixed). Bindings whose text exceeds types.MaxDocTypeLen are
// dropped the same way (see substTemplates), so they cannot grow
// geometrically along the chain.
const maxGenAncestors = 64

// genBindings returns the template bindings of class cls instantiated with
// args, and of every ancestor reached through @extends/@implements/@use
// (lower-case FQN -> bindings). Ancestors without generic arguments are
// listed with no bindings. Cached per class and arguments.
func (e *Env) genBindings(cls string, args []types.Type) map[string]tplBindings {
	key := strings.ToLower(strings.TrimPrefix(cls, `\`))
	if len(args) > 0 {
		var b strings.Builder
		b.WriteString(key)
		for _, a := range args {
			b.WriteByte('|')
			b.WriteString(a.DocString())
		}
		key = b.String()
	}
	if r, ok := e.generics[key]; ok {
		return r
	}
	out := map[string]tplBindings{}
	type item struct {
		fqn  string
		args []types.Type
	}
	queue := []item{{strings.TrimPrefix(cls, `\`), args}}
	for len(queue) > 0 && len(out) < maxGenAncestors {
		it := queue[0]
		queue = queue[1:]
		k := strings.ToLower(strings.TrimPrefix(it.fqn, `\`))
		if _, seen := out[k]; seen {
			continue
		}
		c := e.Index.Class(it.fqn, e.PHP)
		if c == nil {
			continue
		}
		b := e.bindArgs(c, it.args)
		out[k] = b
		for _, s := range c.Supers {
			sargs := make([]types.Type, len(s.Args))
			for i, a := range s.Args {
				sargs[i], _ = substTemplates(a, b, c)
			}
			queue = append(queue, item{s.Class, sargs})
		}
		queue = append(queue, item{c.Parent, nil})
		for _, i := range c.Interfaces {
			queue = append(queue, item{i, nil})
		}
		for _, t := range c.Traits {
			queue = append(queue, item{t, nil})
		}
	}
	if e.generics == nil {
		e.generics = map[string]map[string]tplBindings{}
	}
	e.generics[key] = out
	return out
}

// bindArgs binds the templates of c to args: positionally, except when
// fewer arguments than templates are given. Then the arguments bind, in
// order, the templates without bound or default when there are exactly as
// many (Shopware's `@template TElement` + `@template TKey of array-key =
// array-key`: `Collection<Foo>` is its element type); else the trailing
// templates when the leading ones all have a bound or default; else, for a
// single argument to a Traversable class, the value template (`Generator<
// Foo>`). Templates left unbound take their default.
func (e *Env) bindArgs(c *index.Class, args []types.Type) tplBindings {
	if len(args) == 0 || len(c.Templates) == 0 {
		return nil
	}
	b := tplBindings{}
	pos := make([]int, 0, len(args)) // template index of each argument
	for i := range args {
		pos = append(pos, i)
	}
	if n := len(c.Templates) - len(args); n > 0 {
		var req []int
		lead := true
		for i, t := range c.Templates {
			opt := t.Bound != "" || t.Default != ""
			if !opt {
				req = append(req, i)
			}
			if i < n {
				lead = lead && opt
			}
		}
		switch {
		case len(req) == len(args):
			pos = req
		case lead:
			for i := range pos {
				pos[i] += n
			}
		case len(args) == 1 && e.Index.IsSubtype(c.FQN, "Traversable", e.PHP):
			pos[0] = 1
		}
	}
	for i, p := range pos {
		if p < len(c.Templates) {
			b[c.Templates[p].Name] = args[i]
		}
	}
	for _, t := range c.Templates {
		if _, ok := b[t.Name]; !ok && t.Default != "" {
			b[t.Name] = types.FromDoc(t.Default, nil)
		}
	}
	return b
}

// substTemplates parses doc type text whose `\~T` atoms are templates of
// class c, replacing them by their bindings in b (else their bound, else
// mixed). bound reports whether a template was replaced by an actual
// binding (neither unknown nor mixed).
func substTemplates(text string, b tplBindings, c *index.Class) (t types.Type, bound bool) {
	res := func(w string) string {
		name := strings.TrimPrefix(w, `\`)
		if !strings.HasPrefix(name, "~") {
			return name
		}
		name = name[1:]
		if bt, ok := b[name]; ok && !bt.IsUnknown() && !bt.Has("mixed") {
			// An over-long binding (see maxGenAncestors) is not used.
			if ds := bt.DocString(); len(ds) <= types.MaxDocTypeLen {
				bound = true
				return "=" + ds
			}
		}
		for _, tp := range c.Templates {
			if tp.Name == name && tp.Bound != "" {
				return "=" + tp.Bound
			}
		}
		return ""
	}
	t = types.FromDoc(text, res)
	return t, bound
}

// genMethodReturn types a call to method m on a receiver of class cls with
// generic arguments args (nil: the class's own @extends bindings only),
// when m's documented return type mentions class templates that the
// receiver binds. ok is false otherwise (callers use the plain types).
func (e *Env) genMethodReturn(m *index.Method, cls string, args []types.Type) (types.Type, bool) {
	if m.GenReturn == "" {
		return types.Unknown, false
	}
	b, ok := e.genBindings(cls, args)[strings.ToLower(strings.TrimPrefix(m.Class, `\`))]
	if !ok || len(b) == 0 {
		return types.Unknown, false
	}
	// genBindings lists only the classes the index has: c is non-nil.
	c := e.Index.Class(m.Class, e.PHP)
	t, bound := substTemplates(m.GenReturn, b, c)
	if !bound || t.IsUnknown() {
		return types.Unknown, false
	}
	return t, true
}

// traversal returns the key and value types of iterating a value of class
// cls with generic arguments args: the bindings of Traversable's templates
// (unknown when not bound).
func (e *Env) traversal(cls string, args []types.Type) (k, v types.Type) {
	tc := e.Index.Class("Traversable", e.PHP)
	if tc == nil || len(tc.Templates) < 2 {
		return types.Unknown, types.Unknown
	}
	b := e.genBindings(cls, args)["traversable"]
	return b[tc.Templates[0].Name], b[tc.Templates[1].Name]
}

// iterTypes returns the key and value types of iterating a value of type
// t that has non-array members: `T[]` members, `iterable<…>` and
// Traversable classes with bound generic arguments. Each is unknown when a
// member's is unknown or mixed.
func (e *Env) iterTypes(t types.Type) (key, val types.Type) {
	t = t.Without("null", "false")
	if t.IsUnknown() || len(t.Atoms()) == 0 {
		return types.Unknown, types.Unknown
	}
	var ks, vs []types.Type
	in := t.Intersection()
	if in != nil {
		// An intersection iterates as the side that binds Traversable.
		k, v := types.Unknown, types.Unknown
		for _, a := range in {
			if k, v = e.traversal(a, t.TypeArgs(a)); !v.IsUnknown() {
				break
			}
		}
		ks, vs = append(ks, k), append(vs, v)
	}
	for _, a := range t.Atoms() {
		if slices.Contains(in, a) {
			continue
		}
		var k, v types.Type
		switch {
		case strings.HasSuffix(a, "[]"):
			k, v = t.ArrayKey(), types.Of(strings.TrimSuffix(a, "[]"))
			if k.IsUnknown() {
				k = types.Of("int", "string")
			}
		case a == "iterable":
			switch args := t.TypeArgs(a); len(args) {
			case 1:
				v = args[0]
			case 2:
				k, v = args[0], args[1]
			}
		case strings.HasPrefix(a, `\`):
			k, v = e.traversal(a, t.TypeArgs(a))
		}
		ks, vs = append(ks, k), append(vs, v)
	}
	join := func(ts []types.Type) types.Type {
		for _, x := range ts {
			if x.IsUnknown() || x.Has("mixed") {
				return types.Unknown
			}
		}
		return types.Union(ts...)
	}
	return join(ks), join(vs)
}
