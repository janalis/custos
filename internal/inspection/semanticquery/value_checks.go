package semanticquery

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

// Value checks shared by a rule that judges proven types and its twin that
// judges PHPDoc-aware ones: member access on a non-object, foreach over a
// non-iterable, and operands of arithmetic. Each check takes the view it
// judges. The documented view reports only what the proven view does not,
// and only where the PHPDoc-aware type does not contradict a known proven
// type (a tag the code disproves is stale, not a contract the value breaks)
// and the value is not `$this` in a closure, which may be rebound.

// documentedType returns the PHPDoc-aware type of x, of proven type proven,
// or its proven type where an array follows a write into an offset of false
// (see vivified).
func documentedType(ctx *analysis.Context, x syntax.Expr, proven types.Type) types.Type {
	if vivified(ctx, x, proven) {
		return proven
	}
	return ctx.TypeOf(x)
}

// documentedTrusted reports whether the PHPDoc-aware type given of x may
// stand for its value: x is not `$this` in a closure, and given does not
// contradict proven, its known proven type.
func documentedTrusted(ctx *analysis.Context, x syntax.Expr, proven, given types.Type) bool {
	if reboundThis(x) {
		return false
	}
	return proven.IsUnknown() || types.Accepts(proven, given, fileTypeCheck(ctx).strict) != types.No
}

// judge returns the type of x reject turns into a reported one in the view
// documented selects (see the value checks above).
func judge(ctx *analysis.Context, x syntax.Expr, documented bool, reject func(types.Type) (types.Type, bool)) (types.Type, bool) {
	proven := ctx.Types().ProvenTypeOf(x)
	if t, ok := reject(proven); ok || !documented {
		return t, ok && !documented
	}
	given := documentedType(ctx, x, proven)
	t, ok := reject(given)
	return t, ok && documentedTrusted(ctx, x, proven, given)
}

// NonObjectReceiver returns the receiver of property fetch or method call n
// and its type without null when, in the view documented selects, that
// type holds only scalars and arrays. A receiver that can only be null is
// left to PossiblyNullAccess, and a property read PHP evaluates quietly is
// skipped. n must be one of those two kinds.
func NonObjectReceiver(ctx *analysis.Context, n syntax.Node, documented bool) (syntax.Expr, types.Type, bool) {
	var receiver syntax.Expr
	if fetch, ok := n.(*syntax.PropertyFetch); ok {
		if quietFetch(fetch, false) {
			return nil, types.Unknown, false
		}
		receiver = fetch.Var
	} else {
		receiver = n.(*syntax.MethodCall).Var
	}
	t, ok := judge(ctx, receiver, documented, func(t types.Type) (types.Type, bool) {
		t = t.Without("null")
		return t, !t.IsUnknown() && scalarsOr(t, types.IsArrayAtom)
	})
	return receiver, t, ok
}

// NonIterableSubject returns the type of the subject of foreach n when, in
// the view documented selects, it holds only scalars and null.
func NonIterableSubject(ctx *analysis.Context, n *syntax.Foreach, documented bool) (types.Type, bool) {
	return judge(ctx, n.Expr, documented, func(t types.Type) (types.Type, bool) {
		return t, !t.IsUnknown() && scalarsOr(t, func(a string) bool { return a == "null" })
	})
}

// scalarsOr reports a type each member of which is an int, a float, a
// string, a bool or a member also accepts.
func scalarsOr(t types.Type, also func(string) bool) bool {
	for _, a := range t.Atoms() {
		switch a {
		case "int", "float", "string", "bool", "true", "false":
		default:
			if !also(a) {
				return false
			}
		}
	}
	return true
}

// OperandMismatch is an operand of arithmetic that cannot take part in it.
type OperandMismatch struct {
	Operand syntax.Expr
	Type    types.Type
}

// Message is the report text of m.
func (m OperandMismatch) Message() string {
	names := make([]string, 0, len(m.Type.Atoms()))
	for _, a := range m.Type.Atoms() {
		names = append(names, strings.TrimPrefix(a, `\`))
	}
	return "Use a number here, not " + strings.Join(names, "|") + "."
}

// overloading lists the classes (lower case) whose objects take part in
// arithmetic: by overloading operators or by converting to a number.
var overloading = [...]string{"gmp", `bcmath\number`, `ffi\cdata`, "simplexmlelement"}

// InvalidOperands returns the operands of arithmetic or bitwise operation n
// (or its compound assignment) that, in the view documented selects, are
// only arrays (except both sides of an array `+`) or only objects of plain,
// fully known classes. PHP throws a TypeError on them from 8.0; the caller
// gates the version. n must be a binary operation or an assignment.
func InvalidOperands(ctx *analysis.Context, n syntax.Node, documented bool) []OperandMismatch {
	return ctx.Shared(operandsKey{n, documented}, func() any { return invalidOperands(ctx, n, documented) }).([]OperandMismatch)
}

type operandsKey struct {
	n          syntax.Node
	documented bool
}

func invalidOperands(ctx *analysis.Context, n syntax.Node, documented bool) []OperandMismatch {
	var left, right syntax.Expr
	var plus bool
	if op, ok := n.(*syntax.Binary); ok {
		if !arithmetic(op.Op.Kind) {
			return nil
		}
		left, right, plus = op.Left, op.Right, op.Op.Kind == syntax.TPlus
	} else {
		op := n.(*syntax.Assign)
		if !compoundArithmetic(op.Op.Kind) {
			return nil
		}
		left, right, plus = op.Var, op.Value, op.Op.Kind == syntax.TPlusEqual
	}
	var proven []OperandMismatch
	if documented {
		proven = InvalidOperands(ctx, n, false)
	}
	var out []OperandMismatch
	for _, pair := range [2][2]syntax.Expr{{left, right}, {right, left}} {
		operand, other := pair[0], pair[1]
		if !documented {
			if t := ctx.Types().ProvenTypeOf(operand); badOperand(ctx, t, func() types.Type { return ctx.Types().ProvenTypeOf(other) }, plus) {
				out = append(out, OperandMismatch{Operand: operand, Type: t})
			}
			continue
		}
		if reported(proven, operand) {
			continue
		}
		otherType := func() types.Type { return documentedType(ctx, other, ctx.Types().ProvenTypeOf(other)) }
		t := ctx.Types().ProvenTypeOf(operand)
		if given := documentedType(ctx, operand, t); badOperand(ctx, given, otherType, plus) && documentedTrusted(ctx, operand, t, given) {
			out = append(out, OperandMismatch{Operand: operand, Type: given})
		}
	}
	return out
}

// reported reports whether operand is among ms.
func reported(ms []OperandMismatch, operand syntax.Expr) bool {
	for _, m := range ms {
		if m.Operand == operand {
			return true
		}
	}
	return false
}

// badOperand reports an operand of type t that cannot take part in the
// operation with the other operand, whose type other gives.
func badOperand(ctx *analysis.Context, t types.Type, other func() types.Type, plus bool) bool {
	switch {
	case t.IsUnknown():
		return false
	case onlyArrays(t):
		return !plus || !mayBeArray(other())
	default:
		return onlyPlainObjects(ctx, t)
	}
}

// onlyArrays reports a type made of arrays.
func onlyArrays(t types.Type) bool {
	for _, a := range t.Atoms() {
		if !types.IsArrayAtom(a) {
			return false
		}
	}
	return true
}

// mayBeArray reports an operand that is unknown or admits an array.
func mayBeArray(t types.Type) bool {
	if t.IsUnknown() {
		return true
	}
	for _, a := range t.Atoms() {
		if types.IsArrayAtom(a) || a == "mixed" || a == "iterable" {
			return true
		}
	}
	return false
}

// onlyPlainObjects reports a type made of classes whose whole hierarchy is
// known and none of which overloads operators.
func onlyPlainObjects(ctx *analysis.Context, t types.Type) bool {
	for _, a := range t.Atoms() {
		if !types.IsClassAtom(a) {
			return false
		}
		class := strings.TrimPrefix(a, `\`)
		if ctx.Index().Class(class, ctx.PHP) == nil || !HierarchyResolved(ctx.Index(), class, ctx.PHP) {
			return false
		}
		for _, name := range overloading {
			if ctx.Index().IsSubtype(class, name, ctx.PHP) {
				return false
			}
		}
	}
	return true
}

// arithmetic reports a binary operator that needs numbers.
func arithmetic(k syntax.TokenKind) bool {
	switch k {
	case syntax.TPlus, syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPow,
		syntax.TAmpersand, syntax.TBar, syntax.TCaret, syntax.TSl, syntax.TSr:
		return true
	}
	return false
}

// compoundArithmetic reports a compound assignment of an arithmetic
// operator.
func compoundArithmetic(k syntax.TokenKind) bool {
	switch k {
	case syntax.TPlusEqual, syntax.TMinusEqual, syntax.TMulEqual, syntax.TDivEqual, syntax.TModEqual, syntax.TPowEqual,
		syntax.TAndEqual, syntax.TOrEqual, syntax.TXorEqual, syntax.TSlEqual, syntax.TSrEqual:
		return true
	}
	return false
}
