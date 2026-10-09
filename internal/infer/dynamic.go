package infer

import (
	"sort"
	"strings"

	"custos/internal/syntax"
)

// Dynamic writes and possibly undefined variables.
//
// extract(), parse_str() with one argument, `$$name = …` and include /
// require (the included file runs in the scope) may set any local variable: a read after one of them (in its block or later, or
// anywhere in a loop holding both) of a variable defined before it is
// unknown. A variable that only plain `=` assignments define, read on a
// path that assigns it nowhere, is null there (PHP warns and reads null).

// noteDynamic records in sv the dynamic constructs of node n of the scope.
func (e *Env) noteDynamic(n syntax.Node, sv *scopeVars) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		if isFirstClassCallable(n.Args) {
			return
		}
		f := e.ResolveFunction(n)
		if f == nil || !f.Builtin {
			return
		}
		switch strings.ToLower(strings.TrimPrefix(f.FQN, `\`)) {
		case "extract":
		case "parse_str":
			if len(n.Args.Args) != 1 {
				return
			}
		default:
			return
		}
		sv.clobbers = append(sv.clobbers, n.Span().Start)
		sv.dynamic = true
	case *syntax.Assign:
		if v, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable); ok && v.NameExpr != nil {
			sv.clobbers = append(sv.clobbers, n.Span().Start)
			sv.dynamic = true
		}
	case *syntax.Include:
		// The included file runs in this scope and may reassign any local.
		sv.clobbers = append(sv.clobbers, n.Span().Start)
		sv.dynamic = true
	}
}

// clobbered reports whether a dynamic write of the scope may have replaced
// the value of v: one lies after a definition reaching v and before v, or
// in a loop around v.
func (e *Env) clobbered(sv *scopeVars, fwd, back []varDef, v, scope syntax.Node) bool {
	if len(sv.clobbers) == 0 || (len(fwd) == 0 && len(back) == 0) {
		return false
	}
	at := v.Span().Start
	first := at
	for _, d := range fwd {
		first = min(first, d.pos)
	}
	// sv.clobbers is in source order: one in (first, at), or in the loop
	// around v.
	inRange := func(lo, hi uint32) bool {
		i := sort.Search(len(sv.clobbers), func(i int) bool { return sv.clobbers[i] >= lo })
		return i < len(sv.clobbers) && sv.clobbers[i] < hi
	}
	if inRange(first+1, at) {
		return true
	}
	loop := outermostLoop(v, scope)
	return loop != nil && inRange(loop.Span().Start, loop.Span().End)
}

// maxAssignScan caps the statements maybeUndefined examines for one read;
// beyond it the variable is taken as assigned (no null added).
const maxAssignScan = 64

// maybeUndefined reports whether v may be read before any assignment: all
// definitions of the variable in the scope are plain `$v = …` assignments
// (no parameter, import, binding, reference, element write, inline @var),
// the scope has no dynamic construct, no definition runs unconditionally
// in a block before v, and no statement or condition on the way to v
// always assigns it (if/else, switch with default, try/catch, do-while…).
func (e *Env) maybeUndefined(sv *scopeVars, defs, fwd []varDef, v *syntax.Variable, scope syntax.Node) bool {
	return e.maybeUndefinedAt(sv, defs, fwd, v.Name, v, scope)
}

func (e *Env) maybeUndefinedAt(sv *scopeVars, defs, fwd []varDef, name string, use, scope syntax.Node) bool {
	if sv.dynamic || len(sv.elemWrites[name]) > 0 {
		return false
	}
	if _, arrow := scope.(*syntax.ArrowFunction); arrow {
		return false // captures the enclosing scope
	}
	for _, d := range defs {
		if d.asg == nil && !d.barrier {
			return false
		}
	}
	at := use.Span().Start
	for _, d := range fwd {
		if d.kill.Len() > 0 && d.kill.Start <= at && at < d.kill.End {
			return false
		}
	}
	return !e.definitelyAssigned(use, name, scope)
}

// definitelyAssigned reports whether every path from the start of scope to
// x assigns variable name (or leaves the function): a statement preceding
// x in an enclosing list, or a condition evaluated before x.
func (e *Env) definitelyAssigned(x syntax.Node, name string, scope syntax.Node) bool {
	scanned := 0
	child := x
	for p := x.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		var stmts []syntax.Stmt
		switch n := p.(type) {
		case *syntax.Block:
			stmts = n.Stmts
		case *syntax.Case:
			stmts = n.Stmts
			if child == syntax.Node(n.Cond) {
				stmts = nil
			}
		case *syntax.If:
			if child != syntax.Node(n.Cond) && e.exprAssigns(n.Cond, name) {
				return true
			}
		case *syntax.ElseIf:
			if child != syntax.Node(n.Cond) && e.exprAssigns(n.Cond, name) {
				return true
			}
		case *syntax.While:
			if child == syntax.Node(n.Body) && e.exprAssigns(n.Cond, name) {
				return true
			}
		case *syntax.For:
			for _, c := range append(append([]syntax.Expr{}, n.Init...), n.Cond...) {
				if child != syntax.Node(c) && e.exprAssigns(c, name) {
					return true
				}
			}
		case *syntax.Foreach:
			if child != syntax.Node(n.Expr) && e.exprAssigns(n.Expr, name) {
				return true
			}
		case *syntax.Switch:
			if child != syntax.Node(n.Cond) && e.exprAssigns(n.Cond, name) {
				return true
			}
		case *syntax.Match:
			if child != syntax.Node(n.Cond) && e.exprAssigns(n.Cond, name) {
				return true
			}
		case *syntax.Binary:
			switch n.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
				if child == syntax.Node(n.Right) && e.exprAssigns(n.Left, name) {
					return true
				}
			}
		case *syntax.Ternary:
			if child != syntax.Node(n.Cond) && e.exprAssigns(n.Cond, name) {
				return true
			}
		}
		// The statements ending before x (not its own nor later ones).
		at := x.Span().Start
		first := sort.Search(len(stmts), func(i int) bool { return stmts[i].Span().End > at })
		for i := first - 1; i >= 0; i-- {
			if scanned++; scanned > maxAssignScan {
				return true // hostile bodies: see maxAssignScan
			}
			if e.stmtAssigns(stmts[i], name) {
				return true
			}
		}
	}
	if scope == nil {
		for i := len(e.File.Stmts) - 1; i >= 0; i-- {
			if s := e.File.Stmts[i]; s.Span().End <= x.Span().Start && e.stmtAssigns(s, name) {
				return true
			}
		}
	}
	return false
}

// stmtAssigns reports whether statement s, when it completes, has assigned
// name on every path, or always leaves the function (return, throw, exit).
// A continue, break or goto ends a list without assigning.
func (e *Env) stmtAssigns(s syntax.Stmt, name string) bool {
	k := assignKey{s, name}
	if r, ok := e.assignsMemo[k]; ok {
		return r
	}
	r := e.stmtAssignsUncached(s, name)
	if e.assignsMemo == nil {
		e.assignsMemo = map[assignKey]bool{}
	}
	e.assignsMemo[k] = r
	return r
}

func (e *Env) stmtAssignsUncached(s syntax.Stmt, name string) bool {
	switch n := s.(type) {
	case *syntax.Return:
		return true
	case *syntax.ExprStmt:
		if _, ok := syntax.UnwrapParens(n.Expr).(*syntax.Throw); ok {
			return true
		}
		if syntax.ExitInvocation(n.Expr) {
			return true
		}
		return e.exprAssigns(n.Expr, name)
	case *syntax.Echo:
		for _, x := range n.Exprs {
			if e.exprAssigns(x, name) {
				return true
			}
		}
	case *syntax.Block:
		return e.listAssigns(n.Stmts, name)
	case *syntax.If:
		if e.exprAssigns(n.Cond, name) {
			return true
		}
		if n.Else == nil || !e.stmtAssigns(n.Body, name) || !e.stmtAssigns(n.Else.Body, name) {
			return false
		}
		for _, ei := range n.ElseIfs {
			if !e.exprAssigns(ei.Cond, name) && !e.stmtAssigns(ei.Body, name) {
				return false
			}
		}
		return true
	case *syntax.Switch:
		if e.exprAssigns(n.Cond, name) {
			return true
		}
		hasDefault := false
		for _, c := range n.Cases {
			hasDefault = hasDefault || c.Cond == nil
			if len(c.Stmts) > 0 && !e.listAssigns(c.Stmts, name) {
				return false
			}
		}
		return hasDefault && len(n.Cases) > 0 && len(n.Cases[len(n.Cases)-1].Stmts) > 0
	case *syntax.Try:
		if n.Finally != nil && e.listAssigns(n.Finally.Body.Stmts, name) {
			return true
		}
		if !e.listAssigns(n.Body.Stmts, name) {
			return false
		}
		for _, c := range n.Catches {
			if !e.listAssigns(c.Body.Stmts, name) {
				return false
			}
		}
		return true
	case *syntax.DoWhile:
		return e.stmtAssigns(n.Body, name) || e.exprAssigns(n.Cond, name)
	case *syntax.While:
		return e.exprAssigns(n.Cond, name)
	case *syntax.For:
		for _, c := range append(append([]syntax.Expr{}, n.Init...), n.Cond...) {
			if e.exprAssigns(c, name) {
				return true
			}
		}
	case *syntax.Foreach:
		return e.exprAssigns(n.Expr, name)
	}
	return false
}

// listAssigns is stmtAssigns for a statement list: one statement assigning
// (or leaving) before any continue, break or goto.
func (e *Env) listAssigns(stmts []syntax.Stmt, name string) bool {
	for _, s := range stmts {
		if e.stmtAssigns(s, name) {
			return true
		}
		switch s.(type) {
		case *syntax.Continue, *syntax.Break, *syntax.Goto:
			return false
		}
	}
	return false
}

// exprAssigns reports whether evaluating x always assigns variable name
// with `=`: outside the right operands of && / || / ??, ternary branches,
// match arms and nested function bodies.
func (e *Env) exprAssigns(x syntax.Expr, name string) bool {
	found := false
	syntax.Inspect(x, func(n syntax.Node) bool {
		if found {
			return false
		}
		switch n := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike, *syntax.Match:
			return false
		case *syntax.Assign:
			if v, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name {
				found = true
				return false
			}
		case *syntax.Binary:
			switch n.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
				found = e.exprAssigns(n.Left, name)
				return false
			}
		case *syntax.Ternary:
			found = e.exprAssigns(n.Cond, name)
			return false
		}
		return true
	})
	return found
}

// dynamicRead reports, for a read of variable v, whether a dynamic write
// may have replaced its value (clobbered) and whether it may be read
// before any assignment (undef); the T-rules typer consults it.
func (e *Env) dynamicRead(v *syntax.Variable) (clob, undef bool) {
	if f, ok := e.dynReads[v]; ok {
		return f&1 != 0, f&2 != 0
	}
	scope := syntax.EnclosingVariableScope(v)
	sv := e.scopeVars(scope)
	defs := sv.defs[v.Name]
	if len(defs) == 0 || len(defs) > maxVarDefs {
		return false, false
	}
	fwd, back, _ := e.reaching(defs, v, scope)
	if len(fwd)+len(back) == 0 {
		return false, false
	}
	clob = e.clobbered(sv, fwd, back, v, scope)
	undef = !clob && e.maybeUndefined(sv, defs, fwd, v, scope)
	e.noteDynRead(v, clob, undef)
	return clob, undef
}

// noteDynRead caches the dynamicRead facts of read v.
func (e *Env) noteDynRead(v *syntax.Variable, clob, undef bool) {
	if e.dynReads == nil {
		e.dynReads = map[*syntax.Variable]uint8{}
	}
	var f uint8
	if clob {
		f |= 1
	}
	if undef {
		f |= 2
	}
	e.dynReads[v] = f
}

// assignKey identifies a stmtAssigns query.
type assignKey struct {
	s    syntax.Stmt
	name string
}
