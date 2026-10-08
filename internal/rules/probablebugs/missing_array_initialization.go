package probablebugs

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// missingArrayInitialization reports `$name[] = …` pushes nested in two or
// more loops of a function where $name is never initialised.
type missingArrayInitialization struct{}

func init() { register(missingArrayInitialization{}) }

func (missingArrayInitialization) ID() string { return "MissingArrayInitialization" }

func (missingArrayInitialization) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KArrayDimFetch}
}

func (missingArrayInitialization) Check(ctx *analysis.Context, n syntax.Node) {
	access := n.(*syntax.ArrayDimFetch)
	if access.Dim != nil || access.Span().Len() == 0 { // D1
		return
	}

	// D2, D3: innermost function-like and loops crossed on the way.
	var fn syntax.Node
	loops := 0
	for p := access.Parent(); p != nil && fn == nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			loops++
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			fn = p
		}
	}
	if fn == nil || loops < 2 {
		return
	}

	// D4: base variable.
	var base syntax.Expr = access
	for {
		d, ok := base.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		base = d.Var
	}
	v, ok := base.(*syntax.Variable)
	if !ok || v.Name == "" || util.IsSuperglobal(v.Name) { // superglobals live outside the function (custos)
		return
	}
	name := v.Name

	// D5, D6: braced body, not a parameter or closure import.
	var body *syntax.Block
	var params []*syntax.Param
	switch f := fn.(type) {
	case *syntax.Function:
		body, params = f.Body, f.Params
	case *syntax.Method:
		body, params = f.Body, f.Params
	case *syntax.Closure:
		body, params = f.Body, f.Params
		for _, u := range f.Uses {
			if u.Var != nil && u.Var.Name == name {
				return
			}
		}
	}
	// body is set: an arrow function never encloses two loops (D3 stopped)
	for _, p := range params {
		if p.Var != nil && p.Var.Name == name {
			return
		}
	}

	// D7: any direct assignment operand / foreach header occurrence.
	initialised := false
	syntax.Inspect(body, func(c syntax.Node) bool {
		if initialised {
			return false
		}
		if cv, ok := c.(*syntax.Variable); ok && cv.Name == name && initialisesArray(cv) {
			initialised = true
		}
		return !initialised
	})
	if initialised {
		return
	}
	ctx.ReportNode(access, "Array '$"+name+"' is never initialised; initialise it before the loops.")
}

// initialisesArray reports whether variable v is a direct assignment operand,
// part of a foreach header, a global/static declaration, or a destructuring
// target.
func initialisesArray(v *syntax.Variable) bool {
	switch p := v.Parent().(type) {
	case *syntax.Assign:
		return p.Var == syntax.Expr(v) || p.Value == syntax.Expr(v)
	case *syntax.Foreach:
		return p.Expr == syntax.Expr(v) || p.Key == syntax.Expr(v) || p.Value == syntax.Expr(v)
	case *syntax.Global:
		return true
	case *syntax.StaticVar:
		return p.Var == v
	case *syntax.ArrayItem:
		return isDestructuringItem(p)
	}
	return false
}

// isDestructuringItem reports whether item belongs to a list()/[...] that is
// an assignment target or a foreach value target (any nesting depth).
func isDestructuringItem(item *syntax.ArrayItem) bool {
	var n syntax.Node = item
	for {
		lst := n.Parent() // an array item always sits in an Array or List
		switch p := lst.Parent().(type) {
		case *syntax.Assign:
			return p.Var == lst
		case *syntax.Foreach:
			return p.Value == lst || p.Key == lst
		case *syntax.ArrayItem:
			n = p
		default:
			return false
		}
	}
}
