package controlflow

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// multiAssignmentUsage reports destructuring of a foreach variable in a
// separate statement, and consecutive numbered reads from the same array.
type multiAssignmentUsage struct{}

func init() { register(multiAssignmentUsage{}) }

func (multiAssignmentUsage) ID() string { return "MultiAssignmentUsage" }

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
		if ctx.PHP < phpver.PHP55 { // D1 / E1
			return
		}
		v, ok := as.Value.(*syntax.Variable) // D3
		if !ok || v.NameExpr != nil {
			return
		}
		if declaredByDirectForeach(stmt, v.Name) { // D4
			ctx.ReportNode(as, "Destructure directly in the foreach header.")
		}
	case *syntax.Variable:
		if as.Op.Kind != syntax.TEqual { // D5
			return
		}
		base, key, ok := numberedRead(as.Value) // D7
		if !ok {
			return
		}
		prev, ok := util.PrevStmt(ctx.File, stmt) // D6
		if !ok {
			return
		}
		ps, ok := prev.(*syntax.ExprStmt)
		if !ok {
			return
		}
		pa, ok := ps.Expr.(*syntax.Assign)
		if !ok || pa.Op.Kind != syntax.TEqual {
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
		for i < len(a) && isSpaceByte(a[i]) {
			i++
		}
		for j < len(b) && isSpaceByte(b[j]) {
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

func isSpaceByte(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

// declaredByDirectForeach reports whether stmt is a direct statement of a
// foreach body (braced, alternative syntax or brace-less) whose header
// declares $name.
func declaredByDirectForeach(stmt syntax.Stmt, name string) bool {
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
	return fe != nil && (declaresVar(fe.Key, name) || declaresVar(fe.Value, name))
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
