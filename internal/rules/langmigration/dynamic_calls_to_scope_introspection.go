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
	if l, ok := util.UnwrapParens(target).(*syntax.Literal); ok && l.LitKind == syntax.LitString { // D3
		lit = l
	} else { // D4
		vals, known := dcsiValues(ctx.File, target)
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

// dcsiValues is the possible-values procedure of the spec (R1-R7) where
// variables (R3) only take the assignments that reach the use.
func dcsiValues(f *syntax.File, e syntax.Expr) ([]syntax.Expr, bool) {
	var out []syntax.Expr
	seen := map[syntax.Node]bool{}
	known := true
	var collect func(e syntax.Expr)
	collect = func(e syntax.Expr) {
		e = util.UnwrapParens(e)
		if e == nil || seen[e] || !known {
			return
		}
		seen[e] = true
		switch x := e.(type) {
		case *syntax.Ternary: // R1
			if x.Then != nil {
				collect(x.Then)
			}
			collect(x.Else)
		case *syntax.Binary: // R2
			if x.Op.Kind != syntax.TCoalesce {
				out = append(out, e)
				return
			}
			collect(x.Left)
			collect(x.Right)
		case *syntax.Variable: // R3
			if x.NameExpr != nil || x.Name == "" {
				return
			}
			scope := util.EnclosingFuncLike(x)
			if scope == nil {
				return
			}
			params, body := dcsiScopeParts(scope)
			if util.UnstableVariable(body, x.Name) {
				known = false
				return
			}
			defs, entry := util.ReachingAssignments(scope, x, x.Name)
			if entry {
				for _, p := range params {
					if p.Var != nil && p.Var.Name == x.Name && p.Default != nil {
						collect(p.Default)
					}
				}
			}
			for _, d := range defs {
				v := d.Value
				for {
					inner, ok := util.UnwrapParens(v).(*syntax.Assign)
					if !ok || inner.Op.Kind != syntax.TEqual {
						break
					}
					v = inner.Value
				}
				collect(v)
			}
		default: // R4-R7
			vals, ok := util.PossibleValuesKnown(f, e)
			if !ok {
				known = false
				return
			}
			out = append(out, vals...)
		}
	}
	collect(e)
	if !known {
		return nil, false
	}
	return out, true
}

func dcsiScopeParts(scope syntax.Node) ([]*syntax.Param, syntax.Node) {
	switch s := scope.(type) {
	case *syntax.Function:
		return s.Params, s.Body
	case *syntax.Method:
		if s.Body != nil {
			return s.Params, s.Body
		}
		return s.Params, nil
	case *syntax.Closure:
		return s.Params, s.Body
	case *syntax.ArrowFunction:
		return s.Params, s.Expr
	}
	return nil, nil
}
