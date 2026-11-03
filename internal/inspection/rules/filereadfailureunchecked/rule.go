// Package filereadfailureunchecked implements the native FileReadFailureUnchecked inspection.
package filereadfailureunchecked

import (
	"strconv"
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
		if !ok || infallibleRead(ctx, read) || throwingDecode(ctx, n, input) || decodedFallback(ctx, n, input) || unmatchedSubject(ctx, n, input) || rejectsEmpty(ctx, n, input) || lengthTested(ctx, n, input) {
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
// opened on one of PHP's own memory or standard streams (E7), which
// cannot fail.
func infallibleRead(ctx *analysis.Context, read *syntax.FuncCall) bool {
	if semanticquery.NativeBuiltinName(ctx, read) != "stream_get_contents" {
		return false
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
	var x syntax.Node = call
	for {
		if _, paren := x.Parent().(*syntax.Paren); !paren {
			break
		}
		x = x.Parent()
	}
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
