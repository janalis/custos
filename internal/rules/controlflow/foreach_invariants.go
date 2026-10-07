package controlflow

import (
	"sort"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// foreachInvariants reports counter `for` loops over `count($a)` and
// `while (list(..) = each($a))` loops that should be foreach loops.
type foreachInvariants struct{}

func init() { register(foreachInvariants{}) }

const (
	foreachInvCounterMsg = "Iterate with foreach instead of a counter loop."
	foreachInvEachMsg    = "Replace the each() loop with foreach."
)

func (foreachInvariants) ID() string { return "ForeachInvariants" }

func (foreachInvariants) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFor, syntax.KAssign}
}

func (r foreachInvariants) Check(ctx *analysis.Context, n syntax.Node) {
	if n.Span().Len() == 0 {
		return
	}
	switch x := n.(type) {
	case *syntax.For:
		r.counterLoop(ctx, x)
	case *syntax.Assign:
		r.eachLoop(ctx, x)
	}
}

// foreachInvBody returns the loop body as a braced block holding at
// least one statement.
func foreachInvBody(ctx *analysis.Context, s syntax.Stmt) (*syntax.Block, bool) {
	b, ok := bracedBlock(ctx, s)
	if !ok {
		return nil, false
	}
	for _, st := range b.Stmts {
		if st.Span().Len() > 0 {
			return b, true
		}
	}
	return nil, false
}

func ruleSimpleVar(e syntax.Expr) (*syntax.Variable, bool) {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name == "" {
		return nil, false
	}
	return v, true
}

// ---- Part A ---------------------------------------------------------------------------

func (foreachInvariants) counterLoop(ctx *analysis.Context, l *syntax.For) {
	if len(l.Loop) != 1 || len(l.Cond) != 1 { // D1, D5
		return
	}
	body, ok := foreachInvBody(ctx, l.Body) // D2
	if !ok {
		return
	}
	var counter *syntax.Variable // D3
	for _, in := range l.Init {
		a, ok := in.(*syntax.Assign)
		if !ok || a.Op.Kind != syntax.TEqual || a.ByRef {
			continue
		}
		v, ok := ruleSimpleVar(a.Var)
		if !ok {
			continue
		}
		if ctx.Text(a.Value) == "0" {
			counter = v
			break
		}
	}
	if counter == nil {
		return
	}
	step, ok := l.Loop[0].(*syntax.IncDec) // D4
	if !ok || step.Op.Kind != syntax.TInc {
		return
	}
	if sv, ok := ruleSimpleVar(step.Var); !ok || sv.Name != counter.Name {
		return
	}
	cond, ok := l.Cond[0].(*syntax.Binary) // D5
	if !ok {
		return
	}
	var limit syntax.Expr
	if v, ok := ruleSimpleVar(cond.Left); ok && v.Name == counter.Name {
		limit = cond.Right
	} else if v, ok := ruleSimpleVar(cond.Right); ok && v.Name == counter.Name {
		limit = cond.Left
	} else {
		return
	}
	// D6
	var container syntax.Expr
	distinct := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if distinct {
			return false
		}
		d, ok := x.(*syntax.ArrayDimFetch)
		if !ok || d.Dim == nil {
			return true
		}
		if v, ok := ruleSimpleVar(d.Dim); ok && v.Name == counter.Name {
			if container == nil {
				container = d.Var
			} else if !util.EquivalentFoldNames(ctx.File, container, d.Var) {
				distinct = true
			}
		}
		return true
	})
	if container == nil || distinct {
		return
	}
	// D7/D8
	vals := foreachInvLimitValues(ctx, l, limit)
	if len(vals) != 1 {
		return
	}
	cnt, ok := vals[0].(*syntax.FuncCall)
	if !ok || !ctx.IsGlobalFunctionCall(cnt, "count") {
		return
	}
	if _, isName := cnt.Name.(*syntax.Name); !isName {
		return
	}
	args, ok := util.CallArgValues(cnt)
	if !ok || len(args) != 1 || !util.EquivalentFoldNames(ctx.File, args[0], container) {
		return
	}
	// D8b: write guard on the counter and the limit.
	var counterInit syntax.Node
	for _, in := range l.Init {
		if a, ok := in.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
			if v, ok := ruleSimpleVar(a.Var); ok && v == counter {
				counterInit = a
			}
		}
	}
	skip := map[syntax.Node]bool{counterInit: true}
	isCounter := func(e syntax.Expr) bool {
		v, ok := ruleSimpleVar(e)
		return ok && v.Name == counter.Name
	}
	if foreachInvWritten(ctx, l, isCounter, skip) {
		return
	}
	lim := util.UnwrapParens(limit)
	var isLimit func(syntax.Expr) bool
	if lv, ok := ruleSimpleVar(lim); ok {
		isLimit = func(e syntax.Expr) bool {
			v, ok := ruleSimpleVar(e)
			return ok && v.Name == lv.Name
		}
	} else if _, ok := lim.(*syntax.PropertyFetch); ok {
		isLimit = func(e syntax.Expr) bool {
			_, ok := e.(*syntax.PropertyFetch)
			return ok && util.EquivalentFoldNames(ctx.File, e, lim)
		}
	}
	if isLimit != nil {
		lskip := map[syntax.Node]bool{}
		for _, in := range l.Init {
			if a, ok := in.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef && isLimit(a.Var) {
				lskip[a] = true
				break
			}
		}
		if foreachInvWritten(ctx, l, isLimit, lskip) {
			return
		}
	}
	kw, ok := util.NextSignificant(ctx.File, l.Span().Start)
	if !ok {
		return
	}
	ctx.Report(syntax.Span{Start: kw.Start, End: kw.End}, foreachInvCounterMsg, analysis.Fix{
		Title: "Convert to foreach",
		Edits: func() []analysis.TextEdit {
			return counterLoopFix(ctx, l, body, counter, container, limit)
		},
	})
}

func counterLoopFix(ctx *analysis.Context, l *syntax.For, body *syntax.Block, counter *syntax.Variable,
	container, limit syntax.Expr) []analysis.TextEdit {
	cName := "$" + counter.Name
	valName := cName + "Value"
	// F1.1: qualifying accesses.
	var repl []syntax.Span
	var replaced []syntax.Node
	syntax.Inspect(body, func(x syntax.Node) bool {
		d, ok := x.(*syntax.ArrayDimFetch)
		if !ok || d.Dim == nil || d.Span().Len() == 0 {
			return true
		}
		if !util.EquivalentFoldNames(ctx.File, d.Var, container) || !util.EquivalentFoldNames(ctx.File, d.Dim, counter) {
			return true
		}
		if !accessReplaceable(ctx, d) {
			return true
		}
		sp := d.Span()
		if _, inStr := d.Parent().(*syntax.InterpolatedString); inStr &&
			sp.Start > 0 && ctx.Src[sp.Start-1] == '{' && int(sp.End) < len(ctx.Src) && ctx.Src[sp.End] == '}' {
			sp = syntax.Span{Start: sp.Start - 1, End: sp.End + 1}
		}
		repl = append(repl, sp)
		replaced = append(replaced, d)
		return false
	})
	sort.Slice(repl, func(i, j int) bool { return repl[i].Start < repl[j].Start })
	bs := body.Span()
	var b strings.Builder
	pos := bs.Start
	for _, sp := range repl {
		b.Write(ctx.Src[pos:sp.Start])
		b.WriteString(valName)
		pos = sp.End
	}
	b.Write(ctx.Src[pos:bs.End])
	// F1.3: does the counter remain in the body?
	remains := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if remains {
			return false
		}
		for _, r := range replaced {
			if x == r {
				return false
			}
		}
		if v, ok := ruleSimpleVar(ruleAsExpr(x)); ok && v.Name == counter.Name {
			remains = true
		}
		return true
	})
	header := "foreach (" + ctx.Text(container) + " as "
	if remains {
		header += cName + " => "
	}
	header += valName + ") "
	edits := []analysis.TextEdit{{Span: l.Span(), NewText: header + b.String()}}
	// F1.4: limit cleanup.
	if _, direct := util.UnwrapParens(limit).(*syntax.FuncCall); !direct {
		if del, ok := limitAssignment(ctx, l, limit); ok {
			edits = append(edits, analysis.TextEdit{Span: util.WithLeadingWhitespace(ctx.File, del.Span())})
		}
	}
	return edits
}

func ruleAsExpr(n syntax.Node) syntax.Expr {
	e, _ := n.(syntax.Expr)
	return e
}

// foreachInvLimitValues implements D8 for the limit X: a plain variable takes
// only the assignments that reach the loop condition (its own init
// assignment when present, which dominates the condition).
func foreachInvLimitValues(ctx *analysis.Context, l *syntax.For, limit syntax.Expr) []syntax.Expr {
	v, ok := ruleSimpleVar(limit)
	if !ok {
		return util.PossibleValues(ctx.File, limit)
	}
	scope := util.EnclosingFuncLike(l)
	body := util.FuncLikeBody(scope)
	if scope == nil || body == nil || util.UnstableVariableIn(ctx.File, body, v.Name) {
		return nil
	}
	innermost := func(e syntax.Expr) syntax.Expr {
		for {
			a, ok := util.UnwrapParens(e).(*syntax.Assign)
			if !ok || a.Op.Kind != syntax.TEqual || a.ByRef {
				return e
			}
			e = a.Value
		}
	}
	for i := len(l.Init) - 1; i >= 0; i-- {
		if a, ok := l.Init[i].(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
			if t, ok := ruleSimpleVar(a.Var); ok && t.Name == v.Name {
				return util.PossibleValues(ctx.File, innermost(a.Value))
			}
		}
	}
	var out []syntax.Expr
	defs, entry := util.ReachingAssignmentsIn(ctx.File, scope, v, v.Name)
	if entry {
		for _, p := range util.FuncLikeParams(scope) {
			if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
				out = append(out, util.PossibleValues(ctx.File, p.Default)...)
			}
		}
	}
	for _, d := range defs {
		out = append(out, util.PossibleValues(ctx.File, innermost(d.Value))...)
	}
	return out
}

// limitAssignment returns the statement `$n = count(..);` to delete when the
// limit's only remaining occurrence in the enclosing function is its
// assignment.
func limitAssignment(ctx *analysis.Context, l *syntax.For, limit syntax.Expr) (syntax.Stmt, bool) {
	fn := util.EnclosingFuncLike(l)
	if fn == nil {
		return nil, false
	}
	fbody := util.FuncLikeBody(fn)
	if fbody == nil {
		return nil, false
	}
	header := syntax.Span{Start: l.Span().Start, End: l.Body.Span().Start}
	var found []syntax.Node
	syntax.Inspect(fbody, func(x syntax.Node) bool {
		if x.Span().Len() == 0 || header.Contains(x.Span()) {
			return true
		}
		if util.EquivalentFoldNames(ctx.File, x, limit) {
			found = append(found, x)
		}
		return true
	})
	if len(found) != 1 {
		return nil, false
	}
	a, ok := found[0].Parent().(*syntax.Assign)
	if !ok || a.Var != found[0] || a.Op.Kind != syntax.TEqual {
		return nil, false
	}
	st, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return nil, false
	}
	return st, true
}

// accessReplaceable implements the parent contexts of F1.1.
func accessReplaceable(ctx *analysis.Context, d *syntax.ArrayDimFetch) bool {
	switch p := d.Parent().(type) {
	case *syntax.PropertyFetch:
		return p.Var == syntax.Expr(d)
	case *syntax.MethodCall:
		return p.Var == syntax.Expr(d)
	case *syntax.Binary, *syntax.Unary, *syntax.Paren, *syntax.Echo, *syntax.InterpolatedString:
		return true
	case *syntax.Instanceof:
		return p.Expr == syntax.Expr(d)
	case *syntax.ArrayDimFetch:
		if p.Dim == syntax.Expr(d) {
			return true
		}
		// base of a further access: kept when the chain's outermost access
		// sits directly in an assignment.
		outer := syntax.Node(p)
		for {
			pp, ok := outer.Parent().(*syntax.ArrayDimFetch)
			if !ok || pp.Var != outer {
				break
			}
			outer = pp
		}
		_, inAssign := outer.Parent().(*syntax.Assign)
		return !inAssign
	case *syntax.If:
		return p.Cond == syntax.Expr(d)
	case *syntax.ElseIf:
		return p.Cond == syntax.Expr(d)
	case *syntax.Switch:
		return p.Cond == syntax.Expr(d)
	case *syntax.Case:
		return p.Cond == syntax.Expr(d)
	case *syntax.For:
		return true
	case *syntax.While:
		return p.Cond == syntax.Expr(d)
	case *syntax.DoWhile:
		return p.Cond == syntax.Expr(d)
	case *syntax.Foreach:
		return p.Expr == syntax.Expr(d)
	case *syntax.Assign:
		return p.Value == syntax.Expr(d) && !p.ByRef
	case *syntax.Arg:
		if p.ByRef || p.Unpack {
			return false
		}
		list, ok := p.Parent().(*syntax.ArgList)
		if !ok {
			return false
		}
		return callHasNoByRefParams(ctx, list.Parent())
	}
	return false
}

func callHasNoByRefParams(ctx *analysis.Context, call syntax.Node) bool {
	params, ok := foreachInvCallParams(ctx, call)
	if !ok {
		return false
	}
	for _, p := range params {
		if p.ByRef {
			return false
		}
	}
	return true
}

// foreachInvCallParams returns the parameters of the function or method a
// call resolves to (user code or stubs).
func foreachInvCallParams(ctx *analysis.Context, call syntax.Node) ([]index.Param, bool) {
	switch c := call.(type) {
	case *syntax.FuncCall:
		f := ctx.Types().ResolveFunction(c)
		if f == nil {
			return nil, false
		}
		return f.Params, true
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return nil, false
		}
		cls := ctx.TypeOf(c.Var).Classes()
		if len(cls) != 1 {
			return nil, false
		}
		m := ctx.Index().FindMethod(cls[0], id.Value, ctx.PHP)
		if m == nil {
			return nil, false
		}
		return m.Params, true
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		name, ok2 := c.Class.(*syntax.Name)
		if !ok || !ok2 {
			return nil, false
		}
		fqn := ctx.Names().Class(name.Value, name.Span().Start)
		m := ctx.Index().FindMethod(fqn, id.Value, ctx.PHP)
		if m == nil {
			return nil, false
		}
		return m.Params, true
	}
	return nil, false
}

// foreachInvByRefArg reports whether arg (at position pos of its call)
// binds to a by-reference parameter of a resolved function/method.
func foreachInvByRefArg(ctx *analysis.Context, arg *syntax.Arg, pos int) bool {
	if arg.ByRef {
		return true
	}
	list, ok := arg.Parent().(*syntax.ArgList)
	if !ok {
		return false
	}
	params, ok := foreachInvCallParams(ctx, list.Parent())
	if !ok {
		return false
	}
	if arg.Name != nil {
		for _, p := range params {
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
				return p.ByRef
			}
		}
		return false
	}
	if pos < len(params) {
		return params[pos].ByRef
	}
	if n := len(params); n > 0 && params[n-1].Variadic {
		return params[n-1].ByRef
	}
	return false
}

// foreachInvWritten implements the write guard D8b: it reports whether a
// node matching target is written anywhere in the loop's init, condition
// or body, other than by the nodes in skip.
func foreachInvWritten(ctx *analysis.Context, l *syntax.For, match func(syntax.Expr) bool, skip map[syntax.Node]bool) bool {
	found := false
	var targetHit func(e syntax.Expr) bool
	targetHit = func(e syntax.Expr) bool {
		switch t := util.UnwrapParens(e).(type) {
		case *syntax.List:
			for _, it := range t.Items {
				if it != nil && it.Value != nil && targetHit(it.Value) {
					return true
				}
			}
			return false
		case *syntax.Array:
			for _, it := range t.Items {
				if it != nil && it.Value != nil && targetHit(it.Value) {
					return true
				}
			}
			return false
		case nil:
			return false
		default:
			return match(t)
		}
	}
	visit := func(n syntax.Node) bool {
		if found || n == nil {
			return false
		}
		if skip[n] {
			return false
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.ArrowFunction:
			return false
		case *syntax.Closure:
			for _, u := range x.Uses {
				if u.ByRef && u.Var != nil && match(u.Var) {
					found = true
				}
			}
			return false
		case *syntax.Assign:
			if targetHit(x.Var) || (x.ByRef && match(util.UnwrapParens(x.Value))) {
				found = true
				return false
			}
		case *syntax.IncDec:
			if match(util.UnwrapParens(x.Var)) {
				found = true
				return false
			}
		case *syntax.Foreach:
			if (x.Key != nil && targetHit(x.Key)) || (x.Value != nil && targetHit(x.Value)) {
				found = true
				return false
			}
		case *syntax.Unset:
			for _, v := range x.Vars {
				if match(util.UnwrapParens(v)) {
					found = true
					return false
				}
			}
		case *syntax.Global:
			for _, v := range x.Vars {
				if match(v) {
					found = true
					return false
				}
			}
		case *syntax.StaticVar:
			if x.Var != nil && match(x.Var) {
				found = true
				return false
			}
		case *syntax.ArgList:
			pos := 0
			for _, a := range x.Args {
				arg, ok := a.(*syntax.Arg)
				if !ok {
					continue
				}
				if !arg.Unpack && match(util.UnwrapParens(arg.Value)) && foreachInvByRefArg(ctx, arg, pos) {
					found = true
					return false
				}
				pos++
			}
		}
		return true
	}
	for _, e := range l.Init {
		syntax.Inspect(e, visit)
	}
	for _, e := range l.Cond {
		syntax.Inspect(e, visit)
	}
	syntax.Inspect(l.Body, visit)
	return found
}

// ---- Part B ---------------------------------------------------------------------------

func (foreachInvariants) eachLoop(ctx *analysis.Context, a *syntax.Assign) {
	if a.Op.Kind != syntax.TEqual {
		return
	}
	var items []*syntax.ArrayItem // D10
	switch t := a.Var.(type) {
	case *syntax.List:
		items = t.Items
	case *syntax.Array:
		items = t.Items
	default:
		return
	}
	call, ok := a.Value.(*syntax.FuncCall)
	if !ok {
		return
	}
	if name, ok := call.Name.(*syntax.Name); !ok || name.NameKind != syntax.NameUnqualified || !ctx.IsGlobalFunctionCall(call, "each") {
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 1 {
		return
	}
	subject := args[0]
	var loop syntax.Stmt // D11
	var body syntax.Stmt
	switch p := a.Parent().(type) {
	case *syntax.While:
		if p.Cond != syntax.Expr(a) {
			return
		}
		loop, body = p, p.Body
	case *syntax.For:
		loop, body = p, p.Body
	default:
		return
	}
	blk, ok := foreachInvBody(ctx, body)
	if !ok {
		return
	}
	used := false // D12
	syntax.Inspect(blk, func(x syntax.Node) bool {
		if used {
			return false
		}
		if util.EquivalentFoldNames(ctx.File, x, subject) {
			used = true
		}
		return true
	})
	if used {
		return
	}
	kw, ok := util.NextSignificant(ctx.File, loop.Span().Start)
	if !ok {
		return
	}
	span := syntax.Span{Start: kw.Start, End: kw.End}
	var fixes []analysis.Fix
	if _, isWhile := loop.(*syntax.While); isWhile && len(items) == 2 &&
		items[0] != nil && items[1] != nil && items[0].Key == nil && items[1].Key == nil {
		kv, ok1 := ruleSimpleVar(items[0].Value)
		vv, ok2 := ruleSimpleVar(items[1].Value)
		if ok1 && ok2 {
			fixes = append(fixes, analysis.Fix{
				Title: "Convert to foreach",
				Edits: func() []analysis.TextEdit {
					keyUsed := false
					syntax.Inspect(blk, func(x syntax.Node) bool {
						if v, ok := ruleSimpleVar(ruleAsExpr(x)); ok && v.Name == kv.Name {
							keyUsed = true
						}
						return !keyUsed
					})
					h := "foreach (" + ctx.Text(subject) + " as "
					if keyUsed {
						h += ctx.Text(kv) + " => "
					}
					h += ctx.Text(vv) + ") "
					return []analysis.TextEdit{{Span: loop.Span(), NewText: h + ctx.Text(blk)}}
				},
			})
		}
	}
	ctx.ReportSeverity(span, meta.SeverityError, foreachInvEachMsg, fixes...)
}
