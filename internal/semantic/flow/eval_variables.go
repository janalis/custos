package flow

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// evalVariable reads a variable, named or computed.
func (e *Env) evalVariable(n *syntax.Variable, f *frame, r *scopeResult, depth int) Value {
	var v Value
	if n.NameExpr != nil {
		// The name may be the one of any variable, which a reference
		// to it or a callee given it can then write.
		if name := e.eval(n.NameExpr, f, r, depth+1); !name.plain {
			// An object given for the name is converted to text.
			e.implicit(n, f, r)
		}
		f.definite = false
		v = Value{}
		if assignment, target := n.Parent().(*syntax.Assign); !target || assignment.Var != n || assignment.Op.Kind != syntax.TEqual {
			e.observeParameters(n, f, r)
		}
		if name, known := literalText(n.NameExpr); known && external(name) {
			v = e.request(name, n, f)
		}
	} else if external(n.Name) {
		v = e.request(n.Name, n, f)
	} else {
		var tracked bool
		if v, tracked = f.vars.lookup(n.Name); !tracked {
			// A variable nothing assigned is null, and the globals table
			// is an array. The session table is whatever the application
			// put there.
			v.plain, v.deep = n.Name == "GLOBALS" || f.definite && importable(n.Name), f.definite && importable(n.Name)
			if spread := f.vars.value(imported); importable(n.Name) && tainted(spread) {
				// An import of request data may have created the variable.
				v = Value{Sources: spread.Sources, Safe: spread.Safe}
			}
		}
		v.tested, v.unless, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull, v.entry = nil, nil, false, false, false, false, false
	}
	return v
}

// evalAssign runs an assignment, plain, compound or null-coalescing.
func (e *Env) evalAssign(n *syntax.Assign, f *frame, r *scopeResult, depth int) Value {
	if n.Op.Kind == syntax.TCoalesceEqual {
		// The offsets of the target are always evaluated; the right
		// side runs only when the target is null.
		e.operands(n.Var, f, r, depth+1)
		assigned := clone(f)
		v := e.assign(n, &assigned, r, depth, true)
		*f = rejoin(*f, assigned, r)
		f.unmatchAll(r)
		return v
	}
	v := e.assign(n, f, r, depth, false)
	f.unmatchAll(r)
	return v
}

// evalIncDec steps a variable up or down.
func (e *Env) evalIncDec(n *syntax.IncDec, f *frame, r *scopeResult, depth int) Value {
	v := e.eval(n.Var, f, r, depth+1)
	if target, ok := n.Var.(*syntax.Variable); ok && target.NameExpr == nil {
		// The stepped variable holds the step applied to a known value,
		// which a reader can evaluate, and nothing provable otherwise.
		// Before PHP 8.3 a step leaves an object, or an array, as it is.
		strict := e.types.PHP.AtLeast(phpversion.PHP83)
		next := Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe, plain: v.plain || strict, deep: v.deep || strict}
		delta := int64(1)
		if n.Op.Kind == syntax.TDec {
			delta = -1
		}
		next.interval = arithmetic(syntax.TPlus, v, integerValue(delta))
		if v.Expr != nil {
			next.Expr = n
		}
		f.write(r, target.Name, next)
		if n.Prefix {
			v = next
			v.Expr = n
		}
	} else {
		kill(n.Var, f, r)
		v = Value{}
	}
	return v
}
