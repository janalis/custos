// Package selectwatcharraysnotrestored implements the native SelectWatchArraysNotRestored inspection.
package selectwatcharraysnotrestored

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Restore stream watch arrays before each selection."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SelectWatchArraysNotRestored" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "stream_select") {
		return
	}
	var loop *syntax.While
	for p := c.Parent(); p != nil; p = p.Parent() {
		if l, ok := p.(*syntax.While); ok {
			loop = l
			break
		}
	}
	if loop == nil {
		return
	}
	for _, slot := range []struct {
		position int
		name     string
	}{{0, "read"}, {1, "write"}, {2, "except"}} {
		arg := semanticquery.CallArgument(c.Args, slot.position, slot.name)
		var value *syntax.Array
		valid, budget := true, 256
		scope := syntax.EnclosingVariableScope(c)
		visit := func(node syntax.Node) bool {
			budget--
			if budget < 0 {
				valid = false
				return false
			}
			if node != scope && syntax.IsVariableScope(node) {
				return false
			}
			if node.Span().Start >= loop.Span().Start {
				return false
			}
			if a, ok := node.(*syntax.Assign); ok && semanticquery.ExpansionCSame(ctx, a.Var, arg) {
				value = nil
				if !a.ByRef && semanticquery.NativeDominates(a, loop) {
					value, _ = a.Value.(*syntax.Array)
				}
			}
			return true
		}
		if scope != nil {
			syntax.Inspect(scope, visit)
		} else {
			syntax.InspectFile(ctx.File, visit)
		}
		if !valid {
			return
		}

		ok := value != nil
		if !ok || len(value.Items) == 0 || value.Span().Start >= loop.Span().Start {
			continue
		}
		// A pre-selection assignment in the loop restores the watched set.
		restored := false
		syntax.Inspect(loop.Body, func(node syntax.Node) bool {
			if a, ok := node.(*syntax.Assign); ok && a.Span().Start < c.Span().Start && semanticquery.ExpansionCSame(ctx, a.Var, arg) {
				restored = true
			}
			return true
		})
		if !restored {
			ctx.ReportNode(c, message)
			return
		}
	}
}
