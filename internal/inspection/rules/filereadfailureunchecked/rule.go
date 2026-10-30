// Package filereadfailureunchecked implements the native FileReadFailureUnchecked inspection.
package filereadfailureunchecked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
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
		if x.Op.Kind != syntax.TDot {
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
		if !ok || infallibleRead(ctx, read) || throwingDecode(ctx, n, input) || unmatchedSubject(ctx, n, input) || rejectsEmpty(ctx, n, input) {
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

// infallibleRead reports whether read is stream_get_contents() of a handle
// opened on one of PHP's own memory or standard streams (E7), which
// cannot fail.
func infallibleRead(ctx *analysis.Context, read *syntax.FuncCall) bool {
	if semanticquery.NativeBuiltinName(ctx, read) != "stream_get_contents" {
		return false
	}
	open, ok := semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(read.Args, 0, "stream")).(*syntax.FuncCall)
	return ok && semanticquery.NativeBuiltinName(ctx, open) == "fopen" && semanticquery.InfallibleOpen(open)
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
// the empty text (E9): the failure value, read as the empty text, matches
// nothing, like a text without a match.
func unmatchedSubject(ctx *analysis.Context, n syntax.Node, input syntax.Expr) bool {
	call, ok := n.(*syntax.FuncCall)
	if !ok {
		return false
	}
	name := semanticquery.NativeBuiltinName(ctx, call)
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
