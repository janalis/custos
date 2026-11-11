package argumentcountmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// resolve lists every declaration the call may bind to and names the
// nominal one; false when any of them is unknown or uncertain. callee is
// the call's agreed contract: a resolved name, a receiver of known classes.
func resolve(ctx *analysis.Context, call syntax.Node, callee semanticquery.Callee) (string, []target, bool) {
	switch c := call.(type) {
	case *syntax.FuncCall:
		return function(ctx, c)
	case *syntax.MethodCall:
		id := c.Name.(*syntax.Identifier)
		name := ""
		var all []target
		for _, cls := range ctx.TypeOf(c.Var).Without("null").Classes() {
			n, ts, ok := methodTargets(ctx, strings.TrimPrefix(cls, `\`), id.Value, true)
			if !ok {
				return "", nil, false
			}
			if name == "" {
				name = n
			}
			all = append(all, ts...)
		}
		return name, all, true
	case *syntax.StaticCall:
		id := c.Name.(*syntax.Identifier)
		cls := semanticquery.StaticCallClass(ctx, c.Class)
		// Only `static::` binds late; a named class, self and parent call
		// the method that class resolves.
		late := strings.EqualFold(c.Class.(*syntax.Name).Value, "static")
		return methodTargets(ctx, cls, id.Value, late)
	}
	return construction(ctx, call.(*syntax.New), callee)
}

// function resolves a function call to its single declaration.
func function(ctx *analysis.Context, c *syntax.FuncCall) (string, []target, bool) {
	ix := ctx.Index()
	f := ctx.Types().ResolveFunction(c)
	if f.Builtin && !trustedBuiltin(f.File, f.FQN, ctx.PHP) {
		return "", nil, false
	}
	if !f.Builtin {
		// A polyfill of a function PHP declares in another version, or a
		// function the project declares more than once (conditionally
		// loaded variants), may not be the declaration that runs.
		if other := ix.Function(f.FQN, 0); other != nil && other.Builtin {
			return "", nil, false
		}
		if len(ix.FunctionDecls(f.FQN, 0)) != 1 {
			return "", nil, false
		}
	}
	return strings.TrimPrefix(f.FQN, `\`), []target{{params: f.Params, builtin: f.Builtin}}, true
}

// construction resolves the constructor a `new` runs.
func construction(ctx *analysis.Context, c *syntax.New, callee semanticquery.Callee) (string, []target, bool) {
	cls := ctx.Types().ClassRef(c.Class.(*syntax.Name))
	declared := ctx.Index().Class(cls, ctx.PHP)
	if declared == nil || declared.Kind != syntax.KindClass || declared.Abstract || !semanticquery.HierarchyResolved(ctx.Index(), cls, ctx.PHP) {
		return "", nil, false
	}
	if callee.Builtin {
		if m := ctx.Index().FindMethod(cls, "__construct", ctx.PHP); m == nil || !trustedMethod(ctx, m) {
			return "", nil, false
		}
	}
	return strings.TrimPrefix(declared.FQN, `\`) + "::__construct", []target{{params: callee.Params, builtin: callee.Builtin}}, true
}

// methodTargets lists method name as class cls resolves it and, with
// overrides, as every indexed subtype of cls resolves it, since the
// runtime object may be any of them.
func methodTargets(ctx *analysis.Context, cls, name string, overrides bool) (string, []target, bool) {
	ix := ctx.Index()
	c := ix.Class(cls, ctx.PHP)
	// A trait method runs in the classes using the trait, which may declare
	// the callee themselves.
	if c == nil || c.Kind == syntax.KindTrait {
		return "", nil, false
	}
	m, ok := method(ctx, cls, name)
	if !ok {
		return "", nil, false
	}
	label := strings.TrimPrefix(m.Class, `\`) + "::" + m.Name
	targets := []target{{params: m.Params, builtin: m.Builtin}}
	if !overrides || m.Visibility == index.Private || m.Final || c.Final {
		return label, targets, true
	}
	seen := map[string]bool{}
	queue := []string{cls}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range ix.ChildrenAll(cur) {
			k := strings.ToLower(strings.TrimPrefix(child, `\`))
			if seen[k] {
				continue
			}
			seen[k] = true
			sub, ok := method(ctx, child, name)
			if !ok {
				return "", nil, false
			}
			targets = append(targets, target{params: sub.Params, builtin: sub.Builtin})
			queue = append(queue, child)
		}
	}
	return label, targets, true
}

// method resolves name on class cls when that declaration is certain: the
// class and the one declaring the method are declared once, every ancestor
// is indexed, and neither a `@method` tag nor a magic handler serves it.
func method(ctx *analysis.Context, cls, name string) (*index.Method, bool) {
	ix := ctx.Index()
	if len(ix.ClassDecls(cls, ctx.PHP)) != 1 || !semanticquery.HierarchyResolved(ix, cls, ctx.PHP) {
		return nil, false
	}
	m := ix.FindMethod(cls, name, ctx.PHP)
	if m == nil || m.Magic || len(ix.ClassDecls(m.Class, ctx.PHP)) != 1 || m.Builtin && !trustedMethod(ctx, m) {
		return nil, false
	}
	// A method the caller cannot see is handed to __call or __callStatic.
	if m.Visibility != index.Public && (ix.FindMethod(cls, "__call", ctx.PHP) != nil || ix.FindMethod(cls, "__callStatic", ctx.PHP) != nil) {
		return nil, false
	}
	return m, true
}

// trustedMethod reports whether builtin method m states the parameter
// counts PHP enforces (see trustedBuiltin).
func trustedMethod(ctx *analysis.Context, m *index.Method) bool {
	decls := ctx.Index().ClassDecls(m.Class, 0)
	return len(decls) == 1 && trustedBuiltin(decls[0].File, m.Class+"::"+m.Name, ctx.PHP)
}
