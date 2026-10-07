package performance

import (
	"strconv"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
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
	inLoop := false // G5
	for p := syntax.Node(start); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.Foreach, *syntax.For, *syntax.While, *syntax.DoWhile:
			inLoop = true
		}
		if inLoop {
			break
		}
	}
	if !inLoop {
		return
	}
	for _, arg := range args { // G6
		if util.EquivalentFoldNames(ctx.File, a.Var, arg.Value) {
			ctx.ReportNode(call, "'"+name+"(...)' inside a loop re-copies the accumulator each time; merge once after the loop.")
			return
		}
	}
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
			ctx.ReportNode(b, msg, forLengthFix(ctx, f, call, names.pick(f, ci, forFixBase(other))))
		}
	}
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
