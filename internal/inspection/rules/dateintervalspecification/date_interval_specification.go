package dateintervalspecification

import (
	"regexp"
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// dateIntervalSpecification validates literal DateInterval specifications.
type dateIntervalSpecification struct{}

func (dateIntervalSpecification) ID() string               { return "DateIntervalSpecification" }
func (dateIntervalSpecification) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }

var (
	// R1 without the `T(?=\d)` lookahead, checked separately.
	dateIntervalR1 = regexp.MustCompile(`^P(([0-9]+Y)?([0-9]+M)?([0-9]+W)?([0-9]+D)?)?(T([0-9]+H)?([0-9]+M)?([0-9]+S)?)?$`)
	dateIntervalR2 = regexp.MustCompile(`^P\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`)
)

func validDateIntervalSpec(s string) bool {
	if dateIntervalR2.MatchString(s) {
		return true
	}
	if !dateIntervalR1.MatchString(s) {
		return false
	}
	if i := strings.IndexByte(s, 'T'); i >= 0 && (i+1 >= len(s) || s[i+1] < '0' || s[i+1] > '9') {
		return false
	}
	return true
}

func (dateIntervalSpecification) Check(ctx *analysis.Context, n syntax.Node) {
	nw := n.(*syntax.New)
	lit, discovered := dateIntervalLiteral(ctx, nw)
	if lit == nil {
		return
	}
	content, _, ok := astquery.QuotedStringRaw(lit) // D4
	if !ok {
		return
	}
	msg := "Malformed DateInterval specification."
	if validDateIntervalSpec(content) { // D5
		if ctx.PHP >= phpversion.PHP80 || !weeksWithDays(content) { // D7
			return
		}
		msg = "Weeks and days cannot be combined in a DateInterval specification before PHP 8.0."
	}
	if discovered && dateIntervalReportedEarlier(ctx, nw, lit) { // D6
		return
	}
	ctx.ReportNode(lit, msg)
}

// weeksWithDays reports whether a valid R1 specification has both a week
// and a day part (D7): before PHP 8.0 the weeks are silently discarded.
func weeksWithDays(s string) bool {
	date, _, _ := strings.Cut(s, "T")
	return strings.Contains(date, "W") && strings.Contains(date, "D")
}

// dateIntervalLiteral implements D1-D3: the specification literal of a
// `new DateInterval(...)`, and whether it was found by value discovery.
func dateIntervalLiteral(ctx *analysis.Context, nw *syntax.New) (syntax.Expr, bool) {
	args := astquery.ArgumentNodes(nw.Args)
	if len(args) != 1 { // D1
		return nil, false
	}
	name, ok := nw.Class.(*syntax.Name)
	if !ok || !strings.EqualFold(ctx.Names().Class(name.Value, name.Span().Start), "DateInterval") { // D2 (class names are case-insensitive)
		return nil, false
	}
	arg := semArgValue(args[0])
	if arg == nil {
		return nil, false
	}
	if astquery.IsStringLiteral(arg) { // D3
		return arg, false
	}
	var lits []syntax.Expr
	vals := semanticquery.DiscoverValues(ctx.Types(), arg)
	for _, v := range vals {
		if astquery.IsStringLiteral(v) {
			lits = append(lits, v)
		}
	}
	// custos: another (non-literal) value may be the one that reaches the
	// constructor (`$v = $q['v'] ?? ''; if ($v === '') throw …;`).
	if len(lits) != 1 || len(vals) != 1 {
		return nil, false
	}
	// custos: discovery only follows plain assignments; a value it cannot
	// see (caller-supplied, written through a reference, ...) may be the
	// one used.
	if !specSettled(ctx, arg, map[syntax.Node]bool{}) {
		return nil, false
	}
	return lits[0], true
}

// specSettled reports whether the values discovery collected from e are
// the only ones e can hold: every variable on the way is a local written
// by plain assignments only (D3). Each expression is visited once.
func specSettled(ctx *analysis.Context, e syntax.Expr, seen map[syntax.Node]bool) bool {
	e = syntax.UnwrapParens(e)
	if seen[e] {
		return true
	}
	seen[e] = true
	switch x := e.(type) {
	case *syntax.Ternary:
		first := x.Then
		if first == nil {
			first = x.Cond
		}
		return specSettled(ctx, first, seen) && specSettled(ctx, x.Else, seen)
	case *syntax.Binary:
		if x.Op.Kind == syntax.TCoalesce {
			return specSettled(ctx, x.Left, seen) && specSettled(ctx, x.Right, seen)
		}
	case *syntax.Variable:
		return variableSettled(ctx, x, seen)
	}
	return true
}

// variableSettled reports whether the local v only ever holds the values
// of its plain `=` assignments, themselves settled.
func variableSettled(ctx *analysis.Context, v *syntax.Variable, seen map[syntax.Node]bool) bool {
	scope := syntax.EnclosingFuncLike(v)
	if entersScope(scope, v.Name) || callsExtract(scope) {
		return false
	}
	assigned := false
	for _, acc := range flowquery.VarAccesses(ctx.File, scope, v.Name) {
		if !acc.Write {
			if acc.ElemWrite || boundByReference(ctx, acc) {
				return false
			}
			continue
		}
		a, ok := acc.By.(*syntax.Assign)
		if !ok || a.Op.Kind != syntax.TEqual || a.ByRef || syntax.UnwrapParens(a.Var) != syntax.Expr(acc.Var) {
			return false // foreach target, destructuring, reference, global/static, catch, unset...
		}
		if !specSettled(ctx, astquery.AssignedValue(a), seen) {
			return false
		}
		assigned = true
	}
	return assigned
}

// entersScope reports whether name holds a value on entry of scope: a
// parameter (its default is only one of the values callers pass) or a
// variable captured by a closure.
func entersScope(scope syntax.Node, name string) bool {
	params, _ := astquery.ScopeParts(scope)
	for _, p := range params {
		if p.Var != nil && p.Var.Name == name {
			return true
		}
	}
	if c, ok := scope.(*syntax.Closure); ok {
		for _, u := range c.Uses {
			if u.Var != nil && u.Var.Name == name {
				return true
			}
		}
	}
	return false
}

// callsExtract reports whether scope itself calls extract(), which can
// write any of its variables.
func callsExtract(scope syntax.Node) bool {
	found := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			_, name, _ := astquery.CallName(call)
			found = found || strings.EqualFold(name, "extract")
		}
		return !found && (n == scope || !syntax.IsFuncLike(n))
	})
	return found
}

// boundByReference reports whether a read of a variable hands out a
// reference to it: `$r = &$v`, `[&$v]`, `use (&$v)`, a call-time `&$v` or
// an argument bound to a by-reference parameter of a resolved callee.
func boundByReference(ctx *analysis.Context, acc flowquery.VarAccess) bool {
	if u, ok := acc.By.(*syntax.ClosureUse); ok {
		return u.ByRef
	}
	switch p := acc.Var.Parent().(type) {
	case *syntax.Assign:
		return p.ByRef
	case *syntax.ArrayItem:
		return p.ByRef
	case *syntax.Arg:
		return p.ByRef || argBindsByRef(ctx, p)
	}
	return false
}

// argBindsByRef reports whether arg binds to a by-reference parameter of
// its resolved callee (named, positional or variadic tail). An unresolved
// callee binds nothing: most calls take their arguments by value.
func argBindsByRef(ctx *analysis.Context, arg *syntax.Arg) bool {
	list := arg.Parent().(*syntax.ArgList)
	callee, _ := semanticquery.ResolveCallee(ctx, list.Parent())
	params := callee.Params
	if arg.Name != nil {
		return slices.ContainsFunc(params, func(p index.Param) bool {
			return p.ByRef && strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value)
		})
	}
	i := slices.Index(list.Args, syntax.Expr(arg))
	if n := len(params); i >= n && n > 0 && params[n-1].Variadic {
		i = n - 1
	}
	return i < len(params) && params[i].ByRef
}

// dateIntervalReportedEarlier reports whether a `new DateInterval(...)`
// located before nw in the file resolves to the same literal, which was
// then already reported there (D6). Only runs for discovered literals.
func dateIntervalReportedEarlier(ctx *analysis.Context, nw *syntax.New, lit syntax.Expr) bool {
	found := false
	syntax.InspectFile(ctx.File, func(x syntax.Node) bool {
		if found || x.Span().Start >= nw.Span().Start {
			return false
		}
		if other, ok := x.(*syntax.New); ok {
			if l, _ := dateIntervalLiteral(ctx, other); l == lit {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// Semantic marks the rule as needing the project index.
func (dateIntervalSpecification) Semantic() {}
