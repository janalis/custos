package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// dynamicCallsToScopeIntrospection reports indirect invocations (variable
// callee or callback string) of functions that touch the caller's scope.
type dynamicCallsToScopeIntrospection struct{}

func init() { register(dynamicCallsToScopeIntrospection{}) }

func (dynamicCallsToScopeIntrospection) ID() string { return "DynamicCallsToScopeIntrospection" }

func (dynamicCallsToScopeIntrospection) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

// scopeSensitive is set S of the spec.
var scopeSensitive = map[string]bool{
	"compact": true, "extract": true, "func_get_args": true, "func_get_arg": true,
	"func_num_args": true, "get_defined_vars": true, "mb_parse_str": true, "parse_str": true,
}

// callbackIndex is table C of the spec.
var callbackIndex = map[string]int{
	"call_user_func": 0, "call_user_func_array": 0, "array_map": 0,
	"array_filter": 1, "array_reduce": 1, "array_walk": 1, "array_walk_recursive": 1,
}

func (dynamicCallsToScopeIntrospection) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP71 { // E1
		return
	}
	call := n.(*syntax.FuncCall)
	var target syntax.Expr
	if _, part, ok := util.FuncNamePart(call); !ok { // D1
		target = call.Name
	} else if idx, ok := callbackIndex[strings.ToLower(part)]; ok && ctx.GlobalFunctionName(call) != "" { // D2
		if call.Args == nil || len(call.Args.Args) < idx+1 {
			return
		}
		arg, ok := call.Args.Args[idx].(*syntax.Arg)
		if !ok || arg.Unpack {
			return
		}
		target = arg.Value
	} else {
		return
	}
	if target == nil || target.Span().Start == target.Span().End {
		return
	}
	var lit *syntax.Literal
	if l, ok := syntax.UnwrapParens(target).(*syntax.Literal); ok && l.LitKind == syntax.LitString { // D3
		lit = l
	} else { // D4
		vals, known := util.PossibleValuesReaching(ctx.File, target)
		if !known {
			return
		}
		for _, v := range vals {
			l, ok := v.(*syntax.Literal)
			if !ok || l.LitKind != syntax.LitString {
				continue
			}
			if lit != nil {
				return // E6
			}
			lit = l
		}
	}
	if lit == nil {
		return
	}
	val, ok := util.StringLiteralValue(lit.Raw) // D5
	if !ok {
		return
	}
	val = strings.TrimPrefix(val, `\`)
	if !scopeSensitive[strings.ToLower(val)] {
		return
	}
	ctx.ReportNode(target, "'"+val+"' reads the caller scope and cannot be invoked indirectly since PHP 7.1.")
}
