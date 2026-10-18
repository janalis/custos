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

// handleEscapes rejects lexical uses outside direct calls and boolean tests.
// The shared option proof has already rejected copies and stored handles.
func handleEscapes(ctx *analysis.Context, at syntax.Node, handle syntax.Expr) bool {
	// The option proof only succeeds for a plain local handle variable.
	name, _ := astquery.PlainVariableName(handle)
	escaped := false
	visit := func(n syntax.Node) bool {
		if escaped || n.Span().Start >= at.Span().Start {
			return false
		}
		switch n := n.(type) {
		case *syntax.Closure:
			for _, use := range n.Uses {
				escaped = escaped || use.Var.Name == name
			}
			return false
		case *syntax.ArrowFunction:
			escaped = escaped || astquery.MentionsVariable(n.Expr, name)
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
			} else if n.Name == name {
				escaped = !keepsHandleLocal(n)
			}
		}
		return !syntax.IsVariableScope(n)
	}
	if scope := syntax.EnclosingVariableScope(at); scope != nil {
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return escaped
}

// keepsHandleLocal accepts the positions where a handle variable cannot leave
// the scope unseen. Copies have already invalidated the shared option proof.
func keepsHandleLocal(v *syntax.Variable) bool {
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
		return p.Var == child && p.Op.Kind == syntax.TEqual && !p.ByRef
	}
	return false
}
