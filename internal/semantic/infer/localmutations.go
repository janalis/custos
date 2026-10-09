package infer

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// statementKill is the block in which a standalone mutation always runs.
func (e *Env) statementKill(n syntax.Node) syntax.Span {
	if expr, ok := n.(syntax.Expr); ok {
		if _, arrow := syntax.EnclosingVariableScope(n).(*syntax.ArrowFunction); arrow {
			return e.arrowDefinitionSpan(expr)
		}
	}
	at := n.Span().Start
	for {
		if _, expr := n.(syntax.Expr); !expr {
			break
		}
		p := n.Parent()
		if e.skippedExpressionAt(p, at) {
			return syntax.Span{}
		}
		if es, ok := p.(*syntax.ExprStmt); ok {
			if e.condPart(es.Expr, at, 0) {
				return syntax.Span{}
			}
			n = es
			break
		}
		if _, expr := p.(syntax.Expr); !expr {
			return syntax.Span{}
		}
		n = p
	}
	if b, ok := n.Parent().(*syntax.Block); ok {
		return b.Span()
	}
	if n.Parent() == nil {
		return syntax.Span{End: uint32(len(e.File.Src))}
	}
	return syntax.Span{}
}

// arrowDefinitionSpan identifies definitions that always execute when an
// arrow's expression is evaluated. Walking parents keeps this linear in
// nesting depth rather than rescanning the expression per definition.
func (e *Env) arrowDefinitionSpan(expr syntax.Expr) syntax.Span {
	arrow, ok := syntax.EnclosingVariableScope(expr).(*syntax.ArrowFunction)
	if !ok {
		return syntax.Span{}
	}
	child := syntax.Node(expr)
	for p := expr.Parent(); p != arrow; child, p = p, p.Parent() {
		if e.skippedExpressionAt(p, expr.Span().Start) {
			return syntax.Span{}
		}
		switch n := p.(type) {
		case *syntax.Binary:
			if child == syntax.Node(n.Right) {
				switch n.Op.Kind {
				case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
					return syntax.Span{}
				}
			}
		case *syntax.Ternary:
			if child != syntax.Node(n.Cond) {
				return syntax.Span{}
			}
		case *syntax.MatchArm:
			return syntax.Span{}
		}
	}
	return arrow.Expr.Span()
}

// incDecStored describes the value stored after ++/--. Integer boundaries
// may overflow to float; strings may be numeric or use string increment.
// Invalid/unknown operands remain unknown rather than suggesting fixes.
func (e *Env) incDecStored(n *syntax.IncDec) types.Type {
	t := e.TypeOf(n.Var)
	if t.IsUnknown() {
		return types.Unknown
	}
	var ts []types.Type
	for _, a := range t.Atoms() {
		switch a {
		case "null":
			if n.Op.Kind == syntax.TInc {
				ts = append(ts, types.Int)
			} else {
				ts = append(ts, types.Null)
			}
		case "int":
			ts = append(ts, types.Of("int", "float"))
		case "float":
			ts = append(ts, types.Float)
		case "string":
			ts = append(ts, types.Of("string", "int", "float"))
		case "bool", "true", "false":
			ts = append(ts, types.Of(a))
		default:
			return types.Unknown
		}
	}
	return types.Union(ts...)
}
