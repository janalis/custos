package gettypemissuse

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
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
			text := repl
			// `return'string'===gettype($x)`: the predicate (or its leading
			// backslash) would fuse with the keyword before the literal.
			if start := bin.Span().Start; repl[0] != '!' && start > 0 && gettypeNameByte(ctx.File.Src[start-1]) {
				text = " " + repl
			}
			return []diagnostic.TextEdit{{Span: bin.Span(), NewText: text}}
		},
	})
}

// gettypeNameByte reports whether c can end a keyword or a name.
func gettypeNameByte(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
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
	return ok && !gettypeWrittenIndirectly(ctx, e)
}

// gettypeWrittenIndirectly reports whether a local variable the value of e
// derives from may be written without an assignment naming it: it is bound
// to a by-reference parameter (or passed to a callee that cannot be
// resolved), aliased with `&`, or its scope writes variables by name
// (extract(), `$$n`, include, eval). The shared value discovery calls such
// a variable's assignments complete.
func gettypeWrittenIndirectly(ctx *analysis.Context, e syntax.Expr) bool {
	scope := syntax.EnclosingFuncLike(e)
	if scope == nil {
		return false // only constants and properties resolve at top level
	}
	pending := gettypeVariables(nil, e)
	if len(pending) == 0 {
		return false
	}
	escaped, deps, dynamic := map[string]bool{}, map[string][]string{}, false
	escape := func(x syntax.Expr) {
		if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok {
			escaped[v.Name] = true
		}
	}
	arguments := func(list *syntax.ArgList, f *index.Function) {
		if list == nil {
			return
		}
		for i, a := range list.Args {
			if arg, ok := a.(*syntax.Arg); ok && (f == nil || !gettypeBoundByValue(list, i, f.Params)) {
				escape(arg.Value)
			}
		}
	}
	syntax.Inspect(scope, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Variable:
			dynamic = dynamic || x.NameExpr != nil
		case *syntax.Include, *syntax.Eval:
			dynamic = true
		case *syntax.Assign:
			if x.ByRef {
				escape(x.Value)
			}
			if v, ok := syntax.UnwrapParens(x.Var).(*syntax.Variable); ok {
				deps[v.Name] = gettypeVariables(deps[v.Name], x.Value)
			}
		case *syntax.ArrayItem:
			if x.ByRef {
				escape(x.Value)
			}
		case *syntax.FuncCall:
			switch ctx.FunctionName(x) {
			case "extract", "parse_str":
				dynamic = true
			}
			arguments(x.Args, ctx.Types().ResolveFunction(x))
		case *syntax.MethodCall:
			arguments(x.Args, nil)
		case *syntax.StaticCall:
			arguments(x.Args, nil)
		case *syntax.New:
			arguments(x.Args, nil)
		}
		return !dynamic
	})
	if dynamic {
		return true
	}
	seen := map[string]bool{}
	for len(pending) > 0 {
		name := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if escaped[name] {
			return true
		}
		if !seen[name] {
			seen[name] = true
			pending = append(pending, deps[name]...)
		}
	}
	return false
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

// gettypeBoundByValue reports whether argument i of list provably binds to
// a by-value parameter of params.
func gettypeBoundByValue(list *syntax.ArgList, i int, params []index.Param) bool {
	arg := list.Args[i].(*syntax.Arg)
	if arg.ByRef {
		return false
	}
	if arg.Name != nil {
		for _, p := range params {
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
				return !p.ByRef
			}
		}
		return false
	}
	if i >= len(params) {
		// Extra arguments bind to a variadic last parameter, or are ignored.
		return len(params) == 0 || !params[len(params)-1].Variadic || !params[len(params)-1].ByRef
	}
	return !params[i].ByRef
}
