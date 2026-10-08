package performance

import (
	"strconv"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/syntax"
)

// slowArrayOperationsInLoop reports accumulating array_merge()-like calls in
// loops and length calls re-evaluated in `for` conditions.
type slowArrayOperationsInLoop struct{}

func init() { register(slowArrayOperationsInLoop{}) }

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
	call, name := perfCall(ctx, n, "array_merge", "array_merge_recursive", "array_replace", "array_replace_recursive") // G1
	if call == nil {
		return
	}
	args := perfArgs(call.Args)
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
		if util.EquivalentFoldNames(ctx.File, a.Var, arg.Value) {
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
		if !ok || w == a || !util.EquivalentFoldNames(ctx.File, a.Var, w.Var) {
			return false
		}
		c, _ := perfCall(ctx, w.Value, "array_merge", "array_merge_recursive", "array_replace", "array_replace_recursive")
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
				read = util.EquivalentFoldNames(ctx.File, a.Var, e)
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
			call, name := perfCall(ctx, side, forLengthFuncs...)
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
			if forSubjectMayChange(ctx, f, call) {
				ctx.ReportNode(b, msg) // custos: hoisting would freeze a changing length
				continue
			}
			ctx.ReportNode(b, msg, forLengthFix(ctx, f, call, names.pick(f, ci, forFixBase(other))))
		}
	}
}

// forSubjectMayChange reports whether the measured value may change while
// the loop runs (custos), so computing its length once would change the
// number of iterations: the subject is not a variable or property fetch,
// or the body/step writes it or one of its elements, unsets or passes it
// (or an element) to a by-reference or unresolved parameter, or calls a
// method on the object holding a measured property.
func forSubjectMayChange(ctx *analysis.Context, f *syntax.For, call *syntax.FuncCall) bool {
	args, ok := util.CallArgValues(call)
	if !ok || len(args) == 0 {
		return true
	}
	subj := syntax.UnwrapParens(args[0])
	// An element (`count($grid[$r])`) is measured through its root array:
	// any write into the root, or to an offset expression, counts.
	var dims []syntax.Expr
	for {
		d, ok := subj.(*syntax.ArrayDimFetch)
		if !ok || d.Dim == nil {
			break
		}
		dims = append(dims, syntax.UnwrapParens(d.Dim))
		subj = syntax.UnwrapParens(d.Var)
	}
	var holder syntax.Expr // object whose property is measured
	switch x := subj.(type) {
	case *syntax.Variable:
		if x.NameExpr != nil {
			return true
		}
	case *syntax.PropertyFetch:
		holder = x.Var
	case *syntax.StaticPropertyFetch:
	default:
		return true
	}
	touches := func(e syntax.Expr) bool { // e is subj, one of its elements or an offset
		for _, d := range dims {
			if util.EquivalentFoldNames(ctx.File, syntax.UnwrapParens(e), d) {
				return true
			}
		}
		for {
			e = syntax.UnwrapParens(e)
			if util.EquivalentFoldNames(ctx.File, e, subj) {
				return true
			}
			d, ok := e.(*syntax.ArrayDimFetch)
			if !ok {
				return false
			}
			e = d.Var
		}
	}
	changed := false
	visit := func(n syntax.Node) bool {
		if changed {
			return false
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.Assign:
			changed = touches(x.Var) || (x.ByRef && touches(x.Value))
		case *syntax.IncDec:
			changed = touches(x.Var)
		case *syntax.Unset:
			for _, v := range x.Vars {
				changed = changed || touches(v)
			}
		case *syntax.Foreach:
			changed = x.ByRef && touches(x.Expr)
		case *syntax.MethodCall:
			if holder != nil && util.EquivalentFoldNames(ctx.File, syntax.UnwrapParens(x.Var), holder) {
				changed = true
			}
		}
		if list := callArgList(n); list != nil && !changed {
			var params []index.Param
			resolved := false
			if fc, ok := n.(*syntax.FuncCall); ok {
				if fn := ctx.Types().ResolveFunction(fc); fn != nil {
					params, resolved = fn.Params, true
				}
			}
			pos := 0
			for _, a := range list.Args {
				arg, ok := a.(*syntax.Arg)
				if !ok {
					continue
				}
				if touches(arg.Value) {
					switch {
					case !resolved, arg.Unpack:
						changed = true
					case pos < len(params) && params[pos].ByRef:
						changed = true
					case len(params) > 0 && params[len(params)-1].Variadic && params[len(params)-1].ByRef && pos >= len(params)-1:
						changed = true
					}
				}
				pos++
			}
		}
		return !changed
	}
	syntax.Inspect(f.Body, visit)
	for _, e := range f.Loop {
		syntax.Inspect(e, visit)
	}
	return changed
}

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
	if name, ok := perfPlainVar(other); ok {
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
			if call, _ := perfCall(ctx, side, forLengthFuncs...); call != nil {
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
func forLengthFix(ctx *analysis.Context, f *syntax.For, call *syntax.FuncCall, v string) analysis.Fix {
	init := v + " = " + ctx.Text(call)
	return analysis.Fix{
		Title: "Compute the length once in the initialiser",
		Edits: func() []analysis.TextEdit {
			var at uint32
			text := init
			if len(f.Init) > 0 {
				at = f.Init[len(f.Init)-1].Span().End
				text = ", " + init
			} else if t, ok := util.FindToken(ctx.File, f.Span(), syntax.TLParen); ok {
				at = t.End
			}
			return []analysis.TextEdit{
				{Span: syntax.Span{Start: at, End: at}, NewText: text},
				{Span: call.Span(), NewText: v},
			}
		},
	}
}
