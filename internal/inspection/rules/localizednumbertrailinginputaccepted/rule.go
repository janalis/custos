// Package localizednumbertrailinginputaccepted implements the native LocalizedNumberTrailingInputAccepted inspection.
package localizednumbertrailinginputaccepted

import (
	"regexp"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
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
	if consumedWhole(ctx, semanticquery.CallArgument(parse.Args, 2, "offset"), value, len(text)) {
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

// consumedWhole reports whether the offset the parse returned through the
// local offset is proven, where value is echoed, to equal the input length
// size (or the echo is unreachable): a comparison with the length guards it.
// The offset never passes the end of the input.
func consumedWhole(ctx *analysis.Context, offset, value syntax.Expr, size int) bool {
	v, ok := syntax.UnwrapParens(offset).(*syntax.Variable)
	if !ok || v.NameExpr != nil || semanticquery.NativeLabelled(ctx, value) {
		return false
	}
	t := ctx.Types().Native().NarrowedVarAt(v.Name, value, types.Ints(types.IntRange{Lo: 0, Hi: int64(size)}))
	return t.IsUnknown() || t.IntsOnly(int64(size), int64(size))
}
