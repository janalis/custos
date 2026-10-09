package suspiciousloop

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

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
		if len(f.Cond) >= 2 { // D1
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
				if o == name {
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
				if v, ok := a.Var.(*syntax.Variable); ok {
					add(v)
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
