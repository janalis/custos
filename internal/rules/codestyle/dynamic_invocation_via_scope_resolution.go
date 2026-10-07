package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// dynamicInvocationViaScopeResolution flags instance methods called through
// `::` (`self::run()`, `static::run()`, `$obj::run()`).
type dynamicInvocationViaScopeResolution struct{}

func init() { register(dynamicInvocationViaScopeResolution{}) }

func (dynamicInvocationViaScopeResolution) ID() string { return "DynamicInvocationViaScopeResolution" }

func (dynamicInvocationViaScopeResolution) Semantic() {}

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
	if cls == "" {
		return
	}
	ix := ctx.Index()
	m := ix.FindMethod(cls, id.Value, ctx.PHP)
	if m == nil || m.Static || m.Abstract {
		return
	}
	if decl := ix.Class(m.Class, ctx.PHP); decl == nil || decl.Kind == syntax.KindInterface {
		return
	}
	classSpan := call.Class.Span()
	colons, ok := util.FindToken(ctx.File, syntax.Span{Start: classSpan.End, End: id.Span().Start}, syntax.TPaamayimNekudotayim)
	if !ok {
		return
	}
	colonsSpan := syntax.Span{Start: colons.Start, End: colons.End}
	instanceMsg := "Call '" + id.Value + "' on an instance with '->' instead of '::'."
	left := ctx.SpanText(classSpan)
	if strings.EqualFold(left, "static") || strings.EqualFold(left, "self") || strings.EqualFold(left, util.LastNamePart(m.Class)) { // D3
		meth, ok := syntax.EnclosingFuncLike(call).(*syntax.Method)
		if !ok || meth.Name == nil || strings.EqualFold(meth.Name.Value, id.Value) {
			return
		}
		if owner, ok := meth.Parent().(*syntax.ClassLike); ok {
			if fqn := ctx.Types().ClassFQN(owner); fqn != "" {
				if ix.FindMethod(fqn, id.Value, ctx.PHP) == nil {
					return
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
		ctx.Report(call.Span(), "Call '"+id.Value+"' with '$this->' instead of '::'.", analysis.Fix{ // D3a
			Title: "Use $this->",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{{Span: classSpan, NewText: "$this"}, {Span: colonsSpan, NewText: "->"}}
			},
		})
		return
	}
	switch call.Class.(type) { // D4
	case *syntax.Name, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		return
	}
	ctx.Report(call.Span(), instanceMsg, analysis.Fix{
		Title: "Use ->",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: colonsSpan, NewText: "->"}} },
	})
}
