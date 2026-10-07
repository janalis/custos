package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/syntax"
)

// callee is the resolved declaration of a function or method call.
type callee struct {
	params []index.Param
	byRef  bool
}

// resolveCallee resolves a function, method or static method call to its
// declaration (user code or stubs). Methods are looked up on the receiver's
// inferred classes (first class declaring the method wins).
func resolveCallee(ctx *analysis.Context, call syntax.Node) (callee, bool) {
	switch c := call.(type) {
	case *syntax.FuncCall:
		if f := ctx.Types().ResolveFunction(c); f != nil {
			return callee{params: f.Params, byRef: f.ByRef}, true
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return callee{}, false
		}
		for _, cls := range ctx.TypeOf(c.Var).Classes() {
			if m := ctx.Index().FindMethod(strings.TrimPrefix(cls, `\`), id.Value, ctx.PHP); m != nil {
				return callee{params: m.Params, byRef: m.ByRef}, true
			}
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return callee{}, false
		}
		cls := staticCallClass(ctx, c.Class)
		if cls == "" {
			return callee{}, false
		}
		if m := ctx.Index().FindMethod(cls, id.Value, ctx.PHP); m != nil {
			return callee{params: m.Params, byRef: m.ByRef}, true
		}
	}
	return callee{}, false
}

// staticCallClass resolves the class part of a static call or class
// constant fetch (self/static/parent through the enclosing class); "" when
// dynamic or unresolvable.
func staticCallClass(ctx *analysis.Context, e syntax.Expr) string {
	nm, ok := e.(*syntax.Name)
	if !ok {
		return ""
	}
	switch strings.ToLower(nm.Value) {
	case "self", "static":
		return ctx.Types().ClassFQN(infer.EnclosingClass(nm))
	case "parent":
		if c := ctx.Index().Class(ctx.Types().ClassFQN(infer.EnclosingClass(nm)), ctx.PHP); c != nil {
			return strings.TrimPrefix(c.Parent, `\`)
		}
		return ""
	}
	return ctx.Names().Class(nm.Value, nm.Span().Start)
}

// callArgs returns the argument list of a call node (nil when absent).
func callArgs(call syntax.Node) *syntax.ArgList {
	switch c := call.(type) {
	case *syntax.FuncCall:
		return c.Args
	case *syntax.MethodCall:
		return c.Args
	case *syntax.StaticCall:
		return c.Args
	}
	return nil
}
