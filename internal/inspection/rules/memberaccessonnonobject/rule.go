// Package memberaccessonnonobject implements the native
// MemberAccessOnNonObject inspection.
package memberaccessonnonobject

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "MemberAccessOnNonObject" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KPropertyFetch, syntax.KMethodCall}
}

func (rule) Semantic() {}

// Check reports the receiver of a member access when native declarations
// prove it is never an object.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var receiver syntax.Expr
	switch access := n.(type) {
	case *syntax.PropertyFetch:
		if quietRead(access) {
			return
		}
		receiver = access.Var
	case *syntax.MethodCall:
		receiver = access.Var
	}
	typ := ctx.Types().ProvenTypeOf(receiver).Without("null")
	if typ.IsUnknown() || !neverObject(typ) {
		return
	}
	ctx.ReportNode(receiver, "Access this member on an object, not "+typ.String()+".")
}

// neverObject reports a type made of scalars and arrays only.
func neverObject(t types.Type) bool {
	for _, a := range t.Atoms() {
		switch a {
		case "int", "float", "string", "bool", "true", "false":
		default:
			if !types.IsArrayAtom(a) {
				return false
			}
		}
	}
	return true
}

// quietRead reports a property read PHP evaluates without complaint: inside
// isset() or empty(), or on the left of `??`, at any depth of the chain.
func quietRead(n syntax.Node) bool {
	child := n
	for p := n.Parent(); p != nil; child, p = p, p.Parent() {
		switch parent := p.(type) {
		case *syntax.PropertyFetch, *syntax.ArrayDimFetch, *syntax.Paren:
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return parent.Op.Kind == syntax.TCoalesce && parent.Left == child
		default:
			return false
		}
	}
	return false
}
