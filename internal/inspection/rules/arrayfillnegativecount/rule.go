// Package arrayfillnegativecount implements the native ArrayFillNegativeCount inspection.
package arrayfillnegativecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonnegative array fill count."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayFillNegativeCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "array_fill") {
		return
	}
	count := semanticquery.CallArgument(c.Args, 1, "count")
	v, k := semanticquery.NativeInt(ctx, count)
	if k && v < 0 && !rebound(ctx, c, count) {
		ctx.ReportNode(c, message)
	}
}

// rebound reports whether a count read from a variable may no longer hold the
// assigned literal: the variable, or one whose value was copied into it, is
// changed by `++`/`--`, bound by reference in a closure `use` or an array
// item, or the scope calls extract(). File-level variables are also globals,
// so a `global` declaration or any `$GLOBALS` access in the file counts too.
func rebound(ctx *analysis.Context, call *syntax.FuncCall, count syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(count).(*syntax.Variable)
	if !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(call)
	hidden := map[string]bool{}    // names written outside plain assignments
	feeds := map[string][]string{} // assigned name -> variables read by the values
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
		case *syntax.IncDec:
			if target, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable); ok {
				hidden[target.Name] = true
			}
		case *syntax.ArrayItem:
			if target, ok := n.Value.(*syntax.Variable); ok && n.ByRef {
				hidden[target.Name] = true
			}
		case *syntax.FuncCall:
			if ctx.GlobalFunctionName(n) == "extract" {
				hidden[variable.Name] = true
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
				for _, x := range n.Vars {
					if global, ok := x.(*syntax.Variable); ok {
						hidden[global.Name] = true
					}
				}
			case *syntax.Variable:
				if n.Name == "GLOBALS" {
					hidden[variable.Name] = true
				}
			}
			return true
		})
	}
	seen := map[string]bool{variable.Name: true}
	for queue := []string{variable.Name}; len(queue) > 0; queue = queue[1:] {
		if hidden[queue[0]] {
			return true
		}
		for _, name := range feeds[queue[0]] {
			if !seen[name] {
				seen[name] = true
				queue = append(queue, name)
			}
		}
	}
	return false
}
