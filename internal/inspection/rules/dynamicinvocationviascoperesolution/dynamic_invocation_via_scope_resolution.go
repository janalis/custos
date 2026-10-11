package dynamicinvocationviascoperesolution

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// dynamicInvocationViaScopeResolution flags instance methods called through
// `::` (`self::run()`, `static::run()`, `$obj::run()`).
type dynamicInvocationViaScopeResolution struct{}

func (dynamicInvocationViaScopeResolution) ID() string { return "DynamicInvocationViaScopeResolution" }
func (dynamicInvocationViaScopeResolution) Semantic()  {}
func (dynamicInvocationViaScopeResolution) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KStaticCall}
}

func (dynamicInvocationViaScopeResolution) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.StaticCall)
	id, ok := call.Name.(*syntax.Identifier) // D1
	if !ok || id.Value == "" || call.Args == nil || call.Span().Len() == 0 {
		return
	}
	cls := ctx.Types().ClassRef(call.Class) // D2
	if nm, ok := call.Class.(*syntax.Name); ok && cls == "" && (strings.EqualFold(nm.Value, "self") || strings.EqualFold(nm.Value, "static")) {
		// In a trait self:: is the (unknown) using class; the trait's own
		// methods are members of it.
		cls = ctx.Types().ClassFQN(syntax.EnclosingClass(call))
	}
	if cls == "" {
		return
	}
	ix := ctx.Index()
	m := ix.FindMethod(cls, id.Value, ctx.PHP)
	if m == nil || m.Static || m.Abstract { // interface methods are indexed abstract
		return
	}
	classSpan := call.Class.Span()
	colons, _ := astquery.FindToken(ctx.File, syntax.Span{Start: classSpan.End, End: id.Span().Start}, syntax.TPaamayimNekudotayim)
	colonsSpan := syntax.Span{Start: colons.Start, End: colons.End}
	instanceMsg := "Call '" + id.Value + "' on an instance with '->' instead of '::'."
	left := ctx.SpanText(classSpan)
	if strings.EqualFold(left, "static") || strings.EqualFold(left, "self") || strings.EqualFold(left, astquery.LastNamePart(m.Class)) { // D3
		meth, ok := syntax.EnclosingFuncLike(call).(*syntax.Method)
		if !ok || meth.Name == nil || strings.EqualFold(meth.Name.Value, id.Value) {
			return
		}
		bound := !strings.EqualFold(left, "static") // self::/Foo:: skip late binding
		safe := true
		if owner, ok := meth.Parent().(*syntax.ClassLike); ok {
			if fqn := ctx.Types().ClassFQN(owner); fqn != "" {
				own := ix.FindMethod(fqn, id.Value, ctx.PHP)
				if own == nil {
					return
				}
				if bound {
					safe = boundCallSafe(ctx, fqn, m, own)
				} else {
					// custos: from this scope `$this->m()` prefers the
					// class's private m(); `static::m()` runs the runtime
					// class's redeclaration.
					safe = own.Visibility != index.Private || !overridable(ctx, fqn, own.Name)
				}
				// Spec Divergences: the enclosing class must be the declaring
				// class or one of its descendants.
				if !ix.IsSubtype(fqn, m.Class, ctx.PHP) {
					return
				}
			}
		}
		if meth.Modifiers.Has(syntax.TStatic) { // D3b
			ctx.Report(call.Span(), instanceMsg)
			return
		}
		thisMsg := "Call '" + id.Value + "' with '$this->' instead of '::'."
		if !safe {
			// custos: `self::m()` always runs this class's m(); `$this->m()`
			// runs a subclass override, so no fix then.
			ctx.Report(call.Span(), thisMsg)
			return
		}
		ctx.Report(call.Span(), thisMsg, diagnostic.Fix{ // D3a
			Title: "Use $this->",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: classSpan, NewText: "$this"}, {Span: colonsSpan, NewText: "->"}}
			},
		})
		return
	}
	switch call.Class.(type) { // D4
	case *syntax.Name, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		return
	}
	if !isThis(call.Class) && mayBindThis(ctx, call, m.Class) {
		// custos: PHP runs `$other::m()` on the caller's compatible `$this`,
		// not on `$other`, so `->` would change the receiver.
		ctx.Report(call.Span(), instanceMsg)
		return
	}
	ctx.Report(call.Span(), instanceMsg, diagnostic.Fix{
		Title: "Use ->",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: colonsSpan, NewText: "->"}} },
	})
}

// boundCallSafe reports whether a statically bound call (`self::m()`,
// `Foo::m()`) from class fqn runs the same method as `$this->m()` for
// every instance: the name resolves on fqn to the called method m itself,
// and no subclass can override it (private or final method, final class,
// no indexed descendant declaring it).
func boundCallSafe(ctx *analysis.Context, fqn string, m, own *index.Method) bool {
	if !strings.EqualFold(own.Class, m.Class) {
		return false // `Ancestor::m()` while fqn overrides m: like parent::m()
	}
	return m.Visibility == index.Private || m.Final || !overridable(ctx, fqn, m.Name)
}

// overridable reports whether an instance of fqn can carry another
// declaration of method name than the one fqn sees: fqn is a trait (its
// methods belong to the unknown using class), or it is neither final nor
// an enum and an indexed descendant declares the name.
func overridable(ctx *analysis.Context, fqn, name string) bool {
	ix := ctx.Index()
	if c := ix.Class(fqn, ctx.PHP); c != nil {
		if c.Final || c.Kind == syntax.KindEnum {
			return false
		}
		if c.Kind == syntax.KindTrait {
			return true
		}
	}
	below := ctx.Memo("descendant-methods\x00"+strings.ToLower(fqn), func() any {
		return descendantMethods(ctx, fqn)
	}).(map[string]bool)
	return below[strings.ToLower(name)]
}

// descendantMethods extends semanticquery.DescendantMethods with the
// methods descendants take from traits (declared there or aliased by an
// `as` adaptation), which the shared walk does not see.
func descendantMethods(ctx *analysis.Context, fqn string) map[string]bool {
	ix := ctx.Index()
	out := semanticquery.DescendantMethods(ix, fqn, ctx.PHP)
	seen := map[string]bool{}
	queue := []string{fqn}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, ch := range ix.ChildrenAll(cur) {
			k := strings.ToLower(strings.TrimPrefix(ch, `\`))
			if seen[k] {
				continue
			}
			seen[k] = true
			queue = append(queue, ch)
			for _, a := range ix.Ancestors(ch, ctx.PHP) {
				if a.Kind == syntax.KindTrait {
					for name := range a.Methods {
						out[name] = true
					}
				}
				for _, ad := range a.TraitAdaptations {
					if ad.Alias != "" {
						out[strings.ToLower(ad.Alias)] = true
					}
				}
			}
		}
	}
	return out
}

// isThis reports whether e is `$this`, possibly parenthesized.
func isThis(e syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return ok && v.Name == "this"
}

// mayBindThis reports whether call can run with a `$this` that is an
// instance of the class declaring the called method: it sits in a
// non-static method or property hook (directly or through non-static
// closures and arrow functions) of a class related to that class. PHP then
// passes that `$this` to a `::` call of an instance method.
func mayBindThis(ctx *analysis.Context, call syntax.Node, declaring string) bool {
	for p := call.Parent(); p != nil; p = p.Parent() {
		switch f := p.(type) {
		case *syntax.Function:
			return false
		case *syntax.Closure:
			if f.Static {
				return false
			}
		case *syntax.ArrowFunction:
			if f.Static {
				return false
			}
		case *syntax.Method:
			return !f.Modifiers.Has(syntax.TStatic) && related(ctx, syntax.EnclosingClass(f), declaring)
		case *syntax.PropertyHook:
			return related(ctx, syntax.EnclosingClass(f), declaring)
		}
	}
	return false
}

// related reports whether an instance of owner (or of a class using or
// extending it) can be an instance of the class declaring the method:
// one is a subtype of the other. Traits and anonymous classes on either
// side are not tracked, so they count as related.
func related(ctx *analysis.Context, owner *syntax.ClassLike, declaring string) bool {
	ix := ctx.Index()
	fqn := ctx.Types().ClassFQN(owner)
	d := ix.Class(declaring, ctx.PHP)
	if fqn == "" || owner.ClassKind == syntax.KindTrait || d == nil || d.Kind == syntax.KindTrait {
		return true
	}
	return ix.IsSubtype(fqn, declaring, ctx.PHP) || ix.IsSubtype(declaring, fqn, ctx.PHP)
}
