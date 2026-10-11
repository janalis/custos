// Package explodeemptyseparator implements the ExplodeEmptySeparator inspection.
package explodeemptyseparator

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExplodeEmptySeparator" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Supply a nonempty separator."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "explode")
	if call == nil {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 0, "separator")
	separator, known := semanticquery.NativeString(ctx, arg)
	if known && separator == "" && !rebound(ctx, call, arg) {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}

// rebound reports whether a separator read from a variable may no longer hold
// the assigned empty literal: the variable, or one whose value was copied into
// it, is changed by `++`/`--`, bound by reference in a closure `use` or an
// array item, declared `global`, or the scope imports variables through
// extract() or single-argument parse_str(). File-level variables are also
// globals, so a `global` declaration anywhere in the file or any `$GLOBALS`
// access counts too.
func rebound(ctx *analysis.Context, call *syntax.FuncCall, separator syntax.Expr) bool {
	var queue []string
	syntax.Inspect(separator, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok {
			queue = append(queue, v.Name)
		}
		return true
	})
	if len(queue) == 0 {
		return false
	}
	scope := syntax.EnclosingVariableScope(call)
	imported := false              // any local may be written behind the flow's back
	hidden := map[string]bool{}    // names written outside plain assignments
	feeds := map[string][]string{} // assigned name -> variables read by the values
	globals := func(n *syntax.Global) {
		for _, x := range n.Vars {
			if v, ok := x.(*syntax.Variable); ok {
				hidden[v.Name] = true
			}
		}
	}
	visit := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.ArrowFunction:
			return n == scope
		case *syntax.Closure:
			for _, use := range n.Uses {
				if use.ByRef && use.Var != nil {
					hidden[use.Var.Name] = true
				}
			}
			return n == scope
		case *syntax.Global:
			globals(n)
		case *syntax.IncDec:
			if target, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable); ok {
				hidden[target.Name] = true
			}
		case *syntax.ArrayItem:
			if target, ok := n.Value.(*syntax.Variable); ok && n.ByRef {
				hidden[target.Name] = true
			}
		case *syntax.FuncCall:
			switch ctx.GlobalFunctionName(n) {
			case "extract":
				imported = true
			case "parse_str":
				imported = imported || len(n.Args.Args) == 1
			}
		case *syntax.Assign:
			if target, ok := n.Var.(*syntax.Variable); ok {
				syntax.Inspect(n.Value, func(x syntax.Node) bool {
					if read, ok := x.(*syntax.Variable); ok {
						feeds[target.Name] = append(feeds[target.Name], read.Name)
					}
					return true
				})
			}
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.Global:
				globals(n)
			case *syntax.Variable:
				imported = imported || n.Name == "GLOBALS"
			}
			return true
		})
	}
	if imported {
		return true
	}
	seen := map[string]bool{}
	for ; len(queue) > 0; queue = queue[1:] {
		name := queue[0]
		if hidden[name] {
			return true
		}
		if !seen[name] {
			seen[name] = true
			queue = append(queue, feeds[name]...)
		}
	}
	return false
}
