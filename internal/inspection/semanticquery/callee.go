package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// Callee is the resolved declaration of a function or method call.
type Callee struct {
	Params []index.Param
	ByRef  bool
}

// ResolveCallee resolves a function, method or static method call to its
// declaration (user code or stubs). Methods are looked up on the receiver's
// inferred classes (first class declaring the method wins).
func ResolveCallee(ctx *analysis.Context, call syntax.Node) (Callee, bool) {
	switch c := call.(type) {
	case *syntax.FuncCall:
		if f := ctx.Types().ResolveFunction(c); f != nil {
			return Callee{Params: f.Params, ByRef: f.ByRef}, true
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return Callee{}, false
		}
		for _, cls := range ctx.TypeOf(c.Var).Classes() {
			if m := ctx.Index().FindMethod(strings.TrimPrefix(cls, `\`), id.Value, ctx.PHP); m != nil {
				return Callee{Params: m.Params, ByRef: m.ByRef}, true
			}
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return Callee{}, false
		}
		cls := StaticCallClass(ctx, c.Class)
		if cls == "" {
			return Callee{}, false
		}
		if m := ctx.Index().FindMethod(cls, id.Value, ctx.PHP); m != nil {
			return Callee{Params: m.Params, ByRef: m.ByRef}, true
		}
	}
	return Callee{}, false
}

// StaticCallClass resolves the class part of a static call or class
// constant fetch (self/static/parent through the enclosing class); "" when
// dynamic or unresolvable.
func StaticCallClass(ctx *analysis.Context, e syntax.Expr) string {
	nm, ok := e.(*syntax.Name)
	if !ok {
		return ""
	}
	switch strings.ToLower(nm.Value) {
	case "self", "static":
		return ctx.Types().ClassFQN(syntax.EnclosingClass(nm))
	case "parent":
		if c := ctx.Index().Class(ctx.Types().ClassFQN(syntax.EnclosingClass(nm)), ctx.PHP); c != nil {
			return strings.TrimPrefix(c.Parent, `\`)
		}
		return ""
	}
	return ctx.Names().Class(nm.Value, nm.Span().Start)
}
