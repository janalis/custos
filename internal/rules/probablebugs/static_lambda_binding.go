package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// staticLambdaBinding reports `$this` and `parent::` instance calls inside
// static closures and arrow functions.
type staticLambdaBinding struct{}

func init() { register(staticLambdaBinding{}) }

// Semantic marks the rule as needing the project symbol index.
func (staticLambdaBinding) Semantic() {}

func (staticLambdaBinding) ID() string { return "StaticLambdaBinding" }

func (staticLambdaBinding) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClosure, syntax.KArrowFunction}
}

func (staticLambdaBinding) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP54 { // D1
		return
	}
	var root syntax.Node
	switch f := n.(type) {
	case *syntax.Closure:
		if !f.Static || f.Body == nil {
			return
		}
		root = f.Body
	case *syntax.ArrowFunction:
		if !f.Static || f.Expr == nil {
			return
		}
		root = f.Expr
	}
	kw, ok := util.NextSignificant(ctx.File, n.Span().Start) // D2
	if !ok || kw.Kind != syntax.TStatic || kw.Start != n.Span().Start {
		return
	}

	fix := analysis.Fix{Title: "Remove 'static'", Edits: func() []analysis.TextEdit {
		return []analysis.TextEdit{{Span: util.WithTrailingWhitespace(ctx.File, syntax.Span{Start: kw.Start, End: kw.End})}}
	}}

	// D3, D4: every matching candidate, in pre-order.
	syntax.Inspect(root, func(c syntax.Node) bool {
		switch c := c.(type) {
		case *syntax.Variable:
			if c.NameExpr == nil && c.Name == "this" && staticLambdaOwner(c) == n {
				ctx.Report(c.Span(), "Static closures have no $this; remove 'static' or avoid $this.", fix)
			}
		case *syntax.StaticCall:
			if staticLambdaParentInstanceCall(ctx, c) && staticLambdaOwner(c) == n {
				ctx.ReportNode(c, "Calling an instance method of the parent class requires an object; this closure is static.", fix)
			}
		}
		return true
	})
}

// staticLambdaOwner returns the function whose object binding n uses: the
// nearest enclosing function-like, skipping non-static arrow functions,
// which inherit the binding of the scope they are defined in.
func staticLambdaOwner(n syntax.Node) syntax.Node {
	for f := syntax.EnclosingFuncLike(n); f != nil; f = syntax.EnclosingFuncLike(f) {
		if af, ok := f.(*syntax.ArrowFunction); ok && !af.Static {
			continue
		}
		return f
	}
	return nil
}

func staticLambdaParentInstanceCall(ctx *analysis.Context, c *syntax.StaticCall) bool {
	nm, ok := c.Class.(*syntax.Name)
	if !ok || !strings.EqualFold(nm.Value, "parent") { // keywords are case-insensitive
		return false
	}
	id, ok := c.Name.(*syntax.Identifier)
	if !ok {
		return false
	}
	cls := staticCallClass(ctx, c.Class)
	if cls == "" {
		return false
	}
	m := ctx.Index().FindMethod(cls, id.Value, ctx.PHP)
	return m != nil && !m.Static
}
