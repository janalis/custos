// Package pdoreferencebindingvariablereuse implements the native PdoReferenceBindingVariableReuse inspection.
package pdoreferencebindingvariablereuse

import (
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Bind distinct parameters to distinct scalar identities."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoReferenceBindingVariableReuse" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	_, known := semanticquery.PreparedSQL(ctx, c.Var)
	if !known {
		return
	}

	if c.Args != nil && len(c.Args.Args) > 0 {
		return
	}
	type referenceBinding struct{ param, value string }
	variables := map[string]referenceBinding{}
	report := false
	prior, historyKnown := semanticquery.NativeCallHistory(ctx, c, c.Var, "bindparam").Certain()
	if !historyKnown {
		return
	}
	for _, node := range prior {
		binding := node.(*syntax.MethodCall)
		param, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(binding.Args, 0, "param"))
		variable, ok := semanticquery.CallArgument(binding.Args, 1, "var").(*syntax.Variable)
		if !known || !ok {
			return
		}
		value, known := scalarBinding(ctx, variable)
		if !known {
			return
		}
		if old, exists := variables[variable.Name]; exists && old.param != param && old.value != value {
			report = true
		}
		variables[variable.Name] = referenceBinding{param: param, value: value}
	}

	if report {
		ctx.ReportNode(c, message)
	}
}

// scalarBinding proves a scalar value at one binding point.
func scalarBinding(ctx *analysis.Context, expr syntax.Expr) (string, bool) {
	if value, known := semanticquery.NativeInt(ctx, expr); known {
		return "integer:" + strconv.FormatInt(value, 10), true
	}
	if value, known := semanticquery.NativeString(ctx, expr); known {
		return "string:" + value, true
	}
	return "", false
}
