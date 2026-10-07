package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// deprecatedConstructorStyle reports PHP 4 style constructors (a method named
// after its class).
type deprecatedConstructorStyle struct{}

func init() { register(deprecatedConstructorStyle{}) }

func (deprecatedConstructorStyle) ID() string { return "DeprecatedConstructorStyle" }

func (deprecatedConstructorStyle) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

func (deprecatedConstructorStyle) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	if m.Name == nil || m.Name.Span().Len() == 0 || m.Modifiers.Has(syntax.TStatic) { // D1
		return
	}
	cls, ok := m.Parent().(*syntax.ClassLike)
	if !ok || cls.ClassKind != syntax.KindClass || cls.Name == nil { // D2, E2, E3, E6 (enums: spec Divergences)
		return
	}
	if !strings.EqualFold(m.Name.Value, cls.Name.Value) { // D3 (PHP names are case-insensitive)
		return
	}
	for _, mem := range cls.Members { // D4 (case-insensitive: spec Divergences)
		if o, ok := mem.(*syntax.Method); ok && o.Name != nil && strings.EqualFold(o.Name.Value, "__construct") {
			return
		}
	}
	if inNamedNamespace(cls) { // spec Divergences: not a constructor there
		return
	}
	span := m.Name.Span()
	ctx.Report(span, "Class '"+cls.Name.Value+"' uses an old-style constructor; rename it to __construct.", analysis.Fix{
		Title: "Rename to __construct",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: "__construct"}} },
	})
}

func inNamedNamespace(n syntax.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if ns, ok := p.(*syntax.Namespace); ok {
			return ns.Name != nil
		}
	}
	return false
}
