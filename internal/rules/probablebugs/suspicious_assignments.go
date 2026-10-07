package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// suspiciousAssignments groups six checks for assignments that are likely
// mistakes (see specs/SuspiciousAssignments.md).
type suspiciousAssignments struct{}

func init() { register(suspiciousAssignments{}) }

// Semantic marks the rule as needing the project symbol index.
func (suspiciousAssignments) Semantic() {}

func (suspiciousAssignments) ID() string { return "SuspiciousAssignments" }

func (suspiciousAssignments) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KSwitch, syntax.KAssign, syntax.KFunction, syntax.KMethod, syntax.KClosure}
}

func (r suspiciousAssignments) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Switch:
		saSwitch(ctx, x)
	case *syntax.Assign:
		saCompound(ctx, x)
		saTypo(ctx, x)
		saOverwrite(ctx, x)
		saDestructuring(ctx, x)
	default:
		saParams(ctx, n)
	}
}

func saIsDestructuring(a *syntax.Assign) bool {
	if a.Op.Kind != syntax.TEqual {
		return false
	}
	switch a.Var.(type) {
	case *syntax.Array, *syntax.List:
		return true
	}
	return false
}

func saIsPlain(a *syntax.Assign) bool { return a.Op.Kind == syntax.TEqual && !saIsDestructuring(a) }

// saStmtAssign returns the assignment of an expression statement.
func saStmtAssign(s syntax.Stmt) *syntax.Assign {
	es, ok := s.(*syntax.ExprStmt)
	if !ok {
		return nil
	}
	a, _ := es.Expr.(*syntax.Assign)
	return a
}

// saContainsEquivalent reports whether any node under root (root included
// when withRoot) is equivalent to t.
func saContainsEquivalent(f *syntax.File, root, t syntax.Node, withRoot bool) bool {
	found := false
	syntax.Inspect(root, func(n syntax.Node) bool {
		if found {
			return false
		}
		if (withRoot || n != root) && n.Kind() == t.Kind() && util.EquivalentFoldNames(f, n, t) {
			found = true
		}
		return !found
	})
	return found
}

func saTerminates(s syntax.Stmt, withGoto bool) bool {
	switch s := s.(type) {
	case *syntax.Break, *syntax.Return, *syntax.Continue:
		return true
	case *syntax.Goto:
		return withGoto
	case *syntax.ExprStmt:
		switch s.Expr.(type) {
		case *syntax.Throw, *syntax.Exit:
			return true
		}
	}
	return false
}

// ---- A. switch fall-through -------------------------------------------------------------

const saSwitchMsg = "This write overwrites a value set in a previous case; a 'break' may be missing."

func saSwitch(ctx *analysis.Context, sw *syntax.Switch) {
	var written []syntax.Node
	inW := func(t syntax.Node) bool {
		for _, w := range written {
			if util.EquivalentFoldNames(ctx.File, w, t) {
				return true
			}
		}
		return false
	}
	for _, c := range sw.Cases {
		if len(c.Stmts) == 0 { // D1
			continue
		}
		var local []syntax.Node
		for _, s := range c.Stmts { // D2
			a := saStmtAssign(s)
			if a == nil {
				continue
			}
			if saIsDestructuring(a) { // D2a
				for _, it := range saListItems(a.Var) {
					if inW(it) {
						ctx.ReportNode(it, saSwitchMsg)
					} else {
						local = append(local, it)
					}
				}
				continue
			}
			if !saIsPlain(a) { // D2b
				continue
			}
			if d, ok := a.Var.(*syntax.ArrayDimFetch); ok && d.Dim == nil {
				continue
			}
			if saContainsEquivalent(ctx.File, a.Value, a.Var, true) {
				continue
			}
			if inW(a.Var) {
				ctx.ReportNode(a.Var, saSwitchMsg)
			} else {
				local = append(local, a.Var)
			}
		}
		for _, l := range local { // D3
			if !inW(l) {
				written = append(written, l)
			}
		}
		if saTerminates(c.Stmts[len(c.Stmts)-1], true) { // D4
			written = nil
		}
	}
}

func saListItems(e syntax.Expr) []syntax.Node {
	var items []*syntax.ArrayItem
	switch l := e.(type) {
	case *syntax.Array:
		items = l.Items
	case *syntax.List:
		items = l.Items
	}
	var out []syntax.Node
	for _, it := range items {
		if it == nil || it.Value == nil {
			continue
		}
		switch it.Value.(type) {
		case *syntax.Array, *syntax.List:
			continue
		}
		out = append(out, it.Value)
	}
	return out
}

// ---- B. self-referencing compound assignment ---------------------------------------------

var saCompoundOps = map[syntax.TokenKind]syntax.TokenKind{
	syntax.TPlusEqual: syntax.TPlus, syntax.TMinusEqual: syntax.TMinus, syntax.TMulEqual: syntax.TMul,
	syntax.TDivEqual: syntax.TDiv, syntax.TModEqual: syntax.TMod, syntax.TConcatEqual: syntax.TDot,
	syntax.TAndEqual: syntax.TAmpersand, syntax.TOrEqual: syntax.TBar, syntax.TXorEqual: syntax.TCaret,
	syntax.TSlEqual: syntax.TSl, syntax.TSrEqual: syntax.TSr,
}

func saCompound(ctx *analysis.Context, a *syntax.Assign) {
	op, ok := saCompoundOps[a.Op.Kind] // D5
	if !ok {
		return
	}
	b, ok := a.Value.(*syntax.Binary) // D6
	if !ok || b.Op.Kind != op || !util.EquivalentFoldNames(ctx.File, b.Left, a.Var) {
		return
	}
	ctx.ReportNode(a, "The target is repeated on the right-hand side of the compound assignment; likely a merge mistake.")
}

// ---- C. parameter overwritten before use ---------------------------------------------------

func saParams(ctx *analysis.Context, fn syntax.Node) {
	body := syntax.FuncLikeBody(fn)
	params := syntax.FuncLikeParams(fn)
	if body == nil || len(params) == 0 || len(body.Stmts) == 0 { // D7
		return
	}
	if util.InTestContext(ctx, fn) { // E6
		return
	}
	// D8 for every parameter in one walk (a walk per parameter was
	// quadratic); only the first two accesses of each matter.
	names := map[string]bool{}
	for _, p := range params {
		if !p.ByRef && p.Var != nil && p.Var.Name != "" {
			names[p.Var.Name] = true
		}
	}
	firstTwo := map[string][]*syntax.Variable{}
	saFlowVars(body, func(v *syntax.Variable) {
		if names[v.Name] && len(firstTwo[v.Name]) < 2 {
			firstTwo[v.Name] = append(firstTwo[v.Name], v)
		}
	})
	for _, p := range params {
		if p.ByRef || p.Var == nil || p.Var.Name == "" {
			continue
		}
		name := p.Var.Name
		accesses := firstTwo[name]
		if len(accesses) < 2 {
			continue
		}
		first := accesses[0]
		a, ok := first.Parent().(*syntax.Assign) // D9
		if !ok || a.Var != syntax.Expr(first) || !saIsPlain(a) {
			continue
		}
		es, ok := a.Parent().(*syntax.ExprStmt)
		if !ok || es.Parent() != syntax.Node(body) {
			continue
		}
		count := 0 // D10
		syntax.Inspect(a, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name {
				count++
			}
			return true
		})
		if count == 1 {
			ctx.ReportNode(first, "Parameter is overwritten before its value is used.")
		}
	}
}

// saFlowVars calls fn for the simple variable accesses under n in
// approximate evaluation order: assignment right-hand sides before targets,
// nested function bodies skipped.
func saFlowVars(n syntax.Node, fn func(*syntax.Variable)) {
	switch x := n.(type) {
	case *syntax.Variable:
		if x.NameExpr == nil && x.Name != "" && !util.IsStaticPropName(x) {
			fn(x)
		}
		if x.NameExpr != nil {
			saFlowVars(x.NameExpr, fn)
		}
		return
	case *syntax.Assign:
		saFlowVars(x.Value, fn)
		saFlowVars(x.Var, fn)
		return
	case *syntax.Closure:
		for _, u := range x.Uses {
			saFlowVars(u, fn)
		}
		return
	case *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
		return
	}
	syntax.Children(n, func(c syntax.Node) { saFlowVars(c, fn) })
}

// ---- D. =+ / =- / =! typo ---------------------------------------------------------------

func saTypo(ctx *analysis.Context, a *syntax.Assign) {
	if !saIsPlain(a) || a.ByRef { // D11
		return
	}
	u, ok := a.Value.(*syntax.Unary)
	if !ok {
		return
	}
	switch u.Op.Kind {
	case syntax.TPlus, syntax.TMinus, syntax.TExclaim:
	default:
		return
	}
	if a.Op.Span.End != u.Op.Span.Start || int(u.Op.Span.End) >= len(ctx.Src) {
		return
	}
	switch ctx.Src[u.Op.Span.End] {
	case ' ', '\t', '\n', '\r':
	default:
		return
	}
	ctx.ReportNode(a, "Did you mean '"+ctx.SpanText(u.Op.Span)+"='? Fix the operator or the spacing.")
}

// ---- E. value overwritten immediately ------------------------------------------------------

func saOverwrite(ctx *analysis.Context, a *syntax.Assign) {
	if !saIsPlain(a) { // D12
		return
	}
	stmt, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	t := a.Var
	for e := t; ; { // D13
		d, ok := e.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		if d.Dim == nil {
			return
		}
		switch k := syntax.UnwrapParens(d.Dim).(type) {
		case *syntax.Literal:
		case *syntax.MagicConst:
			return
		case *syntax.Unary:
			if l, ok := k.Expr.(*syntax.Literal); !ok || k.Op.Kind != syntax.TMinus || l.LitKind == syntax.LitString {
				return
			}
		default:
			return
		}
		e = d.Var
	}
	// D14: the right-hand side (itself included) must not read T.
	if saContainsEquivalent(ctx.File, a.Value, t, true) {
		return
	}
	seen := map[string]bool{}
	dup := false
	syntax.Inspect(t, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && n != syntax.Node(t) && v.NameExpr == nil && v.Name != "this" {
			if seen[v.Name] {
				dup = true
			}
			seen[v.Name] = true
		}
		return !dup
	})
	if dup {
		return
	}
	prev, ok := util.PrevStmt(ctx.File, stmt) // D15
	if !ok {
		return
	}
	target := ctx.Text(t)
	if ifs, ok := prev.(*syntax.If); ok { // D16
		if len(ifs.ElseIfs) > 0 || ifs.Else != nil {
			return
		}
		blk, ok := ifs.Body.(*syntax.Block)
		if !ok || blk.Alt || len(blk.Stmts) == 0 || saTerminates(blk.Stmts[len(blk.Stmts)-1], false) {
			return
		}
		idx := -1
		for i, s := range blk.Stmts {
			if pa := saStmtAssign(s); pa != nil && saIsPlain(pa) && util.EquivalentFoldNames(ctx.File, pa.Var, t) {
				idx = i
				break
			}
		}
		if idx < 0 {
			return
		}
		if idx+1 < len(blk.Stmts) {
			var region syntax.Node = blk.Stmts[idx+1]
			if si, ok := region.(*syntax.If); ok {
				region = si.Cond
			}
			if region != nil && saContainsEquivalent(ctx.File, region, t, true) {
				return
			}
		}
		ctx.ReportNode(a, target+" is overwritten right after the 'if'; an 'else' may be missing.")
		return
	}
	pa := saStmtAssign(prev) // D17
	if pa == nil || !saIsPlain(pa) || !util.EquivalentFoldNames(ctx.File, pa.Var, t) || pa.ByRef {
		return
	}
	if blk, ok := prev.Parent().(*syntax.Block); ok {
		if tr, ok := blk.Parent().(*syntax.Try); ok && tr.Body == blk {
			return
		}
	}
	incdec := false
	syntax.Inspect(t, func(n syntax.Node) bool {
		if _, ok := n.(*syntax.IncDec); ok {
			incdec = true
		}
		return !incdec
	})
	if incdec {
		return
	}
	ctx.ReportNode(t, target+" is overwritten right after being assigned.")
}

// ---- F. destructuring a non-array --------------------------------------------------------

func saDestructuring(ctx *analysis.Context, a *syntax.Assign) {
	if !saIsDestructuring(a) { // D18
		return
	}
	if _, ok := a.Parent().(*syntax.ExprStmt); !ok {
		return
	}
	t := ctx.TypeOf(a.Value) // D19
	if t.IsUnknown() {
		return
	}
	// D20a: null and false are failure markers, ignored next to a
	// supporting type.
	supporting, bad, onlyMarkers := false, false, true
	for _, at := range t.Atoms() { // D20
		switch {
		case at == "array" || at == "mixed" || strings.HasSuffix(at, "[]"):
			supporting = true
		case strings.HasPrefix(at, `\`) && saArrayAccess(ctx, strings.TrimPrefix(at, `\`)):
			supporting = true
		case at == "null" || at == "false":
			bad = true
		default:
			bad = true
			onlyMarkers = false
		}
	}
	if !bad || (supporting && onlyMarkers) {
		return
	}
	ctx.ReportNode(a, "Destructuring a value that is not an array.")
}

func saArrayAccess(ctx *analysis.Context, cls string) bool {
	if strings.EqualFold(cls, "ArrayAccess") {
		return true
	}
	return ctx.Index().IsSubtype(cls, "ArrayAccess", ctx.PHP)
}
