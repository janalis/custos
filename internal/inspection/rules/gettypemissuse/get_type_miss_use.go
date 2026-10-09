package gettypemissuse

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// getTypeMissUse reports `gettype($x) === 'type'` comparisons replaceable by
// an is_*() predicate, and type names gettype() never returns.
type getTypeMissUse struct{}

func (getTypeMissUse) ID() string               { return "GetTypeMissUse" }
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
	name, ok := call.Name.(*syntax.Name) // D1
	if !ok {
		return
	}
	if name.NameKind != syntax.NameUnqualified && name.NameKind != syntax.NameFullyQualified ||
		!ctx.IsGlobalFunctionCall(call, "gettype") {
		return
	}
	args, ok := astquery.CallArgValues(call)
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
	t, _, _ := astquery.QuotedStringRaw(lit)
	pred, known := gettypePredicates[t]
	if !known { // D4b
		if t == "unknown type" || t == "resource (closed)" {
			return
		}
		ctx.ReportSeverity(lit.Span(), diagnostic.SeverityError, "gettype() never returns '"+t+"'.")
		return
	}
	if lit != other && !gettypeOnlyValue(ctx, other) {
		return // custos: other values (a subclass default, a caller's argument) may reach the comparison
	}
	// A namespaced function of that name would capture a bare call.
	repl := semanticquery.QualifiedBuiltin(ctx, pred, call.Span().Start) + "(" + ctx.Text(args[0]) + ")"
	if bin.Op.Kind == syntax.TIsNotEqual || bin.Op.Kind == syntax.TIsNotIdentical {
		repl = "!" + repl
	}
	ctx.Report(bin.Span(), "Use '"+repl+"' instead.", diagnostic.Fix{
		Title: "Use " + pred + "()",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: bin.Span(), NewText: repl}}
		},
	})
}

// gettypeOperandLiteral resolves e to a single string literal (D3).
func gettypeOperandLiteral(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	if _, _, ok := astquery.QuotedStringRaw(e); ok {
		return e
	}
	var found syntax.Expr
	var foundText string
	for _, v := range flowquery.PossibleValues(ctx.File, e) {
		t, _, ok := astquery.QuotedStringRaw(v)
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

// gettypeOnlyValue reports whether the complete value set of e is a single
// string literal: only then does the comparison always test that type.
func gettypeOnlyValue(ctx *analysis.Context, e syntax.Expr) bool {
	vals, complete := flowquery.PossibleValuesComplete(ctx.File, e)
	if !complete || len(vals) != 1 {
		return false
	}
	_, _, ok := astquery.QuotedStringRaw(vals[0])
	return ok
}
