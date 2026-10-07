package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/syntax"
)

// callableMethodValidity reports is_callable() member callbacks naming a
// non-public method, or a non-static method without an object.
type callableMethodValidity struct{}

func init() { register(callableMethodValidity{}) }

func (callableMethodValidity) ID() string { return "CallableMethodValidity" }

func (callableMethodValidity) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

// Semantic marks the rule as needing the project index.
func (callableMethodValidity) Semantic() {}

func (callableMethodValidity) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "is_callable") { // D1
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 1 { // D2
		return
	}
	a := args[0]
	vals := util.DiscoverValues(ctx.Types(), a) // D3
	if len(vals) != 1 {
		return
	}
	var m *index.Method
	static := false // the callback is referenced without an object
	switch v := vals[0].(type) {
	case *syntax.Literal: // D4/D5 string form
		if v.LitKind != syntax.LitString {
			return
		}
		s, ok := util.StringLiteralValue(v.Raw)
		if !ok {
			return
		}
		cls, meth, ok := strings.Cut(s, "::")
		if !ok || cls == "" || meth == "" {
			return
		}
		m = ctx.Index().FindMethod(strings.TrimPrefix(cls, `\`), meth, ctx.PHP)
		static = true
	case *syntax.Array: // D4/D5 array form
		if len(v.Items) != 2 {
			return
		}
		for _, it := range v.Items {
			if it == nil || it.Value == nil || it.Key != nil || it.Unpack || it.ByRef {
				return
			}
		}
		recv, nameExpr := util.UnwrapParens(v.Items[0].Value), util.UnwrapParens(v.Items[1].Value)
		lit, ok := nameExpr.(*syntax.Literal)
		if !ok || lit.LitKind != syntax.LitString {
			return
		}
		meth, ok := util.StringLiteralValue(lit.Raw)
		if !ok || meth == "" {
			return
		}
		var classes []string // D6
		switch r := recv.(type) {
		case *syntax.Literal:
			if s, ok := util.StringLiteralValue(r.Raw); ok && r.LitKind == syntax.LitString {
				classes = []string{strings.TrimPrefix(s, `\`)}
			}
		case *syntax.ClassConstFetch:
			if id, ok := r.Name.(*syntax.Identifier); ok && strings.EqualFold(id.Value, "class") {
				classes = semClassesOf(ctx, r.Class)
				static = true
			} else {
				classes = semClassesOf(ctx, r)
			}
		default:
			classes = semClassesOf(ctx, r)
		}
		for _, c := range classes {
			if m = ctx.Index().FindMethod(c, meth, ctx.PHP); m != nil {
				break
			}
		}
		if t := ctx.TypeOf(recv); t.Has("string") || t.Has("callable") {
			static = true
		}
	default:
		return
	}
	if m == nil {
		return
	}
	if m.Visibility != index.Public && !callableVisibleFromScope(ctx, call, m) { // D7
		ctx.ReportNode(a, "Method '"+m.Name+"' is not public, so the callback cannot be invoked from outside.")
	}
	if !m.Static && static { // D8
		ctx.ReportNode(a, "Method '"+m.Name+"' is not static but is referenced without an object.")
	}
}

// callableVisibleFromScope reports whether the non-public method m is
// accessible from the class scope the call is written in (D7): the declaring
// class (or a class using the declaring trait) for a private method; a class
// related to the declaring class by inheritance, in either direction, for a
// protected one. Closures keep the scope of their class; a named function
// declared inside a method has none.
func callableVisibleFromScope(ctx *analysis.Context, call syntax.Node, m *index.Method) bool {
	var cl *syntax.ClassLike
	for p := call.Parent(); p != nil && cl == nil; p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Function:
			return false
		case *syntax.ClassLike:
			cl = x
		}
	}
	if cl == nil {
		return false
	}
	ix := ctx.Index()
	decl := strings.TrimPrefix(m.Class, `\`)
	scope := strings.TrimPrefix(ctx.Types().ClassFQN(cl), `\`)
	if scope == "" { // anonymous class: only its parent can relate it to decl
		if m.Visibility != index.Protected || len(cl.Extends) == 0 {
			return false
		}
		parent := ctx.Names().Class(cl.Extends[0].Value, cl.Extends[0].Span().Start)
		return ix.IsSubtype(parent, decl, ctx.PHP)
	}
	if strings.EqualFold(scope, decl) {
		return true
	}
	if m.Visibility == index.Private {
		c := ix.Class(decl, ctx.PHP)
		return c != nil && c.Kind == syntax.KindTrait && ix.IsSubtype(scope, decl, ctx.PHP)
	}
	return ix.IsSubtype(scope, decl, ctx.PHP) || ix.IsSubtype(decl, scope, ctx.PHP)
}
