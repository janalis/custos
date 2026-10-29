package typeunsafearraysearch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// typeUnsafeArraySearch asks for the strict argument of in_array() and
// array_search().
type typeUnsafeArraySearch struct{}

func (typeUnsafeArraySearch) ID() string               { return "TypeUnsafeArraySearch" }
func (typeUnsafeArraySearch) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (typeUnsafeArraySearch) Semantic()                {}
func (typeUnsafeArraySearch) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	_, name, ok := astquery.CallName(call)
	if lname := strings.ToLower(name); !ok || (lname != "in_array" && lname != "array_search") || astquery.ArgCount(call) != 2 { // D1, D2, E3
		return
	}
	if !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) { // D1
		return
	}
	needle, ok1 := call.Args.Args[0].(*syntax.Arg)
	haystack, ok2 := call.Args.Args[1].(*syntax.Arg)
	if !ok1 || !ok2 { // `...` placeholder (rejected by PHP after another argument)
		return
	}
	if needle.Unpack || haystack.Unpack { // D2: a spread passes an unknown number of arguments, strict flag included
		return
	}
	if literalStringHaystack(ctx, haystack.Value) { // E1
		return
	}
	nt, ht := ctx.TypeOf(needle.Value), ctx.TypeOf(haystack.Value) // E2
	if !nt.IsUnknown() && !ht.IsUnknown() && len(nt.Atoms()) == 1 && len(ht.Atoms()) == 1 &&
		(strings.TrimPrefix(ht.Atoms()[0], `\`) == strings.TrimPrefix(nt.Atoms()[0], `\`)+"[]" || shapeOf(ht, nt.Atoms()[0])) {
		return
	}
	// custos: no fixing. `true` changes the result whenever loose and strict
	// equality differ ('1' == 1, null == '', '1e1' == '10'): a config
	// string searched among int constants stops matching. The only cases
	// known to be equivalent (same scalar type on both sides) are already
	// exempt (E2).
	ctx.Report(call.Span(), "Pass a third argument to say whether this search must be type-strict.")
}

// shapeOf reports whether t is a sealed array shape whose every value has
// type atom alone (`const MODES = ['a', 'b']` with a string needle).
func shapeOf(t types.Type, atom string) bool {
	if !t.IsSealedShape() || len(t.ShapeKeys()) == 0 {
		return false
	}
	for _, k := range t.ShapeKeys() {
		if as := k.Type.Atoms(); k.Type.IsUnknown() || len(as) != 1 || strings.TrimPrefix(as[0], `\`) != strings.TrimPrefix(atom, `\`) {
			return false
		}
	}
	return true
}

// literalStringHaystack implements E1: an array literal of non-numeric
// strings, or a local variable every value of which is one.
func literalStringHaystack(ctx *analysis.Context, h syntax.Expr) bool {
	if v, ok := astquery.PlainVariable(h); ok {
		vals, complete := semanticquery.LocalValuesComplete(ctx.Types(), h)
		if !complete || len(vals) == 0 {
			return false
		}
		// `$known[] = '1'` adds an element the assigned literal lacks.
		for _, a := range flowquery.VarAccesses(ctx.File, syntax.EnclosingVariableScope(v), v.Name) {
			if a.ElemWrite {
				return false
			}
		}
		for _, v := range vals {
			if !literalStringArray(ctx, v) {
				return false
			}
		}
		return true
	}
	return literalStringArray(ctx, h)
}

// literalStringArray reports an array literal of non-numeric strings and
// class names (E1).
func literalStringArray(ctx *analysis.Context, h syntax.Expr) bool {
	arr, ok := h.(*syntax.Array)
	if !ok || len(arr.Items) == 0 {
		return false
	}
	for _, it := range arr.Items {
		if it == nil || it.Key != nil || it.Unpack || it.Value == nil {
			return false
		}
		if classNameLiteral(it.Value) {
			continue // `Foo::class` is a class name: never numeric
		}
		content, ok := stringLiteralContent(ctx, it.Value)
		if !ok {
			return false
		}
		content = strings.TrimSpace(content)
		if content == "" || astquery.IsComparisonNumericString(content) { // numeric strings compare numerically
			return false
		}
	}
	return true
}

// classNameLiteral reports `Name::class` with a written class name (not
// `$object::class`, which reads a value).
func classNameLiteral(e syntax.Expr) bool {
	c, ok := e.(*syntax.ClassConstFetch)
	if !ok {
		return false
	}
	_, named := c.Class.(*syntax.Name)
	id, ok := c.Name.(*syntax.Identifier)
	return named && ok && strings.EqualFold(id.Value, "class")
}

// stringLiteralContent returns the raw text between the quotes of a quoted
// string literal (interpolated ones included).
func stringLiteralContent(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	if raw, _, ok := astquery.QuotedStringRaw(e); ok {
		return raw, true
	}
	// custos: a class constant named through its class (not static::,
	// which a subclass may redefine) stands for its string initializer.
	if cf, ok := e.(*syntax.ClassConstFetch); ok {
		if n, named := cf.Class.(*syntax.Name); named && !strings.EqualFold(n.Value, "static") {
			if lit := semanticquery.LocalStringLiteral(ctx.Types(), cf); lit != nil {
				if raw, _, ok := astquery.QuotedStringRaw(lit); ok {
					return raw, true
				}
			}
			return indexedStringConst(ctx, cf)
		}
		return "", false
	}
	if is, ok := e.(*syntax.InterpolatedString); ok && !is.Heredoc && !is.Backtick {
		if t := ctx.Text(is); len(t) >= 2 && t[0] == '"' && t[len(t)-1] == '"' {
			return t[1 : len(t)-1], true
		}
	}
	return "", false
}

// indexedStringConst reads a class constant declared in another file: its
// indexed initializer, when that is a single plain quoted string (no
// interpolation, no concatenation).
func indexedStringConst(ctx *analysis.Context, cf *syntax.ClassConstFetch) (string, bool) {
	id, ok := cf.Name.(*syntax.Identifier)
	cls := ctx.Types().ClassRef(cf.Class)
	if !ok || cls == "" {
		return "", false
	}
	k := ctx.Index().FindConst(cls, id.Value, ctx.PHP)
	if k == nil || k.Case || len(k.Value) < 2 {
		return "", false
	}
	q, body := k.Value[0], k.Value[1:len(k.Value)-1]
	if q != '\'' && q != '"' || k.Value[len(k.Value)-1] != q || strings.ContainsRune(body, rune(q)) || q == '"' && strings.ContainsAny(body, `$\`) {
		return "", false
	}
	return body, true
}
