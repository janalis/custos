// Package splobjectstorageindirectarraymutation implements the native SplObjectStorageIndirectArrayMutation inspection.
package splobjectstorageindirectarraymutation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Assign the modified storage value back."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "SplObjectStorageIndirectArrayMutation" }
func (rule) Semantic()   {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KAssign, syntax.KIncDec, syntax.KUnset}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	var targets []syntax.Expr
	switch x := n.(type) {
	case *syntax.Assign:
		targets = []syntax.Expr{x.Var}
	case *syntax.IncDec:
		targets = []syntax.Expr{x.Var}
	case *syntax.Unset:
		targets = x.Vars
	}
	for _, target := range targets {
		nested, ok := syntax.UnwrapParens(target).(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		slot, ok := nested.Var.(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		if semanticquery.NativeConstruction(ctx, slot.Var, "SplObjectStorage") == nil {
			continue
		}
		if storedArray(ctx, slot, n) && !replaced(ctx, slot, n) {
			ctx.ReportNode(n, message)
			return
		}
	}
}

// readOnly lists the SplObjectStorage methods that leave every stored value in
// place.
var readOnly = map[string]bool{
	"contains": true, "count": true, "current": true, "gethash": true, "getinfo": true, "key": true,
	"next": true, "offsetexists": true, "offsetget": true, "rewind": true, "seek": true, "valid": true,
}

// replaced reports whether the storage may have received another value for the
// slot after the last assignment that certainly ran. Only a write that provably
// stores an array keeps the fact.
func replaced(ctx *analysis.Context, slot *syntax.ArrayDimFetch, at syntax.Node) bool {
	dirty := false
	scope := syntax.EnclosingVariableScope(at)
	same := func(e syntax.Expr) bool { return semanticquery.NativeSameValue(ctx, e, slot.Var) }
	name, _ := astquery.PlainVariableName(slot.Var)
	direct := func(e syntax.Expr) *syntax.ArrayDimFetch {
		if s, ok := e.(*syntax.ArrayDimFetch); ok && same(s.Var) {
			return s
		}
		return nil
	}
	visit := func(n syntax.Node) bool {
		if n.Span().Start >= at.Span().Start {
			return false
		}
		switch x := n.(type) {
		case *syntax.Closure:
			// A capturing closure can write the storage whenever it runs.
			for _, use := range x.Uses {
				dirty = dirty || use.Var.Name == name
			}
			return false
		case *syntax.ArrowFunction:
			// An arrow function captures the storage by mentioning it.
			syntax.Inspect(x.Expr, func(inner syntax.Node) bool {
				v, ok := inner.(*syntax.Variable)
				dirty = dirty || ok && v.Name == name
				return true
			})
			return false
		case *syntax.Assign:
			s := direct(x.Var)
			if s == nil {
				break
			}
			if semanticquery.NativeSameValue(ctx, s.Dim, slot.Dim) && semanticquery.NativeDominates(x, at) {
				dirty = false
			} else if semanticquery.NativeArray(ctx, x.Value) == nil {
				dirty = true
			}
		case *syntax.Unset:
			for _, v := range x.Vars {
				dirty = dirty || direct(v) != nil
			}
		case *syntax.MethodCall:
			if id, ok := x.Name.(*syntax.Identifier); (!ok || !readOnly[strings.ToLower(id.Value)]) && same(x.Var) {
				dirty = true
			}
		}
		return n == scope || !syntax.IsVariableScope(n)
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return dirty
}

// storedArray proves the last direct storage assignment is an array.
func storedArray(ctx *analysis.Context, slot *syntax.ArrayDimFetch, at syntax.Node) bool {
	found := false
	scope := syntax.EnclosingVariableScope(at)
	scan := func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok && semanticquery.NativeDominates(a, at) {
			if s, ok := a.Var.(*syntax.ArrayDimFetch); ok && semanticquery.NativeSameValue(ctx, s.Var, slot.Var) && semanticquery.NativeSameValue(ctx, s.Dim, slot.Dim) {
				found = semanticquery.NativeArray(ctx, a.Value) != nil
			}
		}
		return true
	}
	for _, event := range semanticquery.ContractEvents(ctx, scope) {
		if event.Span().Start >= at.Span().Start {
			break
		}
		scan(event)
	}
	return found
}
