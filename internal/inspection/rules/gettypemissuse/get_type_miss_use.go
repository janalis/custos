package gettypemissuse

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// getTypeMissUse reports `gettype($x) === 'type'` comparisons replaceable by
// an is_*() predicate, and type names gettype() never returns.
type getTypeMissUse struct{}

func (getTypeMissUse) ID() string               { return "GetTypeMissUse" }
func (getTypeMissUse) Semantic()                {}
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
			text := repl
			// `return'string'===gettype($x)`: the predicate (or its leading
			// backslash) would fuse with the keyword before the literal.
			if start := bin.Span().Start; repl[0] != '!' && start > 0 && astquery.NameByte(ctx.File.Src[start-1]) {
				text = " " + repl
			}
			return []diagnostic.TextEdit{{Span: bin.Span(), NewText: text}}
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
	values, _ := semanticquery.LocalValuesKnown(ctx.Types(), e)
	for _, v := range values {
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
	vals, complete := semanticquery.LocalValuesComplete(ctx.Types(), e)
	if !complete || len(vals) != 1 {
		return false
	}
	_, _, ok := astquery.QuotedStringRaw(vals[0])
	return ok && !gettypeWrittenIndirectly(ctx, e)
}

// gettypeWrittenIndirectly reports whether a local variable the value of e
// derives from may be written without an assignment naming it: it is bound
// to a by-reference parameter (or passed to a callee that cannot be
// resolved), aliased with `&`, or its scope writes variables by name
// (extract(), `$$n`, include, eval). Reference argument and alias checks
// already belong to the shared value discovery contract.
func gettypeWrittenIndirectly(ctx *analysis.Context, e syntax.Expr) bool {
	scope := syntax.EnclosingFuncLike(e)
	if scope == nil {
		return false // only constants and properties resolve at top level
	}
	pending := gettypeVariables(nil, e)
	if len(pending) == 0 {
		return false
	}
	dynamic := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Variable:
			dynamic = dynamic || x.NameExpr != nil
		case *syntax.Include, *syntax.Eval:
			dynamic = true
		case *syntax.FuncCall:
			switch ctx.FunctionName(x) {
			case "extract", "parse_str":
				dynamic = true
			}
		}
		return !dynamic
	})
	return dynamic
}

// gettypeVariables appends the names of the variables read under e.
func gettypeVariables(names []string, e syntax.Expr) []string {
	syntax.Inspect(e, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil {
			names = append(names, v.Name)
		}
		return true
	})
	return names
}
