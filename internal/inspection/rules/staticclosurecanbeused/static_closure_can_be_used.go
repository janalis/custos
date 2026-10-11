package staticclosurecanbeused

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// staticClosureCanBeUsed suggests `static` for closures and arrow functions
// that never use the object they are created in.
type staticClosureCanBeUsed struct{}

func (staticClosureCanBeUsed) ID() string { return "StaticClosureCanBeUsed" }
func (staticClosureCanBeUsed) Semantic()  {}
func (staticClosureCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClosure, syntax.KArrowFunction}
}

func (staticClosureCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP54 || n.Span().Len() == 0 { // D1
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
	sites, unsafe := closureUsageSites(ctx, n) // D6
	if unsafe {
		return
	}
	fixable := true
	for _, site := range sites { // D7
		switch closureSiteSafe(ctx, site) {
		case siteUnsafe:
			return
		case siteUnknown:
			fixable = false
		}
	}
	from := n.Span().Start
	if len(attrs) > 0 {
		from = attrs[len(attrs)-1].Span().End
	}
	tok, _ := astquery.FindToken(ctx.File, syntax.Span{Start: from, End: n.Span().End}, kw) // always present
	span := syntax.Span{Start: tok.Start, End: tok.End}
	const msg = "Closure does not use $this; declare it static."
	if !fixable {
		// A user-defined or unknown consumer may bind the closure (F2).
		ctx.Report(span, msg)
		return
	}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Declare static",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: syntax.Span{Start: span.Start, End: span.Start}, NewText: "static "}}
		},
	})
}

// closureNeedsThis searches region deeply for `$this` (D4), code loaded
// into the scope of a closure created in an object (D4b), a `Name::m()` call resolving to an
// instance method (D5) or a `self::m()`/`static::m()` call not known to
// target a static method (D5b).
func closureNeedsThis(ctx *analysis.Context, region syntax.Node) bool {
	found := false
	home := syntax.EnclosingClass(region)
	syntax.Inspect(region, func(x syntax.Node) bool {
		if found {
			return false
		}
		switch x := x.(type) {
		case *syntax.Variable:
			if x.NameExpr == nil && x.Name == "this" {
				found = true
			}
		case *syntax.Include, *syntax.Eval: // D4b: the loaded code inherits `$this`
			found = closureCapturesThis(region)
		case *syntax.StaticCall:
			cls, ok := x.Class.(*syntax.Name)
			if !ok {
				break
			}
			known, static := false, false
			if id, ok := x.Name.(*syntax.Identifier); ok {
				if fqn := ctx.Types().ClassRef(cls); fqn != "" {
					if m := ctx.Index().FindMethod(fqn, id.Value, ctx.PHP); m != nil {
						known, static = true, m.Static
					}
				}
			}
			if strings.EqualFold(cls.Value, "self") || strings.EqualFold(cls.Value, "static") { // D5b
				// dynamic or unknown target included; not in a nested class
				found = home != nil && syntax.EnclosingClass(x) == home && (!known || !static)
				break
			}
			found = known && !static // D5: `parent::`, `Runner::`, `Base::`
		}
		return !found
	})
	return found
}

// closureCapturesThis reports whether a closure whose search region is
// region is created where `$this` exists: in an instance method or a
// property hook, possibly through enclosing non-static closures. Elsewhere
// (function, static method, static closure, file level) the closure holds
// no object unless a consumer binds one, which D6/D7 rule on.
func closureCapturesThis(region syntax.Node) bool {
	for s := syntax.EnclosingVariableScope(region); s != nil; s = syntax.EnclosingVariableScope(s) {
		switch f := s.(type) {
		case *syntax.Closure:
			if f.Static {
				return false
			}
		case *syntax.ArrowFunction:
			if f.Static {
				return false
			}
		case *syntax.Function:
			return false
		case *syntax.Method:
			return !f.Modifiers.Has(syntax.TStatic)
		default: // property hook
			return true
		}
	}
	return false
}

// closureUsageSites collects the nodes through which the closure is used
// (D6): argument lists, receiver method calls, array pairs or literals.
// unsafe is true when the closure escapes in a way that may bind it later
// (returned, stored, or its variable used in an unrecognised context).
func closureUsageSites(ctx *analysis.Context, n syntax.Node) (sites []syntax.Node, unsafe bool) {
	// D6: the closure's value is also the value of the parentheses, `??`,
	// `?:` branch or `match` arm around it.
	val := n
forward:
	for {
		switch p := val.Parent().(type) {
		case *syntax.Paren:
			val = p
		case *syntax.Binary:
			if p.Op.Kind != syntax.TCoalesce {
				break forward
			}
			val = p
		case *syntax.Ternary:
			if p.Cond == val && p.Then != nil {
				break forward // only tested
			}
			val = p
		case *syntax.MatchArm:
			if p.Body != val {
				break forward
			}
			val = p.Parent()
		default:
			break forward
		}
	}
	switch p := val.Parent().(type) {
	case *syntax.ExprStmt: // discarded
		return nil, false
	case *syntax.FuncCall: // `(function () {…})()` cannot rebind it
		return nil, false
	case *syntax.MethodCall: // D6e
		return []syntax.Node{p}, false
	case *syntax.StaticCall: // D6e
		return []syntax.Node{p}, false
	case *syntax.Arg: // D6a
		return []syntax.Node{p.Parent()}, false
	case *syntax.Assign: // D6b
		target, ok := p.Var.(*syntax.Variable)
		if !ok || p.Op.Kind != syntax.TEqual || target.NameExpr != nil || target.Name == "" {
			// Stored in a property, an array element, a static or dynamic
			// variable: it may be bound later (custos diverges, see spec).
			return nil, true
		}
		if _, stmt := p.Parent().(*syntax.ExprStmt); !stmt {
			// The assignment's own value travels on (`return $v = …`,
			// `$this->hook = $v = …`).
			return nil, true
		}
		var uses []*syntax.Variable
		if body := syntax.FuncLikeBody(syntax.EnclosingFuncLike(p)); body != nil {
			// occurrences of the name in the body, in source order (indexed once per body)
			uses = flowquery.VarOccurrences(ctx.File, body)[target.Name]
		} else {
			// A file-level variable is a global: every use in the file
			// counts, and none may be out of sight.
			scope := closureFileScope(ctx)
			if scope.open {
				return nil, true
			}
			uses = scope.vars[target.Name]
		}
		for _, v := range uses {
			if v == target {
				continue
			}
			switch vp := v.Parent().(type) {
			case *syntax.Arg:
				sites = append(sites, vp.Parent())
				continue
			case *syntax.MethodCall:
				sites = append(sites, vp)
				continue
			case *syntax.StaticCall:
				sites = append(sites, vp)
				continue
			case *syntax.FuncCall:
				if vp.Name == syntax.Expr(v) {
					continue // direct invocation `$v(...)` cannot rebind it
				}
			}
			return nil, true // unrecognised use: the closure may be bound later
		}
		return sites, false
	case *syntax.ArrayItem: // D6c
		if p.Value != val {
			return nil, false
		}
		if p.Key != nil {
			return []syntax.Node{p}, false
		}
		return []syntax.Node{p.Parent()}, false
	}
	// D6d/D6f: returned, yielded, the value of an arrow function, …
	return nil, true
}

// closureFileVars indexes the simple variables of a whole file.
type closureFileVars struct {
	vars map[string][]*syntax.Variable
	// open: the file reaches its globals in ways the index cannot follow.
	open bool
}

// closureFileScope returns the variable index of the current file, built
// once per file (D6b at file level).
func closureFileScope(ctx *analysis.Context) *closureFileVars {
	return ctx.Memo("file-vars", func() any {
		scope := &closureFileVars{vars: map[string][]*syntax.Variable{}}
		for _, st := range ctx.File.Stmts {
			syntax.Inspect(st, func(x syntax.Node) bool {
				switch x := x.(type) {
				case *syntax.Variable:
					if x.NameExpr != nil || x.Name == "GLOBALS" {
						scope.open = true
					} else {
						scope.vars[x.Name] = append(scope.vars[x.Name], x)
					}
				case *syntax.Include, *syntax.Eval:
					scope.open = true
				case *syntax.FuncCall:
					switch ctx.GlobalFunctionName(x) {
					case "compact", "extract", "get_defined_vars":
						scope.open = true
					}
				}
				return true
			})
		}
		return scope
	}).(*closureFileVars)
}

// closureSite is the D7 verdict for one usage site.
type closureSite uint8

const (
	siteUnsafe  closureSite = iota // may bind an object: no report
	siteUnknown                    // consumer not known to PHP itself: report, no fix
	siteSafe
)

// closureSiteSafe implements D7 (S1–S6).
func closureSiteSafe(ctx *analysis.Context, site syntax.Node) closureSite {
	safe := func(ok bool) closureSite {
		if ok {
			return siteSafe
		}
		return siteUnsafe
	}
	switch s := site.(type) {
	case *syntax.ArgList:
		switch call := s.Parent().(type) {
		case *syntax.MethodCall: // S1
			id, ok := call.Name.(*syntax.Identifier)
			return safe(ok && strings.EqualFold(id.Value, "bind") &&
				bindNullScope(ctx, s, 1) && closureMethodResolves(ctx, ctx.Types().ClassRef(call.Var), "bind"))
		case *syntax.StaticCall: // S1
			id, ok := call.Name.(*syntax.Identifier)
			if !ok {
				return siteUnknown
			}
			if strings.EqualFold(id.Value, "bind") {
				return safe(bindNullScope(ctx, s, 1) && closureMethodResolves(ctx, ctx.Types().ClassRef(call.Class), "bind"))
			}
			return closureStaticTarget(ctx, ctx.Types().ClassRef(call.Class), id.Value)
		case *syntax.FuncCall: // S2
			f := ctx.Types().ResolveFunction(call)
			if f == nil || strings.Contains(strings.TrimPrefix(f.FQN, `\`), `\`) {
				return siteUnsafe
			}
			if !f.Builtin {
				return siteUnknown // its body may bind the closure
			}
			return siteSafe
		}
		return siteUnsafe // S3
	case *syntax.MethodCall: // S4
		id, ok := s.Name.(*syntax.Identifier)
		return safe(ok && strings.EqualFold(id.Value, "bindTo") && bindNullScope(ctx, s.Args, 0) &&
			ctx.Index().FindMethod("Closure", "bindTo", ctx.PHP) != nil)
	case *syntax.StaticCall: // S4 (`$v::m()` form)
		id, ok := s.Name.(*syntax.Identifier)
		return safe(ok && strings.EqualFold(id.Value, "bindTo") && bindNullScope(ctx, s.Args, 0) &&
			ctx.Index().FindMethod("Closure", "bindTo", ctx.PHP) != nil)
	}
	return siteUnsafe // S5 / S6: array values
}

// closureStaticTarget classifies cls::name as a consumer: a static method
// of PHP itself is safe; a user-defined one, or an unknown class or method
// (upstream reports those), is reported without a fix; an instance method
// or a magic method reached through __callStatic() (a facade's instance)
// is unsafe.
func closureStaticTarget(ctx *analysis.Context, cls, name string) closureSite {
	ix := ctx.Index()
	if m := ix.FindMethod(cls, name, ctx.PHP); m != nil {
		switch {
		case !m.Static || m.Magic:
			return siteUnsafe
		case m.Builtin:
			return siteSafe
		}
		return siteUnknown
	}
	if cls != "" && ix.FindMethod(cls, "__callStatic", ctx.PHP) != nil {
		return siteUnsafe
	}
	return siteUnknown
}

// bindNullScope reports whether argument i of list exists and is null.
func bindNullScope(ctx *analysis.Context, list *syntax.ArgList, i int) bool {
	args, ok := astquery.ArgValues(list)
	return ok && len(args) > i && syntax.IsNullConst(syntax.UnwrapParens(args[i]))
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
