package flow

import (
	"cmp"
	"slices"

	"custos/internal/php/syntax"
)

// bindings seeds the names of a scope that code outside it can replace: a
// reference parameter, a reference capture of the closure itself and, at file
// level, a variable some function of the file binds with `global`.
func (e *Env) bindings(scope syntax.Node, r *scopeResult, initial *frame) {
	if scope == nil {
		r.global = true
		for name := range e.globals() {
			initial.alias(name, aliasHeld)
		}
		return
	}
	for _, p := range syntax.VariableScopeParams(scope) {
		switch {
		case !p.ByRef || p.Var == nil:
		case p.Variadic:
			// The name holds an array of references: only a write to one of
			// its elements reaches an argument (see Summarize).
			initial.vars[received(p.Var.Name)] = initial.vars[p.Var.Name]
			initial.alias(p.Var.Name, aliasHeld)
		default:
			initial.receive(p.Var)
		}
	}
	if closure, ok := scope.(*syntax.Closure); ok {
		for _, use := range closure.Uses {
			if use.ByRef && use.Var != nil {
				initial.receive(use.Var)
			}
		}
	}
}

// received is the reserved binding, under a name no variable has, of the
// variable a reference parameter or a reference capture stands for when the
// scope starts.
func received(name string) string { return "&" + name }

// receive binds a name to the caller's variable it stands for. The two are
// bound like the names of `$a = &$b`: a write through the name reaches the
// variable until the name is made to stand for something else (`unset`,
// `static`, `global`, `$name = &$other`), after which the variable keeps
// what it held.
func (f *frame) receive(name *syntax.Variable) {
	f.bind(name.Name, received(name.Name), name.Span().Start)
	f.vars[received(name.Name)] = f.vars[name.Name]
}

// globals returns the names functions of the file bind with `global` or
// through a literal globals table entry. The reserved empty name records a
// function that uses the table in any other way. The
// result stays nil for the common file without either.
func (e *Env) globals() map[string]bool {
	if e.globalsKnown {
		return e.global
	}
	e.globalsKnown = true
	syntax.InspectFile(e.file, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Global:
			if scope := syntax.EnclosingVariableScope(n); scope != nil {
				for _, v := range n.Vars {
					name := rootVariable(v)
					e.bindGlobal(name, false)
					e.addWriter(scope, name)
				}
			}
		case *syntax.Variable:
			if scope := syntax.EnclosingVariableScope(n); n.Name == "GLOBALS" && n.NameExpr == nil && scope != nil {
				// A literal entry is that one variable; any other use of the
				// table can reach all of them.
				name := ""
				if entry, ok := n.Parent().(*syntax.ArrayDimFetch); ok && entry.Var == syntax.Expr(n) {
					name = assigned(entry)
				}
				e.bindGlobal(name, name == "")
				e.addWriter(scope, name)
			}
		}
		return true
	})
	// The request data those functions write reaches the file-level variable.
	// The writers are analysed in source order, each once however many names
	// it binds. One that reads a global after a call saw what the writers
	// before it leave only: when the request data written to a global
	// changes, the writers bound to that global are analysed again with it,
	// the others keeping their result, until nothing more is written. Each
	// round carries the data one global further, so the rounds are bounded
	// by the number of globals written, and by the scope transfer budget.
	writers := make([]globalWriter, 0, len(e.writers))
	for w := range e.writers {
		writers = append(writers, w)
	}
	e.writers = nil
	slices.SortFunc(writers, func(a, b globalWriter) int {
		return cmp.Or(cmp.Compare(a.scope.Span().Start, b.scope.Span().Start), cmp.Compare(a.name, b.name))
	})
	names := map[string]bool{}
	for _, w := range writers {
		names[w.name] = true
	}
	e.propagateGlobals(writers, len(names), MaxTransfers)
	return e.global
}

// forget drops the result of the writers bound to one of the globals named:
// they are analysed again, in the next round or when they are asked for, and
// with each the scopes around it, which read what it leaves in the variables
// it captures. The other writers read none of what changed.
func (e *Env) forget(writers []globalWriter, changed map[string]bool) {
	for _, w := range writers {
		if changed[w.name] {
			for scope := w.scope; scope != nil; scope = syntax.EnclosingVariableScope(scope) {
				delete(e.scopes, scope)
			}
		}
	}
}

// changedTaint names the global variables two rounds wrote different
// request data to.
func changedTaint(a, b map[string]Value) map[string]bool {
	var changed map[string]bool
	differ := func(a, b map[string]Value) {
		for name, v := range a {
			if w, ok := b[name]; !ok || !equal(v, w) {
				if changed == nil {
					changed = map[string]bool{}
				}
				changed[name] = true
			}
		}
	}
	differ(a, b)
	differ(b, a)
	return changed
}

// globalWriter is a function scope bound to a global variable.
type globalWriter struct {
	scope syntax.Node
	name  string
}

func (e *Env) addWriter(scope syntax.Node, name string) {
	if name == "" {
		return
	}
	if e.writers == nil {
		e.writers = map[globalWriter]bool{}
	}
	e.writers[globalWriter{scope, name}] = true
}

func (e *Env) bindGlobal(name string, table bool) {
	if name == "" && !table {
		return
	}
	if e.global == nil {
		e.global = map[string]bool{}
	}
	e.global[name] = true
}

func (e *Env) globalsTable() bool { return e.globals()[""] }

// rootVariable names the plain variable an lvalue chain is stored in.
func rootVariable(x syntax.Expr) string {
	for {
		switch n := syntax.UnwrapParens(x).(type) {
		case *syntax.ArrayDimFetch:
			x = n.Var
		case *syntax.Variable:
			if n.NameExpr != nil {
				return ""
			}
			return n.Name
		default:
			return ""
		}
	}
}

// referenceTarget aliases the slots a destructuring target binds by
// reference and reports whether it has any.
func referenceTarget(x syntax.Expr, f *frame) bool {
	found := false
	for _, item := range targetItems(x) {
		if item == nil {
			continue
		}
		if item.ByRef {
			if name := plainVariable(item.Value); name != "" {
				// The name leaves what it stood for, for the slot.
				f.detach(name)
			}
			f.alias(rootVariable(item.Value), aliasAny)
			found = true
		}
		if referenceTarget(item.Value, f) {
			found = true
		}
	}
	return found
}

// rebound lists the plain variables a write to target gives another value:
// the variable itself, or each one a destructuring names, at any depth.
func rebound(target syntax.Expr) []string {
	if name := plainVariable(target); name != "" {
		return []string{name}
	}
	var names []string
	for _, item := range targetItems(target) {
		if item != nil {
			names = append(names, rebound(item.Value)...)
		}
	}
	return names
}

// targetItems returns the slots of a `[...]` or `list(...)` destructuring target.
func targetItems(x syntax.Expr) []*syntax.ArrayItem {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Array:
		return n.Items
	case *syntax.List:
		return n.Items
	}
	return nil
}

// propagateGlobals charges new scope analyses and each writer visit to a
// shared budget, including the fixed-point bookkeeping between analyses.
func (e *Env) propagateGlobals(writers []globalWriter, names, budget int) {
	if e.globalTaint == nil {
		e.globalTaint = map[string]Value{}
	}
	for round, work := 0, 0; ; round++ {
		var taint map[string]Value
		for _, w := range writers {
			work++
			if work > budget {
				// The exact chain exceeded the scope budget. Every participating
				// global may hold external data; a truncated clean suffix is no proof.
				for _, writer := range writers {
					e.globalTaint[writer.name] = Value{Complete: true, Sources: []Source{{Kind: "global", Path: e.file.Path, Parameter: -1}}}
				}
				e.forget(writers, e.global)
				return
			}
			_, cached := e.scopes[w.scope]
			written := e.holderWrites(w.scope, w.name)
			if !cached {
				work += e.scopes[w.scope].operations
			}
			if retained(written) {
				if taint == nil {
					taint = map[string]Value{}
				}
				if old, ok := taint[w.name]; ok {
					written = effects(old, written, e.scopes[w.scope])
				}
				taint[w.name] = written
			}
		}
		changed := changedTaint(taint, e.globalTaint)
		if len(changed) == 0 {
			break
		}
		e.globalTaint = taint
		e.forget(writers, changed)
		if round >= names {
			break
		}
	}
}
