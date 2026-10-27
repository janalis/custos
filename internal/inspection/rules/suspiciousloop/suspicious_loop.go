package suspiciousloop

import (
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
// the inner loop n overwrites it (E6): n ends the outer body, neither loop
// binds by reference, and nothing reads the name after the outer loop (nor
// elsewhere in a loop around it), so the next iteration or the end of the
// loop loses nothing.
func deadOuter(ctx *analysis.Context, n, p syntax.Node, name string) bool {
	fe, ok := p.(*syntax.Foreach)
	if !ok || fe.ByRef {
		return false
	}
	if in, ok := n.(*syntax.Foreach); ok && in.ByRef {
		return false
	}
	last := fe.Body
	if b, ok := last.(*syntax.Block); ok && len(b.Stmts) > 0 {
		last = b.Stmts[len(b.Stmts)-1]
	}
	if last != n {
		return false
	}
	// Reads from the start of the outermost loop around p count: a later
	// iteration of that loop may come back to them.
	var scope syntax.Node = ctx.File
	from := p.Span()
	for a := p.Parent(); a != nil; a = a.Parent() {
		if syntax.IsFuncLike(a) {
			scope = a
			break
		}
		switch a.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			from.Start = a.Span().Start
		}
	}
	read := false
	syntax.Inspect(scope, func(x syntax.Node) bool {
		s := x.Span()
		if s.Start >= from.Start && s.Start < p.Span().Start || s.Start >= from.End {
			if v, ok := x.(*syntax.Variable); ok && (v.NameExpr != nil || v.Name == name) {
				read = true
			}
		}
		return !read
	})
	return !read
}
