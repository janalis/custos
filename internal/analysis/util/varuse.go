package util

import "custos/internal/syntax"

// VarAccess is one occurrence of a named variable inside a scope.
type VarAccess struct {
	Var *syntax.Variable
	// Write is true when the occurrence (re)binds the variable: assignment
	// targets (plain, compound, by-reference, destructuring), foreach
	// key/value targets, global/static declarations, catch variables,
	// unset() arguments and ++/--.
	Write bool
	// Compound marks writes that also read the old value (`$v .= …`,
	// `$v++`).
	Compound bool
	// ElemWrite marks a read of a variable that is the base of an array
	// element / property being assigned (`$v[] = 1`, `$v['k'] = 1`,
	// `$v->p = 1`): the variable itself is not rebound.
	ElemWrite bool
	// By is the construct responsible for the access: the *syntax.Assign,
	// *syntax.Foreach, *syntax.Global, *syntax.StaticVar, *syntax.Catch,
	// *syntax.Unset, *syntax.IncDec or *syntax.ClosureUse; nil for an
	// ordinary read.
	By syntax.Node
}

// VarAccesses returns every access to the variable `name` (without '$') in
// scope, in source order. scope is a function, method, closure or arrow
// function (its parameters are not included), or nil for the file's
// top-level code. Bodies of nested functions, closures and classes are
// separate scopes and are skipped, except that a closure's `use ($v)` and
// any occurrence inside a nested arrow function (which captures by value)
// count as reads of the enclosing scope. Variable-variables are ignored.
// No control-flow reachability is applied.
func VarAccesses(f *syntax.File, scope syntax.Node, name string) []VarAccess {
	w := varWalker{name: name}
	switch s := scope.(type) {
	case nil:
		for _, st := range f.Stmts {
			w.node(st)
		}
	case *syntax.ArrowFunction:
		w.node(s.Expr)
	default:
		if b := FuncLikeBody(scope); b != nil {
			w.node(b)
		}
	}
	return w.out
}

type varWalker struct {
	name    string
	out     []VarAccess
	inArrow int
}

func (w *varWalker) add(v *syntax.Variable, write, compound, elem bool, by syntax.Node) {
	if v == nil || v.NameExpr != nil || v.Name != w.name {
		return
	}
	if w.inArrow > 0 && write {
		return // writes inside an arrow function are local to it
	}
	if w.inArrow > 0 {
		write, compound, elem, by = false, false, false, nil
	}
	w.out = append(w.out, VarAccess{Var: v, Write: write, Compound: compound, ElemWrite: elem, By: by})
}

// target records the variables written by an assignment-like target and
// walks its non-target sub-expressions (keys, indices) as reads.
func (w *varWalker) target(e syntax.Expr, compound bool, by syntax.Node) {
	switch t := e.(type) {
	case *syntax.Variable:
		if t.NameExpr != nil {
			w.node(t.NameExpr)
			return
		}
		w.add(t, true, compound, false, by)
	case *syntax.List:
		w.items(t.Items, by)
	case *syntax.Array:
		w.items(t.Items, by)
	case *syntax.ArrayDimFetch:
		w.elemBase(t.Var, by)
		if t.Dim != nil {
			w.node(t.Dim)
		}
	case *syntax.PropertyFetch:
		w.elemBase(t.Var, by)
		w.node(t.Name)
	default:
		if e != nil {
			w.node(e)
		}
	}
}

// elemBase walks the container of an element/property write.
func (w *varWalker) elemBase(e syntax.Expr, by syntax.Node) {
	switch b := e.(type) {
	case *syntax.Variable:
		if b.NameExpr == nil {
			w.add(b, false, false, true, by)
			return
		}
	case *syntax.ArrayDimFetch:
		w.elemBase(b.Var, by)
		if b.Dim != nil {
			w.node(b.Dim)
		}
		return
	case *syntax.PropertyFetch:
		w.elemBase(b.Var, by)
		w.node(b.Name)
		return
	}
	w.node(e)
}

func (w *varWalker) items(items []*syntax.ArrayItem, by syntax.Node) {
	for _, it := range items {
		if it == nil {
			continue
		}
		if it.Key != nil {
			w.node(it.Key)
		}
		if it.Value != nil {
			w.target(it.Value, false, by)
		}
	}
}

func (w *varWalker) node(n syntax.Node) {
	if n == nil {
		return
	}
	switch n := n.(type) {
	case *syntax.Function, *syntax.Method, *syntax.ClassLike:
		return
	case *syntax.Closure:
		for _, u := range n.Uses {
			w.add(u.Var, false, false, false, u)
		}
		return
	case *syntax.ArrowFunction:
		for _, p := range n.Params {
			if p.Default != nil {
				w.node(p.Default)
			}
		}
		for _, p := range n.Params {
			if p.Var != nil && p.Var.Name == w.name {
				return // parameter shadows the outer variable
			}
		}
		w.inArrow++
		w.node(n.Expr)
		w.inArrow--
		return
	case *syntax.Variable:
		if n.NameExpr != nil {
			w.node(n.NameExpr)
			return
		}
		w.add(n, false, false, false, nil)
		return
	case *syntax.Assign:
		// Source order: the target precedes the value.
		w.target(n.Var, n.Op.Kind != syntax.TEqual, n)
		w.node(n.Value)
		return
	case *syntax.IncDec:
		if v, ok := n.Var.(*syntax.Variable); ok && v.NameExpr == nil {
			w.add(v, true, true, false, n)
			return
		}
		w.target(n.Var, true, n)
		return
	case *syntax.Foreach:
		w.node(n.Expr)
		if n.Key != nil {
			w.target(n.Key, false, n)
		}
		if n.Value != nil {
			w.target(n.Value, false, n)
		}
		w.node(n.Body)
		return
	case *syntax.Global:
		for _, x := range n.Vars {
			if v, ok := x.(*syntax.Variable); ok {
				w.add(v, true, false, false, n)
			} else {
				w.node(x)
			}
		}
		return
	case *syntax.StaticVar:
		w.add(n.Var, true, false, false, n)
		if n.Default != nil {
			w.node(n.Default)
		}
		return
	case *syntax.Catch:
		for _, t := range n.Types {
			w.node(t)
		}
		if n.Var != nil {
			w.add(n.Var, true, false, false, n)
		}
		w.node(n.Body)
		return
	case *syntax.Unset:
		for _, x := range n.Vars {
			if v, ok := x.(*syntax.Variable); ok && v.NameExpr == nil {
				w.add(v, true, false, false, n)
			} else {
				w.node(x)
			}
		}
		return
	}
	syntax.Children(n, w.node)
}

// CountVarAccesses returns the number of reads and writes in accesses.
func CountVarAccesses(accesses []VarAccess) (reads, writes int) {
	for _, a := range accesses {
		if a.Write {
			writes++
		} else {
			reads++
		}
	}
	return reads, writes
}
