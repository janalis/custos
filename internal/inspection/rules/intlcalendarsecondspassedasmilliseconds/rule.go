// Package intlcalendarsecondspassedasmilliseconds implements the native IntlCalendarSecondsPassedAsMilliseconds inspection.
package intlcalendarsecondspassedasmilliseconds

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Convert seconds to calendar milliseconds."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntlCalendarSecondsPassedAsMilliseconds" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "IntlCalendar", "setTime") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "timestamp")
	origin, ok := semanticquery.NativeValue(ctx, arg).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, origin, "time") {
		return
	}
	direct, ok := syntax.UnwrapParens(arg).(*syntax.FuncCall)
	if !ok {
		if !rebound(ctx, c, arg) {
			ctx.ReportNode(c, message)
		}
		return
	}
	// Edit the call only: an argument that is already parenthesized keeps its
	// parentheses and the comments written between them and the call.
	repl := ctx.Text(direct) + " * 1000"
	if arg == syntax.Expr(direct) {
		repl = "(" + repl + ")"
	}
	ctx.ReportNode(c, message, astquery.ReplaceFix(direct.Span(), repl))
}

// rebound reports whether a timestamp read from a variable may no longer hold
// the assigned time() result: the variable, or one whose value was copied into
// it, is captured by reference in a closure `use` or bound by a `catch`
// clause. File-level variables are also globals, so a `global` declaration or
// any `$GLOBALS` access in the file counts too.
func rebound(ctx *analysis.Context, call *syntax.MethodCall, timestamp syntax.Expr) bool {
	variable := syntax.UnwrapParens(timestamp).(*syntax.Variable)
	scope := syntax.EnclosingVariableScope(call)
	hidden := map[string]bool{}    // names written outside plain assignments
	feeds := map[string][]string{} // assigned name -> variables read by the values
	visit := func(n syntax.Node) bool {
		if n == scope {
			return true
		}
		switch n := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.ArrowFunction:
			return false
		case *syntax.Closure:
			for _, use := range n.Uses {
				if use.ByRef && use.Var != nil {
					hidden[use.Var.Name] = true
				}
			}
			return false
		case *syntax.Catch:
			if n.Var != nil {
				hidden[n.Var.Name] = true
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
