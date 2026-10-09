package arraysearchusedasinarray

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// arraySearchUsedAsInArray reports array_search() results used only as a
// boolean, and pointless comparisons of its result with true.
type arraySearchUsedAsInArray struct{}

const (
	arraySearchMsg     = "Use 'in_array(...)' to test membership."
	arraySearchTrueMsg = "array_search() cannot return true; this comparison never changes."
)

func (arraySearchUsedAsInArray) ID() string { return "ArraySearchUsedAsInArray" }
func (arraySearchUsedAsInArray) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (arraySearchUsedAsInArray) Check(ctx *analysis.Context, n syntax.Node) {
	call, ok := astquery.IsFuncNamedFold(n, "array_search")                                    // D1
	if !ok || !ctx.IsGlobalFunctionCall(call, "array_search") || astquery.ArgCount(call) < 2 { // D1, E1
		return
	}
	name := call.Name.(*syntax.Name)
	// D2
	if astquery.IsLogicalOperand(call) {
		span := call.Span()
		ctx.Report(span, arraySearchMsg, diagnostic.Fix{
			Title: "Use in_array()",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: name.Span(), NewText: semanticquery.QualifiedBuiltinFor(ctx, "in_array", call)}}
			},
		})
		return
	}
	// D3
	bin, ok := call.Parent().(*syntax.Binary)
	if !ok || (bin.Op.Kind != syntax.TIsIdentical && bin.Op.Kind != syntax.TIsNotIdentical) {
		return
	}
	other := bin.Right
	if bin.Right == syntax.Expr(call) {
		other = bin.Left
	}
	val, ok := astquery.BoolConst(other)
	if !ok {
		return
	}
	if val { // D3b
		ctx.ReportSeverity(other.Span(), diagnostic.SeverityError, arraySearchTrueMsg)
		return
	}
	span := bin.Span()
	negate := bin.Op.Kind == syntax.TIsIdentical
	ctx.Report(span, arraySearchMsg, diagnostic.Fix{
		Title: "Use in_array()",
		Edits: func() []diagnostic.TextEdit {
			cs := call.Span()
			var b strings.Builder
			if negate {
				b.WriteByte('!')
			}
			b.WriteString(semanticquery.QualifiedBuiltinFor(ctx, "in_array", call))
			b.WriteString(ctx.SpanText(syntax.Span{Start: name.Span().End, End: cs.End}))
			return []diagnostic.TextEdit{{Span: span, NewText: b.String()}}
		},
	})
}
