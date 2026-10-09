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
	if m.Visibility == index.Private || m.Final {
		return true
	}
	ix := ctx.Index()
	c := ix.Class(fqn, ctx.PHP)
	if c != nil && (c.Final || c.Kind == syntax.KindEnum) {
		return true
	}
	if c != nil && c.Kind == syntax.KindTrait { // self is the using class
		return false
	}
	below := ctx.Memo("descendant-methods\x00"+strings.ToLower(fqn), func() any {
		return semanticquery.DescendantMethods(ix, fqn, ctx.PHP)
	}).(map[string]bool)
	return !below[strings.ToLower(m.Name)]
}
