package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// rebinds reports that evaluating x may write a variable tested reads: the
// target of an assignment or a step, or an argument a callee may take by
// reference. An earlier fact about that variable then describes a value that
// may be gone.
//
// A call written in x may also run code that writes a variable tested reads
// when that variable is bound to storage such code reaches: a static
// variable the function assigns, a global, a reference.
func (e *Env) rebinds(x, tested syntax.Expr, f *frame) bool {
	found := false
	if r := e.scopes[syntax.EnclosingVariableScope(tested)]; r != nil {
		reachable := false
		syntax.Inspect(tested, func(read syntax.Node) bool {
			if v, ok := read.(*syntax.Variable); ok && v.NameExpr == nil {
				_, shared := f.aliased[v.Name]
				reachable = reachable || shared || r.statics[v.Name].state != staticKept
			}
			return !reachable
		})
		syntax.Inspect(x, func(n syntax.Node) bool {
			if found {
				return false
			}
			switch n.(type) {
			case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
				found = reachable
			}
			return !found
		})
	}
	mentions := func(written syntax.Expr) {
		if name := rootVariable(written); name != "" {
			syntax.Inspect(tested, func(read syntax.Node) bool {
				if v, ok := read.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name {
					found = true
				}
				return !found
			})
		}
	}
	syntax.Inspect(x, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Assign:
			mentions(n.Var)
		case *syntax.IncDec:
			mentions(n.Var)
		case *syntax.ArgList:
			params, _, known := e.declaration(n.Parent())
			position := 0
			for _, x := range n.Args {
				a, ok := x.(*syntax.Arg)
				if !ok {
					continue
				}
				// Only a resolved by-value parameter proves a read.
				read := false
				for i, param := range params {
					if a.Name == nil && i == position || a.Name != nil && param.Name == a.Name.Value {
						read = known && !param.ByRef && !a.Unpack && !a.ByRef
					}
				}
				if a.Name == nil {
					position++
				}
				if !read {
					mentions(a.Value)
				}
			}
		}
		return !found
	})
	return found
}

// refineAssigned applies what a condition with a known outcome says about a
// value it assigns in place (`($u = f()) !== false`) to that producer alone:
// other producers of the variable reached the read without this test.
//
// Whatever reaches the read, the fact holds of it unless the operand was
// skipped: the binding keeps it for a later test of the left operand to
// settle (see ran).
func (e *Env) refineAssigned(operation *syntax.Binary, f *frame, truth bool) {
	cond := operation.Right
	stored := assignedInPlace(cond)
	if stored == nil {
		return
	}
	name := stored.Var.(*syntax.Variable).Name
	produced := syntax.UnwrapParens(stored.Value)
	alone := frame{owner: f.owner}
	alone.resetVars()
	alone.setVar(name, Value{Complete: true, Expr: produced})
	e.refine(cond, &alone, truth)
	fact := alone.vars.value(name)
	v, tracked := f.vars.lookup(name)
	if !(fact.NonFalse || fact.NonNull) || !tracked {
		return
	}
	if v.unless == operation && (fact.NonFalse && v.skippedFalse || fact.NonNull && v.skippedNull) {
		// The value the skipped operand left in place passed the same test.
		f.excluding(name, v, fact.NonFalse && v.skippedFalse, fact.NonNull && v.skippedNull)
		v = f.vars.value(name)
	}
	v.unless, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull = operation, fact.NonFalse, fact.NonNull, false, false
	f.setVar(name, v)
	if e.scopes[syntax.EnclosingVariableScope(stored)].stale[stored] || v.Expr != nil || v.origins == nil {
		return
	}
	n := &originSet{items: append([]Origin(nil), v.origins.items...), full: v.origins.full, compressed: v.origins.compressed, other: v.origins.other}
	for i, o := range n.items {
		if syntax.UnwrapParens(o.Expr) == produced {
			n.items[i] = Origin{o.Expr, o.NonFalse || fact.NonFalse, o.NonNull || fact.NonNull}
		}
	}
	v.origins = finishOrigins(n, f.owner, v.origins)
	f.setVar(name, v)
}

// assignedInPlace finds the plain assignment a condition tests directly:
// the condition itself, under negations, an operand of a comparison, or the
// first argument of a type predicate.
func assignedInPlace(cond syntax.Expr) *syntax.Assign {
	switch n := syntax.UnwrapParens(cond).(type) {
	case *syntax.Assign:
		if guarded(n) != nil {
			return n
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TExclaim {
			return assignedInPlace(n.Expr)
		}
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
			if stored := assignedInPlace(n.Left); stored != nil {
				return stored
			}
			return assignedInPlace(n.Right)
		}
	case *syntax.FuncCall:
		name, named := n.Name.(*syntax.Name)
		if !named || n.Args == nil || len(n.Args.Args) != 1 {
			return nil
		}
		switch strings.ToLower(strings.TrimPrefix(name.Value, `\`)) {
		case "is_string", "is_array", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_numeric",
			"is_object", "is_resource", "is_iterable", "is_countable", "is_callable", "is_bool", "is_null":
			if arg, ok := n.Args.Args[0].(*syntax.Arg); ok && arg.Name == nil && !arg.Unpack && !arg.ByRef {
				if stored, ok := syntax.UnwrapParens(arg.Value).(*syntax.Assign); ok && guarded(stored) != nil {
					return stored
				}
			}
		}
	}
	return nil
}
