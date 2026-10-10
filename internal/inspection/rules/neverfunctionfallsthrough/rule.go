// Package neverfunctionfallsthrough implements the native NeverFunctionFallsThrough inspection.
package neverfunctionfallsthrough

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Terminate every path in the never function."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "NeverFunctionFallsThrough" }
func (rule) Semantic()   {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	var ret syntax.Expr
	var body *syntax.Block
	switch x := n.(type) {
	case *syntax.Function:
		ret = x.ReturnType
		body = x.Body
	case *syntax.Method:
		ret = x.ReturnType
		body = x.Body
	case *syntax.Closure:
		ret = x.ReturnType
		body = x.Body
	}
	if ret == nil || body == nil || !strings.EqualFold(ctx.Text(ret), "never") || syntax.Terminates(body) {
		return
	}
	unknown := false
	explicit := false

	budget := 256
	syntax.Inspect(body, func(x syntax.Node) bool {
		budget--
		if budget < 0 {
			unknown = true
			return false
		}
		if x != body && syntax.IsVariableScope(x) {
			return false
		}
		switch x := x.(type) {
		case *syntax.Return:
			explicit = true
		case *syntax.Goto, *syntax.BadStmt:
			unknown = true
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			ret, known := callReturn(ctx, x)
			if !known {
				unknown = true
			} else if ret == "never" {
				if _, direct := x.Parent().(*syntax.ExprStmt); !direct {
					unknown = true
				}
			}
		case *syntax.Exit, *syntax.Throw:
			if _, direct := x.Parent().(*syntax.ExprStmt); !direct {
				unknown = true
			}
		case *syntax.While:
			if value, known := semanticquery.NativeTruth(ctx, x.Cond); known && value && !syntax.Terminates(x) {
				unknown = true
			}
		case *syntax.For:
			if len(x.Cond) == 1 {
				if value, known := semanticquery.NativeTruth(ctx, x.Cond[0]); known && value && !syntax.Terminates(x) {
					unknown = true
				}
			}
		}
		return true
	})
	budget = 256
	neverCall := terminates(ctx, body, &budget)
	if budget < 0 {
		unknown = true
	}

	if !unknown && !explicit && !neverCall {
		ctx.ReportNode(ret, message)
	}
}

func callReturn(ctx *analysis.Context, n syntax.Node) (string, bool) {
	switch c := n.(type) {
	case *syntax.FuncCall:
		fn := ctx.Types().ResolveFunction(c)
		if fn != nil {
			return fn.Return, true
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		classes := ctx.TypeOf(c.Var).Classes()
		if ok && len(classes) == 1 {
			m := ctx.Index().FindMethod(classes[0], id.Value, ctx.PHP)
			if m != nil {
				return m.Return, true
			}
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if ok {
			m := ctx.Index().FindMethod(semanticquery.StaticCallClass(ctx, c.Class), id.Value, ctx.PHP)
			if m != nil {
				return m.Return, true
			}
		}
	}
	return "", false
}

func terminates(ctx *analysis.Context, s syntax.Stmt, budget *int) bool {
	*budget--
	if *budget < 0 {
		return false
	}
	if syntax.Terminates(s) {
		return true
	}
	switch s := s.(type) {
	case *syntax.ExprStmt:
		ret, known := callReturn(ctx, syntax.UnwrapParens(s.Expr))
		return known && ret == "never"
	case *syntax.Block:
		for _, st := range s.Stmts {
			if terminates(ctx, st, budget) {
				return true
			}
		}
	case *syntax.If:
		if s.Else == nil || !terminates(ctx, s.Body, budget) || !terminates(ctx, s.Else.Body, budget) {
			return false
		}
		for _, branch := range s.ElseIfs {
			if !terminates(ctx, branch.Body, budget) {
				return false
			}
		}
		return true
	}
	return false
}
