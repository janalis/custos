package multiassignmentusage

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// multiAssignmentUsage reports destructuring of a foreach variable in a
// separate statement, and consecutive numbered reads from the same array.
type multiAssignmentUsage struct{}

func (multiAssignmentUsage) ID() string               { return "MultiAssignmentUsage" }
func (multiAssignmentUsage) Semantic()                {}
func (multiAssignmentUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (multiAssignmentUsage) Check(ctx *analysis.Context, n syntax.Node) {
	as := n.(*syntax.Assign)
	stmt, ok := as.Parent().(*syntax.ExprStmt)
	if !ok || as.Var == nil || as.Value == nil {
		return
	}
	switch target := as.Var.(type) {
	case *syntax.List, *syntax.Array:
		if a, isArr := target.(*syntax.Array); isArr && !a.Short {
			return
		}
		if ctx.PHP < phpversion.PHP55 { // D1 / E1
			return
		}
		v, ok := as.Value.(*syntax.Variable) // D3
		if !ok || v.NameExpr != nil {
			return
		}
		fe := declaredByDirectForeach(stmt, v.Name)                                  // D4
		if fe != nil && targetsFitHeader(as.Var, fe) && !usedElsewhere(ctx, fe, v) { // D4b, D4c
			ctx.ReportNode(as, "Destructure directly in the foreach header.")
		}
	case *syntax.Variable:
		if as.Op.Kind != syntax.TEqual || as.ByRef { // D5; custos: `$q =& $m[6]` keeps a reference
			return
		}
		base, key, ok := numberedRead(as.Value) // D7
		if !ok {
			return
		}
		prev, ok := astquery.PrevStmt(ctx.File, stmt) // D6
		if !ok {
			return
		}
		ps, ok := prev.(*syntax.ExprStmt)
		if !ok {
			return
		}
		pa, ok := ps.Expr.(*syntax.Assign)
		if !ok || pa.Op.Kind != syntax.TEqual || pa.ByRef {
			return
		}
		pbase, pkey, ok := numberedRead(pa.Value)
		if !ok || pkey == key || !sameTextIgnoringSpace(ctx.Text(base), ctx.Text(pbase)) {
			return
		}
		// D7b: the base must denote the same array in both statements and be
		// evaluated without side effects.
		if baseHasSideEffects(base) || writesIntoBase(ctx, pa.Var, base) {
			return
		}
		// E8 (custos): `[$a, $b] = $s` assigns null from a string, where
		// `$s[0]` reads its first byte.
		if ctx.TypeOf(base).Has("string") {
			return
		}
		ctx.ReportNode(as, "Use one destructuring assignment from '"+ctx.Text(base)+"' instead.") // D8
	}
}

// numberedRead returns X and the effective array key when e is `X[N]` with
// N a (possibly negated) number literal. Float keys truncate to int, as PHP
// does; ok is false for other shapes and unparsable literals.
func numberedRead(e syntax.Expr) (base syntax.Expr, key int64, ok bool) {
	d, isDim := e.(*syntax.ArrayDimFetch)
	if !isDim || d.Dim == nil || d.Var == nil {
		return nil, 0, false
	}
	idx, neg := d.Dim, false
	if u, isUn := idx.(*syntax.Unary); isUn && u.Op.Kind == syntax.TMinus {
		idx, neg = u.Expr, true
	}
	lit, isLit := idx.(*syntax.Literal)
	if !isLit {
		return nil, 0, false
	}
	raw := strings.ReplaceAll(lit.Raw, "_", "")
	switch lit.LitKind {
	case syntax.LitInt:
		if len(raw) > 1 && raw[0] == '0' && raw[1] >= '0' && raw[1] <= '9' {
			raw = "0o" + raw[1:] // legacy octal
		}
		v, err := strconv.ParseInt(raw, 0, 64)
		if err != nil {
			return nil, 0, false
		}
		key = v
	case syntax.LitFloat:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil || math.IsInf(f, 0) || math.IsNaN(f) || math.Abs(f) >= 1<<62 {
			return nil, 0, false
		}
		key = int64(f)
	default:
		return nil, 0, false
	}
	if neg {
		key = -key
	}
	return d.Var, key, true
}

func sameTextIgnoringSpace(a, b string) bool {
	i, j := 0, 0
	for {
		for i < len(a) && astquery.IsSpace(a[i]) {
			i++
		}
		for j < len(b) && astquery.IsSpace(b[j]) {
			j++
		}
		if i == len(a) || j == len(b) {
			return i == len(a) && j == len(b)
		}
		if a[i] != b[j] {
			return false
		}
		i++
		j++
	}
}

// declaredByDirectForeach reports whether stmt is a direct statement of a
// foreach body (braced, alternative syntax or brace-less) whose header
// declares $name as its by-value value (or inside a destructuring value):
// the key cannot be a list, and a by-reference value cannot become one.
func declaredByDirectForeach(stmt syntax.Stmt, name string) *syntax.Foreach {
	var fe *syntax.Foreach
	switch p := stmt.Parent().(type) {
	case *syntax.Foreach:
		if p.Body == stmt {
			fe = p
		}
	case *syntax.Block:
		if f, ok := p.Parent().(*syntax.Foreach); ok && f.Body == syntax.Stmt(p) {
			fe = f
		}
	}
	if fe != nil && !fe.ByRef && declaresVar(fe.Value, name) {
		return fe
	}
	return nil
}

// targetsFitHeader implements D4b: the header assigns the value before the
// key and its own targets left to right, so a target that mentions a
// variable declared by the header (`[$out[$k], $b] = $row`) would read it
// too early there; a by-reference target would turn the loop into a
// by-reference iteration.
func targetsFitHeader(target syntax.Expr, fe *syntax.Foreach) bool {
	declared := map[string]bool{}
	collect := func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok {
			declared[v.Name] = true
		}
		return true
	}
	if fe.Key != nil {
		syntax.Inspect(fe.Key, collect)
	}
	syntax.Inspect(fe.Value, collect)
	fits := true
	syntax.Inspect(target, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.ArrayItem:
			fits = fits && !n.ByRef
		case *syntax.Variable:
			fits = fits && n.NameExpr == nil && !declared[n.Name]
		}
		return fits
	})
	return fits
}

// usedElsewhere implements D4c: the header form no longer assigns $v, so
// any other mention of it that could observe the loop's value blocks the
// report. That is every mention in the loop body, every mention after the
// loop, and, when the loop is nested in another loop, every mention in the
// outermost of those loops; plus anything that makes the variable outlive
// the function or reads variables by name.
func usedElsewhere(ctx *analysis.Context, fe *syntax.Foreach, v *syntax.Variable) bool {
	var scope syntax.Node
	region := fe.Span().Start
	for p := fe.Parent(); p != nil && scope == nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure:
			scope = p
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			region = p.Span().Start
		}
	}
	found := false
	var visit func(syntax.Node) bool
	visit = func(n syntax.Node) bool {
		if found {
			return false
		}
		start := n.Span().Start
		switch n := n.(type) {
		case *syntax.Foreach:
			// Another loop over the same name assigns it before its body
			// runs: only its subject can see the earlier value.
			if n != fe && !n.Span().Contains(fe.Span()) && (isVar(n.Value, v.Name) || isVar(n.Key, v.Name)) {
				syntax.Inspect(n.Expr, visit)
				return false
			}
		case *syntax.Variable:
			switch {
			case n.NameExpr != nil:
				found = start >= region
			case n.Name != v.Name || n == v:
			case start >= region:
				found = start < fe.Span().Start || start >= fe.Body.Span().Start
			default:
				found = outlivesScope(n.Parent())
			}
		case *syntax.FuncCall:
			_, part, _ := astquery.FuncNamePart(n)
			switch strings.ToLower(part) {
			case "compact", "extract", "get_defined_vars":
				found = start >= region
			}
		case *syntax.Eval, *syntax.Include:
			found = start >= region
		}
		return !found
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}

func isVar(e syntax.Expr, name string) bool {
	v, ok := e.(*syntax.Variable)
	return ok && v.NameExpr == nil && v.Name == name
}

// outlivesScope reports a declaration that keeps the variable observable
// after the function returns: `global`, `static`, a by-reference parameter
// or a by-reference closure binding.
func outlivesScope(decl syntax.Node) bool {
	switch d := decl.(type) {
	case *syntax.Global, *syntax.StaticVar:
		return true
	case *syntax.Param:
		return d.ByRef
	case *syntax.ClosureUse:
		return d.ByRef
	}
	return false
}

func declaresVar(e syntax.Expr, name string) bool {
	switch e := e.(type) {
	case *syntax.Variable:
		return e.NameExpr == nil && e.Name == name
	case *syntax.Array:
		return itemsDeclare(e.Items, name)
	case *syntax.List:
		return itemsDeclare(e.Items, name)
	}
	return false
}

func itemsDeclare(items []*syntax.ArrayItem, name string) bool {
	for _, it := range items {
		if it != nil && it.Value != nil && declaresVar(it.Value, name) {
			return true
		}
	}
	return false
}

// baseHasSideEffects reports a base expression containing a call or another
// expression with effects: destructuring would evaluate it once instead of
// once per statement.
func baseHasSideEffects(base syntax.Expr) bool {
	found := false
	syntax.Inspect(base, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New,
			*syntax.Assign, *syntax.IncDec, *syntax.Include, *syntax.Eval,
			*syntax.Yield, *syntax.YieldFrom, *syntax.Exit, *syntax.Print:
			found = true
		}
		return !found
	})
	return found
}

// writesIntoBase reports whether the previous assignment's target (array
// element writes stripped: `$x[1]` writes `$x`) occurs inside base, i.e.
// the previous statement changes what the base denotes (`$row = $row[0];
// $b = $row[1];`).
func writesIntoBase(ctx *analysis.Context, target, base syntax.Expr) bool {
	for {
		d, ok := target.(*syntax.ArrayDimFetch)
		if !ok || d.Var == nil {
			break
		}
		target = d.Var
	}
	want := ctx.Text(target)
	found := false
	syntax.Inspect(base, func(n syntax.Node) bool {
		if n.Kind() == target.Kind() && sameTextIgnoringSpace(ctx.Text(n), want) {
			found = true
		}
		return !found
	})
	return found
}
