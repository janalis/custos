package phpunittests

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/phpunit"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// putSuggestion is the outcome of an assertion strategy: the new method
// name and the new argument list (texts; "" slots become `null`).
type putSuggestion struct {
	name  string
	slots []string
	msg   string // overrides the default message when set
	noFix bool   // the rewrite is not provably equivalent: report only
}
type putStrategy func(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool)

func putCheckCall(ctx *analysis.Context, c phpunit.PuCall) {
	m := c.Name
	switch {
	case m == "expects":
		if ctx.Bool("PROMOTE_MOCKING_ONCE") {
			putExpectsOnce(ctx, c)
		}
		return
	case m == "will":
		if ctx.Bool("PROMOTE_MOCKING_WILL_RETURN") {
			putWillReturn(ctx, c)
		}
		return
	case !strings.HasPrefix(strings.ToLower(m), "assert") || m == "assert":
		return
	}
	args, ok := c.Values()
	if !ok {
		return
	}
	strategies := []putStrategy{putInvertedBool, putBoolOfComparison}
	if ctx.Bool("SUGGEST_TO_USE_ASSERTSAME") {
		strategies = append(strategies, putStrictEquality)
	}
	if ctx.Bool("PROMOTE_PHPUNIT_API") {
		strategies = append(strategies, putEmpty, putConstant, putInternalType, putInstanceof, putGetClass,
			putResourceExists, putCount, putContains, putRegexNumeric, putRegexComparison, putFileEquals,
			putStringEqualsFile)
	}
	for _, s := range strategies {
		sug, ok := s(ctx, c, args)
		if !ok {
			continue
		}
		if putLacksMethod(ctx, c, sug.name) { // E12
			return
		}
		msg := sug.msg
		if msg == "" {
			msg = "Use '" + sug.name + "()' instead."
		}
		var fixes []diagnostic.Fix
		if !sug.noFix {
			fixes = append(fixes, putRenameFix(ctx, c, sug.name, sug.slots))
		}
		ctx.ReportNode(c.Node, msg, fixes...)
		return
	}
}

// putLacksMethod reports whether the receiver is a class whose whole
// hierarchy is known, which declares the called assertion but no method
// named name, or one that PHPUnit does not declare: a test base of its own
// (the redis extension's TestSuite, whose assertEquals() is strict), not
// PHPUnit.
func putLacksMethod(ctx *analysis.Context, c phpunit.PuCall, name string) bool {
	var classes []string
	if n, named := c.Recv.(*syntax.Name); named {
		classes = []string{phpunit.PuResolveClassName(ctx, n)}
	} else {
		classes = ctx.TypeOf(c.Recv).Classes()
	}
	if len(classes) != 1 {
		return false
	}
	ix, cls := ctx.Index(), strings.TrimPrefix(classes[0], `\`)
	if !ix.AncestorsComplete(cls, ctx.PHP) || ix.FindMethod(cls, c.Ident.Value, ctx.PHP) == nil {
		return false
	}
	sug := ix.FindMethod(cls, name, ctx.PHP)
	return sug == nil || !strings.HasPrefix(strings.ToLower(strings.TrimPrefix(sug.Class, `\`)), `phpunit\`)
}

// putRenameFix renames the call and rebuilds its argument list.
func putRenameFix(ctx *analysis.Context, c phpunit.PuCall, name string, slots []string) diagnostic.Fix {
	start, end := c.Ident.Span().Start, c.Node.Span().End
	return diagnostic.Fix{
		Title: "Use " + name + "()",
		Edits: func() []diagnostic.TextEdit {
			// every slot holds source text: putSlotsFrom never leaves a gap
			return []diagnostic.TextEdit{{Span: syntax.Span{Start: start, End: end}, NewText: name + "(" + strings.Join(slots, ", ") + ")"}}
		},
	}
}

// putSlotsFrom builds an argument list of exactly n entries from the given
// leading values (extra values are dropped; the last value is the message,
// args[msg]). Slots left over hold the original arguments that follow the
// message, in order (upstream pads them with `null`; see spec Divergences).
func putSlotsFrom(ctx *analysis.Context, args []syntax.Expr, msg, n int, vals ...string) []string {
	out := make([]string, n)
	copy(out, vals)
	for i := len(vals); i < n; i++ {
		if j := msg + 1 + i - len(vals); j < len(args) {
			out[i] = ctx.Text(args[j])
		}
	}
	return out
}

// putOpt returns the text of args[i], or "" when absent.
func putOpt(ctx *analysis.Context, args []syntax.Expr, i int) string {
	if i < len(args) {
		return ctx.Text(args[i])
	}
	return ""
}

func putIn(m string, set ...string) bool {
	for _, s := range set {
		if m == s {
			return true
		}
	}
	return false
}

// putPositive classifies assertTrue/assertNotFalse (true) vs
// assertFalse/assertNotTrue (false); ok false for other names.
func putPositive(m string) (positive, ok bool) {
	switch m {
	case "assertTrue", "assertNotFalse":
		return true, true
	case "assertFalse", "assertNotTrue":
		return false, true
	}
	return false, false
}

// putFunc matches a call resolving to the global function f (names compare
// case-insensitively; a same-named namespaced function does not count) and
// returns its positional arguments.
func putFunc(ctx *analysis.Context, e syntax.Expr, f string) ([]syntax.Expr, bool) {
	call, ok := e.(*syntax.FuncCall)
	if !ok || !ctx.IsGlobalFunctionCall(call, f) {
		return nil, false
	}
	return astquery.CallArgValues(call)
}

// putStrictTypes reports whether the file declares strict_types=1.
func putStrictTypes(ctx *analysis.Context) bool {
	for _, st := range ctx.File.Stmts {
		if d, ok := st.(*syntax.Declare); ok {
			for _, it := range d.Items {
				if strings.EqualFold(it.Key.Value, "strict_types") && ctx.Text(it.Value) == "1" {
					return true
				}
			}
		}
	}
	return false
}

// putFits reports whether e may be moved into a PHPUnit parameter declared
// with the given scalar type: its type is known to be exactly that, or it
// is unknown and the file coerces scalar arguments (no strict_types).
func putFits(ctx *analysis.Context, e syntax.Expr, atom string) bool {
	t := ctx.TypeOf(e)
	if t.IsUnknown() {
		return !putStrictTypes(ctx)
	}
	return t.OnlyOf(atom)
}

// putIsBool reports whether x always evaluates to a bool.
func putIsBool(ctx *analysis.Context, x syntax.Expr) bool {
	return ctx.TypeOf(x).OnlyOf("bool", "true", "false")
}

// D15
func putInvertedBool(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	m := c.Name
	if !putIn(m, "assertTrue", "assertFalse") || len(args) < 1 {
		return putSuggestion{}, false
	}
	u, ok := syntax.UnwrapParens(args[0]).(*syntax.Unary)
	if !ok || u.Op.Kind != syntax.TExclaim {
		return putSuggestion{}, false
	}
	// D15a: `!empty(Y)` names the final assertion when D18 accepts Y.
	if e, ok := syntax.UnwrapParens(u.Expr).(*syntax.Empty); ok && e.Expr != nil && ctx.Bool("PROMOTE_PHPUNIT_API") && putEmptyOperand(ctx, e.Expr) {
		name := "assertEmpty"
		if m == "assertTrue" {
			name = "assertNotEmpty"
		}
		return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 1, len(args), ctx.Text(e.Expr), putOpt(ctx, args, 1))}, true
	}
	name := "assertNotTrue"
	if m == "assertFalse" {
		name = "assertNotFalse"
	}
	x := syntax.UnwrapParens(u.Expr)
	slots := []string{ctx.Text(x)}
	if len(args) == 2 {
		slots = append(slots, ctx.Text(args[1]))
	}
	// assertNotTrue(X) accepts every X other than true, `!X` only falsy ones
	return putSuggestion{name: name, slots: slots, noFix: !putIsBool(ctx, x)}, true
}

// putMayBeFloat reports whether e is not known to exclude float values.
func putMayBeFloat(ctx *analysis.Context, e syntax.Expr) bool {
	t := ctx.TypeOf(e)
	return t.IsUnknown() || t.HasAny("float", "mixed")
}

// D16
func putBoolOfComparison(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	m := c.Name
	if !putIn(m, "assertTrue", "assertNotTrue", "assertFalse", "assertNotFalse") || len(args) < 1 {
		return putSuggestion{}, false
	}
	b, ok := syntax.UnwrapParens(args[0]).(*syntax.Binary)
	if !ok {
		return putSuggestion{}, false
	}
	var opNeg, strict bool
	switch b.Op.Kind {
	case syntax.TIsEqual:
	case syntax.TIsNotEqual:
		opNeg = true
	case syntax.TIsIdentical:
		strict = true
	case syntax.TIsNotIdentical:
		opNeg, strict = true, true
	default:
		return putSuggestion{}, false
	}
	methodNeg := m == "assertFalse" || m == "assertNotTrue"
	name := "assert"
	if methodNeg != opNeg {
		name += "Not"
	}
	if strict {
		name += "Same"
	} else {
		name += "Equals"
	}
	slots := []string{ctx.Text(b.Left), ctx.Text(b.Right)}
	if len(args) == 2 {
		slots = append(slots, ctx.Text(args[1]))
	}
	var noFix bool
	if strict {
		// PHPUnit below 10 compares two floats with a tolerance
		noFix = putVersion(ctx) < 100 && putMayBeFloat(ctx, b.Left) && putMayBeFloat(ctx, b.Right)
	} else {
		// assertEquals() agrees with `==` on these types only (two numeric
		// strings, arrays, objects and floats follow PHPUnit's own rules)
		loose := []string{"int", "bool", "true", "false", "null"}
		noFix = !ctx.TypeOf(b.Left).OnlyOf(loose...) || !ctx.TypeOf(b.Right).OnlyOf(loose...)
	}
	return putSuggestion{name: name, slots: slots, noFix: noFix}, true
}

// D17
func putStrictEquality(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	var name string
	switch c.Name {
	case "assertEquals":
		name = "assertSame"
	case "assertNotEquals":
		name = "assertNotSame"
	default:
		return putSuggestion{}, false
	}
	if len(args) < 2 || !putScalarTyped(ctx, args[0]) || !putScalarTyped(ctx, args[1]) {
		return putSuggestion{}, false
	}
	slots := make([]string, len(args))
	for i, a := range args {
		slots[i] = ctx.Text(a)
	}
	return putSuggestion{name: name, slots: slots, msg: "Loose comparison; use '" + name + "()' instead."}, true
}

func putScalarTyped(ctx *analysis.Context, e syntax.Expr) bool {
	t := ctx.TypeOf(e)
	if t.IsUnknown() {
		return false
	}
	for _, a := range t.Atoms() {
		switch a {
		case "int", "float", "string", "bool", "true", "false", "null":
		default:
			return false
		}
	}
	return true
}

// D18
func putEmpty(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	pos, ok := putPositive(c.Name)
	if !ok || len(args) < 1 {
		return putSuggestion{}, false
	}
	e, ok := args[0].(*syntax.Empty)
	if !ok || e.Expr == nil || !putEmptyOperand(ctx, e.Expr) {
		return putSuggestion{}, false
	}
	name := "assertNotEmpty"
	if pos {
		name = "assertEmpty"
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 1, len(args), ctx.Text(e.Expr), putOpt(ctx, args, 1))}, true
}

// putDefined reports whether the local variable v is certainly defined
// where it is read: it is a parameter, or an assignment always precedes it.
func putDefined(ctx *analysis.Context, v *syntax.Variable) bool {
	scope := syntax.EnclosingFuncLike(v)
	params, _ := astquery.ScopeParts(scope)
	for _, p := range params {
		if p.Var != nil && p.Var.Name == v.Name {
			return true
		}
	}
	_, entry := flowquery.ReachingAssignmentsIn(ctx.File, scope, v, v.Name)
	return !entry
}

// putEmptyOperand reports whether x has a fully known type made only of
// null, bool, int, float, string and array: there assertEmpty() agrees
// with empty() (objects, Countable ones especially, do not). x must also be
// a plain variable or a call: empty() tolerates a missing offset or an
// uninitialized property, reading them as an argument does not.
func putEmptyOperand(ctx *analysis.Context, x syntax.Expr) bool {
	switch v := x.(type) {
	case *syntax.Variable:
		if !putDefined(ctx, v) {
			return false
		}
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
	default:
		return false
	}
	t := ctx.TypeOf(x)
	if t.IsUnknown() {
		return false
	}
	for _, a := range t.Atoms() {
		switch a {
		case "null", "bool", "true", "false", "int", "float", "string", "array":
		default:
			if !strings.HasSuffix(a, "[]") {
				return false
			}
		}
	}
	return true
}

// D19
func putConstant(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	m := c.Name
	if !putIn(m, "assertSame", "assertNotSame") || len(args) < 2 {
		return putSuggestion{}, false
	}
	idx, kind := -1, ""
	for i, a := range args[:2] { // a later argument is the message
		cf, ok := a.(*syntax.ConstFetch)
		if !ok {
			continue
		}
		switch strings.ToLower(cf.Name.Value) {
		case "null":
			kind = "Null"
		case "true":
			kind = "True"
		case "false":
			kind = "False"
		default:
			continue
		}
		idx = i
		break
	}
	if idx < 0 {
		return putSuggestion{}, false
	}
	other := ""
	for i, a := range args {
		if i != idx {
			other = ctx.Text(a)
			break
		}
	}
	name := "assert" + kind
	if m == "assertNotSame" {
		name = "assertNot" + kind
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 2, len(args)-1, other, putOpt(ctx, args, 2))}, true
}

// putInternalTypes maps the is_* checks to PHPUnit's type names. custos:
// is_resource() is left out, since PHPUnit's "resource" type also accepts a
// closed resource, for which is_resource() is false.
var putInternalTypes = map[string]string{
	"is_array": "array", "is_bool": "bool", "is_float": "float", "is_int": "int", "is_null": "null",
	"is_numeric": "numeric", "is_object": "object", "is_string": "string",
	"is_scalar": "scalar", "is_callable": "callable", "is_iterable": "iterable",
}

// D20
func putInternalType(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	pos, ok := putPositive(c.Name)
	if !ok || len(args) < 1 {
		return putSuggestion{}, false
	}
	call, ok := args[0].(*syntax.FuncCall)
	if !ok {
		return putSuggestion{}, false
	}
	typ, known := putInternalTypes[ctx.GlobalFunctionName(call)]
	if !known {
		return putSuggestion{}, false
	}
	inner, ok := astquery.CallArgValues(call)
	if !ok || len(inner) != 1 { // is_callable()'s extra arguments change its meaning
		return putSuggestion{}, false
	}
	a0 := ctx.Text(inner[0])
	if putVersion(ctx) < 80 {
		name := "assertNotInternalType"
		if pos {
			name = "assertInternalType"
		}
		return putSuggestion{
			name:  name,
			slots: putSlotsFrom(ctx, args, 1, len(args)+1, "'"+typ+"'", a0, putOpt(ctx, args, 1)),
			msg:   "Use '" + name + "('" + typ + "', …)' instead.",
		}, true
	}
	var name string
	if typ == "null" { // divergence: assertIsNull does not exist in PHPUnit
		name = "assertNotNull"
		if pos {
			name = "assertNull"
		}
	} else {
		t := strings.ToUpper(typ[:1]) + typ[1:]
		name = "assertIsNot" + t
		if pos {
			name = "assertIs" + t
		}
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 1, len(args), a0, putOpt(ctx, args, 1))}, true
}

// D21
func putInstanceof(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	pos, ok := putPositive(c.Name)
	if !ok || len(args) < 1 {
		return putSuggestion{}, false
	}
	io, ok := args[0].(*syntax.Instanceof)
	if !ok || io.Class == nil || io.Expr == nil {
		return putSuggestion{}, false
	}
	var cd string
	noFix := false
	if nm, ok := io.Class.(*syntax.Name); ok {
		if ctx.PHP >= phpversion.PHP55 {
			cd = nm.Value + "::class"
		} else {
			fqn := phpunit.PuResolveClassName(ctx, nm)
			if fqn == "" {
				fqn = strings.TrimPrefix(nm.Value, `\`)
			}
			cd = "'" + strings.ReplaceAll(`\`+fqn, `\`, `\\`) + "'"
		}
	} else {
		cd = ctx.Text(io.Class)
		// instanceof also takes an object, which no mode coerces to the
		// string assertInstanceOf() declares
		noFix = !ctx.TypeOf(io.Class).OnlyOf("string")
	}
	name := "assertNotInstanceOf"
	if pos {
		name = "assertInstanceOf"
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 1, len(args)+1, cd, ctx.Text(io.Expr), putOpt(ctx, args, 1)), noFix: noFix}, true
}

// D22
func putGetClass(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	var name string
	switch c.Name {
	case "assertSame", "assertEquals":
		name = "assertInstanceOf"
	case "assertNotSame", "assertNotEquals":
		name = "assertNotInstanceOf"
	default:
		return putSuggestion{}, false
	}
	if len(args) < 2 {
		return putSuggestion{}, false
	}
	litIdx := 1
	if _, _, ok := astquery.QuotedStringRaw(args[0]); ok {
		litIdx = 0
	}
	content, quote, ok := astquery.QuotedStringRaw(args[litIdx])
	if !ok || len(content) <= 3 {
		return putSuggestion{}, false
	}
	inner, ok := putFunc(ctx, args[1-litIdx], "get_class")
	if !ok || len(inner) != 1 {
		return putSuggestion{}, false
	}
	// the literal must spell a class name (a lone backslash in double
	// quotes may start an escape sequence)
	lit := strings.ReplaceAll(content, `\\`, `\`)
	if !putIsQualifiedName(lit) || quote == '"' && strings.Contains(strings.ReplaceAll(content, `\\`, ""), `\`) {
		return putSuggestion{}, false
	}
	// get_class() names the exact class, instanceof accepts subclasses too:
	// the two only agree for a final class spelled as declared
	cls := ctx.Index().Class(lit, ctx.PHP)
	exact := cls != nil && cls.FQN == lit && (cls.Final || cls.Kind == syntax.KindEnum)
	fq := `\` + strings.TrimPrefix(lit, `\`) // divergence: no double leading backslash
	var cd string
	if ctx.PHP >= phpversion.PHP55 {
		cd = fq + "::class"
	} else {
		cd = "'" + strings.ReplaceAll(fq, `\`, `\\`) + "'"
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 2, len(args), cd, ctx.Text(inner[0]), putOpt(ctx, args, 2)), noFix: !exact}, true
}

// D23
func putResourceExists(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	pos, ok := putPositive(c.Name)
	if !ok || len(args) < 1 {
		return putSuggestion{}, false
	}
	kind := "File"
	inner, ok := putFunc(ctx, args[0], "file_exists")
	if !ok {
		kind = "Directory"
		inner, ok = putFunc(ctx, args[0], "is_dir")
	}
	if !ok || len(inner) != 1 {
		return putSuggestion{}, false
	}
	name := "assert" + kind + "DoesNotExist"
	if putVersion(ctx) < 91 { // the DoesNotExist names only exist from PHPUnit 9.1
		name = "assert" + kind + "NotExists"
	}
	if pos {
		name = "assert" + kind + "Exists"
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 1, len(args), ctx.Text(inner[0]), putOpt(ctx, args, 1))}, true
}

// D24
func putCount(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	var name string
	switch c.Name {
	case "assertSame", "assertEquals":
		name = "assertCount"
	case "assertNotSame", "assertNotEquals":
		name = "assertNotCount"
	default:
		return putSuggestion{}, false
	}
	if len(args) < 2 {
		return putSuggestion{}, false
	}
	inner, ok := putFunc(ctx, args[1], "count")
	if !ok || len(inner) != 1 {
		return putSuggestion{}, false
	}
	// assertCount() declares an int expected count
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 2, len(args), ctx.Text(args[0]), ctx.Text(inner[0]), putOpt(ctx, args, 2)), noFix: !putFits(ctx, args[0], "int")}, true
}

// D25
func putContains(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	if putVersion(ctx) >= 90 {
		return putSuggestion{}, false
	}
	pos, ok := putPositive(c.Name)
	if !ok || len(args) < 1 {
		return putSuggestion{}, false
	}
	inner, ok := putFunc(ctx, args[0], "in_array")
	if !ok || len(inner) < 2 {
		return putSuggestion{}, false
	}
	name := "assertNotContains"
	if pos {
		name = "assertContains"
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 1, len(args)+1, ctx.Text(inner[0]), ctx.Text(inner[1]), putOpt(ctx, args, 1))}, true
}

// putRegexName returns the regex assertion name for the configured PHPUnit
// version: assertMatchesRegularExpression / assertDoesNotMatchRegularExpression
// from 9.1 (where assertRegExp is deprecated), assertRegExp / assertNotRegExp
// below.
func putRegexName(ctx *analysis.Context, positive bool) string {
	if putVersion(ctx) >= 91 {
		if positive {
			return "assertMatchesRegularExpression"
		}
		return "assertDoesNotMatchRegularExpression"
	}
	if positive {
		return "assertRegExp"
	}
	return "assertNotRegExp"
}

// D26
func putRegexNumeric(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	var expected string
	switch c.Name {
	case "assertSame", "assertEquals":
		expected = "1"
	case "assertNotSame", "assertNotEquals":
		expected = "0"
	default:
		return putSuggestion{}, false
	}
	if len(args) < 2 || !astquery.IsNumberLiteral(args[0]) {
		return putSuggestion{}, false
	}
	inner, ok := putFunc(ctx, args[1], "preg_match")
	if !ok || len(inner) != 2 {
		return putSuggestion{}, false
	}
	name := putRegexName(ctx, ctx.Text(args[0]) == expected)
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 2, len(args), ctx.Text(inner[0]), ctx.Text(inner[1]), putOpt(ctx, args, 2))}, true
}

// D27
func putRegexComparison(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	var name string
	switch c.Name {
	case "assertTrue":
		name = putRegexName(ctx, true)
	case "assertFalse":
		name = putRegexName(ctx, false)
	default:
		return putSuggestion{}, false
	}
	if len(args) < 1 {
		return putSuggestion{}, false
	}
	b, ok := args[0].(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TGreater || !astquery.IsNumberLiteral(b.Right) || ctx.Text(b.Right) != "0" {
		return putSuggestion{}, false
	}
	inner, ok := putFunc(ctx, b.Left, "preg_match")
	if !ok || len(inner) != 2 {
		return putSuggestion{}, false
	}
	return putSuggestion{name: name, slots: putSlotsFrom(ctx, args, 1, len(args)+1, ctx.Text(inner[0]), ctx.Text(inner[1]), putOpt(ctx, args, 1))}, true
}

// putEnclosingFuncName returns the name of the nearest enclosing function
// or method ("" for closures and top-level code).
func putEnclosingFuncName(n syntax.Node) string {
	switch f := syntax.EnclosingFuncLike(n).(type) {
	case *syntax.Function:
		return f.Name.Value
	case *syntax.Method:
		return f.Name.Value
	}
	return ""
}

func putFileContentsArg(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	inner, ok := putFunc(ctx, e, "file_get_contents")
	if !ok || len(inner) != 1 {
		return nil
	}
	return inner[0]
}

// D28a
func putFileEquals(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	m := c.Name
	if !putIn(m, "assertSame", "assertEquals", "assertStringEqualsFile") || len(args) < 2 ||
		strings.EqualFold(putEnclosingFuncName(c.Node), "assertFileEquals") {
		return putSuggestion{}, false
	}
	e0, e1 := putFileContentsArg(ctx, args[0]), putFileContentsArg(ctx, args[1])
	var first string
	switch {
	case m == "assertStringEqualsFile" && e1 != nil:
		first = ctx.Text(args[0])
	case e0 != nil && e1 != nil:
		first = ctx.Text(e0)
	default:
		return putSuggestion{}, false
	}
	return putSuggestion{name: "assertFileEquals", slots: putSlotsFrom(ctx, args, 2, len(args), first, ctx.Text(e1), putOpt(ctx, args, 2))}, true
}

// D28b
func putStringEqualsFile(ctx *analysis.Context, c phpunit.PuCall, args []syntax.Expr) (putSuggestion, bool) {
	if !putIn(c.Name, "assertSame", "assertEquals") || len(args) < 2 {
		return putSuggestion{}, false
	}
	if fn := putEnclosingFuncName(c.Node); strings.EqualFold(fn, "assertFileEquals") || strings.EqualFold(fn, "assertStringEqualsFile") {
		return putSuggestion{}, false
	}
	f := putFileContentsArg(ctx, args[0])
	if f == nil {
		return putSuggestion{}, false
	}
	// assertStringEqualsFile() declares a string actual value
	return putSuggestion{name: "assertStringEqualsFile", slots: putSlotsFrom(ctx, args, 2, len(args), ctx.Text(f), ctx.Text(args[1]), putOpt(ctx, args, 2)), noFix: !putFits(ctx, args[1], "string")}, true
}

// D29
func putExpectsOnce(ctx *analysis.Context, c phpunit.PuCall) {
	args, ok := c.Values()
	if !ok || len(args) != 1 {
		return
	}
	inner, ok := phpunit.PuMethodCallNamed(args[0], "exactly")
	if !ok {
		return
	}
	iargs, ok := inner.Values()
	if !ok || len(iargs) != 1 || !astquery.IsNumberLiteral(iargs[0]) || ctx.Text(iargs[0]) != "1" {
		return
	}
	ctx.ReportNode(inner.Node, "Use '->once()' instead.", putRenameFix(ctx, inner, "once", nil))
}

var putWillMap = map[string]string{
	"returnValue":    "willReturn",
	"returnValueMap": "willReturnMap",
	"returnCallback": "willReturnCallback",
	"returnArgument": "willReturnArgument",
}

// D30
func putWillReturn(ctx *analysis.Context, c phpunit.PuCall) {
	args, ok := c.Values()
	if !ok || len(args) != 1 {
		return
	}
	inner, ok := phpunit.AsPuCall(args[0])
	if !ok {
		return
	}
	name, ok := putWillMap[inner.Name]
	if !ok {
		return
	}
	iargs, ok := inner.Values()
	if !ok || len(iargs) != 1 {
		return
	}
	ctx.ReportNode(c.Node, "Use '->"+name+"()' instead.", putRenameFix(ctx, c, name, []string{ctx.Text(iargs[0])}))
}
