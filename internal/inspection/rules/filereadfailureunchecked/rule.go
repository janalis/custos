// Package filereadfailureunchecked implements the native FileReadFailureUnchecked inspection.
package filereadfailureunchecked

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/phpunit"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/types"
)

const message = "Handle file reading failure before using its result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "FileReadFailureUnchecked" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KStaticCall, syntax.KMethodCall, syntax.KBinary}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var inputs []syntax.Expr
	var args *syntax.ArgList
	switch x := n.(type) {
	case *syntax.FuncCall:
		args = x.Args
	case *syntax.StaticCall:
		args = x.Args
	case *syntax.MethodCall:
		args = x.Args
	case *syntax.Binary:
		if x.Op.Kind != syntax.TDot || failureMessage(ctx, x) {
			return
		}
		inputs = []syntax.Expr{x.Left, x.Right}
	}
	assertion := args != nil && phpunit.PuAssertion(ctx, n)
	library, checked := phpunit.LibraryFailurePassing(ctx, n, args, true)
	switch {
	case assertion:
		inputs = phpunit.FailurePassing(ctx, n, args, true)
	case checked:
		inputs, assertion = library, true
	case args != nil:
		inputs = semanticquery.NativeStringInputs(ctx, n, args)
	}
	for _, input := range inputs {
		read, ok := semanticquery.NativeLocalBuiltinResult(ctx, input, "false", "file_get_contents", "stream_get_contents").(*syntax.FuncCall)
		if !ok || infallibleRead(ctx, read) || throwingDecode(ctx, n, input) || decodedFallback(ctx, n, input) || unmatchedSubject(ctx, n, input) || rejectsEmpty(ctx, n, input) || lengthTested(ctx, n, input) || trimmedEmptyTested(ctx, n, input) || assertedText(ctx, n, input) || searchTested(ctx, n, input) || decodedTested(ctx, n, input) {
			continue
		}

		// A narrowed type describes a static local as its last test left it:
		// only the flow knows the calls that may have assigned it since. In
		// a scope with a `goto` label only the flow says what holds here.
		jumps := semanticquery.NativeFailureFlowOnly(ctx, input)
		if !jumps && !semanticquery.NativeStaticLocal(ctx, input) && ctx.TypeOf(input).OnlyOf("string") || ctx.Flow().Excludes(input, "false") || !jumps && semanticquery.NativeNarrowedExcludes(ctx, input, types.Of("string", "false"), "false") {
			continue
		}
		// A parameter must be `string` in whichever declaration runs; a
		// PHPUnit assertion checks what it is given, unless it lets the
		// failure pass as text.
		if args == nil || assertion || semanticquery.NativeStringConsumer(ctx, n, args, input) {
			ctx.ReportNode(n, message)
			return
		}
	}
}

// decodedFallback reports whether input is the text json_decode() call n
// decodes for an expression that falls back on null (E11): the call, a
// ternary branch or an element or property read on it, as the left operand
// of `??` or the argument of isset() or empty(). The failure value decodes
// to null like invalid JSON, so the read failing takes that fallback.
func decodedFallback(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, call) != "json_decode" || semanticquery.CallArgument(call.Args, 0, "json") != input {
		return false
	}
	var x syntax.Node = call
	for {
		switch p := x.Parent().(type) {
		case *syntax.Paren:
		case *syntax.Ternary:
			if syntax.Node(p.Cond) == x {
				return false
			}
		case *syntax.ArrayDimFetch:
			if syntax.Node(p.Var) != x {
				return false
			}
		case *syntax.PropertyFetch:
			if syntax.Node(p.Var) != x {
				return false
			}
		case *syntax.Binary:
			return p.Op.Kind == syntax.TCoalesce && syntax.Node(p.Left) == x
		case *syntax.Isset, *syntax.Empty:
			return true
		default:
			return false
		}
		x = x.Parent()
	}
}

// infallibleRead reports whether read is stream_get_contents() of a handle
// opened on one of PHP's own memory or standard streams, or
// file_get_contents() of such a stream (E7), which cannot fail.
func infallibleRead(ctx *analysis.Context, read *syntax.FuncCall) bool {
	if semanticquery.NativeBuiltinName(ctx, read) == "file_get_contents" {
		// Read directly, such a stream cannot fail either.
		return semanticquery.InfallibleOpen(read)
	}
	stream := semanticquery.CallArgument(read.Args, 0, "stream")
	if p, ok := syntax.UnwrapParens(stream).(*syntax.PropertyFetch); ok {
		return infallibleProperty(ctx, p)
	}
	open, ok := semanticquery.NativeLocalValue(ctx, stream).(*syntax.FuncCall)
	return ok && infallibleOpen(ctx, open)
}

// infallibleOpen reports an fopen() of one of PHP's own memory or standard
// streams.
func infallibleOpen(ctx *analysis.Context, open *syntax.FuncCall) bool {
	return semanticquery.NativeBuiltinName(ctx, open) == "fopen" && semanticquery.InfallibleOpen(open)
}

// infallibleProperty reports `$this->p` for a private instance property
// of the enclosing class, without hooks and with no default other than
// null, that its class only ever sets to an infallibleOpen() and otherwise
// only reads, passes by value or unsets: a read after unset() fails loudly
// (E7).
func infallibleProperty(ctx *analysis.Context, p *syntax.PropertyFetch) bool {
	id, named := p.Name.(*syntax.Identifier)
	this, ok := p.Var.(*syntax.Variable)
	class := syntax.EnclosingClass(p)
	if !named || !ok || this.Name != "this" || p.NullSafe || class == nil || !privateHandle(class, id.Value) {
		return false
	}
	set, safe := false, true
	syntax.Inspect(class, func(n syntax.Node) bool {
		f, ok := n.(*syntax.PropertyFetch)
		if !ok || !safe {
			return safe
		}
		v, isVar := f.Var.(*syntax.Variable)
		name, isID := f.Name.(*syntax.Identifier)
		if !isVar || v.Name != "this" || !isID || name.Value != id.Value {
			return true
		}
		switch at := f.Parent().(type) {
		case *syntax.Assign:
			if at.Var == syntax.Expr(f) {
				open, isCall := syntax.UnwrapParens(at.Value).(*syntax.FuncCall)
				safe = at.Op.Kind == syntax.TEqual && !at.ByRef && isCall && infallibleOpen(ctx, open)
				set = true
			} else {
				safe = !at.ByRef
			}
		case *syntax.Arg:
			callee, resolved := semanticquery.ResolveCallee(ctx, at.Parent().Parent())
			if !resolved {
				safe = false
				break
			}
			byRef, known := semanticquery.ArgumentBindsByRef(at.Parent().(*syntax.ArgList), at, callee.Params, true)
			safe = known && !byRef
		case *syntax.IncDec, *syntax.ArrayItem, *syntax.Foreach, *syntax.ArrayDimFetch:
			safe = false
		}
		return safe
	})
	return safe && set
}

// privateHandle reports whether class declares name as a private instance
// property without hooks whose default, if any, is null.
func privateHandle(class *syntax.ClassLike, name string) bool {
	for _, m := range class.Members {
		prop, ok := m.(*syntax.Property)
		if !ok {
			continue
		}
		for _, it := range prop.Props {
			if it.Var != nil && it.Var.Name == name {
				return prop.Modifiers.Has(syntax.TPrivate) && !prop.Modifiers.Has(syntax.TStatic) && len(prop.Hooks) == 0 &&
					(it.Default == nil || syntax.IsNullConst(it.Default))
			}
		}
	}
	return false
}

// throwingDecode reports whether input is the text json_decode() call n
// decodes under JSON_THROW_ON_ERROR (E8): the failure value is the empty
// text there (or a TypeError under strict types), which throws a
// JsonException, so the failed read cannot pass unnoticed.
func throwingDecode(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, call) != "json_decode" || semanticquery.CallArgument(call.Args, 0, "json") != input {
		return false
	}
	flags := semanticquery.CallArgument(call.Args, 3, "flags")
	if flags == nil {
		return false
	}
	throws := false
	syntax.Inspect(flags, func(x syntax.Node) bool {
		if c, ok := x.(*syntax.ConstFetch); ok && c.Name != nil && astquery.LastNamePart(c.Name.Value) == "JSON_THROW_ON_ERROR" {
			throws = true
		}
		return !throws
	})
	return throws
}

// unmatchedSubject reports whether input is the subject preg_match() or
// preg_match_all() call n tests against a literal pattern that cannot match
// the empty text, or the text json_validate() call n checks (E9): the
// failure value, read as the empty text, matches nothing, like a text
// without a match, and is no valid JSON.
func unmatchedSubject(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok {
		return false
	}
	name := semanticquery.NativeBuiltinName(ctx, call)
	if name == "json_validate" {
		return semanticquery.CallArgument(call.Args, 0, "json") == input
	}
	return (name == "preg_match" || name == "preg_match_all") && semanticquery.CallArgument(call.Args, 1, "subject") == input &&
		flow.EmptylessPattern(semanticquery.CallArgument(call.Args, 0, "pattern"))
}

// emptyRejecting lists the methods of PHP's own classes that throw a
// ValueError, from PHP 8.0 on, when given empty text to parse (E10).
var emptyRejecting = map[string]bool{
	"domdocument::loadxml": true, "domdocument::loadhtml": true,
	"domdocument::schemavalidatesource": true, "domdocument::relaxngvalidatesource": true,
	"xmlreader::xml": true,
}

// rejectsEmpty reports whether input is the text method call n parses with
// one of the emptyRejecting methods, whichever class it runs on (E10): the
// failure value becomes the empty text there, which throws.
func rejectsEmpty(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.MethodCall)
	if !ok || ctx.PHP < phpversion.PHP80 || semanticquery.CallArgument(call.Args, 0, "source") != input {
		return false
	}
	targets := ctx.Flow().Targets(call)
	for _, t := range targets {
		if !t.Builtin || !emptyRejecting[strings.ToLower(strings.TrimPrefix(t.Class, `\`)+"::"+t.Name)] {
			return false
		}
	}
	return len(targets) > 0
}

// lengthTested reports whether input is the text strlen() or mb_strlen()
// call n measures for a comparison with an integer that tells the empty
// text apart (E12): `strlen($s) > 0`, `48 > strlen($s)`, `strlen($s) === 0`.
// The failure value reads as the empty text, which takes the same branch.
func lengthTested(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok || semanticquery.CallArgument(call.Args, 0, "string") != input {
		return false
	}
	if name := semanticquery.NativeBuiltinName(ctx, call); name != "strlen" && name != "mb_strlen" {
		return false
	}
	x := unparen(call)
	cmp, ok := x.Parent().(*syntax.Binary)
	if !ok {
		return false
	}
	op, bound := cmp.Op.Kind, cmp.Right
	if syntax.Node(cmp.Right) == x {
		bound = cmp.Left
		if mirrored, ordering := syntax.MirroredComparison[op]; ordering {
			op = mirrored
		}
	}
	lit, ok := syntax.UnwrapParens(bound).(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitInt {
		return false
	}
	// An integer literal too large for PHP is a float, never parsed here.
	least, _ := strconv.ParseInt(lit.Raw, 0, 64)
	switch op {
	case syntax.TGreater, syntax.TIsSmallerOrEqual:
		return least >= 0
	case syntax.TIsGreaterOrEqual, syntax.TLess:
		return least >= 1
	case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
		return least == 0
	}
	return false
}

// failureMessage reports whether concatenation x builds, alone or as part of
// a longer concatenation, the `$message` of a PHPUnit assertion (E13): the
// text only describes the failure, a failed read shows there as empty.
func failureMessage(ctx *analysis.Context, x *syntax.Binary) bool {
	var outer syntax.Node = x
	for {
		switch p := outer.Parent().(type) {
		case *syntax.Paren:
		case *syntax.Binary:
			if p.Op.Kind != syntax.TDot {
				return false
			}
		case *syntax.Arg:
			call := p.Parent().Parent()
			return (phpunit.PuAssertion(ctx, call) || phpunit.UnindexedAssertion(ctx, call)) && phpunit.MessageArgument(ctx, call, p)
		default:
			return false
		}
		outer = outer.Parent()
	}
}

// trimmedEmptyTested reports whether input is the text trim(), rtrim() or
// ltrim() call n trims for a result tested for emptiness at once, or stored
// in a variable whose next read is such a test (E14): the failure value
// trims to the empty text, which takes the branch of an empty file.
func trimmedEmptyTested(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok || semanticquery.CallArgument(call.Args, 0, "string") != input {
		return false
	}
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "trim", "rtrim", "ltrim":
	default:
		return false
	}
	return testedAtOnce(ctx, call, emptinessTest)
}

// testedAtOnce reports whether the value of call meets test where it is
// produced, or is assigned by a plain statement to a variable whose next
// read meets it.
func testedAtOnce(ctx *analysis.Context, call syntax.Node, test func(syntax.Node) bool) bool {
	x := unparen(call)
	as, ok := x.Parent().(*syntax.Assign)
	if !ok || syntax.Node(as.Value) != x {
		return test(x)
	}
	v, plain := syntax.UnwrapParens(as.Var).(*syntax.Variable)
	if !plain || v.NameExpr != nil || as.Op.Kind != syntax.TEqual || as.ByRef {
		return false
	}
	if _, statement := unparen(as).Parent().(*syntax.ExprStmt); !statement {
		return test(unparen(as))
	}
	next := nextRead(ctx, as, v)
	return next != nil && test(unparen(next))
}

// nextRead returns the next read of v after the assignment statement as,
// when no write comes first; the value of an assignment to v is read
// before v is written (`$v = trim($v);`).
func nextRead(ctx *analysis.Context, as *syntax.Assign, v *syntax.Variable) *syntax.Variable {
	var after []flowquery.VarAccess
	for _, a := range flowquery.VarAccesses(ctx.File, syntax.EnclosingVariableScope(as), v.Name) {
		if a.Var.Span().Start > as.Span().End {
			after = append(after, a)
		}
	}
	if len(after) == 0 {
		return nil
	}
	if !after[0].Write {
		return after[0].Var
	}
	w, ok := after[0].By.(*syntax.Assign)
	if !ok || len(after) < 2 || after[1].Write {
		return nil
	}
	if read := after[1].Var.Span(); read.Start < w.Value.Span().Start || read.End > w.Value.Span().End {
		return nil
	}
	return after[1].Var
}

// decodedTested reports whether the read is the text unserialize() or
// json_decode() call n reads, for a result tested at once for its kind or
// for the value a corrupt text gives (E17): the failure value, read as the
// empty text, unserializes to false and decodes to null like any corrupt
// text, which the test rejects.
func decodedTested(ctx *analysis.Context, n syntax.Node, _ syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok {
		return false
	}
	// The text is the only string parameter of either function: json_decode()
	// gives null, unserialize() false.
	decodesNull := false
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "unserialize":
	case "json_decode":
		decodesNull = true
	default:
		return false
	}
	return testedAtOnce(ctx, call, func(x syntax.Node) bool {
		switch p := x.Parent().(type) {
		case *syntax.Instanceof:
			return true
		case *syntax.Isset:
			return decodesNull
		case *syntax.Binary:
			if p.Op.Kind == syntax.TCoalesce {
				return false
			}
			other := syntax.UnwrapParens(p.Right)
			if syntax.Node(p.Right) == x {
				other = syntax.UnwrapParens(p.Left)
			}
			v, isBool := astquery.BoolConst(other)
			if decodesNull && syntax.IsNullConst(other) || !decodesNull && isBool && !v {
				return p.Op.Kind == syntax.TIsIdentical || p.Op.Kind == syntax.TIsNotIdentical
			}
		case *syntax.Arg:
			if kind := astquery.ParentFuncCall(x); kind != nil {
				switch semanticquery.NativeBuiltinName(ctx, kind) {
				case "is_array", "is_object", "is_string", "is_int":
					return true
				case "is_null", "is_bool":
					return decodesNull
				}
			}
			return decodedAsserted(ctx, p, decodesNull)
		}
		return emptinessTest(x)
	})
}

// decodedAssertions are PHPUnit's assertions that fail on both false and
// null, with the position of the value they check.
var decodedAssertions = map[string]int{
	"assertisarray": 0, "assertisobject": 0, "assertisstring": 0, "assertisint": 0,
	"assertinstanceof": 1, "assertnotempty": 0, "asserttrue": 0, "assertcount": 1,
	"assertarrayhaskey": 1,
}

// decodedAsserted reports whether arg is the value a PHPUnit assertion
// checks and the assertion fails on the value a corrupt text decodes to:
// a type assertion, assertNotNull for json_decode(), assertNotFalse for
// unserialize(), or assertSame or assertEquals against a non-empty array
// literal.
func decodedAsserted(ctx *analysis.Context, arg *syntax.Arg, decodesNull bool) bool {
	call := arg.Parent().Parent()
	c, ok := phpunit.AsPuCall(call)
	if !ok || arg.Name != nil || !phpunit.PuAssertion(ctx, call) && !phpunit.UnindexedAssertion(ctx, call) {
		return false
	}
	args := c.Args.Args
	switch name := strings.ToLower(c.Ident.Value); name {
	case "assertsame", "assertequals", "assertequalscanonicalizing":
		if len(args) < 2 || args[0] != syntax.Node(arg) && args[1] != syntax.Node(arg) {
			return false
		}
		other := args[0]
		if other == syntax.Node(arg) {
			other = args[1]
		}
		o, ok := other.(*syntax.Arg)
		return ok && nonEmptyArray(o.Value)
	case "assertnotnull":
		return decodesNull && args[0] == syntax.Node(arg)
	case "assertnotfalse":
		return !decodesNull && args[0] == syntax.Node(arg)
	default:
		at, known := decodedAssertions[name]
		return known && at < len(args) && args[at] == syntax.Node(arg)
	}
}

// searchTested reports whether input is the haystack of a str_contains(),
// str_starts_with() or str_ends_with() call n whose needle is non-empty
// literal text, for a result tested at once (E16): the failure value holds
// no such text, which takes the branch of an empty file.
func searchTested(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok || semanticquery.CallArgument(call.Args, 0, "haystack") != input {
		return false
	}
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "str_contains", "str_starts_with", "str_ends_with":
	default:
		return false
	}
	needle, ok := syntax.UnwrapParens(semanticquery.CallArgument(call.Args, 1, "needle")).(*syntax.Literal)
	if !ok {
		return false
	}
	text, ok := syntax.LiteralStringValue(needle)
	return ok && text != "" && emptinessTest(unparen(call))
}

// emptyPreserving names the text parameter, and its position, of builtins
// that turn the empty text into the empty text, as they do with the failure
// value (preg_replace() only when it replaces with the empty text).
var emptyPreserving = map[string]struct {
	at   int
	name string
}{
	"addslashes": {0, "string"}, "escapeshellcmd": {0, "command"}, "html_entity_decode": {0, "string"},
	"htmlspecialchars": {0, "string"}, "lcfirst": {0, "string"}, "ltrim": {0, "string"},
	"mb_strtolower": {0, "string"}, "mb_strtoupper": {0, "string"}, "nl2br": {0, "string"},
	"preg_replace": {2, "subject"}, "rtrim": {0, "string"}, "str_ireplace": {2, "subject"},
	"str_replace": {2, "subject"}, "strip_tags": {0, "string"}, "stripslashes": {0, "string"},
	"strtolower": {0, "string"}, "strtoupper": {0, "string"}, "strtr": {0, "string"},
	"trim": {0, "string"}, "ucfirst": {0, "string"}, "ucwords": {0, "string"},
}

// assertedText reports whether input is the text an emptyPreserving builtin
// call n transforms for a result given to a PHPUnit assertion that fails on
// the empty text (E15), directly, through more such calls or through plain
// assignments to a variable whose next read is the next step: no argument
// of the assertion is the empty text and it does not let the failure pass
// as text.
func assertedText(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	x := n
	for {
		call, ok := x.(*syntax.FuncCall)
		if !ok {
			return false
		}
		name := semanticquery.NativeBuiltinName(ctx, call)
		param, known := emptyPreserving[name]
		if !known || semanticquery.CallArgument(call.Args, param.at, param.name) != input ||
			name == "preg_replace" && !emptyText(semanticquery.CallArgument(call.Args, 1, "replacement")) {
			return false
		}
		result := unparen(call)
		if as, ok := result.Parent().(*syntax.Assign); ok && syntax.Node(as.Value) == result {
			v, plain := syntax.UnwrapParens(as.Var).(*syntax.Variable)
			if _, statement := unparen(as).Parent().(*syntax.ExprStmt); !plain || v.NameExpr != nil || as.Op.Kind != syntax.TEqual || as.ByRef || !statement {
				return false
			}
			next := nextRead(ctx, as, v)
			if next == nil {
				return false
			}
			result = unparen(next)
		}
		arg, ok := result.Parent().(*syntax.Arg)
		if !ok {
			return false
		}
		x, input = arg.Parent().Parent(), arg.Value
		if !phpunit.PuAssertion(ctx, x) && !phpunit.UnindexedAssertion(ctx, x) {
			continue
		}
		var args *syntax.ArgList
		switch c := x.(type) {
		case *syntax.FuncCall:
			args = c.Args
		case *syntax.StaticCall:
			args = c.Args
		case *syntax.MethodCall:
			args = c.Args
		}
		for _, a := range args.Args {
			if a, ok := a.(*syntax.Arg); ok && emptyText(a.Value) {
				return false
			}
		}
		return !slices.Contains(phpunit.FailurePassing(ctx, x, args, true), input)
	}
}

// nonEmptyArray reports an array literal with an item.
func nonEmptyArray(e syntax.Expr) bool {
	a, ok := syntax.UnwrapParens(e).(*syntax.Array)
	return ok && len(a.Items) > 0
}

// emptyText reports the empty text literal.
func emptyText(e syntax.Expr) bool {
	lit, ok := syntax.UnwrapParens(e).(*syntax.Literal)
	if !ok {
		return false
	}
	v, ok := syntax.LiteralStringValue(lit)
	return ok && v == ""
}

// unparen returns the outermost parentheses around x, or x.
func unparen(x syntax.Node) syntax.Node {
	for {
		if _, paren := x.Parent().(*syntax.Paren); !paren {
			return x
		}
		x = x.Parent()
	}
}

// emptinessTest reports whether x is compared with the empty text, given
// to empty(), negated or tested for truth as a condition.
func emptinessTest(x syntax.Node) bool {
	switch p := x.Parent().(type) {
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
			other := p.Right
			if syntax.Node(other) == x {
				other = p.Left
			}
			return emptyText(other)
		}
	case *syntax.Empty:
		return true
	case *syntax.Unary:
		return p.Op.Kind == syntax.TExclaim
	case *syntax.If:
		return syntax.Node(p.Cond) == x
	case *syntax.ElseIf:
		return syntax.Node(p.Cond) == x
	case *syntax.Ternary:
		return syntax.Node(p.Cond) == x
	}
	return false
}
