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
	// custos: only conditions that stop at the limit (`$i < X`, `X > $i`,
	// `$i != X`); `$i <= count($a)` runs once more than foreach would.
	var limit syntax.Expr
	op := cond.Op.Kind
	ne := op == syntax.TIsNotEqual || op == syntax.TIsNotIdentical
	if v, ok := ruleSimpleVar(cond.Left); ok && v.Name == counter.Name && (ne || op == syntax.TLess) {
		limit = cond.Right
	} else if v, ok := ruleSimpleVar(cond.Right); ok && v.Name == counter.Name && (ne || op == syntax.TGreater) {
		limit = cond.Left
	} else {
		return
	}
	// custos: the fix keeps only the counter and a limit variable of the
	// header; any other init expression (`$acc = []`) would be lost.
	for _, in := range l.Init {
		if a, ok := in.(*syntax.Assign); ok {
			if v, ok := ruleSimpleVar(a.Var); ok && v.Name == counter.Name {
				continue
			}
			if v, ok := ruleSimpleVar(a.Var); ok {
				if lv, ok := ruleSimpleVar(syntax.UnwrapParens(limit)); ok && lv.Name == v.Name {
					continue
				}
			}
		}
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
	lim := syntax.UnwrapParens(limit)
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
	// D8d: the container itself changes in the loop (reassigned, passed by
	// reference as to sort()/array_splice(), unset, pushed to, or an element
	// unset): foreach iterates a snapshot, the counter loop the live array.
	if foreachInvContainerChanged(ctx, l, container) || foreachInvElementsChanged(ctx, l, container, counter.Name) {
		return
	}
	// D8c: the counter, or a limit assigned in the header, read after the loop
	// would see the foreach's leftovers instead of the final for values.
	if foreachInvUsedAfter(ctx.File, l, counter.Name) {
		return
	}
	if lv, ok := ruleSimpleVar(lim); ok && foreachInvHeaderAssigns(l, lv.Name) && foreachInvUsedAfter(ctx.File, l, lv.Name) {
		return
	}
	ctx.Report(keywordSpan(ctx, l), foreachInvCounterMsg, analysis.Fix{
		Title: "Convert to foreach",
		Edits: func() []analysis.TextEdit {
			return counterLoopFix(ctx, l, body, counter, container, limit)
		},
	})
}

func counterLoopFix(ctx *analysis.Context, l *syntax.For, body *syntax.Block, counter *syntax.Variable,
	container, limit syntax.Expr,
) []analysis.TextEdit {
	cName := "$" + counter.Name
	valName := cName + "Value"
	// F1.1: qualifying accesses.
	type piece struct {
		sp   syntax.Span
		text string
	}
	var repl []piece
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
		sp, text := d.Span(), valName
		if _, inStr := d.Parent().(*syntax.InterpolatedString); inStr {
			if ctx.Src[sp.Start-1] == '{' && ctx.Src[sp.End] == '}' {
				sp = syntax.Span{Start: sp.Start - 1, End: sp.End + 1}
			}
			// "{$c[$i]}abc", "$c[$i][0]", "{$c[$i]}->p": a bare $iValue
			// would absorb the following characters.
			if interpolationContinues(ctx.Src[sp.End:]) {
				text = "{" + valName + "}"
			}
		}
		repl = append(repl, piece{sp, text})
		replaced = append(replaced, d)
		return false
	})
	sort.Slice(repl, func(i, j int) bool { return repl[i].sp.Start < repl[j].sp.Start })
	bs := body.Span()
	var b strings.Builder
	pos := bs.Start
	for _, r := range repl {
		b.Write(ctx.Src[pos:r.sp.Start])
		b.WriteString(r.text)
		pos = r.sp.End
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
		if v, ok := ruleSimpleVar(util.AsExpr(x)); ok && v.Name == counter.Name {
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
	// custos: only a variable's assignment; a property write is state.
	if _, isVar := ruleSimpleVar(limit); isVar {
		if del, ok := limitAssignment(ctx, l, limit); ok {
			edits = append(edits, analysis.TextEdit{Span: util.WithLeadingWhitespace(ctx.File, del.Span())})
		}
	}
	return edits
}

// interpolationContinues reports whether the string text after a simple
// interpolated variable would be read as part of it (an identifier
// character, an offset or a property access). rest is empty only for an
// unterminated string at the end of the file (fuzz).
func interpolationContinues(rest []byte) bool {
	if len(rest) == 0 {
		return false
	}
	c := rest[0]
	return c == '_' || c == '[' || c >= 0x80 || (c >= '0' && c <= '9') || (c|0x20 >= 'a' && c|0x20 <= 'z') ||
		(c == '-' && len(rest) > 1 && rest[1] == '>')
}

// foreachInvLimitValues implements D8 for the limit X: a plain variable takes
// only the assignments that reach the loop condition (its own init
// assignment when present, which dominates the condition). Not
// util.PossibleValuesReaching: the loop's own init assignment wins,
// assignment chains stop at a by-reference assignment and arrow functions
// yield nothing (the spec's D8).
func foreachInvLimitValues(ctx *analysis.Context, l *syntax.For, limit syntax.Expr) []syntax.Expr {
	v, ok := ruleSimpleVar(limit)
	if !ok {
		if pf, ok := syntax.UnwrapParens(limit).(*syntax.PropertyFetch); ok {
			return foreachInvPropertyLimitValues(ctx, l, pf)
		}
		return util.PossibleValues(ctx.File, limit)
	}
	scope := syntax.EnclosingFuncLike(l)
	body := syntax.FuncLikeBody(scope)
	if scope == nil || body == nil || util.UnstableVariableIn(ctx.File, body, v.Name) {
		return nil
	}
	innermost := func(e syntax.Expr) syntax.Expr {
		for {
			a, ok := syntax.UnwrapParens(e).(*syntax.Assign)
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
		for _, p := range syntax.FuncLikeParams(scope) {
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

// foreachInvPropertyLimitValues discovers a property limit (custos): the
// plain assignments to it located before the loop in the enclosing
// function. A write after the loop, or none before it (the value then
// comes from another method or the default, which other code may change)
// gives no value.
func foreachInvPropertyLimitValues(ctx *analysis.Context, l *syntax.For, limit *syntax.PropertyFetch) []syntax.Expr {
	isLimit := func(e syntax.Expr) bool {
		_, ok := syntax.UnwrapParens(e).(*syntax.PropertyFetch)
		return ok && util.EquivalentFoldNames(ctx.File, syntax.UnwrapParens(e), limit)
	}
	// A header write to the property never gets here: only the counter
	// and a limit variable may be assigned in the init clause.
	body := syntax.FuncLikeBody(syntax.EnclosingFuncLike(l))
	if body == nil {
		return nil
	}
	var out []syntax.Expr
	after := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.Assign:
			if !isLimit(x.Var) {
				return true
			}
			if x.Span().Start >= l.Span().Start || x.Op.Kind != syntax.TEqual || x.ByRef {
				after = true
				return false
			}
			out = append(out, util.PossibleValues(ctx.File, x.Value)...)
		}
		return !after
	})
	if after {
		return nil
	}
	return out
}

// limitAssignment returns the statement `$n = count(..);` to delete when the
// limit's only remaining occurrence in the enclosing function is its
// assignment.
func limitAssignment(ctx *analysis.Context, l *syntax.For, limit syntax.Expr) (syntax.Stmt, bool) {
	// Only variable limits get here, and those have values only inside a
	// function body (D8): fbody is never nil.
	fbody := syntax.FuncLikeBody(syntax.EnclosingFuncLike(l))
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
	if !ok || st.Span().End > l.Span().Start {
		return nil, false // custos: never a write after the loop
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
		// an Arg always sits in an ArgList, whose parent is the call
		return callHasNoByRefParams(ctx, p.Parent().Parent())
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
	params, ok := foreachInvCallParams(ctx, arg.Parent().Parent())
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
		switch t := syntax.UnwrapParens(e).(type) {
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
		default:
			return match(t)
		}
	}
	visit := func(n syntax.Node) bool {
		if found {
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
			if targetHit(x.Var) || (x.ByRef && match(syntax.UnwrapParens(x.Value))) {
				found = true
				return false
			}
		case *syntax.IncDec:
			if match(syntax.UnwrapParens(x.Var)) {
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
				if match(syntax.UnwrapParens(v)) {
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
				// match last: callers may record what it matches.
				if !arg.Unpack && foreachInvByRefArg(ctx, arg, pos) && match(syntax.UnwrapParens(arg.Value)) {
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

// foreachInvContainerChanged implements D8d.
func foreachInvContainerChanged(ctx *analysis.Context, l *syntax.For, container syntax.Expr) bool {
	isContainer := func(e syntax.Expr) bool {
		if d, ok := e.(*syntax.ArrayDimFetch); ok && d.Dim == nil { // `$c[] = …`
			e = d.Var
		}
		return util.EquivalentFoldNames(ctx.File, e, container)
	}
	if foreachInvWritten(ctx, l, isContainer, nil) {
		return true
	}
	found := false
	syntax.Inspect(l.Body, func(n syntax.Node) bool {
		if u, ok := n.(*syntax.Unset); ok {
			for _, v := range u.Vars {
				if d, ok := syntax.UnwrapParens(v).(*syntax.ArrayDimFetch); ok && util.EquivalentFoldNames(ctx.File, d.Var, container) {
					found = true
				}
			}
		}
		return !found
	})
	return found
}

// foreachInvElementsChanged reports whether the loop writes container
// elements in a way foreach's snapshot and the fix's value variable would
// not see (custos): an element other than `$c[$i]` (`$c[$i + 1] = …`,
// `$c[$k]['x'] = …`; a later iteration would read the new value), or a
// path under `$c[$i]` followed in the body by a mention of an overlapping
// path (`$c[$i] = trim($c[$i]); echo $c[$i];`: the fix would read the stale
// `$iValue`). Taking a reference (`$r = &$c[$i]`) is not a write.
func foreachInvElementsChanged(ctx *analysis.Context, l *syntax.For, container syntax.Expr, counter string) bool {
	type write struct {
		end  uint32
		path []syntax.Expr
	}
	var writes []write
	other := false
	match := func(e syntax.Expr) bool {
		path, ok := foreachInvElementPath(ctx, e, container)
		if !ok {
			return false
		}
		if a, ok := e.Parent().(*syntax.Assign); ok && a.ByRef && a.Value == e {
			return false
		}
		if v, ok := ruleSimpleVar(path[0]); !ok || v.Name != counter {
			other = true
			return true
		}
		// The write takes effect at the end of its statement: reads in
		// the assigned value (`$c[$i] = trim($c[$i])`) run before it.
		var n syntax.Node = e
		for {
			if _, ok := n.(syntax.Stmt); ok {
				break
			}
			n = n.Parent()
		}
		writes = append(writes, write{n.Span().End, path})
		return false
	}
	if foreachInvWritten(ctx, l, match, nil) || other {
		return true
	}
	if len(writes) == 0 {
		return false
	}
	stale := false
	syntax.Inspect(l.Body, func(n syntax.Node) bool {
		d, ok := n.(*syntax.ArrayDimFetch)
		if !ok || stale || !util.EquivalentFoldNames(ctx.File, d.Var, container) {
			return !stale
		}
		var outer syntax.Expr = d
		for {
			p, ok := outer.Parent().(*syntax.ArrayDimFetch)
			if !ok || p.Var != outer {
				break
			}
			outer = p
		}
		path, _ := foreachInvElementPath(ctx, outer, container)
		for _, w := range writes {
			if d.Span().Start >= w.end && foreachInvPathsOverlap(ctx, path, w.path) {
				stale = true
			}
		}
		return !stale
	})
	return stale
}

// foreachInvElementPath returns the offsets of an element chain
// `$c[k1][k2]…` of the container, outermost container offset first (a nil
// offset for `$c[]`).
func foreachInvElementPath(ctx *analysis.Context, e, container syntax.Expr) ([]syntax.Expr, bool) {
	var rev []syntax.Expr
	for {
		d, ok := syntax.UnwrapParens(e).(*syntax.ArrayDimFetch)
		if !ok {
			return nil, false
		}
		rev = append(rev, d.Dim)
		if util.EquivalentFoldNames(ctx.File, d.Var, container) {
			for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
				rev[i], rev[j] = rev[j], rev[i]
			}
			return rev, true
		}
		e = d.Var
	}
}

// foreachInvPathsOverlap reports whether two element paths may designate
// the same storage: one is a prefix of the other, unless two offsets at the
// same depth are different literals.
func foreachInvPathsOverlap(ctx *analysis.Context, a, b []syntax.Expr) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		x, y := a[i], b[i]
		if x == nil || y == nil {
			return true
		}
		lx, okx := syntax.UnwrapParens(x).(*syntax.Literal)
		ly, oky := syntax.UnwrapParens(y).(*syntax.Literal)
		if okx && oky && ctx.Text(lx) != ctx.Text(ly) {
			return false
		}
	}
	return true
}

// foreachInvHeaderAssigns reports whether the loop's init clause assigns
// the variable name.
func foreachInvHeaderAssigns(l *syntax.For, name string) bool {
	for _, in := range l.Init {
		if a, ok := in.(*syntax.Assign); ok {
			if v, ok := ruleSimpleVar(a.Var); ok && v.Name == name {
				return true
			}
		}
	}
	return false
}

// foreachInvUsedAfter implements D8c: whether $name is mentioned after the
// loop's end in its scope (function body, or the file at top level; nested
// named functions and classes are other scopes). A first later mention that
// is a plain re-assignment dominating the rest — the target of `$x = …;` or
// of a `for` header's `$x = …`, in the statement list holding the loop or
// one of its enclosing statements, with a value not reading $x — does not
// count.
func foreachInvUsedAfter(f *syntax.File, l *syntax.For, name string) bool {
	end := l.Span().End
	var first *syntax.Variable
	visit := func(x syntax.Node) bool {
		if first != nil {
			return false
		}
		switch x.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike:
			return false
		}
		if v, ok := x.(*syntax.Variable); ok && v.Name == name && v.NameExpr == nil && v.Span().Start >= end {
			first = v
		}
		return true
	}
	if scope := syntax.FuncLikeBody(syntax.EnclosingFuncLike(l)); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		for _, st := range f.Stmts {
			syntax.Inspect(st, visit)
		}
	}
	if first == nil {
		return false
	}
	a, ok := first.Parent().(*syntax.Assign)
	if !ok || a.Var != syntax.Expr(first) || a.Op.Kind != syntax.TEqual || a.ByRef || util.MentionsVariable(a.Value, name) {
		return true
	}
	var stmt syntax.Node
	switch p := a.Parent().(type) {
	case *syntax.ExprStmt:
		stmt = p
	case *syntax.For:
		stmt = p
	default:
		return true
	}
	// Two statements sharing a parent sit in one statement list (block,
	// case, namespace or the file), the later one after the loop.
	for anc := syntax.Node(l); anc != nil; anc = anc.Parent() {
		if anc.Parent() == stmt.Parent() {
			return false
		}
	}
	return true
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
	case *syntax.While: // an expression child of a while is its condition
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
	span := keywordSpan(ctx, loop)
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
						if v, ok := ruleSimpleVar(util.AsExpr(x)); ok && v.Name == kv.Name {
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
