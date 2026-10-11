package callablemethodvalidity

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// callableMethodValidity reports is_callable() member callbacks naming a
// non-public method, or a non-static method without an object.
type callableMethodValidity struct{}

func (callableMethodValidity) ID() string               { return "CallableMethodValidity" }
func (callableMethodValidity) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

// Semantic marks the rule as needing the project index.
func (callableMethodValidity) Semantic() {}

func (callableMethodValidity) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "is_callable") { // D1
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 1 { // D2
		return
	}
	a := args[0]
	vals := semanticquery.DiscoverValues(ctx.Types(), a) // D3
	if len(vals) != 1 {
		return
	}
	var m *index.Method
	class := ""     // the class the callback names
	static := false // the callback is referenced without an object
	switch v := vals[0].(type) {
	case *syntax.Literal: // D4/D5 string form
		if v.LitKind != syntax.LitString {
			return
		}
		s, ok := astquery.StringLiteralValue(v.Raw)
		if !ok {
			return
		}
		cls, meth, ok := strings.Cut(s, "::")
		if !ok || cls == "" || meth == "" {
			return
		}
		class = strings.TrimPrefix(cls, `\`)
		m = ctx.Index().FindMethod(class, meth, ctx.PHP)
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
		recv, nameExpr := syntax.UnwrapParens(v.Items[0].Value), syntax.UnwrapParens(v.Items[1].Value)
		lit, ok := nameExpr.(*syntax.Literal)
		if !ok || lit.LitKind != syntax.LitString {
			return
		}
		meth, ok := astquery.StringLiteralValue(lit.Raw)
		if !ok || meth == "" {
			return
		}
		var classes []string // D6
		switch r := recv.(type) {
		case *syntax.Literal:
			if s, ok := astquery.StringLiteralValue(r.Raw); ok && r.LitKind == syntax.LitString {
				classes = []string{strings.TrimPrefix(s, `\`)}
			}
		case *syntax.ClassConstFetch:
			if id, ok := r.Name.(*syntax.Identifier); ok && strings.EqualFold(id.Value, "class") {
				classes = semanticquery.ReferencedClasses(ctx, r.Class)
				static = true
			} else {
				classes = semanticquery.ReferencedClasses(ctx, r)
			}
		default:
			classes = semanticquery.ReferencedClasses(ctx, r)
		}
		for _, c := range classes {
			if m = ctx.Index().FindMethod(c, meth, ctx.PHP); m != nil {
				class = c
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
	cl, hasThis := callScope(call)
	// PHP binds the current $this to a callback written without an object when
	// the calling class is the named class or one of its subclasses (D8).
	bound := static && hasThis && scopeIsA(ctx, cl, class)
	if m.Visibility != index.Public && !callableVisibleFromScope(ctx, cl, m) { // D7
		// An inaccessible method is routed to the matching magic handler (E7):
		// __call when an object is involved, __callStatic otherwise.
		handler := "__call"
		if static && !bound {
			handler = "__callStatic"
		}
		if ctx.Index().FindMethod(class, handler, ctx.PHP) != nil {
			return
		}
		ctx.ReportNode(a, "Method '"+m.Name+"' is not public, so the callback cannot be invoked from outside.")
	}
	if !m.Static && static && !bound { // D8
		ctx.ReportNode(a, "Method '"+m.Name+"' is not static but is referenced without an object.")
	}
}

// callScope returns the class-like declaration the call is written in and
// whether $this exists there: closures and arrow functions keep the scope of
// their class but lose $this when static; a named function declared inside a
// method has no class scope.
func callScope(call syntax.Node) (cl *syntax.ClassLike, hasThis bool) {
	static := false
	for p := call.Parent(); p != nil; p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Function:
			return nil, false
		case *syntax.Closure:
			static = static || x.Static
		case *syntax.ArrowFunction:
			static = static || x.Static
		case *syntax.Method:
			static = static || x.Modifiers.Has(syntax.TStatic)
		case *syntax.ClassLike:
			return x, !static
		}
	}
	return nil, false
}

// scopeIsA reports whether the scope class cl (never nil where $this exists)
// is class or one of its subclasses. The class a trait ends up in is unknown, so a trait scope
// counts as compatible.
func scopeIsA(ctx *analysis.Context, cl *syntax.ClassLike, class string) bool {
	scope := strings.TrimPrefix(ctx.Types().ClassFQN(cl), `\`)
	return cl.ClassKind == syntax.KindTrait || strings.EqualFold(scope, class) || ctx.Index().IsSubtype(scope, class, ctx.PHP)
}

// callableVisibleFromScope reports whether the non-public method m is
// accessible from the class scope cl the call is written in (D7): the
// declaring class (or a class using the declaring trait) for a private
// method; a class related to the declaring class by inheritance, in either
// direction, for a protected one.
func callableVisibleFromScope(ctx *analysis.Context, cl *syntax.ClassLike, m *index.Method) bool {
	if cl == nil {
		return false
	}
	ix := ctx.Index()
	decl := strings.TrimPrefix(m.Class, `\`)
	scope := strings.TrimPrefix(ctx.Types().ClassFQN(cl), `\`)

	if strings.EqualFold(scope, decl) {
		return true
	}
	if m.Visibility == index.Private {
		c := ix.Class(decl, ctx.PHP)
		return c != nil && c.Kind == syntax.KindTrait && ix.IsSubtype(scope, decl, ctx.PHP)
	}
	return ix.IsSubtype(scope, decl, ctx.PHP) || ix.IsSubtype(decl, scope, ctx.PHP)
}
