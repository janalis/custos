package realpathinstreamcontext

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// realpathInStreamContext reports realpath() calls that break inside stream
// wrappers (phar://): include paths and parent-directory climbing.
type realpathInStreamContext struct{}

const realpathGenericMsg = "realpath() fails inside stream wrappers such as phar://; prefer dirname()."

func (realpathInStreamContext) ID() string { return "RealpathInStreamContext" }
func (realpathInStreamContext) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (realpathInStreamContext) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "realpath") || call.Args == nil || len(call.Args.Args) != 1 {
		return
	}
	arg, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	if realpathTestContext(ctx, call) { // D2
		return
	}
	parent, _ := astquery.ParentSkipParens(call)
	_, isInclude := parent.(*syntax.Include)
	if !isInclude && !hasParentDirLiteral(call) { // D3, D4
		return
	}

	span := call.Span()
	r, ok := realpathReplacement(ctx, arg.Value, isInclude)
	if !ok {
		ctx.Report(span, realpathGenericMsg)
		return
	}
	msg := "Use '" + r + "' instead: realpath() fails inside stream wrappers."
	if realpathResultTested(ctx, call, map[*syntax.Assign]bool{}) || !realpathResolvedBase(ctx, arg.Value) ||
		realpathTightlyBound(call) || astquery.HasComment(ctx.File, span) {
		// custos: the failure check would go dead; dirname() of a path that
		// may cross a symlink names another directory than realpath(); a
		// tighter operator would capture part of the replacement; a comment
		// inside the call would be lost.
		ctx.Report(span, msg)
		return
	}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Drop realpath()",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: r}} },
	})
}

// realpathTestContext implements D2: test files and test classes. Unlike
// semanticquery.InTestContext it skips anonymous classes and takes the namespace
// from the enclosing namespace node (spec wording); kept separate so the
// rule's output does not change.
func realpathTestContext(ctx *analysis.Context, n syntax.Node) bool {
	if ctx.IsTestFile() {
		return true
	}
	var class *syntax.ClassLike
	for a := n.Parent(); a != nil; a = a.Parent() {
		switch a := a.(type) {
		case *syntax.ClassLike:
			if class == nil && a.Name != nil {
				class = a
			}
		case *syntax.Namespace:
			if class == nil {
				return false
			}
			fqn := "\\" + class.Name.Value
			if a.Name != nil {
				fqn = "\\" + strings.TrimPrefix(a.Name.Value, "\\") + fqn
			}
			return astquery.IsTestClassFQN(fqn)
		}
	}
	return class != nil && astquery.IsTestClassFQN("\\"+class.Name.Value)
}

// hasParentDirLiteral reports whether a string literal inside n contains `..`.
func hasParentDirLiteral(n syntax.Node) bool {
	found := false
	syntax.Inspect(n, func(c syntax.Node) bool {
		switch c := c.(type) {
		case *syntax.Literal:
			if c.LitKind == syntax.LitString && strings.Contains(c.Raw, "..") {
				found = true
			}
		case *syntax.StringPart:
			if strings.Contains(c.Raw, "..") {
				found = true
			}
		}
		return !found
	})
	return found
}

// realpathReplacement computes the replacement text R (R1-R3).
func realpathReplacement(ctx *analysis.Context, s syntax.Expr, isInclude bool) (string, bool) {
	switch s := s.(type) {
	case *syntax.Literal: // R2
		// custos: outside an inclusion the literal holds `..` (D4), which
		// realpath() would have resolved: the bare literal is another string.
		if s.LitKind == syntax.LitString && isInclude {
			if raw, _, ok := astquery.QuotedStringRaw(s); !ok || !isAbsolutePathLiteral(raw) {
				return "", false // relative paths resolve differently once realpath() is gone
			} else if len(raw) > 1 && strings.HasSuffix(raw, "/") {
				return "", false // custos: realpath() drops the trailing separator
			}
			return ctx.Text(s), true
		}
	case *syntax.Binary: // R1
		if s.Op.Kind != syntax.TDot || isConcat(s.Left) {
			return "", false
		}
		rest, quote, ok := astquery.QuotedStringRaw(s.Right)
		// custos: `/..` must be a whole path segment (`/..cache` names a
		// directory).
		parent := func(r string) bool { return r == "/.." || strings.HasPrefix(r, "/../") }
		if !ok || !parent(rest) {
			return "", false
		}
		left := ctx.Text(s.Left)
		dirname := semanticquery.QualifiedBuiltin(ctx, "dirname", s.Span().Start) // a namespaced dirname() would capture a bare call
		for parent(rest) {
			rest = rest[3:]
			left = dirname + "(" + left + ")"
		}
		// custos: realpath() drops trailing separators (`/../` names the
		// parent itself, `/../etc/` is `…/etc`).
		rest = strings.TrimRight(rest, "/")
		if rest == "" {
			return left, true // no `. ''` tail
		}
		// custos: realpath() also resolves `.`, `..` and empty segments
		// further down; the tail would keep them.
		for seg := range strings.SplitSeq(strings.ReplaceAll(rest[1:], "\\", "/"), "/") {
			if seg == "" || seg == "." || seg == ".." {
				return "", false
			}
		}
		q := string(quote)
		return left + " . " + q + rest + q, true
	}
	return "", false
}

// isAbsolutePathLiteral reports whether a literal path does not depend on the
// current directory or the include path: `/x`, `\\x`, `C:\x`, `C:/x`, or a
// stream/URL path `scheme://...`.
func isAbsolutePathLiteral(raw string) bool {
	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "\\") || strings.Contains(raw, "://") {
		return true
	}
	if len(raw) >= 3 && raw[1] == ':' && (raw[2] == '\\' || raw[2] == '/') {
		c := raw[0] | 0x20
		return c >= 'a' && c <= 'z'
	}
	return false
}

func isConcat(e syntax.Expr) bool {
	b, ok := e.(*syntax.Binary)
	return ok && b.Op.Kind == syntax.TDot
}

// realpathResultTested reports whether the value of e (a realpath() call,
// or an expression holding its result) is tested for failure: negated,
// compared, used as a condition, a switch/match subject or arm, a logical
// or elvis operand, passed to empty() or a type-testing builtin, or
// assigned to a target that is tested that way anywhere in its scope. The
// replacement never returns false, so such a check would go dead. seen
// holds the assignments already followed (`$a = $b; $b = $a;`).
func realpathResultTested(ctx *analysis.Context, e syntax.Node, seen map[*syntax.Assign]bool) bool {
	p, child := astquery.ParentSkipParens(e)
	switch x := p.(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim || x.Op.Kind == syntax.TBoolCast
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual,
			syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor:
			return true
		}
	case *syntax.Ternary:
		return x.Cond == child
	case *syntax.For:
		for _, c := range x.Cond {
			if c == child {
				return true
			}
		}
	case *syntax.If, *syntax.ElseIf, *syntax.While, *syntax.DoWhile,
		*syntax.Switch, *syntax.Case, *syntax.Match, *syntax.Empty:
		return true // an expression directly below these is their condition or subject
	case *syntax.MatchArm:
		return x.Body != child
	case *syntax.Arg:
		call, ok := x.Parent().Parent().(*syntax.FuncCall)
		return ok && realpathTypeTest(ctx.GlobalFunctionName(call))
	case *syntax.Assign:
		if x.Value != child || x.Op.Kind != syntax.TEqual || seen[x] {
			return false
		}
		seen[x] = true
		if realpathResultTested(ctx, x, seen) {
			return true
		}
		// A local is tested within its function, anything else (property,
		// element, global code) anywhere in the file.
		found := false
		visit := func(n syntax.Node) bool {
			if !found && n != x.Var && n.Kind() == x.Var.Kind() && astquery.EquivalentFoldNames(ctx.File, n, x.Var) && realpathResultTested(ctx, n, seen) {
				found = true
			}
			return !found
		}
		fn := syntax.EnclosingFuncLike(x)
		if _, local := x.Var.(*syntax.Variable); local && fn != nil {
			syntax.Inspect(fn, visit)
		} else {
			syntax.InspectFile(ctx.File, visit)
		}
		return found
	}
	return false
}

// realpathTypeTest reports whether the global function name tells false
// from a string: the is_*() family, boolval(), gettype(), get_debug_type()
// and assert().
func realpathTypeTest(name string) bool {
	switch name {
	case "boolval", "gettype", "get_debug_type", "assert":
		return true
	}
	return strings.HasPrefix(name, "is_")
}

// realpathTightlyBound reports whether an operator binding tighter than
// `.` applies directly to the call (`realpath(…)[0]`, `-realpath(…)`,
// `1 + realpath(…)`): it would capture only part of a concatenation.
func realpathTightlyBound(call *syntax.FuncCall) bool {
	switch x := call.Parent().(type) {
	case *syntax.ArrayDimFetch:
		return true
	case *syntax.Unary:
		return x.Op.Kind != syntax.TAt
	case *syntax.Binary:
		return x.Op.Kind != syntax.TDot && x.Op.Kind != syntax.TCoalesce
	}
	return false
}

// realpathResolvedBase reports whether dropping realpath() keeps the same
// directory: an R2 literal, or an R1 concatenation whose base is __DIR__ or
// dirname() of __DIR__/__FILE__ (PHP resolves symlinks in those). Any other
// base may be a symlink: realpath() climbs from its target, dirname() from
// the link (TYPO3's typo3_src). A bare __FILE__ is not a directory:
// realpath() of `__FILE__ . '/..'` fails.
func realpathResolvedBase(ctx *analysis.Context, s syntax.Expr) bool {
	b, ok := s.(*syntax.Binary)
	if !ok {
		return true
	}
	e := syntax.UnwrapParens(b.Left)
	dir := false
	for {
		call, ok := e.(*syntax.FuncCall)
		if !ok || !ctx.IsGlobalFunctionCall(call, "dirname") {
			break
		}
		dir = true
		args, ok := astquery.CallArgValues(call)
		if !ok || len(args) == 0 {
			return false
		}
		e = syntax.UnwrapParens(args[0])
	}
	m, ok := e.(*syntax.MagicConst)
	return ok && (m.Token.Kind == syntax.TDir || dir && m.Token.Kind == syntax.TFile)
}
