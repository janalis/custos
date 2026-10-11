package unsupportedstringoffsetoperations

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// unsupportedStringOffsetOperations reports nested offset writes and `[]`
// appends on string values, fatal since PHP 7.1.
type unsupportedStringOffsetOperations struct{}

func (unsupportedStringOffsetOperations) ID() string { return "UnsupportedStringOffsetOperations" }
func (unsupportedStringOffsetOperations) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KArrayDimFetch}
}

var superglobals = map[string]bool{
	"_GET": true, "_POST": true, "_SESSION": true, "_REQUEST": true, "_FILES": true,
	"_COOKIE": true, "_ENV": true, "_SERVER": true, "GLOBALS": true, "HTTP_RAW_POST_DATA": true,
}

// usoParentDim returns the array access having e as its base, or nil.
func usoParentDim(e syntax.Node) *syntax.ArrayDimFetch {
	if p, ok := e.Parent().(*syntax.ArrayDimFetch); ok && p.Var == e {
		return p
	}
	return nil
}

// usoIsWriteTarget reports whether t is on the writing side of an assignment
// (directly, or through a destructuring pattern when allowPattern is set).
func usoIsWriteTarget(t syntax.Expr, allowPattern bool) bool {
	ctxNode := t.Parent()
	var child syntax.Node = t
	if allowPattern {
		if item, ok := ctxNode.(*syntax.ArrayItem); ok {
			switch pat := item.Parent().(type) {
			case *syntax.Array, *syntax.List:
				ctxNode, child = pat.Parent(), pat
			}
		}
	}
	a, ok := ctxNode.(*syntax.Assign)
	return ok && a.Var == child
}

// usoType is the type of the accessed base, except that an offset read on a
// string is not considered a string.
func usoType(ctx *analysis.Context, c syntax.Expr) types.Type {
	if d, ok := c.(*syntax.ArrayDimFetch); ok && ctx.TypeOf(d.Var).OnlyOf("string") {
		return types.Unknown
	}
	return ctx.TypeOf(c)
}

func (unsupportedStringOffsetOperations) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP71 { // E1
		return
	}
	e := n.(*syntax.ArrayDimFetch)
	c := e.Var
	switch c.(type) { // D1
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch:
	default:
		return
	}
	root := usoRoot(c) // D2
	rootVar, _ := root.(*syntax.Variable)
	if rootVar != nil && rootVar.NameExpr == nil && superglobals[rootVar.Name] {
		return
	}
	var target syntax.Expr
	var msg string
	if p := usoParentDim(e); p != nil { // D3
		t := p
		for q := usoParentDim(t); q != nil; q = usoParentDim(t) {
			t = q
		}
		if !usoIsWriteTarget(t, true) {
			return
		}
		target, msg = t, "String offsets cannot be written as nested arrays (fatal error)."
	} else if e.Dim == nil && usoIsWriteTarget(e, false) { // D4
		target, msg = e, "Appending with [] is not supported on strings (fatal error)."
	} else {
		return
	}
	fn := syntax.EnclosingFuncLike(target)
	if fn == nil { // D5
		return
	}
	if t := usoType(ctx, c); !t.Equal(types.String) { // D6
		return
	}
	if _, ok := c.(*syntax.ArrayDimFetch); ok && !usoDeclaredRoot(root, target) {
		return
	}
	if rootVar != nil && usoReferenced(ctx, fn, rootVar.Name) { // D7
		return
	}
	ctx.ReportNode(target, msg)
}

// usoRoot descends through nested array-access bases to the accessed root.
func usoRoot(e syntax.Expr) syntax.Expr {
	for {
		d, ok := e.(*syntax.ArrayDimFetch)
		if !ok {
			return e
		}
		e = d.Var
	}
}

// usoReferenced reports whether $name (or one of its offsets) is bound to a
// reference somewhere in fn: a reference assignment, a by-reference array
// item, or an argument of a by-reference parameter such as
// `settype($s, 'array')`. Whoever holds the reference can replace the string
// with an array, so the declared type no longer proves anything.
func usoReferenced(ctx *analysis.Context, fn syntax.Node, name string) bool {
	is := func(e syntax.Expr) bool {
		v, ok := usoRoot(e).(*syntax.Variable)
		return ok && v.NameExpr == nil && v.Name == name
	}
	found := false
	syntax.Inspect(fn, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Assign:
			found = found || x.ByRef && (is(x.Var) || is(x.Value))
		case *syntax.ArrayItem:
			found = found || x.ByRef && is(x.Value)
		case *syntax.ArgList:
			callee, _ := semanticquery.ResolveCallee(ctx, x.Parent())
			for i, a := range x.Args {
				if arg, ok := a.(*syntax.Arg); ok && is(arg.Value) && (arg.ByRef || usoByRefParam(callee.Params, arg, i)) {
					found = true
				}
			}
		}
		return !found
	})
	return found
}

// usoByRefParam reports whether the argument at position i binds a
// by-reference parameter of params (by name, by position, or through a
// variadic last parameter).
func usoByRefParam(params []index.Param, arg *syntax.Arg, i int) bool {
	if arg.Name != nil {
		for _, p := range params {
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
				return p.ByRef
			}
		}
		return false
	}
	if n := len(params); i >= n && n > 0 && params[n-1].Variadic {
		i = n - 1
	}
	return i < len(params) && params[i].ByRef
}

// usoDeclaredRoot reports whether the root of an offset chain has a declared
// type (property, or parameter of the enclosing function). A local variable
// is typed from its assignments, e.g. an array literal whose element type
// says nothing about keys the literal lacks: `$a = ['k' => 's'];
// $a['x'][] = 1;` autovivifies `$a['x']` as an array (custos).
func usoDeclaredRoot(root syntax.Expr, at syntax.Node) bool {
	v, ok := root.(*syntax.Variable)
	if !ok {
		return true // property or static property
	}
	for _, p := range syntax.FuncLikeParams(syntax.EnclosingFuncLike(at)) {
		if p.Var != nil && p.Var.Name == v.Name {
			return true
		}
	}
	return false
}
