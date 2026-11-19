// Package invalidoperandtype implements the native InvalidOperandType
// inspection.
package invalidoperandtype

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "InvalidOperandType" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary, syntax.KAssign} }

func (rule) Semantic() {}

// overloading lists the classes (lower case) whose objects take part in
// arithmetic: by overloading operators or by converting to a number.
var overloading = [...]string{"gmp", `bcmath\number`, `ffi\cdata`, "simplexmlelement"}

// Check reports the operands of an arithmetic or bitwise operation that
// native declarations prove make it throw.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !ctx.PHP.AtLeast(phpversion.PHP80) {
		return
	}
	var left, right syntax.Expr
	var plus bool
	switch op := n.(type) {
	case *syntax.Binary:
		if !arithmetic(op.Op.Kind) {
			return
		}
		left, right, plus = op.Left, op.Right, op.Op.Kind == syntax.TPlus
	case *syntax.Assign:
		if !compound(op.Op.Kind) {
			return
		}
		left, right, plus = op.Var, op.Value, op.Op.Kind == syntax.TPlusEqual
	}
	check(ctx, left, right, plus)
	check(ctx, right, left, plus)
}

// check reports operand when it cannot take part in the operation.
func check(ctx *analysis.Context, operand, other syntax.Expr, plus bool) {
	typ := ctx.Types().ProvenTypeOf(operand)
	if typ.IsUnknown() {
		return
	}
	switch {
	case onlyArrays(typ):
		if plus && mayBeArray(ctx.Types().ProvenTypeOf(other)) {
			return
		}
	case !onlyPlainObjects(ctx, typ):
		return
	}
	ctx.ReportNode(operand, "Use a number here, not "+typeName(typ)+".")
}

// typeName prints t's members, class names without the leading backslash.
func typeName(t types.Type) string {
	names := make([]string, 0, len(t.Atoms()))
	for _, a := range t.Atoms() {
		names = append(names, strings.TrimPrefix(a, `\`))
	}
	return strings.Join(names, "|")
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
		if ctx.Index().Class(class, ctx.PHP) == nil || !semanticquery.HierarchyResolved(ctx.Index(), class, ctx.PHP) {
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

// compound reports a compound assignment of an arithmetic operator.
func compound(k syntax.TokenKind) bool {
	switch k {
	case syntax.TPlusEqual, syntax.TMinusEqual, syntax.TMulEqual, syntax.TDivEqual, syntax.TModEqual, syntax.TPowEqual,
		syntax.TAndEqual, syntax.TOrEqual, syntax.TXorEqual, syntax.TSlEqual, syntax.TSrEqual:
		return true
	}
	return false
}
