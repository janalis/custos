// Package curlreadcallbackexceedsrequestedsize implements CurlReadCallbackExceedsRequestedSize.
package curlreadcallbackexceedsrequestedsize

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return no more than the requested byte count."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlReadCallbackExceedsRequestedSize" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := semanticquery.CurlSetting(ctx, c, "CURLOPT_READFUNCTION")
	if value == nil {
		return
	}
	params, body := semanticquery.NativeCallback(ctx, value)
	if len(params) < 3 {
		return
	}
	returns, complete := semanticquery.NativeReturns(body)
	if !complete {
		return
	}
	for _, ret := range returns {
		r, ok := ret.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, r, "str_repeat") {
			return
		}
		text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(r.Args, 0, "string"))
		if !known || len(text) != 1 {
			return
		}
		sum, ok := syntax.UnwrapParens(semanticquery.CallArgument(r.Args, 1, "times")).(*syntax.Binary)
		if !ok || sum.Op.Kind != syntax.TPlus {
			return
		}
		size, ok := syntax.UnwrapParens(sum.Left).(*syntax.Variable)
		extra, known := semanticquery.NativeContractInt(ctx, sum.Right)
		if !ok || size.Name != params[2].Var.Name || !known || extra <= 0 {
			return
		}
	}
	// Any possible write to the requested-size parameter, wherever it sits in
	// the callback (a loop may run it before a return), invalidates the proof.
	if expansionDCurlSizeRebound(ctx, body, params[2].Var.Name) {
		return
	}
	ctx.ReportNode(c, message)
}

// expansionDCurlSizeRebound reports whether the callback body may give its
// local variable $name another value: direct and destructuring assignments,
// ++/--, foreach targets, global/static/catch/unset, references taken by
// assignment, array item or closure capture, the variable passed as a call
// argument (the parameter may be by-reference), and constructs that write
// the local scope by name (variable-variables, extract, include, eval).
func expansionDCurlSizeRebound(ctx *analysis.Context, body syntax.Node, name string) bool {
	rebound := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if rebound {
			return false
		}
		if syntax.IsVariableScope(n) {
			if closure, ok := n.(*syntax.Closure); ok {
				for _, use := range closure.Uses {
					rebound = rebound || use.ByRef && use.Var.Name == name
				}
			}
			return false
		}
		switch x := n.(type) {
		case *syntax.Variable:
			rebound = x.NameExpr != nil
		case *syntax.Assign:
			rebound = astquery.MentionsVariable(x.Var, name) || x.ByRef && astquery.MentionsVariable(x.Value, name)
		case *syntax.IncDec:
			rebound = astquery.MentionsVariable(x.Var, name)
		case *syntax.Foreach:
			rebound = astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name)
		case *syntax.Global, *syntax.StaticVar, *syntax.Unset:
			rebound = astquery.MentionsVariable(x, name)
		case *syntax.Catch:
			rebound = x.Var != nil && x.Var.Name == name
		case *syntax.ArrayItem:
			rebound = x.ByRef && astquery.MentionsVariable(x.Value, name)
		case *syntax.Arg:
			v, ok := syntax.UnwrapParens(x.Value).(*syntax.Variable)
			rebound = ok && v.Name == name
		case *syntax.Include, *syntax.Eval:
			rebound = true
		case *syntax.FuncCall:
			switch ctx.GlobalFunctionName(x) {
			case "extract", "parse_str", "mb_parse_str", "import_request_variables":
				rebound = true
			}
		}
		return !rebound
	})
	return rebound
}
