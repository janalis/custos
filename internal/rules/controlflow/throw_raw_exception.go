package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/syntax"
)

// throwRawException reports throwing the base \Exception class, and
// exceptions created without any argument.
type throwRawException struct{}

func init() { register(throwRawException{}) }

func (throwRawException) ID() string { return "ThrowRawException" }

func (throwRawException) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KThrow} }

func (throwRawException) Semantic() {}

func (throwRawException) Check(ctx *analysis.Context, n syntax.Node) {
	nw, ok := util.UnwrapParens(n.(*syntax.Throw).Expr).(*syntax.New) // D1 (parentheses looked through)
	if !ok {
		return
	}
	name, ok := nw.Class.(*syntax.Name)
	if !ok || names.IsSpecialClass(name.Value) {
		return
	}
	fqn := ctx.Names().Class(name.Value, name.Span().Start)
	if strings.EqualFold(fqn, "Exception") { // D2
		ctx.ReportNode(name, "Throw a more specific exception class than \\Exception.", analysis.Fix{
			Title: "Throw \\RuntimeException",
			Edits: func() []analysis.TextEdit {
				q := name.Value[:strings.LastIndexByte(name.Value, '\\')+1]
				if q == "" || q == `\` {
					q = `\`
				}
				return []analysis.TextEdit{{Span: name.Span(), NewText: q + "RuntimeException"}}
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
	cls := ix.Class(fqn, ctx.PHP)
	if cls == nil {
		return
	}
	ctor := ix.FindMethod(fqn, "__construct", ctx.PHP)
	if ctor == nil || len(ctor.Params) != 3 {
		return
	}
	if _, own := cls.Props["message"]; own {
		return
	}
	if treUserPresetsMessage(ctx, fqn) {
		return
	}
	ctx.ReportNode(nw, "Pass a message when throwing this exception.")
}

// treUserPresetsMessage reports whether a user (non-stub) parent class of
// fqn, or a trait such a class uses, declares a `$message` property. Built-in
// classes (whose `$message` is the empty default) are not consulted.
func treUserPresetsMessage(ctx *analysis.Context, fqn string) bool {
	ix := ctx.Index()
	declares := func(c *index.Class) bool {
		if c == nil || util.IsBuiltinClass(c, ctx.PHP) {
			return false
		}
		if _, ok := c.Props["message"]; ok {
			return true
		}
		for _, t := range c.Traits {
			if tc := ix.Class(t, ctx.PHP); tc != nil && !util.IsBuiltinClass(tc, ctx.PHP) {
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
