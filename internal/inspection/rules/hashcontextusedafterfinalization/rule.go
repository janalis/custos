// Package hashcontextusedafterfinalization implements HashContextUsedAfterFinalization.
package hashcontextusedafterfinalization

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Create a new hash context after finalization."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HashContextUsedAfterFinalization" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if finalizedBefore(ctx, n.(*syntax.FuncCall)) {
		semanticquery.CheckHashContextUsedAfterFinalization(ctx, n, message)
	}
}

// finalizedBefore requires an earlier hash_final that always runs before c
// and is not followed by a rebinding of the context variable. It only vetoes:
// the shared check still has to prove that both calls use the same value.
func finalizedBefore(ctx *analysis.Context, c *syntax.FuncCall) bool {
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "hash_update", "hash_copy", "hash_final":
	default:
		return false
	}
	name := ""
	if v, ok := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 0, "context")).(*syntax.Variable); ok {
		name = v.Name
	}
	scope := syntax.EnclosingVariableScope(c)
	// A by-reference parameter, global or static variable can be replaced by
	// any call whose body is not known.
	shared := false
	for _, p := range syntax.VariableScopeParams(scope) {
		shared = shared || (p.ByRef && name != "" && p.Var.Name == name)
	}
	var finals []*syntax.FuncCall
	var rebinds, calls []syntax.Node
	mentioned := func(args *syntax.ArgList) bool {
		return args != nil && astquery.MentionsVariable(args, name)
	}
	unknown := func(call syntax.Node, args *syntax.ArgList) {
		if mentioned(args) {
			rebinds = append(rebinds, call)
		} else {
			calls = append(calls, call)
		}
	}
	visit := func(x syntax.Node) bool {
		if syntax.IsVariableScope(x) {
			return false
		}
		switch x := x.(type) {
		case *syntax.Global:
			for _, v := range x.Vars {
				shared = shared || astquery.MentionsVariable(v, name)
			}
		case *syntax.StaticVar:
			shared = shared || (name != "" && x.Var.Name == name)
		case *syntax.Assign:
			switch x.Var.(type) {
			case *syntax.Array, *syntax.List:
				if astquery.MentionsVariable(x.Var, name) {
					rebinds = append(rebinds, x)
				}
			}
		case *syntax.Foreach:
			if astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name) {
				rebinds = append(rebinds, x.Value)
			}
		case *syntax.Catch:
			if x.Var != nil && name != "" && x.Var.Name == name {
				rebinds = append(rebinds, x.Var)
			}
		case *syntax.FuncCall:
			switch semanticquery.NativeBuiltinName(ctx, x) {
			case "hash_final":
				// Finalizing another variable proves nothing about this one.
				if name == "" || mentioned(x.Args) {
					finals = append(finals, x)
				}
			case "extract", "parse_str":
				rebinds = append(rebinds, x)
			case "":
				unknown(x, x.Args)
			}
		case *syntax.MethodCall:
			unknown(x, x.Args)
		case *syntax.StaticCall:
			unknown(x, x.Args)
		case *syntax.New:
			unknown(x, x.Args)
		}
		return true
	}
	if body := syntax.VariableScopeBody(scope); body != nil {
		syntax.Inspect(body, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	for _, p := range finals {
		if p.Span().End > c.Span().Start || !semanticquery.NativeDominates(p, c) {
			continue
		}
		between := func(nodes []syntax.Node) bool {
			for _, n := range nodes {
				if n.Span().Start >= p.Span().End && n.Span().End <= c.Span().Start {
					return true
				}
			}
			return false
		}
		if name == "" || (!between(rebinds) && (!shared || !between(calls))) {
			return true
		}
	}
	return false
}
