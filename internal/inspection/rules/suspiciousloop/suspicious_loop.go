package suspiciousloop

import (
	"slices"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// sideEffectsFirst reports whether every condition expression before the
// last one is an assignment, an increment or a call: run for its effect each
// iteration, not meant as a condition (E5).
func sideEffectsFirst(cond []syntax.Expr) bool {
	for _, e := range cond[:len(cond)-1] {
		switch syntax.UnwrapParens(e).(type) {
		case *syntax.Assign, *syntax.IncDec, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		default:
			return false
		}
	}
	return true
}

// suspiciousLoop reports `for` conditions with several expressions and loop
// variables overwriting parameters or outer loop variables.
type suspiciousLoop struct{}

const (
	suspiciousLoopMultiCondMsg = "Only the last expression of the 'for' condition is evaluated as the condition; combine them with && or ||."
	suspiciousLoopOuterMsg     = "' overwrites a variable of an outer loop."
)

func (suspiciousLoop) ID() string { return "SuspiciousLoop" }
func (suspiciousLoop) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFor, syntax.KForeach}
}

func (suspiciousLoop) Check(ctx *analysis.Context, n syntax.Node) {
	s := n.Span()
	kwLen := uint32(len("foreach"))
	if f, ok := n.(*syntax.For); ok {
		kwLen = uint32(len("for"))
		if len(f.Cond) >= 2 && !sideEffectsFirst(f.Cond) { // D1, E5
			ctx.Report(syntax.Span{Start: s.Start, End: s.Start + kwLen}, suspiciousLoopMultiCondMsg)
		}
	}
	kw := syntax.Span{Start: s.Start, End: s.Start + kwLen}
	if !ctx.Bool("VERIFY_VARIABLES_OVERRIDE") { // E4
		return
	}
	names := loopVarNames(n, nil) // D2
	if len(names) == 0 {
		return
	}
	if fn := syntax.EnclosingFuncLike(n); fn != nil { // D3
		kind := "function"
		if _, ok := fn.(*syntax.Method); ok {
			kind = "method"
		}
		params := syntax.FuncLikeParams(fn)
		for _, name := range names {
			for _, p := range params {
				if p.Var != nil && p.Var.Name == name {
					if !p.ByRef && deadParam(ctx, fn, n, p) || outerBinds(n, name) { // E7, E8
						break
					}
					ctx.Report(kw, "Loop variable '$"+name+"' overwrites a "+kind+" parameter.")
					break
				}
			}
		}
	}
	var outer []string
	for p := n.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() { // D4
		switch p.(type) {
		case *syntax.For, *syntax.Foreach:
		default:
			continue
		}
		outer = loopVarNames(p, outer[:0])
		for _, name := range names {
			for _, o := range outer {
				if o == name && !deadOuter(ctx, n, p, name) {
					ctx.Report(kw, "Loop variable '$"+name+suspiciousLoopOuterMsg)
					break
				}
			}
		}
	}
}

// loopVarNames appends the distinct loop variable names of a for/foreach
// loop to dst, in source order.
func loopVarNames(n syntax.Node, dst []string) []string {
	add := func(v *syntax.Variable) {
		if v == nil || v.NameExpr != nil || v.Name == "" {
			return
		}
		for _, x := range dst {
			if x == v.Name {
				return
			}
		}
		dst = append(dst, v.Name)
	}
	var collect func(e syntax.Expr)
	collect = func(e syntax.Expr) {
		switch e := e.(type) {
		case *syntax.Variable:
			add(e)
		case *syntax.Array:
			for _, it := range e.Items {
				if it != nil && it.Value != nil {
					collect(it.Value)
				}
			}
		case *syntax.List:
			for _, it := range e.Items {
				if it != nil && it.Value != nil {
					collect(it.Value)
				}
			}
		}
	}
	switch l := n.(type) {
	case *syntax.For:
		for _, e := range l.Init {
			if a, ok := e.(*syntax.Assign); ok {
				if v, ok := a.Var.(*syntax.Variable); ok && a.Op.Kind == syntax.TEqual && !readsVar(a.Value, v) {
					add(v) // custos: `$i += $step` or `$i = $i + 1` goes on from the value given
				}
			}
		}
	case *syntax.Foreach:
		if l.Key != nil {
			collect(l.Key)
		}
		if l.Value != nil {
			collect(l.Value)
		}
	}
	return dst
}

// readsVar reports whether e reads variable v by name.
func readsVar(e syntax.Expr, v *syntax.Variable) bool {
	found := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		if u, ok := n.(*syntax.Variable); ok && u.NameExpr == nil && u.Name == v.Name {
			found = true
		}
		return !found
	})
	return found
}

// deadOuter reports whether the outer foreach's variable name is dead when
// the inner loop n overwrites it (E6): neither loop binds by reference,
// nothing reads the name after n within p, nor anywhere in a loop between
// them (a loop condition reads it again), nor after p (or before it inside
// a loop around it). The next outer iteration assigns the variable again.
func deadOuter(ctx *analysis.Context, n, p syntax.Node, name string) bool {
	fe, ok := p.(*syntax.Foreach)
	if !ok || fe.ByRef {
		return false
	}
	if in, ok := n.(*syntax.Foreach); ok && in.ByRef {
		return false
	}
	// Regions where a mention would see the inner loop's value: what runs
	// after n until the outer iteration ends (followers), and loops or try
	// statements between them, which come back round n.
	ns, ps := n.Span(), p.Span()
	regions := followers(n, p)
	for a := n.Parent(); a != p; a = a.Parent() {
		switch a.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			regions = append(regions, syntax.Span{Start: a.Span().Start, End: ns.Start})
		case *syntax.Try:
			regions = append(regions, syntax.Span{Start: a.Span().Start, End: ns.Start}, syntax.Span{Start: ns.End, End: a.Span().End})
		}
	}
	var scope syntax.Node = ctx.File
	after := syntax.Span{Start: ps.Start, End: ^uint32(0)}
	for a := p.Parent(); a != nil; a = a.Parent() {
		if syntax.IsFuncLike(a) {
			scope = a
			break
		}
		switch a.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			after.Start = a.Span().Start
		}
	}
	regions = append(regions, syntax.Span{Start: after.Start, End: ps.Start}, syntax.Span{Start: ps.End, End: after.End})
	read := false
	syntax.Inspect(scope, func(x syntax.Node) bool {
		v, ok := x.(*syntax.Variable)
		if !ok || v.NameExpr == nil && v.Name != name || overwrites(v) {
			return true
		}
		for _, r := range regions {
			if v.Span().Start >= r.Start && v.Span().Start < r.End {
				read = true
			}
		}
		return !read
	})
	return !read
}

// followers returns the spans of the statements that can run after n within
// the current iteration of the enclosing loop p: the statements following n
// and each of its ancestors in their statement lists, up to one that leaves
// the iteration (return, throw, exit, or a break or continue reaching p).
func followers(n, p syntax.Node) []syntax.Span {
	var out []syntax.Span
	for cur := n; cur != p; cur = cur.Parent() {
		var list []syntax.Stmt
		switch par := cur.Parent().(type) {
		case *syntax.Block:
			list = par.Stmts
		case *syntax.Case:
			list = par.Stmts
		}
		i := slices.IndexFunc(list, func(x syntax.Stmt) bool { return syntax.Node(x) == cur })
		for _, next := range list[i+1:] {
			out = append(out, next.Span())
			if terminates(next) || leavesLoop(next, p) {
				return out
			}
		}
	}
	return out
}

// leavesLoop reports a break or continue that reaches loop p (or beyond).
func leavesLoop(st syntax.Stmt, p syntax.Node) bool {
	var num syntax.Expr
	switch st := st.(type) {
	case *syntax.Break:
		num = st.Num
	case *syntax.Continue:
		num = st.Num
	default:
		return false
	}
	level := int64(1)
	if lit, ok := num.(*syntax.Literal); ok {
		if v, ok := astquery.ParseIntLiteral(lit.Raw); ok {
			level = v
		}
	}
	// st lies inside p, so the walk reaches it.
	depth := int64(0)
	for a := st.Parent(); ; a = a.Parent() {
		switch a.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile, *syntax.Switch:
			depth++
		}
		if a == p {
			return level >= depth
		}
	}
}

// overwrites reports whether v only receives a new value: a foreach key or
// value, the target of a plain assignment, or a parameter.
func overwrites(v *syntax.Variable) bool {
	switch p := v.Parent().(type) {
	case *syntax.Foreach:
		return p.Key == syntax.Expr(v) || p.Value == syntax.Expr(v)
	case *syntax.Assign:
		return p.Var == syntax.Expr(v) && p.Op.Kind == syntax.TEqual
	case *syntax.Param:
		return true
	}
	return false
}

// deadParam reports whether the parameter name is dead when loop n
// overwrites it (E7): its value was read before (or as the subject of n),
// n does not bind by reference, every statement after
// it up to a return or throw (or the end of fn) leaves the name alone, no
// loop, switch or try around n mentions it elsewhere, and fn reads no
// variable by name.
func deadParam(ctx *analysis.Context, fn, n syntax.Node, param *syntax.Param) bool {
	name := param.Var.Name
	var subject syntax.Node
	if fe, ok := n.(*syntax.Foreach); ok {
		if fe.ByRef {
			return false
		}
		subject = fe.Expr
	}
	if usesByName(ctx, fn, subject) {
		return false
	}
	mentions := func(x, skip syntax.Node) bool {
		found := false
		syntax.Inspect(x, func(y syntax.Node) bool {
			if v, ok := y.(*syntax.Variable); ok && v.Name == name && !overwrites(v) {
				found = true
			}
			return y != skip && !found
		})
		return found
	}
	// The value given must have been used: stored by constructor promotion,
	// read before the loop or in the foreach subject (by name too, as
	// `func_get_args()`); a parameter never read is lost, not dead.
	used := len(param.Modifiers) > 0 || subject != nil && usesByName(ctx, subject, nil)
	syntax.Inspect(fn, func(y syntax.Node) bool {
		if v, ok := y.(*syntax.Variable); ok && v.Name == name && !overwrites(v) {
			fe, isForeach := n.(*syntax.Foreach)
			used = v.Span().Start < n.Span().Start || isForeach && v.Span().Start < fe.Expr.Span().End
		}
		return !used
	})
	if !used {
		return false
	}
	for cur := n; cur != fn; cur = cur.Parent() {
		switch p := cur.Parent().(type) {
		case *syntax.Block:
			for _, st := range p.Stmts[slices.Index(p.Stmts, cur.(syntax.Stmt))+1:] {
				if mentions(st, nil) {
					return false
				}
				if terminates(st) {
					return true
				}
			}
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile, *syntax.Try, *syntax.Case:
			if mentions(p, cur) {
				return false
			}
		}
	}
	return true
}

// outerBinds reports whether a loop around n, in the same function, already
// binds name: the parameter was overwritten there, and the clash is the
// outer loop's (D4, E6), not the parameter's (E8).
func outerBinds(n syntax.Node, name string) bool {
	for p := n.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach:
			if slices.Contains(loopVarNames(p, nil), name) {
				return true
			}
		}
	}
	return false
}

// usesByName reports whether fn, skip aside, may read a variable by its
// name: a variable-variable or a call such as compact() or func_get_args().
func usesByName(ctx *analysis.Context, fn, skip syntax.Node) bool {
	found := false
	syntax.Inspect(fn, func(y syntax.Node) bool {
		if y == skip {
			return false
		}
		switch y := y.(type) {
		case *syntax.Variable:
			found = y.NameExpr != nil
		case *syntax.FuncCall:
			switch ctx.GlobalFunctionName(y) {
			case "func_get_args", "func_get_arg", "compact", "get_defined_vars":
				found = true
			}
		}
		return !found
	})
	return found
}

// terminates reports a statement that leaves the function: a return, or a
// throw or exit statement.
func terminates(st syntax.Stmt) bool {
	switch st := st.(type) {
	case *syntax.Return:
		return true
	case *syntax.ExprStmt:
		switch st.Expr.(type) {
		case *syntax.Throw, *syntax.Exit:
			return true
		}
	}
	return false
}
