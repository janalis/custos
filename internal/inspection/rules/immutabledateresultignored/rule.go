// Package immutabledateresultignored implements the ImmutableDateResultIgnored inspection.
package immutabledateresultignored

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ImmutableDateResultIgnored" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }

const message = "Assign the returned immutable date."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	call := n.(*syntax.MethodCall)
	if semanticquery.NativeDateClass(ctx, call.Var) != "DateTimeImmutable" {
		return
	}
	name, ok := call.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	switch strings.ToLower(name.Value) {
	case "modify", "add", "sub", "setdate", "settime", "settimestamp", "settimezone":
	default:
		return
	}
	if _, ok := call.Parent().(*syntax.ExprStmt); !ok {
		return
	}
	if freshLocalRead(ctx, call) {
		ctx.ReportNode(call, message, astquery.ReplaceFix(call.Span(), ctx.Text(call.Var)+" = "+ctx.Text(call)))
		return
	}
	ctx.ReportNode(call, message)
}

func (rule) Semantic() {}

// freshLocalRead limits assignment to a fresh, unaliased receiver and an
// adjacent formatting read. Unknown statements, captures and arguments cannot
// prove that changing the receiver is intended.
func freshLocalRead(ctx *analysis.Context, call *syntax.MethodCall) bool {
	v, ok := call.Var.(*syntax.Variable)
	if !ok || v.Name == "" || v.Name == "this" {
		return false
	}
	switch v.Name {
	case "GLOBALS", "_SERVER", "_GET", "_POST", "_FILES", "_COOKIE", "_SESSION", "_REQUEST", "_ENV":
		return false
	}
	if !strings.EqualFold(call.Name.(*syntax.Identifier).Value, "modify") {
		return false
	}
	arg := singleStringArgument(call.Args)
	text, ok := astquery.QuotedStringValue(arg)
	if !ok || !validRelativeDate(text) {
		return false
	}
	st := call.Parent().(*syntax.ExprStmt)
	list, i, ok := astquery.StmtList(ctx.File, st)
	if !ok || i != 1 || len(list) < 3 {
		return false
	}
	// Only whole local bodies or the file/namespace top level qualify;
	// loop and conditional bodies can inherit aliases from earlier paths.
	switch parent := st.Parent().(type) {
	case nil, *syntax.Namespace:
	case *syntax.Block:
		switch scope := parent.Parent().(type) {
		case *syntax.Function, *syntax.Method:
			for _, p := range syntax.VariableScopeParams(scope) {
				if p.Var.Name == v.Name {
					return false
				}
			}
		default:
			return false
		}
	default:
		return false
	}
	previous, ok := list[0].(*syntax.ExprStmt)
	if !ok {
		return false
	}
	assignment, ok := previous.Expr.(*syntax.Assign)
	if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
		return false
	}
	left, ok := assignment.Var.(*syntax.Variable)
	if !ok || left.Name != v.Name {
		return false
	}
	created, ok := assignment.Value.(*syntax.New)
	if !ok || semanticquery.NativeDateClass(ctx, created) != "DateTimeImmutable" {
		return false
	}
	if created.Args != nil && len(created.Args.Args) != 0 && singleStringArgument(created.Args) == nil {
		return false
	}
	echo, ok := list[2].(*syntax.Echo)
	if !ok || len(echo.Exprs) != 1 {
		return false
	}
	read, ok := echo.Exprs[0].(*syntax.MethodCall)
	if !ok {
		return false
	}
	receiver, ok := read.Var.(*syntax.Variable)
	name, named := read.Name.(*syntax.Identifier)
	return ok && receiver.Name == v.Name && named && strings.EqualFold(name.Value, "format") && singleStringArgument(read.Args) != nil
}

func singleStringArgument(args *syntax.ArgList) syntax.Expr {
	if args == nil || len(args.Args) != 1 {
		return nil
	}
	a, ok := args.Args[0].(*syntax.Arg)
	if !ok || a.ByRef || a.Unpack || a.Name != nil || !astquery.IsStringLiteral(a.Value) {
		return nil
	}
	return a.Value
}

func validRelativeDate(text string) bool {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		return false
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil || n < -9999 || n > 9999 {
		return false
	}
	switch parts[1] {
	case "day", "days", "hour", "hours", "minute", "minutes", "second", "seconds", "month", "months", "year", "years":
		return true
	}
	return false
}
