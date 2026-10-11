// Package generatorreturnbeforecompletion implements the native GeneratorReturnBeforeCompletion inspection.
package generatorreturnbeforecompletion

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Finish the generator before reading its return value."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GeneratorReturnBeforeCompletion" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP70 {
		return
	}
	call := n.(*syntax.MethodCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "getReturn") {
		return
	}
	v, ok := call.Var.(*syntax.Variable)
	if !ok || v.Name == "" {
		return
	}
	st, ok := consumingStatement(call)
	if !ok {
		return
	}
	prev, ok := astquery.PrevStmtNoDoc(ctx.File, st)
	if !ok {
		return
	}
	assignment, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	a, ok := assignment.Expr.(*syntax.Assign)
	if !ok || a.ByRef {
		return
	}
	local, ok := a.Var.(*syntax.Variable)
	if !ok || local.Name != v.Name {
		return
	}
	closureCall, ok := syntax.UnwrapParens(a.Value).(*syntax.FuncCall)
	if !ok {
		return
	}
	closure, ok := syntax.UnwrapParens(closureCall.Name).(*syntax.Closure)
	if !ok || closure.Body == nil || len(closure.Body.Stmts) == 0 {
		return
	}
	first, ok := closure.Body.Stmts[0].(*syntax.ExprStmt)
	if !ok {
		return
	}
	if _, ok := first.Expr.(*syntax.Yield); !ok {
		return
	}
	ctx.ReportNode(call, message)
}

// consumingStatement returns the statement whose evaluation starts with call:
// every ancestor up to the statement evaluates the operand holding the call
// first and unconditionally, so nothing can resume the generator before it.
func consumingStatement(call *syntax.MethodCall) (syntax.Stmt, bool) {
	child := syntax.Node(call)
	for {
		switch p := child.Parent().(type) {
		case *syntax.Paren, *syntax.Unary, *syntax.Arg:
		case *syntax.Assign:
			if target, plain := p.Var.(*syntax.Variable); !plain || target.Name == "" || syntax.Node(p.Value) != child {
				return nil, false
			}
		case *syntax.Binary:
			if syntax.Node(p.Left) != child {
				return nil, false
			}
		case *syntax.Ternary:
			if syntax.Node(p.Cond) != child {
				return nil, false
			}
		case *syntax.ArgList:
			if syntax.Node(p.Args[0]) != child {
				return nil, false
			}
		case *syntax.FuncCall:
			if _, named := p.Name.(*syntax.Name); !named {
				return nil, false
			}
		case *syntax.ExprStmt:
			return p, true
		case *syntax.Return:
			return p, true
		case *syntax.Echo:
			return p, syntax.Node(p.Exprs[0]) == child
		default:
			return nil, false
		}
		child = child.Parent()
	}
}
