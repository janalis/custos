package infer

import (
	"strings"

	"custos/internal/index"
	"custos/internal/syntax"
	"custos/internal/types"
)

// Untyped property inference: a property without declared or documented
// type that only its own class can write (private, or protected in a final
// class; not static, not in a trait, the class using no trait) is typed as
// the union of its initial value (the default, else null; a promoted
// constructor parameter's type) and every value assigned to it in the
// class's methods (`$this->p = v`, `$other->p = v` on another instance).
// The implicit null is left out when the constructor assigns the property
// in a top-level statement; a property without default that the class
// never writes is unknown (hydrated through reflection). Unknown when any assigned value is unknown, or
// when the property is written in a way not tracked: by reference,
// destructuring, foreach target, unset, by-reference argument, or a
// dynamic property write (`$this->$name = v`) anywhere in the class.
// Element writes (`$this->p[] = v`) widen its arrays to plain `array`;
// `$this->p++` keeps an int|float|null type (adding int).

// propWrites lists the writes of one property inside its class.
type propWrites struct {
	values   []syntax.Expr    // `= v` values
	compound []*syntax.Assign // `.=`, `+=`, `??=`…
	incDec   bool
	elem     bool // element writes
	untrack  bool // written in an untracked way
	ctorSets bool // assigned by a top-level statement of the constructor
}

// classWrites collects the property writes made in class c (cached).
type classWrites struct {
	props   map[string]*propWrites
	dynamic bool // `$x->$name = v` somewhere: any property may be written
	// The declarations, by property name (built in the same pass: looking
	// each one up by scanning the members was quadratic).
	items    map[string]*syntax.PropertyItem // declared without hooks
	promoted map[string]*syntax.Param        // promoted constructor parameters
}

func (e *Env) classWritesOf(c *syntax.ClassLike) *classWrites {
	if cw, ok := e.propWritesCache[c]; ok {
		return cw
	}
	cw := &classWrites{props: map[string]*propWrites{}, items: map[string]*syntax.PropertyItem{}, promoted: map[string]*syntax.Param{}}
	for _, m := range c.Members {
		switch m := m.(type) {
		case *syntax.Property:
			if len(m.Hooks) == 0 {
				for _, it := range m.Props {
					if it.Var != nil {
						if _, dup := cw.items[it.Var.Name]; !dup {
							cw.items[it.Var.Name] = it
						}
					}
				}
			}
		case *syntax.Method:
			if strings.EqualFold(m.Name.Value, "__construct") {
				for _, p := range m.Params {
					if p.Var != nil && len(p.Modifiers) > 0 {
						if _, dup := cw.promoted[p.Var.Name]; !dup {
							cw.promoted[p.Var.Name] = p
						}
					}
				}
			}
		}
	}
	if e.propWritesCache == nil {
		e.propWritesCache = map[*syntax.ClassLike]*classWrites{}
	}
	e.propWritesCache[c] = cw
	get := func(name string) *propWrites {
		w := cw.props[name]
		if w == nil {
			w = &propWrites{}
			cw.props[name] = w
		}
		return w
	}
	// prop returns the property written by target x (with element writes
	// resolved to their base): its name and whether x is an element.
	prop := func(x syntax.Expr) (string, bool, bool) {
		x = unparen(x)
		elem := false
		for {
			d, ok := x.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			x, elem = unparen(d.Var), true
		}
		pf, ok := x.(*syntax.PropertyFetch)
		if !ok {
			return "", false, false
		}
		id, ok := pf.Name.(*syntax.Identifier)
		if !ok {
			return "", false, true
		}
		return id.Value, elem, false
	}
	untrack := func(x syntax.Expr) {
		if name, _, dyn := prop(x); dyn {
			cw.dynamic = true
		} else if name != "" {
			get(name).untrack = true
		}
	}
	var targets func(x syntax.Expr)
	targets = func(x syntax.Expr) {
		switch t := unparen(x).(type) {
		case *syntax.List:
			for _, it := range t.Items {
				if it != nil && it.Value != nil {
					targets(it.Value)
				}
			}
		case *syntax.Array:
			for _, it := range t.Items {
				if it != nil && it.Value != nil {
					targets(it.Value)
				}
			}
		default:
			untrack(t)
		}
	}
	args := func(call syntax.Expr, al *syntax.ArgList) {
		if al == nil {
			return
		}
		for i, x := range al.Args {
			a, ok := x.(*syntax.Arg)
			if !ok {
				continue
			}
			if name, _, _ := prop(a.Value); name == "" {
				continue
			}
			if a.ByRef || e.byRefArg(mutation{call: call, arg: a, idx: i}) {
				untrack(a.Value)
			}
		}
	}
	for _, m := range c.Members {
		meth, ok := m.(*syntax.Method)
		if !ok || meth.Body == nil {
			continue
		}
		ctor := strings.EqualFold(meth.Name.Value, "__construct")
		if ctor {
			for _, st := range meth.Body.Stmts {
				es, ok := st.(*syntax.ExprStmt)
				if !ok {
					continue
				}
				if a, ok := es.Expr.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
					if pf, ok := unparen(a.Var).(*syntax.PropertyFetch); ok && narrowKey(unparen(pf.Var)) == "this" {
						if id, ok := pf.Name.(*syntax.Identifier); ok {
							get(id.Value).ctorSets = true
						}
					}
				}
			}
		}
		syntax.Inspect(meth.Body, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.ClassLike, *syntax.Function:
				return false
			case *syntax.Assign:
				switch t := unparen(n.Var).(type) {
				case *syntax.List, *syntax.Array:
					targets(t)
				default:
					name, elem, dyn := prop(t)
					switch {
					case dyn:
						cw.dynamic = true
					case name == "":
					case n.ByRef:
						get(name).untrack = true
					case elem:
						get(name).elem = true
					case n.Op.Kind == syntax.TEqual:
						get(name).values = append(get(name).values, n.Value)
					default:
						get(name).compound = append(get(name).compound, n)
					}
				}
				if n.ByRef {
					untrack(n.Value)
				}
			case *syntax.IncDec:
				if name, elem, dyn := prop(n.Var); dyn {
					cw.dynamic = true
				} else if name != "" {
					if elem {
						get(name).elem = true
					} else {
						get(name).incDec = true
					}
				}
			case *syntax.Foreach:
				if n.ByRef {
					untrack(n.Expr)
				}
				if n.Key != nil {
					targets(n.Key)
				}
				targets(n.Value)
			case *syntax.Unset:
				for _, x := range n.Vars {
					if name, elem, dyn := prop(x); dyn {
						cw.dynamic = true
					} else if name != "" {
						if elem {
							get(name).elem = true
						} else {
							get(name).untrack = true
						}
					}
				}
			case *syntax.FuncCall:
				args(n, n.Args)
			case *syntax.MethodCall:
				args(n, n.Args)
			case *syntax.StaticCall:
				args(n, n.Args)
			case *syntax.New:
				args(n, n.Args)
			}
			return true
		})
	}
	return cw
}

// inferredProp is the type of untyped property p (see the comment above):
// computed from the class body when it is declared in this file, else the
// type the index computed (AnnotateReturns).
func (e *Env) inferredProp(p *index.Property) types.Type {
	if p.Type != "" || p.DocType != "" || p.Magic || p.Static {
		return types.Unknown
	}
	c := e.Index.Class(p.Class, e.PHP)
	if c == nil || c.File != e.File.Path {
		return types.FromDoc(p.Inferred, nil)
	}
	return e.propFromBody(p, c)
}

// propEligible reports whether only the class's own methods can write p.
func propEligible(p *index.Property, c *index.Class) bool {
	if p.Type != "" || p.DocType != "" || p.Magic || p.Static || c.Kind == syntax.KindTrait || len(c.Traits) > 0 {
		return false
	}
	return p.Visibility == index.Private || (p.Visibility == index.Protected && c.Final)
}

func (e *Env) propFromBody(p *index.Property, c *index.Class) types.Type {
	if !propEligible(p, c) {
		return types.Unknown
	}
	if t, ok := e.props[p.Span]; ok {
		return t
	}
	if e.propBusy[p.Span] {
		return types.Unknown
	}
	if e.propBusy == nil {
		e.propBusy = map[syntax.Span]bool{}
		e.props = map[syntax.Span]types.Type{}
	}
	e.propBusy[p.Span] = true
	t := e.computeProp(p, c)
	delete(e.propBusy, p.Span)
	e.props[p.Span] = t
	return t
}

func (e *Env) computeProp(p *index.Property, c *index.Class) types.Type {
	cl, _ := e.classAt(c.Span).(*syntax.ClassLike)
	if cl == nil {
		return types.Unknown
	}
	cw := e.classWritesOf(cl)
	if cw.dynamic {
		return types.Unknown
	}
	w := cw.props[p.Name]
	if w == nil {
		w = &propWrites{}
	}
	if w.untrack {
		return types.Unknown
	}
	var ts []types.Type
	// Initial value.
	if p.Promoted {
		param := cw.promoted[p.Name]
		if param == nil || param.Hooks != nil {
			return types.Unknown
		}
		ts = append(ts, e.paramType(param.Parent(), param))
	} else {
		item := cw.items[p.Name]
		if item == nil {
			return types.Unknown
		}
		switch {
		case item.Default != nil:
			ts = append(ts, e.TypeOf(item.Default))
		case len(w.values)+len(w.compound) == 0 && !w.incDec && !w.elem:
			// Never written by the class: set from outside (an ORM or
			// serializer using reflection), not null.
			return types.Unknown
		case !w.ctorSets:
			ts = append(ts, types.Null)
		}
	}
	for _, v := range w.values {
		ts = append(ts, e.TypeOf(v))
	}
	for _, a := range w.compound {
		ts = append(ts, e.TypeOf(a))
	}
	if len(ts) == 0 {
		return types.Unknown
	}
	t := types.Union(ts...)
	if t.IsUnknown() || t.Has("mixed") {
		return types.Unknown
	}
	if w.incDec {
		if !t.OnlyOf("int", "float", "null") {
			return types.Unknown
		}
		t = types.Union(t, types.Int)
	}
	if w.elem {
		atoms := []string{"array"}
		for _, a := range t.Atoms() {
			switch {
			case a == "array" || strings.HasSuffix(a, "[]"):
			case a == "null":
				atoms = append(atoms, a)
			default:
				return types.Unknown // string offsets, ArrayAccess objects…
			}
		}
		t = types.Of(atoms...)
	}
	return t
}

// classAt returns the class-like declared at span in this file.
func (e *Env) classAt(span syntax.Span) syntax.Node {
	if e.classes == nil {
		e.classes = map[syntax.Span]syntax.Node{}
		syntax.InspectFile(e.File, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.ClassLike); ok {
				e.classes[c.Span()] = c
			}
			return true
		})
	}
	return e.classes[span]
}
