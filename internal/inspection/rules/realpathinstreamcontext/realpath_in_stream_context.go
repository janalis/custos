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
	if _, isInclude := parent.(*syntax.Include); !isInclude && !hasParentDirLiteral(call) { // D3, D4
		return
	}

	span := call.Span()
	r, ok := realpathReplacement(ctx, arg.Value)
	if !ok {
		ctx.Report(span, realpathGenericMsg)
		return
	}
	msg := "Use '" + r + "' instead: realpath() fails inside stream wrappers."
	if realpathResultTested(ctx, call) || !realpathResolvedBase(ctx, arg.Value) {
		// custos: the failure check would go dead; dirname() of a path that
		// may cross a symlink names another directory than realpath().
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
func realpathReplacement(ctx *analysis.Context, s syntax.Expr) (string, bool) {
	switch s := s.(type) {
	case *syntax.Literal: // R2
		if s.LitKind == syntax.LitString {
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
// or an assignment holding its result) is tested for failure: negated,
// compared, used as a condition or a logical/elvis operand, or assigned to
// a target that a later statement of the same list tests that way. The
// replacement never returns false, so such a check would go dead.
func realpathResultTested(ctx *analysis.Context, e syntax.Node) bool {
	p, child := astquery.ParentSkipParens(e)
	switch x := p.(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim || x.Op.Kind == syntax.TBoolCast
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual,
			syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return true
		}
	case *syntax.Ternary:
		return x.Cond == child
	case *syntax.If:
		return x.Cond == child
	case *syntax.ElseIf:
		return x.Cond == child
	case *syntax.While:
		return x.Cond == child
	case *syntax.Assign:
		if x.Value != child || x.Op.Kind != syntax.TEqual {
			return false
		}
		if realpathResultTested(ctx, x) {
			return true
		}
		es, ok := x.Parent().(*syntax.ExprStmt)
		if !ok {
			return false
		}
		list, i, ok := astquery.StmtList(ctx.File, es)
		if !ok {
			return false
		}
		found := false
		for _, next := range list[i+1:] {
			syntax.Inspect(next, func(n syntax.Node) bool {
				if !found && n.Kind() == x.Var.Kind() && astquery.EquivalentFoldNames(ctx.File, n, x.Var) && realpathResultTested(ctx, n) {
					found = true
				}
				return !found
			})
		}
		return found
	}
	return false
}

// realpathResolvedBase reports whether dropping realpath() keeps the same
// directory: an R2 literal, or an R1 concatenation whose base is __DIR__,
// __FILE__ or dirname() of them (PHP resolves symlinks in those). Any other
// base may be a symlink: realpath() climbs from its target, dirname() from
// the link (TYPO3's typo3_src).
func realpathResolvedBase(ctx *analysis.Context, s syntax.Expr) bool {
	b, ok := s.(*syntax.Binary)
	if !ok {
		return true
	}
	e := syntax.UnwrapParens(b.Left)
	for {
		call, ok := e.(*syntax.FuncCall)
		if !ok || !ctx.IsGlobalFunctionCall(call, "dirname") {
			break
		}
		args, ok := astquery.CallArgValues(call)
		if !ok || len(args) == 0 {
			return false
		}
		e = syntax.UnwrapParens(args[0])
	}
	m, ok := e.(*syntax.MagicConst)
	return ok && (m.Token.Kind == syntax.TDir || m.Token.Kind == syntax.TFile)
}
