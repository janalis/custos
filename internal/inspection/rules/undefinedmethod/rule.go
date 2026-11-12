// Package undefinedmethod implements the UndefinedMethod inspection.
package undefinedmethod

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "UndefinedMethod" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall}
}

func (rule) Semantic() {}

// Check reports a method call (`$x->m()`, `$x?->m()` or `C::m()`) whose
// receiver's classes native declarations prove, none of which declares the
// method, nor `__call` or `__callStatic`, anywhere in its hierarchy (traits,
// trait aliases and `@method` tags included, at any PHP version), unless a
// check of the method's existence guards the call.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var (
		name    *syntax.Identifier
		classes []string
		exact   bool
	)
	switch c := n.(type) {
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return
		}
		name, classes = id, semanticquery.MemberReceiverClasses(ctx, c.Var)
	default:
		s := n.(*syntax.StaticCall)
		id, ok := s.Name.(*syntax.Identifier)
		if !ok {
			return
		}
		name = id
		classes, exact = semanticquery.StaticMemberClasses(ctx, s.Class)
	}
	lname := strings.ToLower(name.Value)
	label, ok := semanticquery.MissingMember(ctx, classes, exact, func(c *index.Class) bool {
		if c.Methods[lname] != nil || c.Methods["__call"] != nil || c.Methods["__callstatic"] != nil {
			return true
		}
		for _, a := range c.TraitAdaptations {
			if strings.EqualFold(a.Alias, name.Value) {
				return true
			}
		}
		return false
	})
	if !ok || semanticquery.MemberGuarded(ctx, n.(syntax.Expr), name.Value, false) {
		return
	}
	ctx.ReportNode(name, "Fix the method name: "+label+" declares no "+name.Value+"().")
}
