// Package zipentryindexpastend implements the native ZipEntryIndexPastEnd inspection.
package zipentryindexpastend

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an existing ZIP entry index."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipEntryIndexPastEnd" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	valid := false
	for _, method := range []string{"getNameIndex", "statIndex", "getFromIndex", "deleteIndex", "getStreamIndex"} {
		valid = valid || semanticquery.NativeMethod(ctx, c, "ZipArchive", method)
	}
	if !valid {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "index")
	if v, ok := semanticquery.NativeInt(ctx, arg); ok && v < 0 && stableIndex(ctx, n, arg) {
		ctx.ReportNode(c, message)
		return
	}
	property, ok := syntax.UnwrapParens(arg).(*syntax.PropertyFetch)
	if !ok || !semanticquery.ExpansionCSame(ctx, property.Var, c.Var) {
		return
	}
	name, ok := property.Name.(*syntax.Identifier)
	if ok && name.Value == "numFiles" {
		ctx.ReportNode(c, message)
	}
}

// stableIndex reports whether the value resolved for an index expression can
// be trusted. Value resolution does not model ++/-- or writes made through a
// by-reference closure capture, and it reads a property as its declared
// default, so those shapes are not proof of the value passed at the call.
func stableIndex(ctx *analysis.Context, at syntax.Node, arg syntax.Expr) bool {
	switch v := syntax.UnwrapParens(arg).(type) {
	case *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
		return false
	case *syntax.Variable:
		// An arrow function shares its parent's variables by value, so
		// both its own writes and the enclosing ones are relevant.
		scope := syntax.EnclosingFuncLike(at)
		for scope != nil && scope.Kind() == syntax.KArrowFunction {
			if !stableVariable(flowquery.VarAccessesByName(ctx.File, scope), v.Name, map[string]bool{}) {
				return false
			}
			scope = syntax.EnclosingFuncLike(scope)
		}
		return stableVariable(flowquery.VarAccessesByName(ctx.File, scope), v.Name, map[string]bool{})
	}
	return true
}

// stableVariable follows plain variable-to-variable copies so that an alias
// of a counter is treated like the counter itself.
func stableVariable(accesses map[string][]flowquery.VarAccess, name string, seen map[string]bool) bool {
	if seen[name] {
		return true
	}
	seen[name] = true
	for _, access := range accesses[name] {
		switch by := access.By.(type) {
		case *syntax.IncDec:
			return false
		case *syntax.ClosureUse:
			if by.ByRef {
				return false
			}
		case *syntax.Assign:
			source, ok := syntax.UnwrapParens(astquery.AssignedValue(by)).(*syntax.Variable)
			if access.Write && ok && !stableVariable(accesses, source.Name, seen) {
				return false
			}
		}
	}
	return true
}
