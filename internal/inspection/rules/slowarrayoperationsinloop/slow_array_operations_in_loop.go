package slowarrayoperationsinloop

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// slowArrayOperationsInLoop reports accumulating array_merge()-like calls in
// loops and length calls re-evaluated in `for` conditions.
type slowArrayOperationsInLoop struct{}

func (slowArrayOperationsInLoop) ID() string { return "SlowArrayOperationsInLoop" }
func (slowArrayOperationsInLoop) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KFor}
}

func (r slowArrayOperationsInLoop) Check(ctx *analysis.Context, n syntax.Node) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		r.checkMerge(ctx, n)
	case *syntax.For:
		r.checkFor(ctx, n)
	}
}

func (slowArrayOperationsInLoop) checkMerge(ctx *analysis.Context, n *syntax.FuncCall) {
	call, name := semanticquery.GlobalCall(ctx, n, "array_merge", "array_merge_recursive", "array_replace", "array_replace_recursive") // G1
	if call == nil {
		return
	}
	args := astquery.WrittenArguments(call.Args)
	if len(args) < 2 { // G2
		return
	}
	if _, ok := args[0].Value.(*syntax.ArrayDimFetch); ok {
		return
	}
	a, ok := call.Parent().(*syntax.Assign) // G3
	if !ok || a.Value != syntax.Expr(call) {
		return
	}
	start := a.Parent()
	if st, ok := start.(*syntax.ExprStmt); ok && mergeRunsOnce(st) { // G4
		return
	}
	var loop syntax.Node // G5
	for p := start; p != nil && !syntax.IsFuncLike(p) && loop == nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.Foreach, *syntax.For, *syntax.While, *syntax.DoWhile:
			loop = p
		}
	}
	if loop == nil {
		return
	}
	for _, arg := range args { // G6
		if astquery.EquivalentFoldNames(ctx.File, a.Var, arg.Value) {
			if mergeNotAccumulating(ctx, loop, a, arg.Value) { // G7
				return
			}
			ctx.ReportNode(call, "'"+name+"(...)' inside a loop re-copies the accumulator each time; merge once after the loop.")
			return
		}
	}
}

// mergeNotAccumulating implements G7 (custos): the merge cannot be hoisted
// out of the innermost loop when the target's base variable is (re)assigned
// elsewhere in the loop or in its header (the foreach value: a fresh value
// per iteration), when the target's index mentions such a variable (a
// different element per iteration), or when the accumulator is read
// elsewhere in the loop (the next iteration needs the merged value).
func mergeNotAccumulating(ctx *analysis.Context, loop syntax.Node, a *syntax.Assign, matched syntax.Expr) bool {
	base, dims := a.Var, []syntax.Expr(nil)
	for {
		d, ok := base.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		if d.Dim != nil {
			dims = append(dims, d.Dim)
		}
		base = d.Var
	}
	// Other accumulating merges into the same target (`$r = array_merge($r,
	// …)` twice in one loop) neither reset nor read it in between.
	sibling := func(x syntax.Node) bool {
		w, ok := x.(*syntax.Assign)
		if !ok || w == a || !astquery.EquivalentFoldNames(ctx.File, a.Var, w.Var) {
			return false
		}
		c, _ := semanticquery.GlobalCall(ctx, w.Value, "array_merge", "array_merge_recursive", "array_replace", "array_replace_recursive")
		return c != nil
	}
	written := mergeLoopWrites(loop, a, sibling)
	if bv, ok := base.(*syntax.Variable); ok && bv.NameExpr == nil && written[bv.Name] {
		return true
	}
	for _, d := range dims {
		hit := false
		syntax.Inspect(d, func(x syntax.Node) bool {
			if v, ok := x.(*syntax.Variable); ok && v.NameExpr == nil && written[v.Name] {
				hit = true
			}
			return !hit
		})
		if hit {
			return true
		}
	}
	plain, _ := a.Var.(*syntax.Variable)
	read := false
	syntax.Inspect(loop, func(x syntax.Node) bool {
		if x == syntax.Node(a.Var) || x == syntax.Node(matched) || sibling(x) {
			return false
		}
		if e, ok := x.(syntax.Expr); ok && e.Kind() == a.Var.Kind() {
			if plain != nil {
				v := e.(*syntax.Variable)
				read = v.NameExpr == nil && v.Name == plain.Name
			} else {
				read = astquery.EquivalentFoldNames(ctx.File, a.Var, e)
			}
		}
		return !read
	})
	return read
}

// mergeLoopWrites returns the plain variables assigned by the loop header
// (foreach key/value, for initialiser/step, assignments in a while
// condition) or by a plain assignment in its body other than a.
func mergeLoopWrites(loop syntax.Node, a *syntax.Assign, skip func(syntax.Node) bool) map[string]bool {
	out := map[string]bool{}
	vars := func(n syntax.Node) {
		if n == nil {
			return
		}
		syntax.Inspect(n, func(x syntax.Node) bool {
			if v, ok := x.(*syntax.Variable); ok && v.NameExpr == nil {
				out[v.Name] = true
			}
			return true
		})
	}
	switch l := loop.(type) {
	case *syntax.Foreach:
		vars(l.Key)
		vars(l.Value)
	case *syntax.For:
		for _, e := range l.Init {
			vars(e)
		}
		for _, e := range l.Loop {
			vars(e)
		}
	}
	syntax.Inspect(loop, func(x syntax.Node) bool {
		if skip(x) {
			return false
		}
		if w, ok := x.(*syntax.Assign); ok && w != a && w.Op.Kind == syntax.TEqual {
			if v, ok := w.Var.(*syntax.Variable); ok && v.NameExpr == nil {
				out[v.Name] = true
			}
		}
		return true
	})
	return out
}

// mergeRunsOnce implements G4: walking from st up to the nearest enclosing
// loop, the merge is exempt when it sits in a conditional branch (if, elseif,
// else, switch case) or in a block whose last statement leaves the loop
// (break, return, throw).
func mergeRunsOnce(st *syntax.ExprStmt) bool {
	for p := st.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		switch p := p.(type) {
		case *syntax.Foreach, *syntax.For, *syntax.While, *syntax.DoWhile:
			return false
		case *syntax.If, *syntax.ElseIf, *syntax.Else, *syntax.Case: // G4b
			return true
		case *syntax.Block:
			if len(p.Stmts) > 0 && leavesLoop(p.Stmts[len(p.Stmts)-1]) { // G4a
				return true
			}
		}
	}
	return false
}

// leavesLoop reports whether s is a break, return or throw statement.
func leavesLoop(s syntax.Stmt) bool {
	switch s := s.(type) {
	case *syntax.Break, *syntax.Return:
		return true
	case *syntax.ExprStmt:
		_, ok := s.Expr.(*syntax.Throw)
		return ok
	}
	return false
}

// isComparisonOp reports whether k is a relational or equality operator
// (F-1).
func isComparisonOp(k syntax.TokenKind) bool {
	switch k {
	case syntax.TLess, syntax.TIsSmallerOrEqual, syntax.TGreater, syntax.TIsGreaterOrEqual,
		syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
		return true
	}
	return false
}

// forLengthFuncs are the length functions of F-2 (`sizeof` is count's alias).
var forLengthFuncs = []string{"count", "sizeof", "strlen", "mb_strlen"}

func (slowArrayOperationsInLoop) checkFor(ctx *analysis.Context, f *syntax.For) {
	var names *forNames
	for ci, cond := range f.Cond { // F-1
		b, ok := cond.(*syntax.Binary)
		if !ok || !isComparisonOp(b.Op.Kind) {
			continue
		}
		fixed := false
		for _, side := range [2]syntax.Expr{b.Left, b.Right} { // F-2
			call, name := semanticquery.GlobalCall(ctx, side, forLengthFuncs...)
			if call == nil {
				continue
			}
			msg := "'" + name + "(...)' is re-evaluated on every iteration; compute it once before the loop."
			if fixed { // both sides: fix only the left one (spec Divergences)
				ctx.ReportNode(b, msg)
				continue
			}
			fixed = true
			other := b.Right
			if side == b.Right {
				other = b.Left
			}
			if names == nil {
				names = newForNames(ctx, f)
			}
			switch forSubjectChange(ctx, f, call, name) {
			case forSubjectWritten:
				continue // custos: the loop changes the length on purpose
			case forSubjectMayChange:
				ctx.ReportNode(b, msg) // custos: hoisting could freeze a changing length
				continue
			}
			ctx.ReportNode(b, msg, forLengthFix(ctx, f, call, names.pick(f, ci, forFixBase(other))))
		}
	}
}

// forSubjectChange tells whether the measured value may change while the
// loop runs (custos), so computing its length once would change the number
// of iterations.
//
// It is written when the body, the step or the condition visibly changes it:
// assigns (also through list()/[] destructuring or as a foreach key/value),
// pushes to, increments or unsets the measured expression, one of its
// elements, or an expression it is reached through (the root array, the
// object holding a measured property, a variable of an offset), or passes
// one of them to a by-reference parameter of a resolved function. The loop
// then measures a length it changes on purpose, and the condition is not
// reported.
//
// It may change when the subject is not a variable or property fetch, when
// the loop iterates it by reference, binds a reference to it, passes it to
// an unresolved callee or unpacks it into a call, calls a method on it or on
// an object it is reached through, hands such an object to user code, or
// runs code that can reach a shared subject (a property, or a variable that
// is not purely local: any call but a builtin without callable parameter),
// and when a reference to it exists anywhere in the enclosing scope. The
// condition is reported without a fix.
func forSubjectChange(ctx *analysis.Context, f *syntax.For, call *syntax.FuncCall, name string) forChange {
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) == 0 {
		return forSubjectMayChange
	}
	full := syntax.UnwrapParens(args[0])
	// An element (`count($grid[$r])`) is measured through its root array.
	subj := full
	for {
		d, ok := subj.(*syntax.ArrayDimFetch)
		if !ok || d.Dim == nil {
			break
		}
		subj = syntax.UnwrapParens(d.Var)
	}
	var holders []syntax.Expr // objects the measured property is reached through
	// shared: the measured value is reachable by code the loop calls (a
	// property, or a variable that is not purely local).
	shared, plain := true, false
	switch x := subj.(type) {
	case *syntax.Variable:
		if x.NameExpr != nil {
			return forSubjectMayChange
		}
		plain, shared = true, forSharedVariable(f, x.Name)
	case *syntax.PropertyFetch:
		for h := syntax.UnwrapParens(x.Var); h != nil; {
			holders = append(holders, h)
			switch p := h.(type) {
			case *syntax.PropertyFetch:
				h = syntax.UnwrapParens(p.Var)
			case *syntax.ArrayDimFetch:
				h = syntax.UnwrapParens(p.Var)
			default:
				h = nil
			}
		}
	case *syntax.StaticPropertyFetch:
	default:
		return forSubjectMayChange
	}
	parts := forMeasuredParts(full, nil)
	// reach returns the part of the measured expression that e, or the array
	// e is an element of, denotes, and whether e is that part itself.
	reach := func(e syntax.Expr) (part syntax.Expr, exact bool) {
		exact = true
		for {
			e = syntax.UnwrapParens(e)
			for _, p := range parts {
				if astquery.EquivalentFoldNames(ctx.File, e, p) {
					return p, exact
				}
			}
			d, ok := e.(*syntax.ArrayDimFetch)
			if !ok {
				return nil, false
			}
			e, exact = d.Var, false
		}
	}
	touches := func(e syntax.Expr) bool {
		p, _ := reach(e)
		return p != nil
	}
	// sibling reports whether e is the measured property, or one of its
	// elements, read through another object expression: `$me->rows` for
	// `$this->rows`, the same array when `$me` is `$this`.
	sibling := func(e syntax.Expr) bool {
		p, ok := subj.(*syntax.PropertyFetch)
		for ok {
			e = syntax.UnwrapParens(e)
			if q, isProp := e.(*syntax.PropertyFetch); isProp {
				return astquery.EquivalentFoldNames(ctx.File, q.Name, p.Name)
			}
			var d *syntax.ArrayDimFetch
			if d, ok = e.(*syntax.ArrayDimFetch); ok {
				e = d.Var
			}
		}
		return false
	}
	var assigns func(e syntax.Expr) bool // e as a write target, destructuring included
	assigns = func(e syntax.Expr) bool {
		var items []*syntax.ArrayItem
		switch x := syntax.UnwrapParens(e).(type) {
		case *syntax.Array:
			items = x.Items
		case *syntax.List:
			items = x.Items
		default:
			return touches(e)
		}
		for _, it := range items {
			if it != nil && it.Value != nil && assigns(it.Value) {
				return true
			}
		}
		return false
	}
	// object reports whether passing part by value may hand over an object
	// the callee can change: a holder, or a subject not known to be an
	// array or a string.
	object := func(part syntax.Expr) bool {
		for _, h := range holders {
			if astquery.EquivalentFoldNames(ctx.File, part, h) {
				return true
			}
		}
		if part != full && part != subj {
			return false
		}
		if part == full && (name == "strlen" || name == "mb_strlen") {
			return false
		}
		return !ctx.TypeOf(subj).IsArrayLike()
	}
	state := forSubjectStable
	mark := func(c forChange) {
		state = max(state, c)
	}
	visit := func(n syntax.Node) bool {
		if state == forSubjectWritten {
			return false
		}
		var fn *index.Function
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike:
			return false
		case *syntax.Closure, *syntax.ArrowFunction:
			if plain { // captured by value; by-reference captures: see below
				return false
			}
		case *syntax.Assign:
			if assigns(x.Var) {
				mark(forSubjectWritten)
			} else if x.ByRef && touches(x.Value) || sibling(x.Var) {
				mark(forSubjectMayChange)
			}
		case *syntax.IncDec:
			if touches(x.Var) {
				mark(forSubjectWritten)
			}
		case *syntax.Unset:
			for _, v := range x.Vars {
				if touches(v) {
					mark(forSubjectWritten)
				} else if sibling(v) {
					mark(forSubjectMayChange)
				}
			}
		case *syntax.Foreach:
			if assigns(x.Value) || x.Key != nil && assigns(x.Key) {
				mark(forSubjectWritten)
			} else if x.ByRef && touches(x.Expr) {
				mark(forSubjectMayChange)
			}
		case *syntax.Variable:
			if plain && x.NameExpr != nil { // `$$name` may be the measured variable
				mark(forSubjectMayChange)
			}
		case *syntax.Include, *syntax.Eval:
			mark(forSubjectMayChange) // runs in this scope
		case *syntax.MethodCall:
			if _, exact := reach(x.Var); exact || shared {
				mark(forSubjectMayChange)
			}
		case *syntax.StaticCall, *syntax.New, *syntax.Yield, *syntax.YieldFrom:
			if shared { // user code runs before the next length check
				mark(forSubjectMayChange)
			}
		case *syntax.FuncCall:
			fn = ctx.Types().ResolveFunction(x)
			if shared && (fn == nil || !fn.Builtin || forCallsBack(fn)) || plain && forWritesScope(ctx, x) {
				mark(forSubjectMayChange)
			}
		}
		if list := callArgList(n); list != nil {
			pos := 0
			for _, a := range list.Args {
				arg, ok := a.(*syntax.Arg)
				if !ok {
					continue
				}
				if part, exact := reach(arg.Value); part != nil {
					switch {
					case fn == nil, arg.Unpack:
						mark(forSubjectMayChange)
					case forParamByRef(fn.Params, arg, pos):
						mark(forSubjectWritten)
					case !fn.Builtin && exact && object(part):
						mark(forSubjectMayChange)
					}
				} else if sibling(arg.Value) && (fn == nil || forParamByRef(fn.Params, arg, pos)) {
					mark(forSubjectMayChange)
				}
				pos++
			}
		}
		return state != forSubjectWritten
	}
	syntax.Inspect(f.Body, visit)
	for _, e := range f.Loop {
		syntax.Inspect(e, visit)
	}
	for _, e := range f.Cond {
		syntax.Inspect(e, visit)
	}
	if state != forSubjectStable {
		return state
	}
	// A reference made anywhere in the scope lets a write under another name
	// change the measured value. A reference to an element of the measured
	// value itself cannot change its length.
	scope := syntax.EnclosingFuncLike(f)
	aliased := func(e syntax.Expr) bool {
		p, exact := reach(e)
		return p != nil && (exact || p != full)
	}
	alias := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.ArrowFunction:
			return n == scope
		case *syntax.Closure:
			if n != scope {
				for _, u := range x.Uses {
					if u.ByRef && touches(u.Var) {
						mark(forSubjectMayChange)
					}
				}
				return false
			}
		case *syntax.Assign:
			if x.ByRef && (aliased(x.Var) || aliased(x.Value)) {
				mark(forSubjectMayChange)
			}
		case *syntax.Foreach:
			if p, exact := reach(x.Expr); x.ByRef && exact && p != full {
				mark(forSubjectMayChange)
			}
		}
		return state == forSubjectStable
	}
	if scope != nil {
		syntax.Inspect(scope, alias)
	} else {
		for _, st := range ctx.File.Stmts {
			syntax.Inspect(st, alias)
		}
	}
	return state
}

// forMeasuredParts appends to out the expressions the measured expression e
// is read through: e itself, the arrays and objects it is fetched from and
// the variables of its offsets.
func forMeasuredParts(e syntax.Expr, out []syntax.Expr) []syntax.Expr {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Variable, *syntax.StaticPropertyFetch:
		out = append(out, x)
	case *syntax.PropertyFetch:
		out = forMeasuredParts(x.Var, append(out, x))
	case *syntax.ArrayDimFetch:
		out = forMeasuredParts(x.Var, append(out, x))
		if x.Dim != nil {
			syntax.Inspect(x.Dim, func(n syntax.Node) bool {
				switch v := n.(type) {
				case *syntax.Variable, *syntax.StaticPropertyFetch:
					out = append(out, v.(syntax.Expr))
					return false
				case *syntax.PropertyFetch:
					out = append(out, v)
				}
				return true
			})
		}
	}
	return out
}

// forSuperglobals are the variables every scope shares.
var forSuperglobals = []string{"GLOBALS", "_GET", "_POST", "_COOKIE", "_REQUEST", "_SERVER", "_SESSION", "_ENV", "_FILES"}

// forSharedVariable reports whether code called from the loop f can write
// the plain variable name: it lives outside the function (top-level code,
// `global`, `static`, a superglobal) or is bound to a variable of the caller
// (by-reference parameter or closure capture).
func forSharedVariable(f *syntax.For, name string) bool {
	scope := syntax.EnclosingFuncLike(f)
	if scope == nil || slices.Contains(forSuperglobals, name) {
		return true
	}
	shared := false
	is := func(e syntax.Expr) bool {
		v, ok := e.(*syntax.Variable)
		return ok && v.Name == name
	}
	syntax.Inspect(scope, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
			if n != scope {
				return false
			}
		case *syntax.Param:
			shared = shared || x.ByRef && is(x.Var)
		case *syntax.ClosureUse:
			shared = shared || x.ByRef && is(x.Var)
		case *syntax.Global:
			shared = shared || slices.ContainsFunc(x.Vars, is)
		case *syntax.StaticVar:
			shared = shared || is(x.Var)
		}
		return !shared
	})
	return shared
}

// forCallsBack reports whether the builtin fn takes a callable, which runs
// user code.
func forCallsBack(fn *index.Function) bool {
	return slices.ContainsFunc(fn.Params, func(p index.Param) bool {
		return strings.Contains(strings.ToLower(p.Type+"|"+p.DocType), "callable")
	})
}

// forWritesScope reports whether call reaches a builtin that creates or
// overwrites variables of the calling scope.
func forWritesScope(ctx *analysis.Context, call *syntax.FuncCall) bool {
	c, _ := semanticquery.GlobalCall(ctx, call, "extract", "parse_str", "mb_parse_str")
	return c != nil
}

// forParamByRef reports whether arg, the pos-th argument of a call, binds
// to a by-reference parameter of params.
func forParamByRef(params []index.Param, arg *syntax.Arg, pos int) bool {
	if arg.Name != nil {
		for _, p := range params {
			if p.ByRef && strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
				return true
			}
		}
		return false
	}
	if pos < len(params) {
		return params[pos].ByRef
	}
	return len(params) > 0 && params[len(params)-1].Variadic && params[len(params)-1].ByRef
}

// forChange grades how the measured value of a `for` condition may change
// while the loop runs (see forSubjectChange).
type forChange int

const (
	forSubjectStable forChange = iota
	forSubjectMayChange
	forSubjectWritten
)

// callArgList returns the argument list of a call or instantiation.
func callArgList(n syntax.Node) *syntax.ArgList {
	switch c := n.(type) {
	case *syntax.FuncCall:
		return c.Args
	case *syntax.MethodCall:
		return c.Args
	case *syntax.StaticCall:
		return c.Args
	case *syntax.New:
		return c.Args
	}
	return nil
}

// forFixBase is the preferred name of the cached length variable (F1).
func forFixBase(other syntax.Expr) string {
	if name, ok := astquery.PlainVariableName(other); ok {
		return name + "Max"
	}
	return "loopsMax"
}

// forFixBases lists, per condition expression of f, the preferred variable
// name of the fix that condition would get ("" when it gets none).
func forFixBases(ctx *analysis.Context, f *syntax.For) []string {
	out := make([]string, len(f.Cond))
	for i, cond := range f.Cond {
		b, ok := cond.(*syntax.Binary)
		if !ok || !isComparisonOp(b.Op.Kind) {
			continue
		}
		for _, side := range [2]syntax.Expr{b.Left, b.Right} {
			if call, _ := semanticquery.GlobalCall(ctx, side, forLengthFuncs...); call != nil {
				other := b.Right
				if side == b.Right {
					other = b.Left
				}
				out[i] = forFixBase(other)
				break
			}
		}
	}
	return out
}

// forNames chooses fresh names for the cached length variable (F1a): a name
// already used anywhere in the enclosing scope is skipped, and loops nested
// in each other never get the same generated name.
type forNames struct {
	ctx   *analysis.Context
	used  map[string]bool
	scope syntax.Node
}

func newForNames(ctx *analysis.Context, f *syntax.For) *forNames {
	scope := syntax.EnclosingFuncLike(f)
	used := map[string]bool{}
	collect := func(n syntax.Node) {
		syntax.Inspect(n, func(c syntax.Node) bool {
			if v, ok := c.(*syntax.Variable); ok && v.NameExpr == nil {
				used[v.Name] = true
			}
			return true
		})
	}
	if scope != nil {
		collect(scope)
	} else {
		for _, st := range ctx.File.Stmts {
			collect(st)
		}
	}
	return &forNames{ctx: ctx, used: used, scope: scope}
}

// pick returns the variable (with `$`) for condition ci of f. The k-th
// claimant of a base name among the enclosing `for` loops of the same scope
// (outermost first) and the earlier conditions of f gets the k-th free
// candidate among base, base1, base2, ...
func (n *forNames) pick(f *syntax.For, ci int, base string) string {
	k := 0
	for i, b := range forFixBases(n.ctx, f) {
		if i < ci && b == base {
			k++
		}
	}
	for p := f.Parent(); p != nil && p != n.scope && !syntax.IsFuncLike(p); p = p.Parent() {
		if outer, ok := p.(*syntax.For); ok {
			for _, b := range forFixBases(n.ctx, outer) {
				if b == base {
					k++
				}
			}
		}
	}
	for i := 0; ; i++ {
		cand := base
		if i > 0 {
			cand += strconv.Itoa(i)
		}
		if n.used[cand] {
			continue
		}
		if k == 0 {
			return "$" + cand
		}
		k--
	}
}

// forLengthFix implements F1: cache the length call in the initialiser.
func forLengthFix(ctx *analysis.Context, f *syntax.For, call *syntax.FuncCall, v string) diagnostic.Fix {
	init := v + " = " + ctx.Text(call)
	return diagnostic.Fix{
		Title: "Compute the length once in the initialiser",
		Edits: func() []diagnostic.TextEdit {
			var at uint32
			text := init
			if len(f.Init) > 0 {
				at = f.Init[len(f.Init)-1].Span().End
				text = ", " + init
			} else if t, ok := astquery.FindToken(ctx.File, f.Span(), syntax.TLParen); ok {
				at = t.End
			}
			return []diagnostic.TextEdit{
				{Span: syntax.Span{Start: at, End: at}, NewText: text},
				{Span: call.Span(), NewText: v},
			}
		},
	}
}
