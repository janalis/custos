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
		// Any rewrite of the watch variable in the loop condition or body, before
		// or after the selection, runs between two consecutive selections.
		restored := false
		syntax.Inspect(loop, func(node syntax.Node) bool {
			switch x := node.(type) {
			case *syntax.Assign:
				restored = restored || writes(ctx, x.Var, arg)
			case *syntax.Foreach:
				restored = restored || writes(ctx, x.Key, arg) || writes(ctx, x.Value, arg)
			case *syntax.Arg:
				// Only a proven builtin is known not to refill the array it receives.
				if semanticquery.ExpansionCSame(ctx, x.Value, arg) {
					call, ok := x.Parent().Parent().(*syntax.FuncCall)
					restored = restored || !ok || semanticquery.NativeBuiltinName(ctx, call) == ""
				}
			}
			return !restored
		})
		if !restored {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// writes reports whether assigning to target rebinds watch, directly or as a
// slot of a destructuring pattern.
func writes(ctx *analysis.Context, target, watch syntax.Expr) bool {
	var items []*syntax.ArrayItem
	switch x := target.(type) {
	case *syntax.Array:
		items = x.Items
	case *syntax.List:
		items = x.Items
	default:
		return semanticquery.ExpansionCSame(ctx, target, watch)
	}
	for _, item := range items {
		if item != nil && writes(ctx, item.Value, watch) {
			return true
		}
	}
	return false
}
