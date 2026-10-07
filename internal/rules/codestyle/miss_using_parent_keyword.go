package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// missUsingParentKeyword flags `parent::m()` calls to an inherited method
// other than the current one that nothing in the hierarchy overrides.
type missUsingParentKeyword struct{}

func init() { register(missUsingParentKeyword{}) }

func (missUsingParentKeyword) ID() string { return "MissUsingParentKeyword" }

func (missUsingParentKeyword) Semantic() {}

func (missUsingParentKeyword) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KStaticCall}
}

func (missUsingParentKeyword) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.StaticCall)
	cls, ok := call.Class.(*syntax.Name)
	if !ok || !strings.EqualFold(cls.Value, "parent") || call.Args == nil { // D1 / E6
		return
	}
	id, ok := call.Name.(*syntax.Identifier)
	if !ok || id.Value == "" {
		return
	}
	meth, ok := syntax.EnclosingFuncLike(call).(*syntax.Method) // D2
	if !ok || meth.Name == nil || meth.Modifiers.Has(syntax.TStatic) {
		return
	}
	owner, ok := meth.Parent().(*syntax.ClassLike)
	if !ok || owner.ClassKind == syntax.KindTrait || owner.ClassKind == syntax.KindInterface || len(owner.Extends) == 0 {
		return
	}
	if strings.EqualFold(id.Value, meth.Name.Value) { // D3
		return
	}
	if mupkDeclaresMethod(ctx, owner, id.Value) { // D4
		return
	}
	ownerFQN := ctx.Types().ClassFQN(owner)
	ix := ctx.Index()
	if ownerFQN != "" && !owner.Modifiers.Has(syntax.TFinal) { // D5
		lname := strings.ToLower(id.Value)
		for _, sub := range ix.Subclasses(ownerFQN) {
			if c := ix.Class(sub, ctx.PHP); c != nil {
				if _, ok := c.Methods[lname]; ok {
					return
				}
			}
		}
	}
	parent := ctx.Names().Class(owner.Extends[0].Value, owner.Span().Start)
	m := ix.FindMethod(parent, id.Value, ctx.PHP) // D6
	if m == nil {
		return
	}
	argsSpan := call.Args.Span()
	args := ""
	if argsSpan.Len() >= 2 {
		args = ctx.SpanText(syntax.Span{Start: argsSpan.Start + 1, End: argsSpan.End - 1})
	}
	repl := "$this->" + id.Value + "(" + args + ")"
	if m.Static {
		repl = "self::" + id.Value + "(" + args + ")"
	}
	span := call.Span()
	ctx.Report(span, "Call it as '"+repl+"' instead of through 'parent::'.", analysis.Fix{
		Title: "Drop parent::",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}

// mupkDeclaresMethod reports whether the class declares a method named name
// itself, or imports it from one of its traits (spec Divergences).
func mupkDeclaresMethod(ctx *analysis.Context, owner *syntax.ClassLike, name string) bool {
	for _, mem := range owner.Members {
		switch mem := mem.(type) {
		case *syntax.Method:
			if mem.Name != nil && strings.EqualFold(mem.Name.Value, name) {
				return true
			}
		case *syntax.TraitUse:
			for _, t := range mem.Traits {
				fqn := ctx.Names().Class(t.Value, t.Span().Start)
				if ctx.Index().FindMethod(fqn, name, ctx.PHP) != nil {
					return true
				}
			}
			for _, ad := range mem.Adaptations {
				if ad.Alias != nil && strings.EqualFold(ad.Alias.Value, name) {
					return true
				}
			}
		}
	}
	return false
}
