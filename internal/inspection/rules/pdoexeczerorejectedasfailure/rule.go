// Package pdoexeczerorejectedasfailure implements PdoExecZeroRejectedAsFailure.
package pdoexeczerorejectedasfailure

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare database execution failure strictly with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoExecZeroRejectedAsFailure" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if semanticquery.NativeMethod(ctx, c, "PDO", "exec") && pdoExecRejectsZero(ctx, c) {
		ctx.ReportNode(c, message)
	}
}

// pdoExecRejectsZero reports whether the falsy branch of a PDO::exec result
// check throws or returns false without first consulting the connection's
// error state, which is how code tells zero affected rows from a failure.
func pdoExecRejectsZero(ctx *analysis.Context, c *syntax.MethodCall) bool {
	var e syntax.Expr = c
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			e = p
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return false
			}
			e = p
		case *syntax.Binary:
			if p.Op.Kind != syntax.TIsEqual || !semanticquery.IsFalse(p.Right) {
				return false
			}
			e = p
		case *syntax.If:
			neg, ok := syntax.UnwrapParens(e).(*syntax.Unary)
			loose, lok := syntax.UnwrapParens(e).(*syntax.Binary)
			if (!ok || neg.Op.Kind != syntax.TExclaim) && (!lok || loose.Op.Kind != syntax.TIsEqual) {
				return false
			}
			return pdoExecBranchRejects(ctx, p.Body, c.Var)
		default:
			return false
		}
	}
}

func pdoExecBranchRejects(ctx *analysis.Context, body syntax.Stmt, receiver syntax.Expr) bool {
	// Start offset of the first errorCode()/errorInfo() read on the receiver.
	consulted, seen := uint32(0), false
	var exits []syntax.Node
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch x := n.(type) {
		case *syntax.MethodCall:
			id, ok := x.Name.(*syntax.Identifier)
			if ok && !seen && (strings.EqualFold(id.Value, "errorCode") || strings.EqualFold(id.Value, "errorInfo")) && astquery.Equivalent(ctx.File, x.Var, receiver) {
				consulted, seen = x.Span().Start, true
			}
		case *syntax.Throw:
			if semanticquery.NativeCallbackReachable(ctx, x) {
				exits = append(exits, x)
			}
		case *syntax.Return:
			if semanticquery.IsFalse(x.Expr) && semanticquery.NativeCallbackReachable(ctx, x) {
				exits = append(exits, x)
			}
		}
		return true
	})
	for _, exit := range exits {
		if !seen || consulted > exit.Span().Start || pdoExecStraightLine(body, exit) {
			return true
		}
	}
	return false
}

// pdoExecStraightLine reports whether exit is a statement of the branch block
// preceded only by plain expression or output statements, so nothing in the
// branch decides whether it runs. A braceless branch needs no case: an earlier
// error-state read there always shares a construct with the exit.
func pdoExecStraightLine(body syntax.Stmt, exit syntax.Node) bool {
	stmt := exit
	if _, ok := exit.(*syntax.Throw); ok {
		stmt = exit.Parent()
	}
	block, ok := body.(*syntax.Block)
	if !ok || stmt.Parent() != syntax.Node(block) {
		return false
	}
	for _, s := range block.Stmts {
		if syntax.Node(s) == stmt {
			break
		}
		switch s.(type) {
		case *syntax.ExprStmt, *syntax.Echo:
		default:
			return false
		}
	}
	return true
}
