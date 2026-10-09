package accessmodifierpresented

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// accessModifierPresented reports class members relying on implicit public
// visibility (methods, properties and, from PHP 7.1, constants).
type accessModifierPresented struct{}

func (accessModifierPresented) ID() string { return "AccessModifierPresented" }
func (accessModifierPresented) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClassLike}
}

func (r accessModifierPresented) Check(ctx *analysis.Context, n syntax.Node) {
	cls := n.(*syntax.ClassLike)
	if cls.ClassKind == syntax.KindInterface && !ctx.Bool("ANALYZE_INTERFACES") { // D4
		return
	}
	consts := ctx.Bool("ANALYZE_CONSTANTS") && ctx.PHP >= phpversion.PHP71
	for _, m := range cls.Members {
		switch m := m.(type) {
		case *syntax.Method: // D1
			if m.Name == nil || m.Name.Span().Len() == 0 {
				continue
			}
			if m.Modifiers.Has(syntax.TPrivate) || m.Modifiers.Has(syntax.TProtected) || m.Modifiers.Has(syntax.TPublic) {
				continue
			}
			r.report(ctx, m.Name.Span(), m.Name.Value, r.modifierFix(ctx, m.Modifiers, m.Span(), m.Name.Span().Start, syntax.TFunction))
		case *syntax.Property: // D2
			if ampHasVisibility(m.Modifiers) {
				continue
			}
			var limit uint32
			if len(m.Props) > 0 && m.Props[0].Var != nil {
				limit = m.Props[0].Var.Span().Start
			}
			if m.Type != nil {
				limit = m.Type.Span().Start
			}
			fix := r.modifierFix(ctx, m.Modifiers, m.Span(), limit, 0)
			for _, it := range m.Props {
				if it.Var == nil || it.Var.Span().Len() == 0 || it.Var.Name == "" {
					continue
				}
				r.report(ctx, it.Var.Span(), it.Var.Name, fix)
			}
		case *syntax.ClassConst: // D3
			if !consts || ampHasVisibility(m.Modifiers) || len(m.Consts) == 0 {
				continue
			}
			name := m.Consts[0].Name
			if name == nil || name.Span().Len() == 0 {
				continue
			}
			r.report(ctx, name.Span(), name.Value, r.modifierFix(ctx, m.Modifiers, m.Span(), name.Span().Start, syntax.TConst))
		}
	}
}

func ampHasVisibility(m syntax.Modifiers) bool {
	for _, t := range m {
		switch t.Kind {
		case syntax.TPublic, syntax.TProtected, syntax.TPrivate,
			syntax.TPublicSet, syntax.TProtectedSet, syntax.TPrivateSet:
			return true
		}
	}
	return false
}

// modifierFix rewrites the modifier list canonically (F1). With an empty
// list, `public ` is inserted before the kw token (or at limit when kw is 0).
func (accessModifierPresented) modifierFix(ctx *analysis.Context, mods syntax.Modifiers, decl syntax.Span, limit uint32, kw syntax.TokenKind) diagnostic.Fix {
	f := ctx.File
	return diagnostic.Fix{
		Title: "Declare 'public'",
		Edits: func() []diagnostic.TextEdit {
			if len(mods) == 0 {
				at := limit
				if kw != 0 {
					// The parser builds a method/constant only after reading
					// its keyword, which follows the attributes and modifiers.
					t, _ := astquery.FindToken(f, syntax.Span{Start: decl.Start, End: limit}, kw)
					at = t.Start
				}
				return []diagnostic.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: "public "}}
			}
			parts := make([]string, 0, 5)
			if mods.Has(syntax.TFinal) {
				parts = append(parts, "final")
			}
			if mods.Has(syntax.TAbstract) {
				parts = append(parts, "abstract")
			}
			parts = append(parts, "public")
			if mods.Has(syntax.TStatic) {
				parts = append(parts, "static")
			}
			if mods.Has(syntax.TReadonly) {
				parts = append(parts, "readonly")
			}
			span := syntax.Span{Start: mods[0].Span.Start, End: mods[len(mods)-1].Span.End}
			return []diagnostic.TextEdit{{Span: span, NewText: strings.Join(parts, " ")}}
		},
	}
}

func (accessModifierPresented) report(ctx *analysis.Context, span syntax.Span, name string, fix diagnostic.Fix) {
	ctx.Report(span, "Declare the visibility of '"+name+"' explicitly.", fix)
}
