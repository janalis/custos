package astquery

import (
	"custos/internal/php/syntax"
)

// LogicalOperand reports whether e (through parentheses) is a condition
// of if/elseif/while/do-while, the operand of `!`, the condition of a full
// ternary, or (withBinary) an operand of a logical binary operator.
func LogicalOperand(e syntax.Node, withBinary bool) bool {
	cur := e
	for {
		p := cur.Parent()
		switch p := p.(type) {
		case *syntax.Paren:
			cur = p
			continue
		case *syntax.If:
			return p.Cond == cur
		case *syntax.ElseIf:
			return p.Cond == cur
		case *syntax.While:
			return p.Cond == cur
		case *syntax.DoWhile:
			return p.Cond == cur
		case *syntax.Unary:
			return p.Op.Kind == syntax.TExclaim
		case *syntax.Ternary:
			return p.Then != nil && p.Cond == cur
		case *syntax.Binary:
			if !withBinary {
				return false
			}
			switch p.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
				return true
			}
		}
		return false
	}
}
