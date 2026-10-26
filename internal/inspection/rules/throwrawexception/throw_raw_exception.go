package throwrawexception

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
)

// throwRawException reports throwing the base \Exception class, and
// exceptions created without any argument.
type throwRawException struct{}

func (throwRawException) ID() string               { return "ThrowRawException" }
func (throwRawException) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KThrow} }
func (throwRawException) Semantic()                {}
func (throwRawException) Check(ctx *analysis.Context, n syntax.Node) {
	nw, ok := syntax.UnwrapParens(n.(*syntax.Throw).Expr).(*syntax.New) // D1 (parentheses looked through)
	if !ok {
		return
	}
	name, ok := nw.Class.(*syntax.Name)
	if !ok || names.IsSpecialClass(name.Value) {
		return
	}
	fqn := ctx.Names().Class(name.Value, name.Span().Start)
	if strings.EqualFold(fqn, "Exception") { // D2
		const msg = "Throw a more specific exception class than \\Exception."
		if treCaughtAsRuntime(ctx, n) {
			ctx.ReportNode(name, msg)
			return
		}
		ctx.ReportNode(name, msg, diagnostic.Fix{
			Title: "Throw \\RuntimeException",
			Edits: func() []diagnostic.TextEdit {
				q := name.Value[:strings.LastIndexByte(name.Value, '\\')+1]
				if q == "" || q == `\` {
					q = `\`
				}
				return []diagnostic.TextEdit{{Span: name.Span(), NewText: q + "RuntimeException"}}
			},
		})
		return
	}
	if !ctx.Bool("REPORT_MISSING_ARGUMENTS") || (nw.Args != nil && len(nw.Args.Args) > 0) { // D3
		return
	}
	ix := ctx.Index()
	if ix.ClassCount(fqn) != 1 {
		return
	}
	cls := ix.Class(fqn, ctx.PHP) // never nil once the class is indexed
	ctor := ix.FindMethod(fqn, "__construct", ctx.PHP)
	if ctor == nil || len(ctor.Params) != 3 {
		return
	}
	if !ctor.Builtin && !ctor.ForwardsMessage { // E7: the constructor may set its own message
		return
	}
	if _, own := cls.Props["message"]; own {
		return
	}
	// custos: `__construct(string $message = 'Resource not found.', …)`
	// presets the message as surely as a `$message` property does.
	if d := strings.TrimSpace(ctor.Params[0].Default); len(d) > 2 && (d[0] == '\'' || d[0] == '"') {
		return
	}
	if treUserPresetsMessage(ctx, fqn) {
		return
	}
	ctx.ReportNode(nw, "Pass a message when throwing this exception.")
}

// treCaughtAsRuntime reports whether the throw sits in the body of a try,
// in the same function-like, with a catch naming \RuntimeException: F1
// would route the exception to that handler instead.
func treCaughtAsRuntime(ctx *analysis.Context, n syntax.Node) bool {
	child := n
	for p := n.Parent(); p != nil && !syntax.IsFuncLike(p); child, p = p, p.Parent() {
		t, ok := p.(*syntax.Try)
		if !ok || child != syntax.Node(t.Body) {
			continue
		}
		for _, c := range t.Catches {
			for _, ty := range c.Types {
				if strings.EqualFold(ctx.Names().Class(ty.Value, ty.Span().Start), "RuntimeException") {
					return true
				}
			}
		}
	}
	return false
}

// treUserPresetsMessage reports whether a user (non-stub) parent class of
// fqn, or a trait such a class uses, declares a `$message` property. Built-in
// classes (whose `$message` is the empty default) are not consulted.
func treUserPresetsMessage(ctx *analysis.Context, fqn string) bool {
	ix := ctx.Index()
	declares := func(c *index.Class) bool {
		if c == nil || isBuiltinClass(c, ctx.PHP) {
			return false
		}
		if _, ok := c.Props["message"]; ok {
			return true
		}
		for _, t := range c.Traits {
			if tc := ix.Class(t, ctx.PHP); tc != nil && !isBuiltinClass(tc, ctx.PHP) {
				if _, ok := tc.Props["message"]; ok {
					return true
				}
			}
		}
		return false
	}
	if declares(ix.Class(fqn, ctx.PHP)) {
		return true
	}
	for _, c := range ix.ParentChain(fqn, ctx.PHP) {
		if declares(c) {
			return true
		}
	}
	return false
}
