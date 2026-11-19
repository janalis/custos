package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// Value checks shared by a rule that judges proven types and its twin that
// judges PHPDoc-aware ones: member access on a non-object, foreach over a
// non-iterable, and operands of arithmetic.

// Judged is a value type a check rejects, and whether native declarations
// alone give it (see infer.Env.ProvenTypeOf).
type Judged struct {
	Type   types.Type
	Proven bool
}

// documentedType returns the PHPDoc-aware type of x, or its proven type
// where an array follows a write into an offset of false (see vivified).
func documentedType(ctx *analysis.Context, x syntax.Expr) types.Type {
	if proven := ctx.Types().ProvenTypeOf(x); vivified(ctx, x, proven) {
		return proven
	}
	return ctx.TypeOf(x)
}

// judge returns the type of x reject turns into a reported one, judging the
// proven type first and the PHPDoc-aware one next.
func judge(ctx *analysis.Context, x syntax.Expr, reject func(types.Type) (types.Type, bool)) (Judged, bool) {
	if t, ok := reject(ctx.Types().ProvenTypeOf(x)); ok {
		return Judged{Type: t, Proven: true}, true
	}
	if t, ok := reject(documentedType(ctx, x)); ok {
		return Judged{Type: t}, true
	}
	return Judged{}, false
}

// NonObjectReceiver returns the receiver of property fetch or method call n
// and its type without null when that type holds only scalars and arrays.
// A receiver that can only be null is left to PossiblyNullAccess, and a
// property read PHP evaluates quietly is skipped. n must be one of those
// two kinds.
func NonObjectReceiver(ctx *analysis.Context, n syntax.Node) (syntax.Expr, Judged, bool) {
	var receiver syntax.Expr
	if fetch, ok := n.(*syntax.PropertyFetch); ok {
		if quietRead(fetch) {
			return nil, Judged{}, false
		}
		receiver = fetch.Var
	} else {
		receiver = n.(*syntax.MethodCall).Var
	}
	j, ok := judge(ctx, receiver, func(t types.Type) (types.Type, bool) {
		t = t.Without("null")
		return t, !t.IsUnknown() && scalarsOr(t, types.IsArrayAtom)
	})
	return receiver, j, ok
}

// NonIterableSubject returns the subject of foreach n and its type when that
// type holds only scalars and null.
func NonIterableSubject(ctx *analysis.Context, n *syntax.Foreach) (Judged, bool) {
	return judge(ctx, n.Expr, func(t types.Type) (types.Type, bool) {
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

// quietRead reports a property read PHP evaluates without complaint: the
// base of a chain inside isset() or empty(), or on the left of `??`.
func quietRead(n syntax.Node) bool {
	child := n
	for p := n.Parent(); ; child, p = p, p.Parent() {
		switch parent := p.(type) {
		case *syntax.Paren:
		case *syntax.PropertyFetch:
			if parent.Var != child {
				return false // a dynamic name is read as usual
			}
		case *syntax.ArrayDimFetch:
			if parent.Var != child {
				return false // an index is read as usual
			}
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return parent.Op.Kind == syntax.TCoalesce && parent.Left == child
		default:
			return false
		}
	}
}

// OperandMismatch is an operand of arithmetic that cannot take part in it.
type OperandMismatch struct {
	Operand syntax.Expr
	Judged
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
// (or its compound assignment) that are only arrays (except both sides of
// an array `+`) or only objects of plain, fully known classes. PHP throws a
// TypeError on them from 8.0; the caller gates the version. n must be a
// binary operation or an assignment.
func InvalidOperands(ctx *analysis.Context, n syntax.Node) []OperandMismatch {
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
	var out []OperandMismatch
	for _, pair := range [2][2]syntax.Expr{{left, right}, {right, left}} {
		if j, ok := operandMismatch(ctx, pair[0], pair[1], plus); ok {
			out = append(out, OperandMismatch{Operand: pair[0], Judged: j})
		}
	}
	return out
}

// operandMismatch judges operand against other, both in the proven view
// first and both in the PHPDoc-aware view next.
func operandMismatch(ctx *analysis.Context, operand, other syntax.Expr, plus bool) (Judged, bool) {
	if t := ctx.Types().ProvenTypeOf(operand); badOperand(ctx, t, ctx.Types().ProvenTypeOf(other), plus) {
		return Judged{Type: t, Proven: true}, true
	}
	if t := documentedType(ctx, operand); badOperand(ctx, t, documentedType(ctx, other), plus) {
		return Judged{Type: t}, true
	}
	return Judged{}, false
}

// badOperand reports an operand of type t that cannot take part in the
// operation with an operand of type other.
func badOperand(ctx *analysis.Context, t, other types.Type, plus bool) bool {
	switch {
	case t.IsUnknown():
		return false
	case onlyArrays(t):
		return !plus || !mayBeArray(other)
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
