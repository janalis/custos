package flowquery

import "custos/internal/php/syntax"

// ReceiverEffect describes the storage a method call may mutate. Conditional
// means a nullsafe access may skip the call; it never proves that storage was
// definitely modified. A fluent return may alias the root receiver.
type ReceiverEffect struct {
	Root        syntax.Expr
	Conditional bool
}

// MethodReceiverEffect follows a receiver through member/element accesses and
// fluent calls. A missing or excessively deep chain supplies no effect proof.
func MethodReceiverEffect(call *syntax.MethodCall) ReceiverEffect {
	if call == nil {
		return ReceiverEffect{}
	}
	effect := ReceiverEffect{Root: call.Var, Conditional: call.NullSafe}
	for range MaxAssignScan {
		switch node := effect.Root.(type) {
		case *syntax.Paren:
			effect.Root = node.Expr
		case *syntax.PropertyFetch:
			effect.Conditional = effect.Conditional || node.NullSafe
			effect.Root = node.Var
		case *syntax.ArrayDimFetch:
			effect.Root = node.Var
		case *syntax.MethodCall:
			effect.Conditional = effect.Conditional || node.NullSafe
			effect.Root = node.Var
		default:
			return effect
		}
	}
	return ReceiverEffect{}
}
