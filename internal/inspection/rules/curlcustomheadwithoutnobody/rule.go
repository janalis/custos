// Package curlcustomheadwithoutnobody implements the native CurlCustomHeadWithoutNoBody inspection.
package curlcustomheadwithoutnobody

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use CURLOPT_NOBODY for a HEAD transfer."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlCustomHeadWithoutNoBody" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	method, known := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_CUSTOMREQUEST")
	name, nk := semanticquery.NativeString(ctx, method)
	if !known || !nk || name != "HEAD" || handleEscapes(ctx, c, h) {
		return
	}
	noBody, explicit := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_NOBODY")
	truth, tk := semanticquery.NativeTruth(ctx, noBody)
	if get, known := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_HTTPGET"); known {
		value, known := semanticquery.NativeTruth(ctx, get)
		if !known {
			return
		}
		if value && (!explicit || get.Span().Start > noBody.Span().Start) {
			ctx.ReportNode(c, message)
			return
		}
	}
	if explicit && tk && !truth || !explicit && semanticquery.NativeCurlDefault(ctx, c, h, "CURLOPT_NOBODY") {
		ctx.ReportNode(c, message)
	}
}

// handleEscapes reports whether the handle variable, or a local copy of it,
// is used before at in a way the call-based option proof cannot follow: a
// closure capture, an arrow-function body, an array item, a store into a
// property or element, a reference, compact() or get_defined_vars(), or any
// other value position. Only direct call arguments, plain copies and boolean
// tests keep the handle local.
func handleEscapes(ctx *analysis.Context, at syntax.Node, handle syntax.Expr) bool {
	// The option proof only succeeds for a plain local handle variable.
	name, _ := astquery.PlainVariableName(handle)
	names := map[string]bool{name: true}
	escaped := false
	visit := func(n syntax.Node) bool {
		if escaped || n.Span().Start >= at.Span().Start {
			return false
		}
		switch n := n.(type) {
		case *syntax.Closure:
			for _, use := range n.Uses {
				escaped = escaped || names[use.Var.Name]
			}
			return false
		case *syntax.ArrowFunction:
			for alias := range names {
				escaped = escaped || astquery.MentionsVariable(n.Expr, alias)
			}
			return false
		case *syntax.FuncCall:
			// These builtins read the scope's variables by name.
			switch semanticquery.NativeBuiltinName(ctx, n) {
			case "compact", "get_defined_vars":
				escaped = true
			}
		case *syntax.Variable:
			if n.NameExpr != nil {
				escaped = true
			} else if names[n.Name] {
				escaped = !keepsHandleLocal(n, names)
			}
		}
		return !syntax.IsVariableScope(n)
	}
	// A copy adds a name, so repeat until earlier uses of that name are seen.
	for known := 0; !escaped && known != len(names); {
		known = len(names)
		if scope := syntax.EnclosingVariableScope(at); scope != nil {
			syntax.Inspect(syntax.VariableScopeBody(scope), visit)
		} else {
			syntax.InspectFile(ctx.File, visit)
		}
	}
	return escaped
}

// keepsHandleLocal accepts the positions where a handle variable cannot leave
// the scope unseen; a plain copy registers its target as another handle name.
func keepsHandleLocal(v *syntax.Variable, names map[string]bool) bool {
	parent, child := astquery.ParentSkipParens(v)
	switch p := parent.(type) {
	case *syntax.Arg, *syntax.Isset, *syntax.Empty, *syntax.If, *syntax.ElseIf:
		return true
	case *syntax.Unary:
		return p.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual, syntax.TBooleanAnd, syntax.TBooleanOr:
			return true
		}
	case *syntax.Assign:
		if p.Var == child {
			return p.Op.Kind == syntax.TEqual && !p.ByRef
		}
		target, plain := astquery.PlainVariableName(p.Var)
		if plain && p.Op.Kind == syntax.TEqual && !p.ByRef {
			names[target] = true
			return true
		}
	}
	return false
}
