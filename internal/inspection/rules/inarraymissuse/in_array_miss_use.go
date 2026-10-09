package inarraymissuse

import (
	"regexp"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// inArrayMissUse reports in_array() over array_keys() (use
// array_key_exists) and over a one-element array literal (use a comparison).
type inArrayMissUse struct{}

func (inArrayMissUse) ID() string               { return "InArrayMissUse" }
func (inArrayMissUse) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (inArrayMissUse) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "in_array") // D1
	if call == nil {
		return
	}
	args, ok := astquery.ArgValues(call.Args)
	if !ok || len(args) < 2 || len(args) > 3 { // D2
		return
	}
	needle, hay := args[0], args[1]

	// Pattern K.
	if keys, _ := semanticquery.GlobalCall(ctx, hay, "array_keys"); keys != nil { // D3
		kargs, ok := astquery.ArgValues(keys.Args)
		if !ok || len(kargs) != 1 {
			return
		}
		repl := semanticquery.QualifiedBuiltin(ctx, "array_key_exists", call.Span().Start) + "(" + ctx.Text(needle) + ", " + ctx.Text(kargs[0]) + ")"
		span := call.Span()
		msg := "Look the key up directly with '" + repl + "'."
		if !iamKeyLookupSafe(ctx, needle, len(args) == 3) {
			ctx.Report(span, msg)
			return
		}
		ctx.Report(span, msg, astquery.ReplaceFix(span, repl))
		return
	}

	// Pattern C.
	arr, ok := hay.(*syntax.Array) // D4
	if !ok || len(arr.Items) != 1 || arr.Items[0] == nil {
		return
	}
	item := arr.Items[0]
	if item.Unpack || item.ByRef || item.Value == nil { // spec Divergences
		return
	}
	value := item.Value

	var target syntax.Node = call // D5
	exists := true
	switch p := call.Parent().(type) {
	case *syntax.Unary:
		if p.Op.Kind == syntax.TExclaim {
			target, exists = p, false
		}
	case *syntax.Binary:
		other := p.Left
		if other == syntax.Expr(call) {
			other = p.Right
		}
		if v, isBool := astquery.BoolConst(other); isBool {
			switch p.Op.Kind {
			case syntax.TIsEqual, syntax.TIsIdentical:
				target, exists = p, v
			case syntax.TIsNotEqual, syntax.TIsNotIdentical:
				target, exists = p, !v
			}
		}
	}

	strict := ctx.Bool("FORCE_STRICT_COMPARISON") // D6
	if len(args) == 3 {
		if v, isBool := astquery.BoolConst(args[2]); isBool && v {
			strict = true
		}
	}
	op := "=="
	if !exists {
		op = "!="
	}
	if strict {
		op += "="
	}
	left := ctx.Text(needle)
	if astquery.NeedsParensAsUnaryOperand(needle) {
		left = "(" + left + ")"
	}
	right := ctx.Text(value)
	if lowerThanEquality(value) {
		right = "(" + right + ")"
	}
	cmp := left + " " + op + " " + right
	if ctx.ComparisonStyle == analysis.StyleYoda {
		cmp = right + " " + op + " " + left
	}
	repl := cmp
	if astquery.NeedsParensAsComparison(target) {
		repl = "(" + cmp + ")"
	}
	ctx.Report(target.Span(), "Compare directly: '"+cmp+"'.", astquery.ReplaceFix(target.Span(), repl))
}

var (
	iamNumeric      = regexp.MustCompile(`^[ \t\n\r\v\f]*[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?[ \t\n\r\v\f]*$`)
	iamCanonicalInt = regexp.MustCompile(`^(0|-?[1-9]\d{0,17})$`)
)

// iamKeyLookupSafe reports whether array_key_exists() answers like the
// in_array() over array_keys() it replaces (custos, see Divergences): keys
// are ints or strings, `'5'` is stored as the int 5, and loose comparison
// matches numeric strings (`'1.0' == 1`), so only these needles qualify:
// with strict comparison an int-typed needle or a string literal that is
// not a canonical integer; with loose comparison (PHP 8, where `'abc' == 0`
// is false) a non-numeric string literal. Anything else keeps the report
// without a fixing.
func iamKeyLookupSafe(ctx *analysis.Context, needle syntax.Expr, strict bool) bool {
	if strict {
		if v, ok := astquery.QuotedStringValue(syntax.UnwrapParens(needle)); ok {
			return !iamCanonicalInt.MatchString(v)
		}
		t := ctx.TypeOf(needle)
		return !t.IsUnknown() && t.OnlyOf("int")
	}
	v, ok := astquery.QuotedStringValue(syntax.UnwrapParens(needle))
	return ok && ctx.PHP >= phpversion.PHP80 && !iamNumeric.MatchString(v)
}

// lowerThanEquality reports whether e binds looser than (or as tight as) a
// comparison and must be parenthesised as an operand of `==`. Not
// astquery.NeedsParensAsEqualityOperand: every binary operator outside the
// tighter list counts here (e.g. the PHP 8.5 pipe operator).
func lowerThanEquality(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Binary:
		return !bindsTighterThanEquality(e.Op.Kind)
	case *syntax.Assign, *syntax.Ternary, *syntax.Print, *syntax.Yield, *syntax.YieldFrom,
		*syntax.Include, *syntax.Throw, *syntax.ArrowFunction:
		return true
	}
	return false
}

// bindsTighterThanEquality reports whether a binary operator has a higher
// precedence than `==`.
func bindsTighterThanEquality(k syntax.TokenKind) bool {
	switch k {
	case syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPlus, syntax.TMinus, syntax.TDot,
		syntax.TSl, syntax.TSr, syntax.TPow, syntax.TLess, syntax.TGreater,
		syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
		return true
	}
	return false
}
