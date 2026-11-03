package jsonencodingapiusage

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/phpunit"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// jsonEncodingAPIUsage asks json_decode() for an explicit result type and
// json_decode()/json_encode() for JSON_THROW_ON_ERROR.
type jsonEncodingAPIUsage struct{}

const (
	jsonTypeMsg  = "Pass the second argument to state whether JSON decodes to arrays or objects."
	jsonThrowMsg = "Pass JSON_THROW_ON_ERROR in the flags of this call."
)

func (jsonEncodingAPIUsage) ID() string               { return "JsonEncodingApiUsage" }
func (jsonEncodingAPIUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (jsonEncodingAPIUsage) Semantic()                {}

// jsonArg looks an argument up by name, else by unnamed position.
func jsonArg(call *syntax.FuncCall, name string, pos int) (*syntax.Arg, bool) {
	for _, a := range call.Args.Args {
		if arg, ok := a.(*syntax.Arg); ok && arg.Name != nil && arg.Name.Value == name {
			return arg, true
		}
	}
	if pos < len(call.Args.Args) {
		if arg, ok := call.Args.Args[pos].(*syntax.Arg); ok && arg.Name == nil && !arg.Unpack && arg.Value != nil {
			return arg, true
		}
	}
	return nil, false
}

func (jsonEncodingAPIUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	qual, name, ok := astquery.CallName(call)
	lname := strings.ToLower(name)
	if !ok || (lname != "json_decode" && lname != "json_encode") || call.Args == nil || len(call.Args.Args) == 0 { // E1
		return
	}
	if !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) { // D0, E2
		return
	}
	for _, a := range call.Args.Args { // E6: `...$args` and `f(...)` hide the bound arguments
		if arg, ok := a.(*syntax.Arg); !ok || arg.Unpack {
			return
		}
	}
	decode := lname == "json_decode"
	span := call.Span()
	text := func(a *syntax.Arg) string { return ctx.Text(a.Value) }

	if decode && ctx.Bool("HARDEN_DECODING_RESULT_TYPE") { // D1
		_, has := jsonArg(call, "associative", 1)
		arrayFlag := jsonFlagAbsent
		if f, ok := jsonArg(call, "flags", 3); ok && !has {
			arrayFlag = jsonObjectAsArray(ctx, f.Value)
		}
		if !has && arrayFlag != jsonFlagPresent { // E7
			var fixes []diagnostic.Fix
			b := "false"
			if ctx.Bool("DECODE_AS_ARRAY") {
				b = "true"
			}
			if jsonHasNamedArg(call) { // F1b: keep the named arguments
				// F1c: no fix when the flags may hold JSON_OBJECT_AS_ARRAY,
				// which an explicit boolean would override.
				if at := call.Args.Args[len(call.Args.Args)-1].Span().End; arrayFlag == jsonFlagAbsent {
					fixes = append(fixes, diagnostic.Fix{Title: "Pass the result type", Edits: func() []diagnostic.TextEdit {
						return []diagnostic.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: ", associative: " + b}}
					}})
				}
			} else if first, ok := call.Args.Args[0].(*syntax.Arg); ok && first.Value != nil {
				newText := qual + name + "(" + text(first) + ", " + b + ")"
				fixes = append(fixes, diagnostic.Fix{Title: "Pass the result type", Edits: func() []diagnostic.TextEdit {
					return []diagnostic.TextEdit{{Span: span, NewText: newText}}
				}})
			}
			ctx.Report(span, jsonTypeMsg, fixes...)
		}
	}

	if !ctx.Bool("HARDEN_ERRORS_HANDLING") || ctx.PHP.Below(phpversion.PHP73) { // D2
		return
	}
	subjectName, flagsPos := "value", 1
	if decode {
		subjectName, flagsPos = "json", 3
	}
	subject, ok := jsonArg(call, subjectName, 0) // D3
	if !ok {
		return
	}
	if !decode && jsonSafeEncode(ctx, call) { // E10
		return
	}
	if decode && jsonDecodesSafeEncode(ctx, call, subject.Value) { // E11
		return
	}
	flags, hasFlags := jsonArg(call, "flags", flagsPos)
	if hasFlags && jsonStrictFlags(ctx, flags.Value) { // D4, D5
		return
	}
	if jsonFailureHandled(ctx, call, decode) || !decode && jsonRoundTrip(ctx, call) { // E8
		return
	}
	if decode && jsonMayBeNull(ctx, subject.Value) { // F7
		// With the flag, the null text a caller may rely on decoding to
		// null throws instead.
		ctx.Report(span, jsonThrowMsg)
		return
	}
	throwConst := jsonThrowConst(ctx, span.Start)
	if hasFlags && flags.Name != nil { // F4: add the constant to the named flags
		val := flags.Value
		old := ctx.Text(val)
		if jsonNeedsParens(val) {
			old = "(" + old + ")"
		}
		ctx.Report(span, jsonThrowMsg, diagnostic.Fix{Title: "Add JSON_THROW_ON_ERROR", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: val.Span(), NewText: throwConst + " | " + old}}
		}})
		return
	}
	if jsonHasNamedArg(call) { // F5: append a named flags argument
		at := call.Args.Args[len(call.Args.Args)-1].Span().End
		ctx.Report(span, jsonThrowMsg, diagnostic.Fix{Title: "Add JSON_THROW_ON_ERROR", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: ", flags: " + throwConst}}
		}})
		return
	}
	newFlags := throwConst
	if hasFlags {
		old := text(flags)
		if jsonNeedsParens(flags.Value) {
			old = "(" + old + ")"
		}
		newFlags += " | " + old
	}
	var newText string
	if decode { // F2
		assoc := "false"
		if a, ok := jsonArg(call, "associative", 1); ok {
			assoc = text(a)
		} else if ctx.Bool("HARDEN_DECODING_RESULT_TYPE") && ctx.Bool("DECODE_AS_ARRAY") {
			assoc = "true"
		}
		depth := "512"
		if d, ok := jsonArg(call, "depth", 2); ok {
			depth = text(d)
		}
		newText = qual + name + "(" + text(subject) + ", " + assoc + ", " + depth + ", " + newFlags + ")"
	} else { // F3
		newText = qual + name + "(" + text(subject) + ", " + newFlags
		if d, ok := jsonArg(call, "depth", 2); ok {
			newText += ", " + text(d)
		}
		newText += ")"
	}
	ctx.Report(span, jsonThrowMsg, diagnostic.Fix{Title: "Add JSON_THROW_ON_ERROR", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: span, NewText: newText}}
	}})
}

// State of JSON_OBJECT_AS_ARRAY in a flags expression.
const (
	jsonFlagAbsent = iota
	jsonFlagPresent
	jsonFlagUnknown
)

// jsonDecodeFlags are the built-in flags json_decode() accepts besides
// JSON_OBJECT_AS_ARRAY; none of them shares its bit.
var jsonDecodeFlags = map[string]bool{
	"JSON_BIGINT_AS_STRING":        true,
	"JSON_INVALID_UTF8_IGNORE":     true,
	"JSON_INVALID_UTF8_SUBSTITUTE": true,
	"JSON_PARTIAL_OUTPUT_ON_ERROR": true,
	"JSON_THROW_ON_ERROR":          true,
}

// jsonObjectAsArray tells whether a flags expression sets
// JSON_OBJECT_AS_ARRAY, reading only `|` chains of built-in constants and
// the literal 0; anything else is unknown.
func jsonObjectAsArray(ctx *analysis.Context, e syntax.Expr) int {
	switch x := e.(type) {
	case *syntax.Paren:
		return jsonObjectAsArray(ctx, x.Expr)
	case *syntax.Binary:
		if x.Op.Kind != syntax.TBar {
			return jsonFlagUnknown
		}
		l, r := jsonObjectAsArray(ctx, x.Left), jsonObjectAsArray(ctx, x.Right)
		switch {
		case l == jsonFlagPresent || r == jsonFlagPresent:
			return jsonFlagPresent
		case l == jsonFlagAbsent && r == jsonFlagAbsent:
			return jsonFlagAbsent
		}
	case *syntax.ConstFetch:
		name := semanticquery.GlobalConstName(ctx, x)
		if name == "JSON_OBJECT_AS_ARRAY" {
			return jsonFlagPresent
		}
		if jsonDecodeFlags[name] {
			return jsonFlagAbsent
		}
	case *syntax.Literal:
		if x.LitKind == syntax.LitInt && x.Raw == "0" {
			return jsonFlagAbsent
		}
	}
	return jsonFlagUnknown
}

// jsonNeedsParens reports whether flags must be parenthesised as the right
// operand of `JSON_THROW_ON_ERROR | …`: operators binding more loosely
// than `|` (logical, `??`), ternaries, assignments and the keyword
// expressions. Bitwise, comparison and arithmetic operands keep their value.
func jsonNeedsParens(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TOr, syntax.TXor, syntax.TAnd, syntax.TCoalesce, syntax.TBooleanOr, syntax.TBooleanAnd:
			return true
		}
		return false
	case *syntax.Ternary, *syntax.Assign, *syntax.Print, *syntax.Yield,
		*syntax.YieldFrom, *syntax.Include, *syntax.Throw, *syntax.ArrowFunction:
		return true
	}
	return false
}

// jsonHasNamedArg reports whether the call passes any argument by name.
func jsonHasNamedArg(call *syntax.FuncCall) bool {
	for _, a := range call.Args.Args {
		if arg, ok := a.(*syntax.Arg); ok && arg.Name != nil {
			return true
		}
	}
	return false
}

// jsonThrowConst spells JSON_THROW_ON_ERROR for a fix at offset at (F6): with
// a leading backslash when the file already writes a global constant that
// way (`\PHP_EOL`, `\JSON_PRETTY_PRINT`), or when a bare name there would
// not reach the global constant.
func jsonThrowConst(ctx *analysis.Context, at uint32) string {
	const name = "JSON_THROW_ON_ERROR"
	if q := semanticquery.QualifiedGlobalConst(ctx, name, at); q != name {
		return q
	}
	qualified := ctx.Memo("qualifiedConsts", func() any {
		found := false
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.ConstFetch); ok && c.Name != nil && strings.HasPrefix(c.Name.Value, `\`) && semanticquery.GlobalConstName(ctx, c) != "" {
				switch strings.ToLower(c.Name.Value) {
				case `\true`, `\false`, `\null`:
				default:
					found = true
				}
			}
			return !found
		})
		return found
	}).(bool)
	if qualified {
		return `\` + name
	}
	return name
}

// jsonStrictFlags implements D5; an unknown discovery result counts as
// strict so that kind E is not reported (D5a).
func jsonStrictFlags(ctx *analysis.Context, f syntax.Expr) bool {
	values, known := semanticquery.DiscoverValuesKnown(ctx.Types(), f)
	if !known { // D5a
		return true
	}
	if len(values) != 1 {
		return false
	}
	v := values[0]
	num := v
	if u, ok := v.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		num = u.Expr
	}
	if lit, ok := num.(*syntax.Literal); ok && lit.LitKind != syntax.LitString {
		return lit.Raw == "4194304" || lit.Raw == "512"
	}
	strict := false
	syntax.Inspect(v, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ConstFetch); ok && c.Name != nil {
			switch astquery.LastNamePart(c.Name.Value) {
			case "JSON_THROW_ON_ERROR", "JSON_PARTIAL_OUTPUT_ON_ERROR":
				strict = true
			}
		}
		return !strict
	})
	return strict
}

// jsonTypeChecks are the type tests that tell a JSON result from the
// failure value (E8).
var jsonTypeChecks = map[string]bool{
	"is_array": true, "is_string": true, "is_object": true, "is_null": true, "is_bool": true,
	"is_int": true, "is_float": true, "is_scalar": true, "is_iterable": true, "is_countable": true,
}

// jsonFailureHandled reports whether the result of call is tested for the
// failure value where it is produced (E8): the call itself (or the ternary
// it is a branch of), or, when a plain assignment stores it in a variable,
// the assignment or that variable's first read after it in the same scope;
// or the statement producing it is followed by an `if` asking
// json_last_error().
func jsonFailureHandled(ctx *analysis.Context, call *syntax.FuncCall, decode bool) bool {
	x := jsonOuter(call)
	// A branch of a ternary is produced where the ternary is.
	for t, ok := x.Parent().(*syntax.Ternary); ok && syntax.Node(t.Cond) != x; t, ok = x.Parent().(*syntax.Ternary) {
		x = jsonOuter(t)
	}
	a, assigned := x.Parent().(*syntax.Assign)
	assigned = assigned && a.Op.Kind == syntax.TEqual && !a.ByRef && syntax.Node(a.Value) == x
	stmt := x
	if assigned {
		stmt = jsonOuter(a)
	}
	if es, ok := stmt.Parent().(*syntax.ExprStmt); ok && jsonLastErrorTested(ctx, es) {
		return true
	}
	if assigned {
		if jsonFailureTest(ctx, jsonOuter(a), decode) {
			return true // `if ($data = json_decode(…))`
		}
		var read syntax.Expr
		switch t := syntax.UnwrapParens(a.Var).(type) {
		case *syntax.Variable:
			if t.Name == "" || t.NameExpr != nil {
				return false
			}
			read = jsonFirstRead(ctx.File, t.Name, a)
		case *syntax.PropertyFetch:
			read = jsonFirstPropertyRead(ctx, t, a)
		}
		if read == nil {
			return false
		}
		x = jsonOuter(read)
	}
	return jsonFailureTest(ctx, x, decode)
}

// jsonOuter returns x with the parentheses and `@` operators around it.
func jsonOuter(x syntax.Expr) syntax.Expr {
	for {
		switch p := x.Parent().(type) {
		case *syntax.Paren:
			x = p
		case *syntax.Unary:
			if p.Op.Kind != syntax.TAt {
				return x
			}
			x = p
		default:
			return x
		}
	}
}

// jsonLastErrorTested reports whether the statement after es asks
// json_last_error() or json_last_error_msg(): in the condition of an `if`,
// or of a ternary it returns or evaluates.
func jsonLastErrorTested(ctx *analysis.Context, es *syntax.ExprStmt) bool {
	next, ok := astquery.NextStmt(ctx.File, es)
	if !ok {
		return false
	}
	var conds []syntax.Expr
	switch st := next.(type) {
	case *syntax.If:
		conds = append(conds, st.Cond)
	case *syntax.Return, *syntax.ExprStmt:
		syntax.Inspect(st, func(n syntax.Node) bool {
			switch t := n.(type) {
			case *syntax.Closure, *syntax.ArrowFunction:
				return false
			case *syntax.Ternary:
				conds = append(conds, t.Cond)
			}
			return true
		})
	}
	found := false
	for _, cond := range conds {
		syntax.Inspect(cond, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && (ctx.IsGlobalFunctionCall(c, "json_last_error") || ctx.IsGlobalFunctionCall(c, "json_last_error_msg")) {
				found = true
			}
			return !found
		})
	}
	return found
}

// jsonFirstPropertyRead returns the first access to the property p is
// (`$this->data`) after assignment a in its function, nil when there is
// none or it is written there.
func jsonFirstPropertyRead(ctx *analysis.Context, p *syntax.PropertyFetch, a *syntax.Assign) syntax.Expr {
	scope := syntax.EnclosingVariableScope(a)
	var first *syntax.PropertyFetch
	syntax.Inspect(scope, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return n == scope
		case *syntax.PropertyFetch:
			if x.Span().Start >= a.Span().End && (first == nil || x.Span().Start < first.Span().Start) && astquery.Equivalent(ctx.File, x, p) {
				first = x
			}
		}
		return true
	})
	if first == nil {
		return nil
	}
	if w, ok := jsonOuter(first).Parent().(*syntax.Assign); ok && syntax.Node(w.Var) == jsonOuter(first) {
		return nil
	}
	return first
}

// jsonFirstRead returns the first occurrence of variable name after
// assignment a in its scope, the file's top-level code included (nil when
// there is none or it is written there).
func jsonFirstRead(f *syntax.File, name string, a *syntax.Assign) syntax.Expr {
	scope := syntax.EnclosingVariableScope(a)
	var first *syntax.Variable
	visit := func(n syntax.Node) bool {
		switch v := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return n == scope
		case *syntax.If:
			// The failure value (null or false) is no instance of anything:
			// `if ($data instanceof \stdClass) { … }` without other arms
			// never runs for it, so the test comes after.
			if t, ok := syntax.UnwrapParens(v.Cond).(*syntax.Instanceof); ok && len(v.ElseIfs) == 0 && v.Else == nil {
				if tested, ok := syntax.UnwrapParens(t.Expr).(*syntax.Variable); ok && tested.Name == name && v.Span().Start >= a.Span().End {
					return false
				}
			}
		case *syntax.Variable:
			if v.Name == name && v.Span().Start >= a.Span().End && (first == nil || v.Span().Start < first.Span().Start) {
				first = v
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(f, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	if first == nil {
		return nil
	}
	if w, ok := jsonOuter(first).Parent().(*syntax.Assign); ok && syntax.Node(w.Var) == jsonOuter(first) {
		return nil
	}
	return first
}

// jsonEncodable reports whether json_encode cannot fail on e: an integer,
// a string literal of valid UTF-8, true, false, null, or an array literal
// of those (keys included) nested less than 64 levels deep, or a call of
// one of PHP's functions that return only ASCII text or an integer (E10).
// jsonSafeEncode reports a json_encode() call that cannot fail (E10): its
// value is encodable, it keeps the default depth and its flags cannot
// hold JSON_NUMERIC_CHECK, which turns a string such as '1e999' into an
// infinite number.
func jsonSafeEncode(ctx *analysis.Context, call *syntax.FuncCall) bool {
	value, ok := jsonArg(call, "value", 0)
	if _, depth := jsonArg(call, "depth", 2); !ok || depth || !jsonEncodable(ctx, value.Value, 0) {
		return false
	}
	flags, has := jsonArg(call, "flags", 1)
	return !has || jsonFlagsWithout(ctx, flags.Value, "JSON_NUMERIC_CHECK")
}

// jsonDecodesSafeEncode reports a json_decode() call with the default
// depth whose text is a json_encode() call that cannot fail (E11): the
// text is valid JSON nested less deeply than the depth.
func jsonDecodesSafeEncode(ctx *analysis.Context, call *syntax.FuncCall, text syntax.Expr) bool {
	if _, depth := jsonArg(call, "depth", 2); depth {
		return false
	}
	inner, ok := syntax.UnwrapParens(text).(*syntax.FuncCall)
	return ok && ctx.GlobalFunctionName(inner) == "json_encode" && jsonSafeEncode(ctx, inner)
}

// jsonFlagsWithout reports whether a flags expression is a `|` chain of
// built-in constants other than name and the literal 0.
func jsonFlagsWithout(ctx *analysis.Context, e syntax.Expr, name string) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Binary:
		return x.Op.Kind == syntax.TBar && jsonFlagsWithout(ctx, x.Left, name) && jsonFlagsWithout(ctx, x.Right, name)
	case *syntax.ConstFetch:
		c := semanticquery.GlobalConstName(ctx, x)
		return strings.HasPrefix(c, "JSON_") && c != name
	case *syntax.Literal:
		return x.LitKind == syntax.LitInt && x.Raw == "0"
	}
	return false
}

func jsonEncodable(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if jsonInt(ctx, e) {
		return true
	}
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.FuncCall:
		args, ok := astquery.ArgValues(x.Args)
		switch ctx.GlobalFunctionName(x) {
		case "base64_encode", "bin2hex", "dechex", "decbin", "decoct":
			return true
		case "md5", "sha1": // a second argument may ask for the raw binary digest
			return ok && len(args) == 1
		}
		return false
	case *syntax.Variable:
		// Every value that may reach the read is itself encodable.
		return jsonLocalValues(ctx, x, false, func(v syntax.Expr) bool { return jsonEncodable(ctx, v, depth) })
	case *syntax.Literal:
		v, ok := syntax.LiteralStringValue(x)
		return ok && utf8.ValidString(v)
	case *syntax.Unary:
		return x.Op.Kind == syntax.TMinus && jsonEncodable(ctx, x.Expr, depth)
	case *syntax.Binary:
		// Text joined from such values is valid UTF-8 too.
		return x.Op.Kind == syntax.TDot && jsonEncodableText(ctx, x.Left, depth) && jsonEncodableText(ctx, x.Right, depth)
	case *syntax.ClassConstFetch:
		// `Name::class` is a class name of valid UTF-8 (not `static::class`,
		// known only at run time, nor `static::NAME`, which a subclass may
		// redefine).
		n, named := x.Class.(*syntax.Name)
		id, ok := x.Name.(*syntax.Identifier)
		if !named || !ok || strings.EqualFold(n.Value, "static") {
			return false
		}
		if !strings.EqualFold(id.Value, "class") {
			return jsonEncodableConst(ctx, x, depth)
		}
		cls := ctx.Types().ClassRef(x.Class)
		return cls != "" && utf8.ValidString(cls)
	case *syntax.ConstFetch:
		name := strings.ToLower(strings.TrimPrefix(x.Name.Value, `\`))
		return name == "true" || name == "false" || name == "null" || jsonEncodableConst(ctx, x, depth)
	case *syntax.Array:
		if depth >= 64 {
			return false
		}
		for _, it := range x.Items {
			if it == nil || it.Unpack || it.ByRef || it.Key != nil && !jsonEncodable(ctx, it.Key, depth+1) || !jsonEncodable(ctx, it.Value, depth+1) {
				return false
			}
		}
		return true
	}
	return false
}

// jsonEncodableText reports an operand of `.` that converts to valid
// UTF-8: an encodable value other than an array (`'T' . count($a)`).
func jsonEncodableText(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if _, array := syntax.UnwrapParens(e).(*syntax.Array); array {
		return false
	}
	return jsonEncodable(ctx, e, depth)
}

// jsonIntFunctions are PHP's functions that return only an integer.
var jsonIntFunctions = map[string]bool{
	"crc32": true, "strlen": true, "count": true, "time": true, "intval": true,
	"intdiv": true, "rand": true, "mt_rand": true, "random_int": true,
}

// jsonInt reports an integer expression (E10): an integer literal, a call
// of a jsonIntFunctions function, or `+`, `-`, `*` or a negation of those,
// whose overflow yields a finite float, which still encodes.
func jsonInt(ctx *analysis.Context, e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Literal:
		return x.LitKind == syntax.LitInt
	case *syntax.FuncCall:
		return jsonIntFunctions[ctx.GlobalFunctionName(x)]
	case *syntax.Unary:
		return x.Op.Kind == syntax.TMinus && jsonInt(ctx, x.Expr)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TPlus, syntax.TMinus, syntax.TMul:
			return jsonInt(ctx, x.Left) && jsonInt(ctx, x.Right)
		}
	}
	return false
}

// jsonEncodableConst reports a constant whose every known initializer is
// encodable (E10). Each constant read counts as one level of depth; value
// discovery ends a cycle of constants.
func jsonEncodableConst(ctx *analysis.Context, c syntax.Expr, depth int) bool {
	values, _ := semanticquery.LocalValuesKnown(ctx.Types(), c)
	if len(values) == 0 {
		// Declared in another file or by PHP: its indexed initializer.
		return jsonIndexedConst(ctx, c)
	}
	for _, v := range values {
		if !jsonEncodable(ctx, v, depth+1) {
			return false
		}
	}
	return true
}

// jsonInteger matches the text of a decimal integer.
var jsonInteger = regexp.MustCompile(`^-?\d+$`)

// jsonIndexedConst reports a constant declared outside the file whose
// indexed initializer is an integer or, declared in the project, one plain
// quoted string of valid UTF-8 without escapes or interpolation. PHP's own
// string constants are left out: their stub values are those of the
// machine that built the stubs.
func jsonIndexedConst(ctx *analysis.Context, c syntax.Expr) bool {
	var value string
	builtin := false
	if x, ok := c.(*syntax.ConstFetch); ok {
		k := semanticquery.ResolveConstant(ctx.Types(), x)
		if k == nil {
			return false
		}
		value, builtin = k.Value, k.Builtin
	} else {
		x := c.(*syntax.ClassConstFetch)
		k := ctx.Index().FindConst(ctx.Types().ClassRef(x.Class), x.Name.(*syntax.Identifier).Value, ctx.PHP)
		if k == nil || k.Case {
			return false
		}
		value = k.Value
	}
	if jsonInteger.MatchString(value) {
		return true
	}
	if builtin || len(value) < 2 {
		return false
	}
	q, body := value[0], value[1:len(value)-1]
	return (q == '\'' || q == '"') && value[len(value)-1] == q && !strings.ContainsAny(body, "'\"$\\") && utf8.ValidString(body)
}

// jsonFailureTest reports whether x is tested for a failed JSON call: the
// condition of a ternary (`?:` included) or of an if, elseif or while,
// negated with `!`, given to empty(), compared with false or null, tested
// with instanceof, given to a type test (is_array() and the like) or to a PHPUnit assertion that
// rejects the failure value, or, for json_decode (null on failure), given
// to isset() or the left operand of `??`.
func jsonFailureTest(ctx *analysis.Context, x syntax.Expr, decode bool) bool {
	if decode && jsonFetchedFallback(ctx, x) {
		return true
	}
	switch p := x.Parent().(type) {
	case *syntax.Ternary:
		return syntax.Node(p.Cond) == x
	case *syntax.If:
		return syntax.Node(p.Cond) == x
	case *syntax.ElseIf:
		return syntax.Node(p.Cond) == x
	case *syntax.While:
		return syntax.Node(p.Cond) == x
	case *syntax.Unary:
		return p.Op.Kind == syntax.TExclaim
	case *syntax.Empty:
		return true
	case *syntax.Instanceof:
		return syntax.Node(p.Expr) == x
	case *syntax.Isset:
		return decode
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TCoalesce:
			return decode && syntax.Node(p.Left) == x
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
			other := p.Right
			if syntax.Node(other) == x {
				other = p.Left
			}
			c, ok := syntax.UnwrapParens(other).(*syntax.ConstFetch)
			name := ""
			if ok {
				name = strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`))
			}
			return name == "false" || name == "null"
		}
	case *syntax.Arg:
		call := p.Parent().Parent()
		if f, ok := call.(*syntax.FuncCall); ok {
			if nm, ok := f.Name.(*syntax.Name); ok && jsonTypeChecks[strings.ToLower(strings.TrimPrefix(nm.Value, `\`))] {
				return true
			}
		}
		if decode && jsonArrayOnly(ctx, call, x) {
			return true
		}
		return jsonAsserted(ctx, call, p, decode)
	}
	return false
}

// jsonMayBeNull reports text that may be null: its type holds null, or it
// reads a property declared without a type or a default, null until set.
func jsonMayBeNull(ctx *analysis.Context, e syntax.Expr) bool {
	if ctx.TypeOf(e).Has("null") {
		return true
	}
	p, ok := syntax.UnwrapParens(e).(*syntax.PropertyFetch)
	if !ok {
		return false
	}
	id, named := p.Name.(*syntax.Identifier)
	classes := ctx.TypeOf(p.Var).Classes()
	if !named || len(classes) != 1 {
		return false
	}
	prop := ctx.Index().FindProperty(strings.TrimPrefix(classes[0], `\`), id.Value, ctx.PHP)
	return prop != nil && prop.Type == "" && !prop.HasDefault
}

// jsonRoundTrip reports a json_encode call that is, parentheses and `@`
// aside, the text of a json_decode call whose own failure is tested or
// throws: a failed encode gives the empty text, which fails that decode.
func jsonRoundTrip(ctx *analysis.Context, call *syntax.FuncCall) bool {
	var x syntax.Node = call
	for {
		switch p := x.Parent().(type) {
		case *syntax.Paren:
			x = p
			continue
		case *syntax.Unary:
			if p.Op.Kind == syntax.TAt {
				x = p
				continue
			}
		}
		break
	}
	arg, ok := x.Parent().(*syntax.Arg)
	if !ok {
		return false
	}
	dec, ok := arg.Parent().Parent().(*syntax.FuncCall)
	if !ok || ctx.GlobalFunctionName(dec) != "json_decode" {
		return false
	}
	if text, ok := jsonArg(dec, "json", 0); !ok || text != arg {
		return false
	}
	if flags, ok := jsonArg(dec, "flags", 3); ok && jsonStrictFlags(ctx, flags.Value) {
		return true
	}
	return jsonFailureHandled(ctx, dec, true)
}

// jsonArrayOnly reports whether x is given, outside a `try` with a
// `catch`, to a parameter declared `array` alone (`array_keys(json_decode($s,
// true))`): the failure value null is a TypeError there. A parameter of
// PHP's own functions rejects it from PHP 8.0 on; one of a function of the
// project, or of a method a class name, `self` or `parent` names (no
// override may widen it), always.
func jsonArrayOnly(ctx *analysis.Context, call syntax.Node, x syntax.Expr) bool {
	var args *syntax.ArgList
	switch c := call.(type) {
	case *syntax.FuncCall:
		args = c.Args
	case *syntax.StaticCall:
		if n, named := c.Class.(*syntax.Name); !named || strings.EqualFold(n.Value, "static") {
			return false
		}
		args = c.Args
	default:
		return false
	}
	callee, ok := semanticquery.ResolveCallee(ctx, call)
	if !ok || callee.Builtin && ctx.PHP.Below(phpversion.PHP80) {
		return false
	}
	for p := call.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if t, ok := p.(*syntax.Try); ok && len(t.Catches) > 0 {
			return false
		}
	}
	for i, param := range callee.Params {
		if param.Type == "array" && !param.Variadic && semanticquery.CallArgument(args, i, param.Name) == x {
			return true
		}
	}
	return false
}

// jsonAssertions are PHPUnit's assertions that fail on the failure value
// of both JSON functions (null, false), with the position of the checked
// argument: type checks, assertNotEmpty, assertTrue, and the assertions
// that need an array, a countable or an iterable.
var jsonAssertions = map[string]int{
	"assertisarray": 0, "assertisobject": 0, "assertisstring": 0, "assertisiterable": 0,
	"assertiscountable": 0, "assertinstanceof": 1, "assertnotempty": 0,
	"assertarrayhaskey": 1, "assertarraynothaskey": 1, "assertcount": 1, "assertnotcount": 1,
	"assertcontains": 1, "assertnotcontains": 1, "asserttrue": 0,
}

// jsonStringAssertions are PHPUnit's assertions whose second parameter is
// declared string: null there is a TypeError, whatever strict_types says,
// so a failed json_decode errors the test.
var jsonStringAssertions = map[string]bool{
	"assertstringcontainsstring": true, "assertstringcontainsstringignoringcase": true,
	"assertstringnotcontainsstring": true, "assertstringnotcontainsstringignoringcase": true,
	"assertstringstartswith": true, "assertstringstartsnotwith": true,
	"assertstringendswith": true, "assertstringendsnotwith": true,
	"assertmatchesregularexpression": true, "assertdoesnotmatchregularexpression": true,
}

// jsonLooseAssertions compare their two values loosely: assertEquals, and
// the variants that first sort arrays or lowercase strings, which still
// tell null from a non-empty array or string.
var jsonLooseAssertions = map[string]bool{
	"assertequals": true, "assertequalscanonicalizing": true, "assertequalsignoringcase": true,
}

// jsonAsserted reports whether arg is the value a PHPUnit assertion checks
// and the assertion fails on the failure value: a type assertion,
// assertNotNull or assertFalse for json_decode or assertNotFalse for
// json_encode.
func jsonAsserted(ctx *analysis.Context, call syntax.Node, arg *syntax.Arg, decode bool) bool {
	c, ok := phpunit.AsPuCall(call)
	if !ok || arg.Name != nil || !phpunit.PuAssertion(ctx, call) && !phpunit.UnindexedAssertion(ctx, call) {
		return false
	}
	name := strings.ToLower(c.Ident.Value)
	if name == "assertsame" || jsonLooseAssertions[name] {
		// Compared with a literal the failure value does not equal.
		if len(c.Args.Args) < 2 {
			return false
		}
		other := c.Args.Args[0]
		if other == syntax.Node(arg) {
			other = c.Args.Args[1]
		} else if c.Args.Args[1] != syntax.Node(arg) {
			return false
		}
		o, ok := other.(*syntax.Arg)
		return ok && o.Name == nil && !o.Unpack && jsonRejectsFailure(ctx, o.Value, name == "assertsame", decode)
	}
	at, known := jsonAssertions[name]
	switch {
	case known:
	case decode && jsonStringAssertions[name]:
		at = 1
	case name == "assertnotnull" && decode, name == "assertnotfalse" && !decode, name == "assertfalse" && decode:
		at = 0
	default:
		return false
	}
	return at < len(c.Args.Args) && c.Args.Args[at] == syntax.Node(arg)
}

// jsonRejectsFailure reports whether e is a literal that the failure value
// (null for json_decode, false for json_encode) is not identical to
// (strict) or, compared loosely, not equal to: any array, string or number
// literal or boolean constant other than the failure value itself, and
// loosely only a non-empty array, a string other than "" and "0", a
// non-zero number or true.
// A local variable counts when each value it may hold does.
func jsonRejectsFailure(ctx *analysis.Context, e syntax.Expr, strict, decode bool) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Variable:
		if strict && x.NameExpr == nil && jsonTypedParam(ctx, syntax.EnclosingVariableScope(x), x.Name, decode) {
			return true
		}
		if outer := jsonCaptured(ctx.File, x); outer != nil {
			return jsonRejectsFailure(ctx, outer, strict, decode)
		}
		return jsonLocalValues(ctx, x, true, func(v syntax.Expr) bool { return jsonRejectsFailure(ctx, v, strict, decode) })
	case *syntax.Array:
		return strict || len(x.Items) > 0
	case *syntax.Literal:
		if strict {
			return true
		}
		if v, ok := syntax.LiteralStringValue(x); ok {
			return v != "" && v != "0"
		}
		f, err := strconv.ParseFloat(strings.ReplaceAll(x.Raw, "_", ""), 64)
		return err == nil && f != 0
	case *syntax.ConstFetch:
		switch strings.ToLower(strings.TrimPrefix(x.Name.Value, `\`)) {
		case "true":
			return true
		case "false":
			return strict && decode
		case "null":
			return strict && !decode
		}
	}
	return false
}

// jsonCaptured returns the `use` clause variable through which x reads a
// variable of the enclosing scope by value, when the closure neither
// writes it nor passes it on or reads its elements; nil otherwise.
func jsonCaptured(f *syntax.File, x *syntax.Variable) *syntax.Variable {
	fn, ok := syntax.EnclosingVariableScope(x).(*syntax.Closure)
	if !ok {
		return nil
	}
	i := slices.IndexFunc(fn.Uses, func(u *syntax.ClosureUse) bool { return u.Var != nil && u.Var.Name == x.Name && !u.ByRef })
	if i < 0 {
		return nil
	}
	for _, a := range flowquery.VarAccesses(f, fn, x.Name) {
		if a.Write || a.ElemWrite {
			return nil
		}
		switch a.Var.Parent().(type) {
		case *syntax.ArrayDimFetch, *syntax.Arg:
			if a.Var != x {
				return nil
			}
		}
	}
	return fn.Uses[i].Var
}

// jsonTypedParam reports whether name is a parameter of scope, or a
// variable a closure or arrow function takes from such a parameter, that
// nothing writes and whose declared type excludes the failure value: a
// single type other than mixed and null, and for json_encode other than
// bool and false, without a null default.
func jsonTypedParam(ctx *analysis.Context, scope syntax.Node, name string, decode bool) bool {
	for _, a := range flowquery.VarAccesses(ctx.File, scope, name) {
		if a.Write {
			return false
		}
	}
	params, _ := astquery.ScopeParts(scope)
	for _, p := range params {
		if p.Var == nil || p.Var.Name != name {
			continue
		}
		t, ok := p.Type.(*syntax.Name)
		if d, null := p.Default.(*syntax.ConstFetch); null && strings.EqualFold(strings.TrimPrefix(d.Name.Value, `\`), "null") || !ok || p.ByRef {
			return false
		}
		switch strings.ToLower(t.Value) {
		case "mixed", "null":
			return false
		case "bool", "false":
			return decode
		}
		return true
	}
	switch fn := scope.(type) {
	case *syntax.Closure:
		for _, u := range fn.Uses {
			if u.Var != nil && u.Var.Name == name && !u.ByRef {
				return jsonTypedParam(ctx, syntax.EnclosingVariableScope(fn), name, decode)
			}
		}
	case *syntax.ArrowFunction:
		return jsonTypedParam(ctx, syntax.EnclosingVariableScope(fn), name, decode)
	}
	return false
}

// jsonFetchedFallback reports an element or property read from a decoded
// value that falls back when the value is missing: `$doc[0] ?? []`,
// `isset($doc->tags)`, `empty($doc['k'])`, that a PHPUnit assertion
// rejecting null checks, or given to an array-only parameter of PHP's own
// functions (`array_column($doc['rows'], 'id')`). A failed decode yields null, whose elements read
// as missing, that is null.
func jsonFetchedFallback(ctx *analysis.Context, x syntax.Expr) bool {
	var outer syntax.Node = x
	for {
		switch p := outer.Parent().(type) {
		case *syntax.ArrayDimFetch:
			if p.Var != outer {
				return false
			}
			outer = p
			continue
		case *syntax.PropertyFetch:
			if p.Var != outer {
				return false
			}
			outer = p
			continue
		case *syntax.Binary:
			return outer != syntax.Node(x) && p.Op.Kind == syntax.TCoalesce && syntax.Node(p.Left) == outer
		case *syntax.Isset, *syntax.Empty:
			return outer != syntax.Node(x)
		case *syntax.Arg:
			if outer == syntax.Node(x) {
				return false
			}
			if jsonArrayOnly(ctx, p.Parent().Parent(), p.Value) {
				return true
			}
			return jsonAsserted(ctx, p.Parent().Parent(), p, true)
		case *syntax.Assign:
			// `$html = $data['html'];` copies the read: the copy's first read
			// tested like the read itself counts.
			v, ok := syntax.UnwrapParens(p.Var).(*syntax.Variable)
			if outer == syntax.Node(x) || !ok || v.Name == "" || v.NameExpr != nil || p.Op.Kind != syntax.TEqual || p.ByRef || syntax.Node(p.Value) != outer {
				return false
			}
			read := jsonFirstRead(ctx.File, v.Name, p)
			return read != nil && jsonFailureTest(ctx, jsonOuter(read), true)
		}
		return false
	}
}

// jsonLocalValues reports whether every value that may reach the read of
// local variable x is known, other than a variable, and satisfies ok, and
// no element of x is written in place (`$v[] = $raw`). With whole, nothing
// may take anything out of x either: no element read (`unset($v[0])`), no
// other read passed as an argument (`array_pop($v)`).
func jsonLocalValues(ctx *analysis.Context, x *syntax.Variable, whole bool, ok func(syntax.Expr) bool) bool {
	values, complete := semanticquery.LocalValuesComplete(ctx.Types(), x)
	if !complete || len(values) == 0 {
		return false
	}
	scope := syntax.EnclosingVariableScope(x)
	if u, ok := x.Parent().(*syntax.ClosureUse); ok {
		scope = syntax.EnclosingVariableScope(u.Parent())
	}
	for _, a := range flowquery.VarAccesses(ctx.File, scope, x.Name) {
		if a.ElemWrite {
			return false
		}
		if !whole || a.Write || a.Var == x {
			continue
		}
		switch a.Var.Parent().(type) {
		case *syntax.ArrayDimFetch, *syntax.Arg:
			return false
		}
	}
	for _, v := range values {
		// A value copied from another variable is not followed further:
		// two variables may copy each other in a loop.
		if _, copied := syntax.UnwrapParens(v).(*syntax.Variable); copied || !ok(v) {
			return false
		}
	}
	return true
}
