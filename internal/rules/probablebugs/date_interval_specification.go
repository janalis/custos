package probablebugs

import (
	"regexp"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// dateIntervalSpecification validates literal DateInterval specifications.
type dateIntervalSpecification struct{}

func init() { register(dateIntervalSpecification{}) }

func (dateIntervalSpecification) ID() string { return "DateIntervalSpecification" }

func (dateIntervalSpecification) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }

var (
	// R1 without the `T(?=\d)` lookahead, checked separately.
	dateIntervalR1 = regexp.MustCompile(`^P(([0-9]+Y)?([0-9]+M)?([0-9]+D)?([0-9]+W)?)?(T([0-9]+H)?([0-9]+M)?([0-9]+S)?)?$`)
	dateIntervalR2 = regexp.MustCompile(`^P[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}$`)
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

func isStringLiteralNode(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Literal:
		return x.LitKind == syntax.LitString
	case *syntax.InterpolatedString:
		return !x.Backtick
	}
	return false
}

func (dateIntervalSpecification) Check(ctx *analysis.Context, n syntax.Node) {
	nw := n.(*syntax.New)
	lit, discovered := dateIntervalLiteral(ctx, nw)
	if lit == nil {
		return
	}
	content, _, ok := util.QuotedStringRaw(lit) // D4
	if !ok {
		return
	}
	if validDateIntervalSpec(content) { // D5
		return
	}
	if discovered && dateIntervalReportedEarlier(ctx, nw, lit) { // D6
		return
	}
	ctx.ReportNode(lit, "Malformed DateInterval specification.")
}

// dateIntervalLiteral implements D1-D3: the specification literal of a
// `new DateInterval(...)`, and whether it was found by value discovery.
func dateIntervalLiteral(ctx *analysis.Context, nw *syntax.New) (syntax.Expr, bool) {
	args := semArgs(nw.Args)
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
	if isStringLiteralNode(arg) { // D3
		return arg, false
	}
	var lits []syntax.Expr
	for _, v := range util.DiscoverValues(ctx.Types(), arg) {
		if isStringLiteralNode(v) {
			lits = append(lits, v)
		}
	}
	if len(lits) != 1 {
		return nil, false
	}
	return lits[0], true
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
