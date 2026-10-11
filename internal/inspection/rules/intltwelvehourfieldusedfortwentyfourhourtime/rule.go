// Package intltwelvehourfieldusedfortwentyfourhourtime implements the native IntlTwelveHourFieldUsedForTwentyFourHourTime inspection.
package intltwelvehourfieldusedfortwentyfourhourtime

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the 24-hour calendar field."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntlTwelveHourFieldUsedForTwentyFourHourTime" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "IntlCalendar", "set") {
		return
	}
	field := semanticquery.CallArgument(c.Args, 0, "field")
	hour := semanticquery.CallArgument(c.Args, 1, "value")
	value, ok := semanticquery.NativeInt(ctx, hour)
	if !ok || rewrittenOutsideAssignments(ctx.File, hour) || value < 12 || value > 23 || !semanticquery.ExpansionCConstant(ctx, field, "IntlCalendar", "FIELD_HOUR") {
		return
	}
	constant := syntax.UnwrapParens(field).(*syntax.ClassConstFetch)
	ctx.ReportNode(c, message, astquery.ReplaceFix(constant.Name.Span(), "FIELD_HOUR_OF_DAY"))
}

// rewrittenOutsideAssignments reports whether the local variable e is changed
// in its scope by something other than an assignment: `++`/`--`, a catch
// binding or a closure importing it by reference. The resolved assignment
// then no longer proves the value at the call.
func rewrittenOutsideAssignments(f *syntax.File, e syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		return false
	}
	named := func(x syntax.Expr) bool {
		t, ok := syntax.UnwrapParens(x).(*syntax.Variable)
		return ok && t.NameExpr == nil && t.Name == v.Name
	}
	found := false
	visit := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.IncDec:
			found = found || named(x.Var)
		case *syntax.Catch:
			found = found || x.Var != nil && named(x.Var)
		case *syntax.Closure:
			for _, use := range x.Uses {
				found = found || use.ByRef && named(use.Var)
			}
		}
		// Nested scopes own their variables; only a by-reference import shares one.
		return !found && !syntax.IsVariableScope(n)
	}
	if scope := syntax.EnclosingVariableScope(v); scope != nil {
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	} else {
		syntax.InspectFile(f, visit)
	}
	return found
}
