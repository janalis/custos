package caseinsensitivestringfunctionsmissuse

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// caseInsensitiveStringFunctionsMissUse reports stripos()/strripos()/
// stristr() whose needle has no letters, where the case-sensitive variant
// gives the same result.
type caseInsensitiveStringFunctionsMissUse struct{}

func (caseInsensitiveStringFunctionsMissUse) ID() string {
	return "CaseInsensitiveStringFunctionsMissUse"
}

func (caseInsensitiveStringFunctionsMissUse) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var caseInsensitiveCounterparts = map[string]string{"stripos": "strpos", "strripos": "strrpos", "stristr": "strstr"}

func (caseInsensitiveStringFunctionsMissUse) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	nameNode, _, ok := astquery.FuncNamePart(call) // D1
	if !ok {
		return
	}
	// D1: the call must reach the global function (any case).
	counterpart, ok := caseInsensitiveCounterparts[ctx.GlobalFunctionName(call)]
	if !ok {
		return
	}
	// D2: positional arguments only; a named or unpacked argument does not
	// say which written position holds the needle.
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) < 2 || len(args) > 3 {
		return
	}
	lit := needleLiteral(ctx, args[1]) // D3
	if lit == nil {
		return
	}
	content, ok := astquery.StringLiteralValue(lit.Raw)
	if !ok || content == "" || !utf8.ValidString(content) || hasLetter(content) { // D4
		return
	}
	span := astquery.NamePartSpan(nameNode)
	newName := counterpart
	if q, _, _ := astquery.CallName(call); q == "" { // a namespaced or imported counterpart would capture a bare call
		newName = semanticquery.QualifiedBuiltin(ctx, counterpart, call.Span().Start)
	}
	ctx.ReportSeverity(call.Span(), diagnostic.SeverityInfo, "Needle has no letters; use '"+counterpart+"(...)' instead.", diagnostic.Fix{
		Title: "Use " + counterpart + "()",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: newName}} },
	})
}

// needleLiteral returns the string literal the needle is proven to be (D3):
// e itself, a constant declared once in this file, a class constant, or a
// local variable whose only value is that literal. nil when another value
// may reach the call, since the renamed call would then compare letters
// exactly.
func needleLiteral(ctx *analysis.Context, e syntax.Expr) *syntax.Literal {
	e = syntax.UnwrapParens(e)
	if _, _, ok := astquery.QuotedStringRaw(e); ok {
		return e.(*syntax.Literal)
	}
	switch x := e.(type) {
	case *syntax.Variable:
		return localLiteral(ctx, x)
	case *syntax.ConstFetch:
		if !declaredOnce(ctx.File, astquery.LastNamePart(x.Name.Value)) {
			return nil
		}
	case *syntax.ClassConstFetch:
	default:
		return nil
	}
	return onlyLiteral(ctx.File, e)
}

// onlyLiteral returns the quoted string literal that is the complete value
// set of e, or nil.
func onlyLiteral(f *syntax.File, e syntax.Expr) *syntax.Literal {
	vals, complete := flowquery.PossibleValuesComplete(f, e)
	if !complete || len(vals) != 1 {
		return nil
	}
	if _, _, ok := astquery.QuotedStringRaw(vals[0]); !ok {
		return nil
	}
	return vals[0].(*syntax.Literal)
}

// localLiteral is onlyLiteral for a local variable. The literal must be
// assigned to the variable itself (no chain through another variable), and
// nothing in the function may rebind the variable behind the assignments.
func localLiteral(ctx *analysis.Context, v *syntax.Variable) *syntax.Literal {
	lit := onlyLiteral(ctx.File, v)
	if lit == nil {
		return nil
	}
	if a, ok := lit.Parent().(*syntax.Assign); !ok || !isLocal(a.Var, v.Name) {
		return nil
	}
	// A complete value set implies a function scope with a body.
	if _, body := astquery.ScopeParts(syntax.EnclosingFuncLike(v)); !stableLocal(ctx, body, v.Name) {
		return nil
	}
	return lit
}

func isLocal(e syntax.Expr, name string) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return ok && v.NameExpr == nil && v.Name == name
}

// scopeWriters are the builtins that create or overwrite variables of the
// calling scope by name.
var scopeWriters = map[string]bool{"extract": true, "parse_str": true, "mb_parse_str": true}

// stableLocal reports whether nothing under body can change $name other
// than the assignments value discovery already accounts for: no
// variable-variable, eval, include or scope-writing builtin, no reference
// taken to the variable, and no call that may receive it by reference
// (unresolved callees included). Nested functions are scanned too, which
// only errs on the silent side.
func stableLocal(ctx *analysis.Context, body syntax.Node, name string) bool {
	unstable := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Variable:
			unstable = unstable || x.NameExpr != nil
		case *syntax.Eval, *syntax.Include:
			unstable = true
		case *syntax.Assign:
			unstable = unstable || x.ByRef && isLocal(x.Value, name)
		case *syntax.ArrayItem:
			unstable = unstable || x.ByRef && isLocal(x.Value, name)
		case *syntax.Yield:
			unstable = unstable || isLocal(x.Value, name) // a by-reference generator hands out the variable
		case *syntax.FuncCall:
			unstable = unstable || scopeWriters[ctx.GlobalFunctionName(x)]
		case *syntax.ArgList:
			unstable = unstable || mayBindByRef(ctx, x, name)
		}
		return !unstable
	})
	return !unstable
}

// mayBindByRef reports whether list passes $name to a callee that is
// unknown or takes one of the arguments by reference.
func mayBindByRef(ctx *analysis.Context, list *syntax.ArgList, name string) bool {
	passed := false
	for _, a := range list.Args {
		if arg, ok := a.(*syntax.Arg); ok && isLocal(arg.Value, name) {
			passed = true
		}
	}
	if !passed {
		return false
	}
	callee, known := semanticquery.ResolveCallee(ctx, list.Parent())
	return !known || semanticquery.ArgBindsByRef(list, callee.Params, true)
}

// declaredOnce reports whether this file declares the global constant
// name exactly once (`const` or define(); a define() whose name is not a
// literal counts as a declaration of any name). Two declarations, as in
// the branches of a conditional, leave the value unknown.
func declaredOnce(f *syntax.File, name string) bool {
	count := 0
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.ConstStmt:
			for _, it := range x.Consts {
				if it.Name != nil && it.Name.Value == name {
					count++
				}
			}
		case *syntax.FuncCall:
			if !strings.EqualFold(astquery.CallLastName(x), "define") {
				break
			}
			args, _ := astquery.CallArgValues(x)
			if len(args) > 0 {
				if raw, _, ok := astquery.QuotedStringRaw(args[0]); ok && astquery.LastNamePart(raw) != name {
					break
				}
			}
			count++
		}
		return true
	})
	return count == 1
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
