// Package assertioncontainsrequiredsideeffect implements AssertionContainsRequiredSideEffect.
package assertioncontainsrequiredsideeffect

import (
	"slices"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Move required assignments outside assertions."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AssertionContainsRequiredSideEffect" }
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
	if !semanticquery.NativeBuiltin(ctx, c, "assert") {
		return
	}
	scope := syntax.EnclosingVariableScope(c)
	params := syntax.VariableScopeParams(scope)
	assignments := []*syntax.Assign{}
	syntax.Inspect(c.Args, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if a, ok := n.(*syntax.Assign); ok && !a.ByRef && a.Op.Kind == syntax.TEqual {
			assignments = append(assignments, a)
		}
		return true
	})
	for _, a := range assignments {
		v, ok := a.Var.(*syntax.Variable)
		if !ok {
			continue
		}
		parameter := false
		for _, p := range params {
			if p.Var.Name == v.Name {
				parameter = true
			}
		}
		if parameter {
			continue
		}
		earlier := false
		later := false
		nodes := 0
		visit := func(node syntax.Node) bool {
			if node != scope && syntax.IsVariableScope(node) {
				return false
			}
			nodes++
			if nodes > 512 {
				return false
			}
			if node.Span().Start >= c.Span().Start && node.Span().End <= c.Span().End {
				return false
			}
			if !later && node.Span().Start > c.Span().End && assertionRebindsLocal(ctx, node, v.Name) {
				earlier = true
				return false
			}
			if variable, ok := node.(*syntax.Variable); ok && variable.Name == v.Name {
				if variable.Span().Start < c.Span().Start {
					earlier = true
					return false
				}
				if variable.Span().Start > c.Span().End && semanticquery.NativeCallbackReachable(ctx, variable) {
					required := true
					for parent := variable.Parent(); parent != nil && parent != scope; parent = parent.Parent() {
						switch x := parent.(type) {
						case *syntax.Isset, *syntax.Empty, *syntax.Unset:
							required = false
						case *syntax.FuncCall:
							// A read in another assertion is compiled out with this one.
							if semanticquery.NativeBuiltin(ctx, x, "assert") {
								required = false
							}
						case *syntax.Binary:
							if x.Op.Kind == syntax.TCoalesce && x.Left.Span().Contains(variable.Span()) {
								required = false
							}
						}
						if _, ok := parent.(syntax.Stmt); ok {
							if parent.Parent() != c.Parent().Parent() {
								required = false
							}
							break
						}
					}
					if required && semanticquery.NativeDominates(c, variable) {
						later = true
					}
				}
			}
			if assign, ok := node.(*syntax.Assign); ok && assign.Span().Start > c.Span().End && astquery.MentionsVariable(assign.Var, v.Name) {
				if !later {
					earlier = true
				}
				return false
			}
			if node.Span().Start > c.Span().End {
				if u, ok := node.(*syntax.Unset); ok && astquery.MentionsVariable(u, v.Name) && !later {
					earlier = true
				}
				if a, ok := node.(*syntax.Assign); ok && a.ByRef && astquery.MentionsVariable(a.Value, v.Name) && !later {
					earlier = true
				}
			}
			if node.Span().Start < c.Span().Start {
				switch node.(type) {
				case *syntax.Global, *syntax.StaticStmt:
					if astquery.MentionsVariable(node, v.Name) {
						earlier = true
					}
				}
			}
			return !earlier
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		if !earlier && later && nodes <= 512 {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// assertionRebindsLocal reports whether node gives the named local a new value
// without reading it: a foreach target, a catch variable or a positional
// write-only argument of a resolved builtin.
func assertionRebindsLocal(ctx *analysis.Context, node syntax.Node, name string) bool {
	switch x := node.(type) {
	case *syntax.Foreach:
		return !astquery.MentionsVariable(x.Expr, name) && (astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name))
	case *syntax.Catch:
		return x.Var != nil && x.Var.Name == name
	case *syntax.FuncCall:
		positions := assertionOutParameters[semanticquery.NativeBuiltinName(ctx, x)]
		if positions == nil {
			return false
		}
		written := false
		for i, expr := range x.Args.Args {
			arg, ok := expr.(*syntax.Arg)
			if !ok || arg.Name != nil || arg.Unpack {
				return false
			}
			target, ok := arg.Value.(*syntax.Variable)
			if ok && target.NameExpr == nil && target.Name == name && slices.Contains(positions, i) {
				written = true
			} else if astquery.MentionsVariable(arg.Value, name) {
				return false
			}
		}
		return written
	}
	return false
}

// assertionOutParameters lists zero-based by-reference parameters that builtins
// overwrite without reading. In-out parameters (sort, array_push, the output
// list of exec, ...) are deliberately absent.
var assertionOutParameters = map[string][]int{
	"preg_match":             {2},
	"preg_match_all":         {2},
	"preg_replace":           {4},
	"preg_replace_callback":  {4},
	"str_replace":            {3},
	"str_ireplace":           {3},
	"similar_text":           {2},
	"parse_str":              {1},
	"mb_parse_str":           {1},
	"exec":                   {2},
	"system":                 {1},
	"passthru":               {1},
	"getimagesize":           {1},
	"getimagesizefromstring": {1},
	"headers_sent":           {0, 1},
	"getmxrr":                {1, 2},
	"is_callable":            {2},
	"fsockopen":              {2, 3},
}
