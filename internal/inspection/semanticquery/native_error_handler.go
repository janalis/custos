package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// nativeJoinedWarningUnknown declines newly retained producer paths across
// explicit error-handler changes. Such callbacks can turn a warning into an
// exception; this bounded lexical scan does not model their execution. It
// does not establish that an opening succeeded or change single-value facts.
func nativeJoinedWarningUnknown(ctx *analysis.Context, at *syntax.FuncCall) bool {
	scope := syntax.EnclosingVariableScope(at)
	unknown := false
	stack := []bool{}
	budget := 2048
	seen := map[string]bool{}
	scalar := map[string]bool{}
	registered := false
	cyclic := NativeLabelled(ctx, at)
	for parent := at.Parent(); parent != nil && parent != scope; parent = parent.Parent() {
		switch parent.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			cyclic = true
		}
	}
	visit := func(n syntax.Node) bool {
		budget--
		if budget < 0 || !cyclic && n.Span().Start >= at.Span().Start {
			return false
		}
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		if cyclic {
			if call, ok := n.(*syntax.FuncCall); ok && NativeBuiltin(ctx, call, "set_error_handler") {
				unknown = true
			}
			return true
		}
		if branch, ok := n.Parent().(*syntax.If); ok && branch.Span().Contains(at.Span()) && n != branch.Cond && !n.Span().Contains(at.Span()) {
			return false
		}
		switch node := n.(type) {
		case *syntax.Variable:
			seen[node.Name] = true
		case *syntax.Assign:
			if variable, ok := node.Var.(*syntax.Variable); ok {
				safe := variable.NameExpr == nil && (!seen[variable.Name] || scalar[variable.Name])
				scalar[variable.Name] = ctx.TypeOf(node.Value).OnlyOf("string", "int", "float", "bool", "null", "resource")
				if registered && !safe {
					unknown = true
					stack = nil
				}
			} else if registered {
				unknown = true
				stack = nil
			}
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.PropertyFetch, *syntax.ArrayDimFetch, *syntax.Include, *syntax.Eval, *syntax.Unset, *syntax.Yield, *syntax.YieldFrom, *syntax.IncDec:
			if registered {
				unknown = true
				stack = nil
			}
		case *syntax.Binary:
			if registered && node.Op.Kind == syntax.TDot && (!ctx.TypeOf(node.Left).OnlyOf("string", "int", "float", "bool", "null") || !ctx.TypeOf(node.Right).OnlyOf("string", "int", "float", "bool", "null")) {
				unknown = true
				stack = nil
			}
		case *syntax.Unary:
			if registered && !ctx.TypeOf(node.Expr).OnlyOf("string", "int", "float", "bool", "null") {
				unknown = true
				stack = nil
			}
		}
		call, ok := n.(*syntax.FuncCall)
		if !ok || call.Span().End > at.Span().Start {
			return true
		}
		switch NativeBuiltinName(ctx, call) {
		case "set_error_handler":
			registered = true
			if !NativeDominates(call, at) || NativeLabelled(ctx, at) {
				unknown = true
				stack = nil
				return true
			}
			stack = append(stack, unknown)
			unknown = !nativeHandlerReturnsFalse(ctx, call)
		case "restore_error_handler":
			if NativeDominates(call, at) && len(stack) > 0 {
				unknown = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
		default:
			if registered {
				switch NativeBuiltinName(ctx, call) {
				case "random_bytes", "base64_encode", "str_replace", "str_contains":
					for _, argument := range call.Args.Args {
						arg, ok := argument.(*syntax.Arg)
						if !ok || !ctx.TypeOf(arg.Value).OnlyOf("string", "int", "float", "bool", "null") {
							unknown = true
							stack = nil
						}
					}
				default:
					unknown = true
					stack = nil
				}
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	if budget < 0 {
		ctx.Gap("query-budget", at.Span())
	}
	return budget < 0 || unknown
}

func nativeHandlerReturnsFalse(ctx *analysis.Context, call *syntax.FuncCall) bool {
	if _, bound := ctx.Flow().BoundArguments(call); !bound {
		return false
	}
	var params []*syntax.Param
	var result syntax.Expr
	switch handler := syntax.UnwrapParens(CallArgument(call.Args, 0, "callback")).(type) {
	case *syntax.ArrowFunction:
		params = handler.Params
		result = handler.Expr
	case *syntax.Closure:
		params = handler.Params
		if len(handler.Body.Stmts) != 1 {
			return false
		}
		statement, ok := handler.Body.Stmts[0].(*syntax.Return)
		if !ok {
			return false
		}
		result = statement.Expr
	default:
		return false
	}
	if len(params) > 4 {
		return false
	}
	for _, param := range params {
		if param.Type != nil {
			return false
		}
	}
	value, known := syntax.UnwrapParens(result).(*syntax.ConstFetch)
	return known && value.Name != nil && strings.EqualFold(strings.TrimPrefix(value.Name.Value, `\`), "false")
}
