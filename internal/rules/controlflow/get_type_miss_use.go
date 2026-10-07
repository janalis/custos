package controlflow

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// getTypeMissUse reports `gettype($x) === 'type'` comparisons replaceable by
// an is_*() predicate, and type names gettype() never returns.
type getTypeMissUse struct{}

func init() { register(getTypeMissUse{}) }

func (getTypeMissUse) ID() string { return "GetTypeMissUse" }

func (getTypeMissUse) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

var gettypePredicates = map[string]string{
	"boolean":  "is_bool",
	"integer":  "is_int",
	"double":   "is_float",
	"string":   "is_string",
	"array":    "is_array",
	"object":   "is_object",
	"resource": "is_resource",
	"NULL":     "is_null",
}

func (getTypeMissUse) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if call.Span().Len() == 0 {
		return
	}
	name, ok := call.Name.(*syntax.Name) // D1
	if !ok {
		return
	}
	if name.NameKind != syntax.NameUnqualified && name.NameKind != syntax.NameFullyQualified ||
		!ctx.IsGlobalFunctionCall(call, "gettype") {
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 1 {
		return
	}
	bin, ok := call.Parent().(*syntax.Binary) // D2
	if !ok {
		return
	}
	switch bin.Op.Kind {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
	default:
		return
	}
	other := bin.Right
	if bin.Right == syntax.Expr(call) {
		other = bin.Left
	}
	lit := gettypeOperandLiteral(ctx, other) // D3
	if lit == nil {
		return
	}
	t, _, _ := util.QuotedStringRaw(lit)
	pred, known := gettypePredicates[t]
	if !known { // D4b
		if t == "unknown type" || t == "resource (closed)" {
			return
		}
		ctx.ReportSeverity(lit.Span(), meta.SeverityError, "gettype() never returns '"+t+"'.")
		return
	}
	// A namespaced function of that name would capture a bare call.
	repl := util.QualifiedBuiltin(ctx, pred, call.Span().Start) + "(" + ctx.Text(args[0]) + ")"
	if bin.Op.Kind == syntax.TIsNotEqual || bin.Op.Kind == syntax.TIsNotIdentical {
		repl = "!" + repl
	}
	ctx.Report(bin.Span(), "Use '"+repl+"' instead.", analysis.Fix{
		Title: "Use " + pred + "()",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: bin.Span(), NewText: repl}}
		},
	})
}

// gettypeOperandLiteral resolves e to a single string literal (D3).
func gettypeOperandLiteral(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	if _, _, ok := util.QuotedStringRaw(e); ok {
		return e
	}
	var found syntax.Expr
	var foundText string
	for _, v := range util.PossibleValues(ctx.File, e) {
		t, _, ok := util.QuotedStringRaw(v)
		if !ok {
			continue
		}
		if found != nil && t != foundText {
			return nil
		}
		if found == nil {
			found, foundText = v, t
		}
	}
	return found
}
