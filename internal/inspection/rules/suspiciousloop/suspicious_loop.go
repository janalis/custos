package suspiciousloop

import (
	"slices"

	"custos/internal/inspection/analysis"
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
					if !p.ByRef && deadParam(ctx, fn, n, name) { // E7
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
	// Regions where a mention would see the inner loop's value.
	ns, ps := n.Span(), p.Span()
	regions := []syntax.Span{{Start: ns.End, End: ps.End}}
	for a := n.Parent(); a != p; a = a.Parent() {
		switch a.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			regions = append(regions, syntax.Span{Start: a.Span().Start, End: ns.Start})
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
func deadParam(ctx *analysis.Context, fn, n syntax.Node, name string) bool {
	if fe, ok := n.(*syntax.Foreach); ok && fe.ByRef || usesByName(ctx, fn) {
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
	// The value given must have been used: read before the loop or as the
	// foreach subject; a parameter never read is lost, not dead.
	used := false
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

// usesByName reports whether fn may read a variable by its name: a
// variable-variable or a call such as compact() or func_get_args().
func usesByName(ctx *analysis.Context, fn syntax.Node) bool {
	found := false
	syntax.Inspect(fn, func(y syntax.Node) bool {
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
