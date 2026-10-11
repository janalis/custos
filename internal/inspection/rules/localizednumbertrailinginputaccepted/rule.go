// Package localizednumbertrailinginputaccepted implements the native LocalizedNumberTrailingInputAccepted inspection.
package localizednumbertrailinginputaccepted

import (
	"regexp"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject unparsed number suffixes."

// decimalStyle is the value of NumberFormatter::DECIMAL.
const decimalStyle = 1

var trailingNumber = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?[A-DF-Za-df-z][A-Za-z]*$`)

// passiveMethods neither change what a formatter accepts nor let it escape.
var passiveMethods = map[string]bool{
	"parse": true, "parsecurrency": true, "format": true, "formatcurrency": true,
	"geterrorcode": true, "geterrormessage": true, "getlocale": true, "getpattern": true,
	"getattribute": true, "gettextattribute": true, "getsymbol": true,
}

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "LocalizedNumberTrailingInputAccepted" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	e := n.(*syntax.Echo)
	if len(e.Exprs) != 1 {
		return
	}
	value := e.Exprs[0]
	parse, ok := semanticquery.NativeValue(ctx, value).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, parse, "NumberFormatter", "parse") || !plainDecimal(ctx, parse) {
		return
	}
	input := semanticquery.CallArgument(parse.Args, 0, "string")
	text, known := semanticquery.NativeString(ctx, input)
	if !known || !trailingNumber.MatchString(text) {
		return
	}
	g := guard{ctx: ctx, parse: parse, input: input, offset: semanticquery.CallArgument(parse.Args, 2, "offset")}
	if g.protects(e) {
		return
	}

	ctx.Report(syntax.Span{Start: e.Span().Start, End: value.Span().End}, message)
}

// plainDecimal proves that the receiver is a two-argument DECIMAL formatter
// that nothing reconfigured: a pattern, affix, symbol or leniency change can
// make the formatter consume the suffix.
func plainDecimal(ctx *analysis.Context, parse *syntax.MethodCall) bool {
	construct := semanticquery.NativeConstruction(ctx, parse.Var, "NumberFormatter")
	if construct == nil || construct.Args == nil || len(construct.Args.Args) != 2 {
		return false
	}
	style, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(construct.Args, 1, "style"))
	if !known || style != decimalStyle {
		return false
	}
	receiver, ok := syntax.UnwrapParens(parse.Var).(*syntax.Variable)
	if !ok {
		return syntax.UnwrapParens(parse.Var) == syntax.Expr(construct)
	}
	// A call later in an enclosing loop also precedes the next iteration's parse.
	limit := parse.Span().Start
	scope := syntax.EnclosingVariableScope(parse)
	for p := parse.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			limit = p.Span().End
		}
	}
	proven := true
	visit := func(n syntax.Node) bool {
		v, ok := n.(*syntax.Variable)
		if !ok || v.Name != receiver.Name || v.Span().Start >= limit {
			return proven
		}
		switch p := v.Parent().(type) {
		case *syntax.Assign:
			proven = p.Var == syntax.Expr(v)
		case *syntax.MethodCall:
			name, named := p.Name.(*syntax.Identifier)
			proven = named && p.Var == syntax.Expr(v) && passiveMethods[strings.ToLower(name.Value)]
		default:
			proven = false
		}
		return proven
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return proven
}

// guard recognizes checks that the parse consumed its whole input.
type guard struct {
	ctx           *analysis.Context
	parse         *syntax.MethodCall
	input, offset syntax.Expr
}

// protects reports whether the echo only runs once the returned offset was
// compared with the input length: inside the branch such a comparison selects,
// or after a statement that leaves when the comparison fails.
func (g guard) protects(e *syntax.Echo) bool {
	scope := syntax.EnclosingVariableScope(e)
	for at := syntax.Node(e); at != nil && at != scope; at = at.Parent() {
		if stmt, ok := at.(syntax.Stmt); ok {
			for _, prior := range flowquery.NativePriorStatements(g.ctx.File, stmt) {
				check, ok := prior.(*syntax.If)
				if ok && check.Else == nil && len(check.ElseIfs) == 0 && check.Span().Start >= g.parse.Span().End && syntax.Terminates(check.Body) && g.implied(check.Cond, false) {
					return true
				}
			}
		}
		switch p := at.Parent().(type) {
		case *syntax.If:
			if at == syntax.Node(p.Body) && g.implied(p.Cond, true) {
				return true
			}
		case *syntax.ElseIf:
			if at == syntax.Node(p.Body) && g.implied(p.Cond, true) {
				return true
			}
		case *syntax.Else:
			if owner, ok := p.Parent().(*syntax.If); ok && g.implied(owner.Cond, false) {
				return true
			}
		}
	}
	return false
}

// implied reports whether cond evaluating to truth proves that the offset
// equals the byte length of the input.
func (g guard) implied(cond syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && g.implied(c.Expr, !truth)
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (g.implied(c.Left, true) || g.implied(c.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (g.implied(c.Left, false) || g.implied(c.Right, false))
		case syntax.TIsIdentical, syntax.TIsNotIdentical:
			return (c.Op.Kind == syntax.TIsIdentical) == truth && (g.compares(c.Left, c.Right) || g.compares(c.Right, c.Left))
		}
	}
	return false
}

func (g guard) compares(offset, length syntax.Expr) bool {
	call, ok := syntax.UnwrapParens(length).(*syntax.FuncCall)
	return ok && semanticquery.ExpansionCSame(g.ctx, offset, g.offset) &&
		semanticquery.NativeBuiltin(g.ctx, call, "strlen") &&
		semanticquery.ExpansionCSame(g.ctx, semanticquery.CallArgument(call.Args, 0, "string"), g.input)
}
