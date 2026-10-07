package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/infer"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// staticClosureCanBeUsed suggests `static` for closures and arrow functions
// that never use the object they are created in.
type staticClosureCanBeUsed struct{}

func init() { register(staticClosureCanBeUsed{}) }

func (staticClosureCanBeUsed) ID() string { return "StaticClosureCanBeUsed" }

func (staticClosureCanBeUsed) Semantic() {}

func (staticClosureCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClosure, syntax.KArrowFunction}
}

func (staticClosureCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP54 || n.Span().Len() == 0 { // D1
		return
	}
	var region syntax.Node
	var attrs []*syntax.AttributeGroup
	kw := syntax.TFunction
	switch f := n.(type) {
	case *syntax.Closure:
		if f.Static || f.Body == nil || len(f.Body.Stmts) == 0 { // D2 / D3
			return
		}
		region, attrs = f.Body, f.Attrs
	case *syntax.ArrowFunction:
		if f.Static || !ctx.Bool("SUGGEST_FOR_SHORT_FUNCTIONS") { // D2 / D3
			return
		}
		region, attrs, kw = f, f.Attrs, syntax.TFn
	}
	if closureNeedsThis(ctx, region) { // D4 / D5
		return
	}
	sites, unsafe := closureUsageSites(n) // D6
	if unsafe {
		return
	}
	for _, site := range sites { // D7
		if !closureSiteSafe(ctx, site) {
			return
		}
	}
	from := n.Span().Start
	if len(attrs) > 0 {
		from = attrs[len(attrs)-1].Span().End
	}
	tok, ok := util.FindToken(ctx.File, syntax.Span{Start: from, End: n.Span().End}, kw)
	if !ok {
		return
	}
	span := syntax.Span{Start: tok.Start, End: tok.End}
	ctx.Report(span, "Closure does not use $this; declare it static.", analysis.Fix{
		Title: "Declare static",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: syntax.Span{Start: span.Start, End: span.Start}, NewText: "static "}}
		},
	})
}

// closureNeedsThis searches region deeply for `$this` (D4), a `parent::m()`
// call resolving to an instance method (D5) or a `self::m()`/`static::m()`
// call not known to target a static method (D5b).
func closureNeedsThis(ctx *analysis.Context, region syntax.Node) bool {
	found := false
	home := infer.EnclosingClass(region)
	syntax.Inspect(region, func(x syntax.Node) bool {
		if found {
			return false
		}
		switch x := x.(type) {
		case *syntax.Variable:
			if x.NameExpr == nil && x.Name == "this" {
				found = true
			}
		case *syntax.StaticCall:
			if cls, ok := x.Class.(*syntax.Name); ok && home != nil && infer.EnclosingClass(x) == home &&
				(strings.EqualFold(cls.Value, "self") || strings.EqualFold(cls.Value, "static")) { // D5b
				id, ok := x.Name.(*syntax.Identifier)
				if !ok {
					found = true // dynamic method name: unknown target
					break
				}
				fqn := ctx.Types().ClassRef(cls)
				m := ctx.Index().FindMethod(fqn, id.Value, ctx.PHP)
				if fqn == "" || m == nil || !m.Static {
					found = true
				}
				break
			}
			if cls, ok := x.Class.(*syntax.Name); ok && strings.EqualFold(cls.Value, "parent") {
				if id, ok := x.Name.(*syntax.Identifier); ok {
					if fqn := ctx.Types().ClassRef(cls); fqn != "" {
						if m := ctx.Index().FindMethod(fqn, id.Value, ctx.PHP); m != nil && !m.Static {
							found = true
						}
					}
				}
			}
		}
		return !found
	})
	return found
}

// closureUsageSites collects the nodes through which the closure is used
// (D6): argument lists, receiver method calls, array pairs or literals.
// unsafe is true when the closure escapes in a way that may bind it later
// (returned, or its variable used in an unrecognised context).
func closureUsageSites(n syntax.Node) (sites []syntax.Node, unsafe bool) {
	if p, _ := util.ParentSkipParens(n); p != nil {
		if _, ok := p.(*syntax.Return); ok { // D6d
			return nil, true
		}
	}
	switch p := n.Parent().(type) {
	case *syntax.Arg: // D6a
		if p.Value == n {
			if list, ok := p.Parent().(*syntax.ArgList); ok {
				return []syntax.Node{list}, false
			}
		}
	case *syntax.Assign: // D6b
		target, ok := p.Var.(*syntax.Variable)
		if !ok || p.Value != n || p.Op.Kind != syntax.TEqual || target.NameExpr != nil || target.Name == "" {
			return nil, false
		}
		body := util.FuncLikeBody(util.EnclosingFuncLike(p))
		if body == nil {
			return nil, false
		}
		abandoned := false
		syntax.Inspect(body, func(x syntax.Node) bool {
			if abandoned {
				return false
			}
			v, ok := x.(*syntax.Variable)
			if !ok || v == target || v.NameExpr != nil || v.Name != target.Name {
				return true
			}
			switch vp := v.Parent().(type) {
			case *syntax.Arg:
				if list, ok := vp.Parent().(*syntax.ArgList); ok {
					sites = append(sites, list)
					return true
				}
			case *syntax.MethodCall:
				sites = append(sites, vp)
				return true
			case *syntax.StaticCall:
				sites = append(sites, vp)
				return true
			case *syntax.FuncCall:
				if vp.Name == syntax.Expr(v) {
					return true // direct invocation `$v(...)` cannot rebind it
				}
			}
			abandoned = true // unrecognised use: the closure may be bound later
			return false
		})
		if abandoned {
			return nil, true
		}
		return sites, false
	case *syntax.ArrayItem: // D6c
		if p.Value != n {
			return nil, false
		}
		if p.Key != nil {
			return []syntax.Node{p}, false
		}
		if arr := p.Parent(); arr != nil {
			return []syntax.Node{arr}, false
		}
	}
	return nil, false
}

// closureSiteSafe implements D7 (S1–S6).
func closureSiteSafe(ctx *analysis.Context, site syntax.Node) bool {
	switch s := site.(type) {
	case *syntax.ArgList:
		switch call := s.Parent().(type) {
		case *syntax.MethodCall: // S1
			id, ok := call.Name.(*syntax.Identifier)
			if ok && strings.EqualFold(id.Value, "bind") {
				return bindNullScope(ctx, s, 1) && closureMethodResolves(ctx, ctx.Types().ClassRef(call.Var), "bind")
			}
			return false
		case *syntax.StaticCall: // S1
			id, ok := call.Name.(*syntax.Identifier)
			if ok && strings.EqualFold(id.Value, "bind") {
				return bindNullScope(ctx, s, 1) && closureMethodResolves(ctx, ctx.Types().ClassRef(call.Class), "bind")
			}
			return true
		case *syntax.FuncCall: // S2
			f := ctx.Types().ResolveFunction(call)
			return f != nil && !strings.Contains(strings.TrimPrefix(f.FQN, `\`), `\`)
		}
		return false // S3
	case *syntax.MethodCall: // S4
		id, ok := s.Name.(*syntax.Identifier)
		return ok && strings.EqualFold(id.Value, "bindTo") && bindNullScope(ctx, s.Args, 0) &&
			ctx.Index().FindMethod("Closure", "bindTo", ctx.PHP) != nil
	case *syntax.StaticCall: // S4 (`$v::m()` form)
		id, ok := s.Name.(*syntax.Identifier)
		return ok && strings.EqualFold(id.Value, "bindTo") && bindNullScope(ctx, s.Args, 0) &&
			ctx.Index().FindMethod("Closure", "bindTo", ctx.PHP) != nil
	case *syntax.ArrayItem: // S5
		return util.EnclosingFuncLike(s) == nil
	}
	return false // S6
}

// bindNullScope reports whether argument i of list exists and is null.
func bindNullScope(ctx *analysis.Context, list *syntax.ArgList, i int) bool {
	if list == nil {
		return false
	}
	args, ok := util.ArgValues(list)
	return ok && len(args) > i && pdvNullConst(args[i])
}

// closureMethodResolves reports whether cls::name resolves to the built-in
// Closure method.
func closureMethodResolves(ctx *analysis.Context, cls, name string) bool {
	if cls == "" {
		return false
	}
	m := ctx.Index().FindMethod(cls, name, ctx.PHP)
	return m != nil && strings.EqualFold(strings.TrimPrefix(m.Class, `\`), "Closure")
}
