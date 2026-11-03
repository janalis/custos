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
	if literalStringHaystack(ctx, haystack.Value) || literalStringNeedle(ctx, needle.Value) { // E1
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
		if it == nil || it.Key != nil || it.Unpack || it.Value == nil || !nonNumericString(ctx, it.Value) {
			return false
		}
	}
	return true
}

// literalStringNeedle reports a needle written as a non-numeric string
// without interpolation, whose value is known (E1).
func literalStringNeedle(ctx *analysis.Context, n syntax.Expr) bool {
	if v, ok := astquery.PlainVariable(n); ok {
		if foreachOverLiteral(ctx, v) {
			return true
		}
		// A local variable every value of which is such a string.
		vals, complete := semanticquery.LocalValuesComplete(ctx.Types(), n)
		if !complete || len(vals) == 0 {
			return false
		}
		for _, v := range vals {
			if !literalStringNeedle(ctx, v) {
				return false
			}
		}
		return true
	}
	if b, ok := n.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
		return nonNumericConcat(b)
	}
	if objectClassName(ctx, n) {
		return true
	}
	_, interpolated := n.(*syntax.InterpolatedString)
	return !interpolated && nonNumericString(ctx, n)
}

// nonNumericConcat reports a concatenation that ends with a string literal
// whose last character, or starts with one whose first character, cannot
// end or start a numeric string: numeric strings allow only whitespace
// around the sign, digits, `.` and exponent.
func nonNumericConcat(b *syntax.Binary) bool {
	if last, ok := syntax.UnwrapParens(b.Right).(*syntax.Literal); ok {
		if v, ok := syntax.LiteralStringValue(last); ok && v != "" && !strings.ContainsAny(v[len(v)-1:], " \t\n\r\v\f.0123456789") {
			return true
		}
	}
	first := b.Left
	for {
		inner, ok := syntax.UnwrapParens(first).(*syntax.Binary)
		if !ok || inner.Op.Kind != syntax.TDot {
			break
		}
		first = inner.Left
	}
	lit, ok := syntax.UnwrapParens(first).(*syntax.Literal)
	if !ok {
		return false
	}
	v, ok := syntax.LiteralStringValue(lit)
	return ok && v != "" && !strings.ContainsAny(v[:1], " \t\n\r\v\f.+-0123456789")
}

// foreachOverLiteral reports a needle that is the value variable of an
// enclosing `foreach` over an array literal of non-numeric strings, which
// nothing in the loop body writes or passes to a function, and nothing
// binds by reference outside it (E1).
func foreachOverLiteral(ctx *analysis.Context, needle *syntax.Variable) bool {
	accesses := flowquery.VarAccesses(ctx.File, syntax.EnclosingVariableScope(needle), needle.Name)
	var loop *syntax.Foreach
	for _, a := range accesses {
		if f, ok := a.By.(*syntax.Foreach); ok && a.Write && syntax.Node(f.Value) == a.Var && flowquery.NodeContains(f.Body, needle) {
			loop = f // the innermost one comes last
		}
	}
	if loop == nil || loop.ByRef || !literalStringHaystack(ctx, loop.Expr) {
		return false
	}
	for _, a := range accesses {
		inside := flowquery.NodeContains(loop.Body, a.Var)
		switch {
		case !inside:
			// Each iteration assigns the variable afresh: only a reference
			// bound outside the loop (`$alias = &$k`) could change it inside.
			if referenceBound(a) {
				return false
			}
		case a.Write, a.ElemWrite:
			return false
		case a.Var != needle:
			if _, arg := a.Var.Parent().(*syntax.Arg); arg {
				return false // `sort($k)` may write it
			}
		}
	}
	return true
}

// referenceBound reports an access that may tie the variable to another
// one: a reference taken to it (`&$k` assigned or in an array) or bound to
// it, or any binding other than a plain assignment or a foreach value.
func referenceBound(a flowquery.VarAccess) bool {
	switch by := a.By.(type) {
	case nil:
		switch p := a.Var.Parent().(type) {
		case *syntax.Assign:
			return p.ByRef
		case *syntax.ArrayItem:
			return p.ByRef
		}
		return false
	case *syntax.Assign:
		return by.ByRef
	case *syntax.Foreach:
		return false
	}
	return true
}

// nonNumericString reports a string literal (or class name) whose trimmed
// contents are non-empty and not numeric: a loose comparison with it acts
// as a strict one for every string (E1).
func nonNumericString(ctx *analysis.Context, e syntax.Expr) bool {
	if classNameLiteral(e) {
		return true // `Foo::class` is a class name: never numeric
	}
	content, ok := stringLiteralContent(ctx, e)
	content = strings.TrimSpace(content)
	return ok && content != "" && !astquery.IsComparisonNumericString(content) // numeric strings compare numerically
}

// objectClassName reports `get_class($o)` and `$o::class`: the class name
// of an object, which is never a numeric string.
func objectClassName(ctx *analysis.Context, e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.FuncCall:
		args, ok := astquery.CallArgValues(x)
		return ok && len(args) == 1 && ctx.GlobalFunctionName(x) == "get_class"
	case *syntax.ClassConstFetch:
		id, ok := x.Name.(*syntax.Identifier)
		return ok && strings.EqualFold(id.Value, "class")
	}
	return false
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
	// custos: a class constant named through its class (static:: only in a
	// class no other class extends, which could redefine it) stands for its
	// string initializer.
	if cf, ok := e.(*syntax.ClassConstFetch); ok {
		if n, named := cf.Class.(*syntax.Name); named && (!strings.EqualFold(n.Value, "static") || unextended(ctx, cf.Class)) {
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

// unextended reports whether `static` names a class that is final or that
// no class of the project extends, under its name or an alias.
func unextended(ctx *analysis.Context, static syntax.Expr) bool {
	fqn := ctx.Types().ClassRef(static)
	c := ctx.Index().Class(fqn, ctx.PHP)
	return c != nil && c.Kind == syntax.KindClass && (c.Final || len(ctx.Index().Subclasses(fqn)) == 0 && len(ctx.Index().AliasesOf(fqn)) == 0)
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
