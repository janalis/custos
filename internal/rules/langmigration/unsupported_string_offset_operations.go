package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

// unsupportedStringOffsetOperations reports nested offset writes and `[]`
// appends on string values, fatal since PHP 7.1.
type unsupportedStringOffsetOperations struct{}

func init() { register(unsupportedStringOffsetOperations{}) }

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
	if ctx.PHP < phpver.PHP71 { // E1
		return
	}
	e := n.(*syntax.ArrayDimFetch)
	c := e.Var
	switch c.(type) { // D1
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch:
	default:
		return
	}
	root := c // D2
	for {
		d, ok := root.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		root = d.Var
	}
	if v, ok := root.(*syntax.Variable); ok && v.NameExpr == nil && superglobals[v.Name] {
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
	if util.EnclosingFuncLike(target) == nil { // D5
		return
	}
	if t := usoType(ctx, c); !t.Equal(types.String) { // D6
		return
	}
	ctx.ReportNode(target, msg)
}
