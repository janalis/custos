// Package splfixedarrayshrinkdiscardsvalues implements the native SplFixedArrayShrinkDiscardsValues inspection.
package splfixedarrayshrinkdiscardsvalues

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve populated entries before shrinking."

// readers are the SplFixedArray methods that leave every slot untouched.
var readers = map[string]bool{
	"getsize": true, "offsetget": true, "offsetexists": true, "toarray": true, "count": true,
	"current": true, "key": true, "next": true, "rewind": true, "valid": true,
	"getiterator": true, "jsonserialize": true,
}

// inspectors are builtins that read an object argument without keeping it,
// changing it or handing it to user code.
var inspectors = map[string]bool{
	"count": true, "sizeof": true, "iterator_count": true, "iterator_to_array": true,
	"var_dump": true, "print_r": true, "var_export": true,
}

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SplFixedArrayShrinkDiscardsValues" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "SplFixedArray", "setSize") {
		return
	}
	size, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 0, "size"))
	old, ok := semanticquery.NativeFixedArraySize(ctx, c.Var, c)
	if !known || !ok || size < 0 || size >= old {
		return
	}
	slots := fixedArraySlots(ctx, c.Var, c)
	if v, ok := c.Var.(*syntax.Variable); ok && !replay(ctx, v, c, slots) {
		return
	}
	for key := range slots {
		if key >= size {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// replay walks every earlier use of the receiver variable in source order,
// applies the slot writes and releases to slots, and reports whether the
// occupancy is still proven when the shrink runs.
func replay(ctx *analysis.Context, recv *syntax.Variable, at *syntax.MethodCall, slots map[int64]bool) bool {
	proven := true
	visit := func(n syntax.Node) bool {
		if n.Span().Start >= at.Span().Start {
			return false
		}
		switch x := n.(type) {
		case *syntax.Closure:
			// A capturing closure may release slots whenever it runs.
			for _, use := range x.Uses {
				proven = proven && use.Var.Name != recv.Name
			}
		case *syntax.ArrowFunction:
			syntax.Inspect(x.Expr, func(inner syntax.Node) bool {
				v, ok := inner.(*syntax.Variable)
				proven = proven && !(ok && v.Name == recv.Name)
				return true
			})
		case *syntax.Variable:
			if x.Name == recv.Name {
				proven = proven && step(ctx, x, at, slots)
			}
		}
		return !syntax.IsVariableScope(n)
	}
	if scope := syntax.EnclosingVariableScope(at); scope != nil {
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return proven
}

// step applies one use of the receiver variable; false means the use may
// change the occupancy in a way that cannot be followed.
func step(ctx *analysis.Context, v *syntax.Variable, at *syntax.MethodCall, slots map[int64]bool) bool {
	same := semanticquery.NativeSameValue(ctx, v, at.Var)
	switch p := v.Parent().(type) {
	case *syntax.Assign:
		return p.Var == v && !p.ByRef
	case *syntax.Foreach:
		return p.Expr == v
	case *syntax.Clone:
		return true
	case *syntax.Arg:
		call, ok := p.Parent().Parent().(*syntax.FuncCall)
		return ok && !p.Unpack && !p.ByRef && inspectors[strings.ToLower(semanticquery.NativeBuiltinName(ctx, call))]
	case *syntax.ArrayDimFetch:
		if p.Var != v {
			return false
		}
		slot, known := semanticquery.NativeInt(ctx, p.Dim)
		switch owner := p.Parent().(type) {
		case *syntax.Assign:
			if owner.Var == p && known && same && stateDominates(owner, at) {
				slots[slot] = true
			}
		case *syntax.Unset:
			if !known || !same {
				return false
			}
			delete(slots, slot)
		}
		return true
	case *syntax.MethodCall:
		name, ok := p.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		arg, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(p.Args, 0, "index"))
		switch method := strings.ToLower(name.Value); method {
		case "setsize":
			// Mirror the earlier resize: it already dropped the slots past its bound.
			if !same || !semanticquery.NativeDominates(p, at) {
				return false
			}
			size, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(p.Args, 0, "size"))
			for key := range slots {
				if !known || key >= size {
					delete(slots, key)
				}
			}
			return true
		case "offsetunset":
			if !known || !same {
				return false
			}
			delete(slots, arg)
			return true
		case "offsetset":
			if known && same && semanticquery.NativeDominates(p, at) {
				slots[arg] = true
			}
			return true
		default:
			return readers[method]
		}
	}
	return false
}

// fixedArraySlots tracks explicit local population and unsets.
func fixedArraySlots(ctx *analysis.Context, e syntax.Expr, at syntax.Node) map[int64]bool {
	slots := map[int64]bool{}
	if c, ok := semanticquery.NativeValue(ctx, e).(*syntax.StaticCall); ok {
		entries, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(c.Args, 0, "array"))
		if known {
			keep := true
			if preserve, state := semanticquery.CallArgumentBinding(c.Args, 1, "preserveKeys"); state != semanticquery.ArgumentMissing {
				var known bool
				if keep, known = semanticquery.NativeTruth(ctx, preserve); !known {
					return slots
				}
			}
			if !keep {
				for i := range len(entries) {
					slots[int64(i)] = true
				}
			} else {
				for key := range entries {
					if strings.HasPrefix(key, "i:") {
						i, err := strconv.ParseInt(key[2:], 10, 64)
						if err == nil {
							slots[i] = true
						}
					}
				}
			}
		}
	}
	scope := syntax.EnclosingVariableScope(at)
	if scope == nil {
		scope = ctx.File.Stmts[0].Parent()
	}
	scan := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Assign:
			if slot, ok := x.Var.(*syntax.ArrayDimFetch); ok && semanticquery.NativeSameValue(ctx, slot.Var, e) && stateDominates(x, at) {
				i, k := semanticquery.NativeInt(ctx, slot.Dim)
				if k {
					slots[i] = true
				}
			}
		case *syntax.MethodCall:
			if semanticquery.NativeMethod(ctx, x, "SplFixedArray", "setSize") && semanticquery.NativeSameValue(ctx, x.Var, e) && semanticquery.NativeDominates(x, at) {
				size, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(x.Args, 0, "size"))
				if !known {
					clear(slots)
				} else if size >= 0 {
					for key := range slots {
						if key >= size {
							delete(slots, key)
						}
					}
				}
			}
		case *syntax.Unset:
			for _, v := range x.Vars {
				if slot, ok := v.(*syntax.ArrayDimFetch); ok && semanticquery.NativeSameValue(ctx, slot.Var, e) && stateDominates(x, at) {
					i, k := semanticquery.NativeInt(ctx, slot.Dim)
					if k {
						delete(slots, i)
					}
				}
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
	return slots
}

// stateDominates permits ordinary lifecycle statement effects such as
// unset, in addition to expression statements.
func stateDominates(before, at syntax.Node) bool {
	if semanticquery.NativeDominates(before, at) {
		return true
	}
	if before.Span().End > at.Span().Start {
		return false
	}
	stmt, ok := before.(syntax.Stmt)
	if !ok {
		return false
	}
	owner := stmt.Parent()
	if _, ok := syntax.StmtListOf(owner); !ok {
		return false
	}
	for p := at; p != nil; p = p.Parent() {
		if p == owner {
			return true
		}
	}
	return false
}
