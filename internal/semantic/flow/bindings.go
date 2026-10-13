package flow

import "custos/internal/php/syntax"

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
	// Each writer scope is analysed once, however many names it binds.
	for w := range e.writers {
		if written := e.holderWrites(w.scope, w.name); tainted(written) {
			if e.globalTaint == nil {
				e.globalTaint = map[string]Value{}
			}
			if old, ok := e.globalTaint[w.name]; ok {
				written = merge(old, written)
			}
			e.globalTaint[w.name] = written
		}
	}
	e.writers = nil
	return e.global
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
