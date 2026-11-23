// Package generatorrewindafteradvance implements the native
// GeneratorRewindAfterAdvance inspection.
package generatorrewindafteradvance

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const message = "Create another generator instead of rewinding it."

type rule struct{}

// New returns the GeneratorRewindAfterAdvance inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GeneratorRewindAfterAdvance" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || ctx.PHP.Below(phpversion.PHP55) || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "rewind") {
		return
	}
	origin, ok := semanticquery.NativeLocalValue(ctx, c.Var).(*syntax.FuncCall)
	if !ok {
		return
	}
	_, body := semanticquery.NativeCallback(ctx, origin.Name)
	if body == nil {
		body = semanticquery.NativeFunctionBody(ctx, origin)
	}
	block, ok := body.(*syntax.Block)
	if !ok {
		return
	}
	yields := 0
	for _, stmt := range block.Stmts {
		expr, ok := stmt.(*syntax.ExprStmt)
		if !ok {
			return
		}
		if _, ok := expr.Expr.(*syntax.Yield); !ok {
			return
		}
		yields++
	}
	if yields < 2 {
		return
	}
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		prior, ok := event.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeDominates(prior, c) || !semanticquery.SameStableLocal(ctx, prior.Var, c.Var) {
			continue
		}
		name, ok := prior.Name.(*syntax.Identifier)
		if ok && (strings.EqualFold(name.Value, "next") || strings.EqualFold(name.Value, "send")) {
			method := ctx.Index().FindMethod("Generator", name.Value, ctx.PHP)
			if method != nil && method.Builtin {
				ctx.ReportNode(c, message)
				return
			}
		}
	}
}
