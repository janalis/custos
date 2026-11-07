package flow

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func (c *pureCompiler) byteTransform(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	size, roles := 2, 2
	kind := pureRepeat
	if name == "str_replace" {
		size, roles, kind = 4, 3, pureReplace
	}
	if len(params) != size {
		return nil
	}
	binding := make([]pureParam, roles)
	for i := range binding {
		binding[i].name = params[i].Name
	}
	args := c.arguments(node, binding, params[:roles])
	if args == nil {
		return nil
	}
	var third *pureExpr
	if kind == pureReplace {
		third = args[2]
	}
	result := pureNode(kind, args[0], args[1], third)
	if result != nil && (kind != pureReplace || third != nil) {
		result.projects = true
		return result
	}
	return nil
}

func (e *Env) evalByteTransform(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	right, ok := e.evalPure(expr.right, slots, r, depth-1)
	a, ak := knownScalar(left)
	b, bk := knownScalar(right)
	if expr.kind == pureRepeat {
		if !ok || !ak || !bk || a.k != 's' || len(a.s) > 4096 || b.k != 'i' || b.i < 0 || len(left.Sources) > MaxSources {
			return Value{}, false
		}
		if a.s != "" && b.i > 4096/int64(len(a.s)) {
			return Value{}, false
		}
		size := 0
		if a.s != "" {
			size = len(a.s) * int(b.i)
		}
		r.operations += len(a.s) + size
		r.copies += size
		if r.operations > MaxTransfers || r.copies > MaxCopied {
			return Value{}, false
		}
		value := scalarValue(constantScalar{k: 's'})
		if size != 0 {
			value = scalarValue(constantScalar{k: 's', s: strings.Repeat(a.s, int(b.i))})
			value.Sources = left.Sources
		}
		return value, true
	}
	subject, sk := e.evalPure(expr.other, slots, r, depth-1)
	text, tk := knownScalar(subject)
	if !ok || !sk || !ak || !bk || !tk || a.k != 's' || b.k != 's' || text.k != 's' || a.s == "" || len(a.s) > 4096 || len(b.s) > 4096 || len(text.s) > 4096 || len(subject.Sources) > MaxSources || len(right.Sources) > MaxSources {
		return Value{}, false
	}
	count := strings.Count(text.s, a.s)
	size := len(text.s) + count*(len(b.s)-len(a.s))
	r.operations += len(a.s) + len(b.s) + len(text.s) + size
	r.copies += size
	if size > 4096 || r.operations > MaxTransfers || r.copies > MaxCopied {
		return Value{}, false
	}
	sources := subject.Sources
	if count != 0 {
		var complete bool
		sources, complete = union(sources, right.Sources)
		if !complete {
			return Value{}, false
		}
	}
	value := scalarValue(constantScalar{k: 's', s: strings.ReplaceAll(text.s, a.s, b.s)})
	value.Sources = sources
	return value, true
}
