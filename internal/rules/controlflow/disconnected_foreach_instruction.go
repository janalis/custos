package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// disconnectedForeachInstruction reports foreach body statements that do not
// depend on the loop, and (optionally) per-iteration object creation.
type disconnectedForeachInstruction struct{}

func init() { register(disconnectedForeachInstruction{}) }

const (
	disconnectedMsg   = "Statement does not depend on the loop; move it out."
	disconnectedClone = "Create the object once before the loop and clone it here."
)

func (disconnectedForeachInstruction) ID() string { return "DisconnectedForeachInstruction" }

func (disconnectedForeachInstruction) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach}
}

type dfiClass uint8

const (
	dfiOther dfiClass = iota
	dfiControl
	dfiIncDec
	dfiNew
	dfiClone
	dfiDomCreate
	dfiAssign
	dfiAccumulate
)

func (disconnectedForeachInstruction) Check(ctx *analysis.Context, n syntax.Node) {
	loop := n.(*syntax.Foreach)
	if loop.Span().Len() == 0 {
		return
	}
	body, ok := bracedBlock(ctx, loop.Body) // D1
	if !ok {
		return
	}
	var stmts []syntax.Stmt // D2/D3
	for _, s := range body.Stmts {
		if _, html := s.(*syntax.InlineHTML); html {
			return
		}
		if s.Span().Len() > 0 {
			stmts = append(stmts, s)
		}
	}
	if len(stmts) == 0 {
		return
	}
	modified := map[string]bool{} // D4
	for p := syntax.Node(loop); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		fe, ok := p.(*syntax.Foreach)
		if !ok {
			continue
		}
		for _, e := range []syntax.Expr{fe.Key, fe.Value} {
			if e == nil {
				continue
			}
			syntax.Inspect(e, func(x syntax.Node) bool {
				if v, ok := x.(*syntax.Variable); ok && v.Name != "" {
					modified[v.Name] = true
				}
				return true
			})
		}
	}
	deps := make([]map[string]bool, len(stmts)) // pass 1
	bound := map[string]map[int]bool{}          // D6a: names bound by catch / inner foreach headers, per statement
	for i, s := range stmts {
		deps[i] = dfiCollect(ctx, s, modified, func(name string) {
			if bound[name] == nil {
				bound[name] = map[int]bool{}
			}
			bound[name][i] = true
		})
	}
	suggestClone := ctx.Bool("SUGGEST_USING_CLONE")
	for i, s := range stmts { // pass 2
		connected := false
		for name := range deps[i] {
			if modified[name] || dfiBoundElsewhere(bound[name], i) {
				connected = true
				break
			}
		}
		if connected {
			continue
		}
		switch dfiClassify(ctx, s) {
		case dfiOther:
			if dfiDisconnected(s) {
				ctx.ReportSeverity(dfiRange(ctx, s), meta.SeverityInfo, disconnectedMsg)
			}
		case dfiNew, dfiDomCreate:
			if suggestClone {
				ctx.ReportSeverity(s.Span(), meta.SeverityInfo, disconnectedClone)
			}
		}
	}
}

// dfiParent returns the parent of n, seeing through call argument wrappers.
func dfiParent(n syntax.Node) syntax.Node {
	p := n.Parent()
	if a, ok := p.(*syntax.Arg); ok {
		return a.Parent()
	}
	return p
}

// dfiBoundElsewhere reports whether a statement other than i binds the name
// (catch variable or inner foreach header).
func dfiBoundElsewhere(binders map[int]bool, i int) bool {
	for j := range binders {
		if j != i {
			return true
		}
	}
	return false
}

// dfiCollect runs pass 1 on statement s: returns Dep(s), adds writes to m and
// reports names bound by catch clauses / inner foreach headers to bind.
func dfiCollect(ctx *analysis.Context, s syntax.Stmt, m map[string]bool, bind func(string)) map[string]bool {
	dep := map[string]bool{}
	syntax.Inspect(s, func(x syntax.Node) bool {
		if call, ok := x.(*syntax.FuncCall); ok { // D11
			if ctx.IsGlobalFunctionCall(call, "compact") {
				if args, ok := util.CallArgValues(call); ok {
					for _, a := range args {
						if c, _, ok := util.QuotedStringRaw(a); ok && c != "" {
							dep[c] = true
						}
					}
				}
			}
		}
		v, ok := x.(*syntax.Variable)
		if !ok || v.Name == "" || util.IsStaticPropName(v) {
			return true
		}
		dfiVariable(ctx, v, m, dep, bind)
		return true
	})
	return dep
}

func dfiVariable(ctx *analysis.Context, v *syntax.Variable, m, dep map[string]bool, bind func(string)) {
	name := v.Name
	var c syntax.Node = v
	for {
		pf, ok := c.Parent().(*syntax.PropertyFetch)
		if !ok || pf.Var != c {
			break
		}
		c = pf
	}
	p := dfiParent(c)
	if item, ok := p.(*syntax.ArrayItem); ok {
		if lit := item.Parent(); lit != nil {
			p = lit
			if pp := dfiParent(lit); pp != nil {
				p = pp
			}
		}
	}
	// D5
	if a, ok := p.(*syntax.Assign); ok {
		switch syntax.UnwrapParens(a.Var).(type) {
		case *syntax.List, *syntax.Array:
			if a.Value != syntax.Expr(v) {
				m[name], dep[name] = true, true
				return
			}
		}
		// D6
		if a.Var == c {
			m[name] = true
			_, isProp := c.(*syntax.PropertyFetch)
			_, inArg := a.Parent().(*syntax.Arg)
			if a.Op.Kind != syntax.TEqual || isProp || inArg {
				dep[name] = true
			}
			return
		}
	}
	// D6a: catch variables and inner foreach key/value targets are bindings.
	switch x := p.(type) {
	case *syntax.Catch:
		if x.Var == v {
			bind(name)
			return
		}
	case *syntax.Foreach:
		if x.Key != nil && x.Key.Span().Contains(v.Span()) || x.Value != nil && x.Value.Span().Contains(v.Span()) {
			bind(name)
			return
		}
	}
	// D7: only element writes modify the array
	if d, ok := p.(*syntax.ArrayDimFetch); ok && d.Var == c {
		dep[name] = true
		if dfiElementWrite(ctx, d) {
			m[name] = true
		}
	}
	// D8
	if list, ok := p.(*syntax.ArgList); ok && c == syntax.Node(v) {
		switch call := list.Parent().(type) {
		case *syntax.MethodCall:
			if !call.NullSafe {
				obj := call.Var
				for {
					switch o := obj.(type) {
					case *syntax.Paren:
						obj = o.Expr
						continue
					case *syntax.PropertyFetch:
						obj = o.Var
						continue
					case *syntax.ArrayDimFetch:
						obj = o.Var
						continue
					}
					break
				}
				if ov, ok := obj.(*syntax.Variable); ok && ov.Name != "" {
					m[ov.Name] = true
					return
				}
			}
		case *syntax.FuncCall:
			if f := ctx.Types().ResolveFunction(call); f != nil && len(f.Params) > 0 {
				idx := -1
				for i, a := range list.Args {
					if arg, ok := a.(*syntax.Arg); ok && arg.Value == syntax.Expr(v) {
						idx = i
						break
					}
				}
				if idx >= 0 {
					byRef := false
					if idx < len(f.Params) {
						byRef = f.Params[idx].ByRef
					} else if last := f.Params[len(f.Params)-1]; last.Variadic {
						byRef = last.ByRef
					}
					if byRef {
						m[name], dep[name] = true, true
						return
					}
				}
			}
		}
	}
	// D9
	if _, ok := p.(*syntax.IncDec); ok {
		m[name], dep[name] = true, true
		return
	}
	dep[name] = true // D10
}

func dfiClassify(ctx *analysis.Context, s syntax.Stmt) dfiClass {
	switch st := s.(type) {
	case *syntax.Break, *syntax.Continue, *syntax.Return:
		return dfiControl
	case *syntax.ExprStmt:
		switch e := st.Expr.(type) {
		case *syntax.IncDec:
			return dfiIncDec
		case *syntax.Assign:
			if d, ok := e.Var.(*syntax.ArrayDimFetch); ok && d.Dim == nil {
				return dfiAccumulate
			}
			if _, ok := ruleSimpleVar(e.Var); !ok {
				return dfiOther
			}
			if e.Op.Kind == syntax.TEqual {
				switch val := e.Value.(type) {
				case *syntax.New:
					return dfiNew
				case *syntax.Clone:
					return dfiClone
				case *syntax.MethodCall:
					if dfiIsDomCreate(ctx, val) {
						return dfiDomCreate
					}
				}
			}
			return dfiAssign
		}
	}
	return dfiOther
}

func dfiIsDomCreate(ctx *analysis.Context, call *syntax.MethodCall) bool {
	id, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "createElement") {
		return false
	}
	for _, cls := range ctx.TypeOf(call.Var).Classes() {
		if m := ctx.Index().FindMethod(cls, id.Value, ctx.PHP); m != nil &&
			strings.EqualFold(strings.TrimPrefix(m.Class, `\`), "DOMDocument") {
			return true
		}
	}
	return false
}

// dfiDisconnected checks the D12 shape conditions of an "other" statement.
func dfiDisconnected(s syntax.Stmt) bool {
	if e, ok := s.(*syntax.Echo); ok && e.Short {
		return false
	}
	exits, hasVar := false, false
	syntax.Inspect(s, func(x syntax.Node) bool {
		if x == syntax.Node(s) {
			return true
		}
		switch v := x.(type) {
		case *syntax.Break, *syntax.Continue, *syntax.Return, *syntax.Throw:
			exits = true
			return false
		case *syntax.Variable:
			if !util.IsStaticPropName(v) {
				hasVar = true
			}
		}
		return true
	})
	return !exits && hasVar
}

func dfiRange(ctx *analysis.Context, s syntax.Stmt) syntax.Span {
	switch s.(type) {
	case *syntax.If, *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach, *syntax.Switch, *syntax.Try:
		if t, ok := util.NextSignificant(ctx.File, s.Span().Start); ok {
			return syntax.Span{Start: t.Start, End: t.End}
		}
	}
	return s.Span()
}

// dfiElementWrite reports whether the element access d (outermost of its
// `[…]`/`->` chain) is written: assigned (any operator, or bound by
// reference), incremented/decremented, unset, used as the receiver of a
// method call, or passed to a by-reference parameter of a resolved function.
func dfiElementWrite(ctx *analysis.Context, d *syntax.ArrayDimFetch) bool {
	var o syntax.Node = d
	for {
		switch q := o.Parent().(type) {
		case *syntax.ArrayDimFetch:
			if q.Var == o {
				o = q
				continue
			}
		case *syntax.PropertyFetch:
			if q.Var == o {
				o = q
				continue
			}
		}
		break
	}
	switch q := dfiParent(o).(type) {
	case *syntax.Assign:
		return q.Var == o || q.ByRef && q.Value == o
	case *syntax.IncDec, *syntax.Unset:
		return true
	case *syntax.MethodCall:
		return q.Var == o
	case *syntax.ArgList:
		call, ok := q.Parent().(*syntax.FuncCall)
		if !ok {
			return false
		}
		f := ctx.Types().ResolveFunction(call)
		if f == nil || len(f.Params) == 0 {
			return false
		}
		for i, a := range q.Args {
			if arg, ok := a.(*syntax.Arg); ok && arg.Value == o {
				if i < len(f.Params) {
					return f.Params[i].ByRef
				}
				last := f.Params[len(f.Params)-1]
				return last.Variadic && last.ByRef
			}
		}
	}
	return false
}
