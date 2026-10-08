package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// realpathInStreamContext reports realpath() calls that break inside stream
// wrappers (phar://): include paths and parent-directory climbing.
type realpathInStreamContext struct{}

func init() { register(realpathInStreamContext{}) }

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
	parent, _ := util.ParentSkipParens(call)
	if _, isInclude := parent.(*syntax.Include); !isInclude && !hasParentDirLiteral(call) { // D3, D4
		return
	}

	span := call.Span()
	r, ok := realpathReplacement(ctx, arg.Value)
	if !ok {
		ctx.Report(span, realpathGenericMsg)
		return
	}
	ctx.Report(span, "Use '"+r+"' instead: realpath() fails inside stream wrappers.", analysis.Fix{
		Title: "Drop realpath()",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: r}} },
	})
}

// realpathTestContext implements D2: test files and test classes. Unlike
// util.InTestContext it skips anonymous classes and takes the namespace
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
			return util.IsTestClassFQN(fqn)
		}
	}
	return class != nil && util.IsTestClassFQN("\\"+class.Name.Value)
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
			if raw, _, ok := util.QuotedStringRaw(s); !ok || !isAbsolutePathLiteral(raw) {
				return "", false // relative paths resolve differently once realpath() is gone
			}
			return ctx.Text(s), true
		}
	case *syntax.Binary: // R1
		if s.Op.Kind != syntax.TDot || isConcat(s.Left) {
			return "", false
		}
		rest, quote, ok := util.QuotedStringRaw(s.Right)
		// custos: `/..` must be a whole path segment (`/..cache` names a
		// directory).
		parent := func(r string) bool { return r == "/.." || strings.HasPrefix(r, "/../") }
		if !ok || !parent(rest) {
			return "", false
		}
		left := ctx.Text(s.Left)
		dirname := util.QualifiedBuiltin(ctx, "dirname", s.Span().Start) // a namespaced dirname() would capture a bare call
		for parent(rest) {
			rest = rest[3:]
			left = dirname + "(" + left + ")"
		}
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
