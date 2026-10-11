package flow

import "custos/internal/php/syntax"

// bindings seeds the names of a scope that code outside it can replace: a
// reference parameter, a reference capture of the closure itself and, at file
// level, a variable some function of the file binds with `global`.
func (e *Env) bindings(scope syntax.Node, r *scopeResult, initial *frame) {
	if scope == nil {
		r.global = true
		for name := range e.globals() {
			initial.alias(name, false)
		}
		return
	}
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.ByRef && p.Var != nil {
			initial.alias(p.Var.Name, false)
		}
	}
	if closure, ok := scope.(*syntax.Closure); ok {
		for _, use := range closure.Uses {
			if use.ByRef && use.Var != nil {
				initial.alias(use.Var.Name, false)
			}
		}
	}
}

// globals returns the names functions of the file bind with `global`. The
// reserved empty name records a function that reads the globals table. The
// result stays nil for the common file without either.
func (e *Env) globals() map[string]bool {
	if e.globalsKnown {
		return e.global
	}
	e.globalsKnown = true
	syntax.InspectFile(e.file, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Global:
			if syntax.EnclosingVariableScope(n) != nil {
				for _, v := range n.Vars {
					e.bindGlobal(rootVariable(v), false)
				}
			}
		case *syntax.Variable:
			if n.Name == "GLOBALS" && n.NameExpr == nil && syntax.EnclosingVariableScope(n) != nil {
				e.bindGlobal("", true)
			}
		}
		return true
	})
	return e.global
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
			f.alias(rootVariable(item.Value), true)
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
