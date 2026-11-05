package flow

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// observeParameter records an implicit read at the point its value is taken,
// rather than treating every occurrence elsewhere in the scope as a read.
func observeParameter(name string, value Value, scope syntax.Node, r *scopeResult) {
	if r.parameterObserved == nil {
		r.parameterObserved = map[string]uint8{}
	}
	if !value.Complete || value.replaced {
		r.parameterObserved[name] |= 1
	}
	params := syntax.VariableScopeParams(scope)
	for _, source := range value.Sources {
		if source.Kind == "parameter" && source.Parameter >= 0 && source.Parameter < len(params) && params[source.Parameter].Var != nil {
			r.parameterObserved[params[source.Parameter].Var.Name] |= 2
		}
	}
}

func (e *Env) observeParameters(at syntax.Node, f *frame, r *scopeResult) {
	scope := syntax.EnclosingVariableScope(at)
	for _, param := range syntax.VariableScopeParams(scope) {
		r.operations++
		if r.operations > MaxTransfers {
			r.complete = false
			r.gap("flow-budget", at.Span())
			return
		}
		if param.Var != nil {
			value := Value{}
			if f != nil {
				value = f.vars[param.Var.Name]
			}
			observeParameter(param.Var.Name, value, scope, r)
		}
	}
}

func (e *Env) observeIntrospection(at syntax.Node, name string, args *syntax.ArgList, f *frame, r *scopeResult) {
	switch name {
	case "func_get_args", "func_get_arg", "func_num_args", "get_defined_vars", "debug_backtrace", "debug_print_backtrace", "gettrace", "gettraceasstring":
		r.argumentSensitive = true
	}
	switch name {
	case "compact", "get_defined_vars", "func_get_args", "func_get_arg":
	default:
		return
	}
	// Older func_get_arg(s) may expose the original argument rather than its
	// current binding. Keep those entry values possible on old PHP targets.
	if e.types.PHP < phpversion.PHP70 && (name == "func_get_arg" || name == "func_get_args") {
		e.observeParameters(at, nil, r)
		return
	}
	if name == "func_get_arg" {
		if call, ok := at.(*syntax.FuncCall); ok {
			bound, known := e.types.Native().BuiltinArguments(call)
			if known && len(bound) > 0 {
				if position, known := literalInteger(r.values.at(bound[0]).Expr); known {
					params := syntax.VariableScopeParams(syntax.EnclosingVariableScope(at))
					if position >= 0 && position < int64(len(params)) && params[position].Var != nil {
						param := params[position]
						observeParameter(param.Var.Name, f.vars[param.Var.Name], syntax.EnclosingVariableScope(at), r)
					}
					return
				}
			}
		}
	}
	if name == "compact" && args != nil {
		for _, node := range args.Args {
			arg, ok := node.(*syntax.Arg)
			if !ok || arg.Unpack {
				e.observeParameters(at, f, r)
				return
			}
			text, known := literalText(r.values.at(arg.Value).Expr)
			if !known {
				e.observeParameters(at, f, r)
				return
			}
			observeParameter(text, f.vars[text], syntax.EnclosingVariableScope(at), r)
		}
		return
	}
	e.observeParameters(at, f, r)
}

type captureNames struct {
	names    []string
	complete bool
}

func (e *Env) observeArrow(arrow *syntax.ArrowFunction, f *frame, r *scopeResult) {
	if e.arrowNames == nil {
		e.arrowNames = map[*syntax.ArrowFunction]captureNames{}
	}
	captures, cached := e.arrowNames[arrow]
	if !cached {
		captures = arrowCaptureNames(arrow)
		e.arrowNames[arrow] = captures
	}
	if !captures.complete {
		e.observeParameters(arrow, f, r)
		return
	}
	for _, name := range captures.names {
		observeParameter(name, f.vars[name], syntax.EnclosingVariableScope(arrow), r)
	}
}

func arrowCaptureNames(arrow *syntax.ArrowFunction) captureNames {
	bound, seen := map[string]int{}, map[string]bool{}
	out := captureNames{complete: true}
	work := 0
	var visit func(syntax.Node)
	visit = func(node syntax.Node) {
		if node == nil || !out.complete {
			return
		}
		work++
		if work > MaxTransfers {
			out.complete = false
			return
		}
		switch n := node.(type) {
		case *syntax.Function, *syntax.ClassLike:
			if class, ok := node.(*syntax.ClassLike); ok && class.Args != nil {
				visit(class.Args)
			}
			return
		case *syntax.Closure:
			for _, use := range n.Uses {
				if use.Var == nil {
					out.complete = false
					return
				}
				visit(use.Var)
			}
			return
		case *syntax.ArrowFunction:
			for _, p := range n.Params {
				if p.Var != nil {
					bound[p.Var.Name]++
				}
			}
			visit(n.Expr)
			for _, p := range n.Params {
				if p.Var != nil {
					bound[p.Var.Name]--
				}
			}
			return
		case *syntax.Variable:
			if property, ok := n.Parent().(*syntax.StaticPropertyFetch); ok && property.Name == n {
				return
			}
			if n.NameExpr != nil {
				out.complete = false
				return
			}
			if n.Name != "" && n.Name != "this" && bound[n.Name] == 0 && !seen[n.Name] {
				seen[n.Name] = true
				out.names = append(out.names, n.Name)
			}
			return
		}
		syntax.Children(node, visit)
	}
	visit(arrow)
	return out
}
