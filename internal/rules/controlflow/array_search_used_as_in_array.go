package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// arraySearchUsedAsInArray reports array_search() results used only as a
// boolean, and pointless comparisons of its result with true.
type arraySearchUsedAsInArray struct{}

func init() { register(arraySearchUsedAsInArray{}) }

const (
	arraySearchMsg     = "Use 'in_array(...)' to test membership."
	arraySearchTrueMsg = "array_search() cannot return true; this comparison never changes."
)

func (arraySearchUsedAsInArray) ID() string { return "ArraySearchUsedAsInArray" }

func (arraySearchUsedAsInArray) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (arraySearchUsedAsInArray) Check(ctx *analysis.Context, n syntax.Node) {
	call, ok := util.IsFuncNamedFold(n, "array_search")                                    // D1
	if !ok || !ctx.IsGlobalFunctionCall(call, "array_search") || util.ArgCount(call) < 2 { // D1, E1
		return
	}
	name := call.Name.(*syntax.Name)
	// D2
	if isLogicalOperand(call) {
		span := call.Span()
		ctx.Report(span, arraySearchMsg, analysis.Fix{
			Title: "Use in_array()",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{{Span: name.Span(), NewText: util.QualifiedBuiltinFor(ctx, "in_array", call)}}
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
	val, ok := boolConstant(other)
	if !ok {
		return
	}
	if val { // D3b
		ctx.ReportSeverity(other.Span(), meta.SeverityError, arraySearchTrueMsg)
		return
	}
	span := bin.Span()
	negate := bin.Op.Kind == syntax.TIsIdentical
	ctx.Report(span, arraySearchMsg, analysis.Fix{
		Title: "Use in_array()",
		Edits: func() []analysis.TextEdit {
			cs := call.Span()
			var b strings.Builder
			if negate {
				b.WriteByte('!')
			}
			b.WriteString(util.QualifiedBuiltinFor(ctx, "in_array", call))
			b.WriteString(ctx.SpanText(syntax.Span{Start: name.Span().End, End: cs.End}))
			return []analysis.TextEdit{{Span: span, NewText: b.String()}}
		},
	})
}

// isLogicalOperand implements D2.
func isLogicalOperand(call *syntax.FuncCall) bool {
	parent, child := util.ParentSkipParens(call)
	switch p := parent.(type) {
	case *syntax.If:
		return p.Cond == child
	case *syntax.ElseIf:
		return p.Cond == child
	case *syntax.While:
		return p.Cond == child
	case *syntax.DoWhile:
		return p.Cond == child
	case *syntax.Unary:
		return p.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return true
		}
	case *syntax.Ternary:
		return p.Then != nil && p.Cond == child
	}
	return false
}
