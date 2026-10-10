package semanticquery

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// NativeConstruction resolves an unchanged locally constructed object.
func NativeConstruction(ctx *analysis.Context, e syntax.Expr, class string) *syntax.New {
	v, ok := NativeLocalValue(ctx, e).(*syntax.New)
	if !ok {
		return nil
	}
	n, ok := v.Class.(*syntax.Name)
	if !ok {
		return nil
	}
	fqn := ctx.Names().Class(n.Value, n.Span().Start)
	c := ctx.Index().Class(fqn, ctx.PHP)
	if c == nil || !strings.HasPrefix(c.File, "stubs/") || !strings.EqualFold(fqn, class) {
		return nil
	}
	return v
}

// NativeNewClass resolves the declaration instantiated by a lexical class name
// or a proven constant class string. Dynamic and ambiguous names stay unknown.
func NativeNewClass(ctx *analysis.Context, n *syntax.New) *index.Class {
	if name, ok := n.Class.(*syntax.Name); ok {
		return ctx.Index().Class(ctx.Names().Class(name.Value, name.Span().Start), ctx.PHP)
	}
	if name, ok := astquery.QuotedStringValue(NativeLocalValue(ctx, n.Class)); ok {
		return ctx.Index().Class(strings.TrimPrefix(name, `\`), ctx.PHP)
	}
	return nil
}

// NativeSameValue establishes object identity without source-name comparisons.
func NativeSameValue(ctx *analysis.Context, a, b syntax.Expr) bool {
	if a == nil || b == nil {
		return false
	}
	x, y := ctx.Flow().Value(a), ctx.Flow().Value(b)
	if x.Complete && y.Complete && x.Identity != 0 && x.Identity == y.Identity && x.Invalidated == y.Invalidated {
		return true
	}
	av, bv := NativeLocalValue(ctx, a), NativeLocalValue(ctx, b)
	return av != nil && av == bv
}

// NativeConditionUse returns the truthiness expression containing a call.
// Integer sentinel comparisons are deliberately excluded.
func NativeConditionUse(ctx *analysis.Context, call *syntax.FuncCall, booleanComparisons bool) syntax.Expr {
	var e syntax.Expr = call
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			e = p
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return nil
			}
			e = p
		case *syntax.Binary:
			if !booleanComparisons {
				return nil
			}
			if p.Op.Kind != syntax.TIsEqual && p.Op.Kind != syntax.TIsNotEqual && p.Op.Kind != syntax.TIsIdentical && p.Op.Kind != syntax.TIsNotIdentical {
				return nil
			}
			other := p.Right
			if other == e {
				other = p.Left
			}
			c, ok := other.(*syntax.ConstFetch)
			if !ok || c.Name == nil || (!strings.EqualFold(c.Name.Value, "true") && !strings.EqualFold(c.Name.Value, "false")) {
				return nil
			}
			e = p
		case *syntax.If:
			return e
		case *syntax.While:
			return e
		case *syntax.DoWhile:
			return e
		default:
			return nil
		}
	}
}

// NativeAttributeFlags resolves the independently declared attribute contract.
func NativeAttributeFlags(ctx *analysis.Context, a *syntax.Attribute) (string, int64, bool) {
	fqn := ctx.Names().Class(a.Name.Value, a.Name.Span().Start)
	d := ClassDecl(ctx.File, ctx.Index().Class(fqn, ctx.PHP))
	if d == nil {
		return fqn, 0, false
	}
	for _, g := range d.Attrs {
		for _, mark := range g.Attrs {
			name := ctx.Names().Class(mark.Name.Value, mark.Name.Span().Start)
			c := ctx.Index().Class(name, ctx.PHP)
			if !strings.EqualFold(name, "Attribute") || c == nil || !strings.HasPrefix(c.File, "stubs/") {
				continue
			}
			arg := CallArgument(mark.Args, 0, "flags")
			if arg == nil {
				if ctx.PHP >= phpversion.PHP85 {
					return fqn, 127, true
				}
				return fqn, 63, true
			}
			n, ok := NativeContractInt(ctx, arg)
			return fqn, n, ok
		}
	}
	return fqn, 0, false
}

// NativeReflectionTarget resolves a known reflected callable declaration.
func NativeReflectionTarget(ctx *analysis.Context, e syntax.Expr) ([]index.Param, string, bool) {
	if c := NativeConstruction(ctx, e, "ReflectionFunction"); c != nil {
		name, ok := NativeString(ctx, CallArgument(c.Args, 0, "function"))
		if !ok {
			return nil, "", false
		}
		f := ctx.Index().Function(strings.TrimPrefix(name, `\`), ctx.PHP)
		if f != nil {
			return f.Params, f.Return, true
		}
		return nil, "", false
	}
	if c := NativeConstruction(ctx, e, "ReflectionMethod"); c != nil {
		target := CallArgument(c.Args, 0, "objectOrMethod")
		name, ok := NativeString(ctx, CallArgument(c.Args, 1, "method"))
		if !ok {
			return nil, "", false
		}
		class, ok := NativeString(ctx, target)
		if !ok {
			classes := ctx.Types().Native().TypeOf(target).Classes()
			if len(classes) != 1 {
				return nil, "", false
			}
			class = classes[0]
		}
		m := ctx.Index().FindMethod(class, name, ctx.PHP)
		if m != nil {
			return m.Params, m.Return, true
		}
	}
	return nil, "", false
}

// NativeFixedArraySize computes an exact local fixed-array bound.
func NativeFixedArraySize(ctx *analysis.Context, e syntax.Expr, at syntax.Node) (int64, bool) {
	var size int64
	known := false
	if c := NativeConstruction(ctx, e, "SplFixedArray"); c != nil {
		arg := CallArgument(c.Args, 0, "size")
		if arg == nil {
			size = 0
			known = true
		} else {
			size, known = NativeInt(ctx, arg)
		}
	} else if c, ok := NativeValue(ctx, e).(*syntax.StaticCall); ok {
		cl, cok := c.Class.(*syntax.Name)
		method, mok := c.Name.(*syntax.Identifier)
		if cok && mok && strings.EqualFold(ctx.Names().Class(cl.Value, cl.Span().Start), "SplFixedArray") && strings.EqualFold(method.Value, "fromArray") {
			entries, ok := NativeArrayEntries(ctx, CallArgument(c.Args, 0, "array"))
			if !ok {
				return 0, false
			}
			keep := CallArgument(c.Args, 1, "preserveKeys")
			if keep != nil {
				v, k := NativeTruth(ctx, keep)
				if !k {
					return 0, false
				}
				if !v {
					size = int64(len(entries))
					known = true
				}
			}
			if !known {
				known = true
				for key := range entries {
					if !strings.HasPrefix(key, "i:") {
						return 0, false
					}
					i, err := strconv.ParseInt(key[2:], 10, 64)
					if err != nil || i < 0 {
						return 0, false
					}
					if i >= size {
						size = i + 1
					}
				}
			}
		}
	}
	if !known || size < 0 {
		return 0, false
	}
	for _, c := range NativeMethodStateCalls(ctx, at, e, "setSize") {
		size, known = NativeInt(ctx, CallArgument(c.Args, 0, "size"))
		if !known || size < 0 {
			return 0, false
		}
	}
	return size, true
}

// NativeFixedArraySlots tracks explicit local population and unsets.
func NativeFixedArraySlots(ctx *analysis.Context, e syntax.Expr, at syntax.Node) map[int64]bool {
	slots := map[int64]bool{}
	if c, ok := NativeValue(ctx, e).(*syntax.StaticCall); ok {
		entries, known := NativeArrayEntries(ctx, CallArgument(c.Args, 0, "array"))
		if known {
			preserve := CallArgument(c.Args, 1, "preserveKeys")
			keep := true
			if preserve != nil {
				keep, _ = NativeTruth(ctx, preserve)
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
			if slot, ok := x.Var.(*syntax.ArrayDimFetch); ok && NativeSameValue(ctx, slot.Var, e) && NativeStateDominates(x, at) {
				i, k := NativeInt(ctx, slot.Dim)
				if k {
					slots[i] = true
				}
			}
		case *syntax.MethodCall:
			if NativeMethod(ctx, x, "SplFixedArray", "setSize") && NativeSameValue(ctx, x.Var, e) && NativeDominates(x, at) {
				size, known := NativeInt(ctx, CallArgument(x.Args, 0, "size"))
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
				if slot, ok := v.(*syntax.ArrayDimFetch); ok && NativeSameValue(ctx, slot.Var, e) && NativeStateDominates(x, at) {
					i, k := NativeInt(ctx, slot.Dim)
					if k {
						delete(slots, i)
					}
				}
			}
		}
		return true
	}
	for _, event := range nativeContractEvents(ctx, scope) {
		if event.Span().Start >= at.Span().Start {
			break
		}
		scan(event)
	}
	return slots
}

// NativeStoredArray proves the last direct storage assignment is an array.
func NativeStoredArray(ctx *analysis.Context, slot *syntax.ArrayDimFetch, at syntax.Node) bool {
	found := false
	scope := syntax.EnclosingVariableScope(at)
	scan := func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok && NativeDominates(a, at) {
			if s, ok := a.Var.(*syntax.ArrayDimFetch); ok && NativeSameValue(ctx, s.Var, slot.Var) && NativeSameValue(ctx, s.Dim, slot.Dim) {
				found = NativeArray(ctx, a.Value) != nil
			}
		}
		return true
	}
	for _, event := range nativeContractEvents(ctx, scope) {
		if event.Span().Start >= at.Span().Start {
			break
		}
		scan(event)
	}
	return found
}

// NativeStreamMode proves a guarded successful local fopen mode.
func NativeStreamMode(ctx *analysis.Context, h syntax.Expr, at syntax.Node) (string, bool) {
	c, ok := NativeValue(ctx, h).(*syntax.FuncCall)
	if !ok || !NativeBuiltin(ctx, c, "fopen") {
		return "", false
	}
	mode, known := NativeString(ctx, CallArgument(c.Args, 1, "mode"))
	if !known || mode == "" {
		return "", false
	}
	if path, k := NativeString(ctx, CallArgument(c.Args, 0, "filename")); k && strings.Contains(path, "://") && !strings.HasPrefix(path, "file://") && !strings.HasPrefix(path, "php://memory") && !strings.HasPrefix(path, "php://temp") {
		return "", false
	}
	if path, known := NativeString(ctx, CallArgument(c.Args, 0, "filename")); known && (strings.HasPrefix(path, "php://memory") || strings.HasPrefix(path, "php://temp")) {
		if mode[0] == 'a' {
			mode = "a+"
		} else if mode[0] == 'w' || strings.Contains(mode, "+") {
			mode = "w+"
		} else {
			mode = "r"
		}
	}
	if !ctx.Flow().Value(h).NonFalse && !NativeSentinelGuard(ctx, h, "false") {
		return "", false
	}
	if len(NativeStreamCalls(ctx, at, h, "fclose")) > 0 {
		return "", false
	}
	return mode, true
}

// NativeCallSucceeded recognizes a direct successful condition dominating at.
func NativeCallSucceeded(ctx *analysis.Context, c *syntax.FuncCall, at syntax.Node) bool {
	accepts := func(cond syntax.Expr, truth bool) bool {
		e := syntax.UnwrapParens(cond)
		if e == c {
			return truth && NativeBuiltinName(ctx, c) != "fseek"
		}
		if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim && syntax.UnwrapParens(u.Expr) == c {
			return !truth && NativeBuiltinName(ctx, c) != "fseek"
		}
		if b, ok := e.(*syntax.Binary); ok {
			equality := (truth && (b.Op.Kind == syntax.TIsIdentical || b.Op.Kind == syntax.TIsEqual)) || (!truth && (b.Op.Kind == syntax.TIsNotIdentical || b.Op.Kind == syntax.TIsNotEqual))
			if !equality {
				return false
			}
			for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
				if syntax.UnwrapParens(pair[0]) != c {
					continue
				}
				if NativeBuiltin(ctx, c, "fseek") {
					v, k := NativeInt(ctx, pair[1])
					return k && v == 0
				}
				v, k := NativeTruth(ctx, pair[1])
				return k && v
			}
		}
		return false
	}
	for p := at; p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if branch, ok := p.(*syntax.If); ok && branch.Body.Span().Contains(at.Span()) && accepts(branch.Cond, true) {
			return true
		}
		if st, ok := p.(syntax.Stmt); ok {
			prev, exists := astquery.PrevStmt(ctx.File, st)
			if guard, ok := prev.(*syntax.If); exists && ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && accepts(guard.Cond, false) {
				return true
			}
		}
	}
	return false
}

// NativeStreamCalls includes guarded stream operations preceding the use.
func NativeStreamCalls(ctx *analysis.Context, at syntax.Node, h syntax.Expr, names ...string) []*syntax.FuncCall {
	var out []*syntax.FuncCall
	scope := syntax.EnclosingVariableScope(at)
	for _, event := range nativeContractEvents(ctx, scope) {
		if event.Span().Start >= at.Span().Start {
			break
		}
		c, ok := event.(*syntax.FuncCall)
		if !ok || c.Span().End >= at.Span().Start || (!NativeDominates(c, at) && !NativeCallSucceeded(ctx, c, at)) || !NativeSameValue(ctx, CallArgument(c.Args, 0, "stream"), h) {
			continue
		}
		for _, name := range names {
			if NativeBuiltin(ctx, c, name) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// NativeCSVSignature identifies the separator slot of each resolved CSV API.
func NativeCSVSignature(ctx *analysis.Context, n syntax.Node) (*syntax.ArgList, int, string, bool) {
	if c, ok := n.(*syntax.FuncCall); ok {
		switch NativeBuiltinName(ctx, c) {
		case "fgetcsv":
			return c.Args, 2, "separator", true
		case "fputcsv":
			return c.Args, 2, "separator", true
		case "str_getcsv":
			return c.Args, 1, "separator", true
		}
		return nil, 0, "", false
	}
	if c, ok := n.(*syntax.MethodCall); ok {
		for _, name := range []string{"fgetcsv", "fputcsv", "setCsvControl"} {
			if NativeMethod(ctx, c, "SplFileObject", name) {
				pos := 0
				if name == "fputcsv" {
					pos = 1
				}
				return c.Args, pos, "separator", true
			}
		}
	}
	return nil, 0, "", false
}

// NativeContractInt includes resolved integer class constants and flag unions.
func NativeContractInt(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if b, ok := syntax.UnwrapParens(e).(*syntax.Binary); ok && b.Op.Kind == syntax.TBar {
		l, lk := NativeContractInt(ctx, b.Left)
		r, rk := NativeContractInt(ctx, b.Right)
		return l | r, lk && rk
	}
	if v, k := NativeInt(ctx, e); k {
		return v, true
	}
	v := NativeValue(ctx, e)
	if v == nil {
		v = e
	}
	if b, ok := v.(*syntax.Binary); ok && b.Op.Kind == syntax.TBar {
		l, lk := NativeContractInt(ctx, b.Left)
		r, rk := NativeContractInt(ctx, b.Right)
		return l | r, lk && rk
	}
	if c, ok := v.(*syntax.ClassConstFetch); ok {
		cl, ok := c.Class.(*syntax.Name)
		name, nok := c.Name.(*syntax.Identifier)
		if ok && nok {
			decl := ctx.Index().FindConst(ctx.Names().Class(cl.Value, cl.Span().Start), name.Value, ctx.PHP)
			if decl != nil {
				owner := ctx.Index().Class(ctx.Names().Class(cl.Value, cl.Span().Start), ctx.PHP)
				if owner != nil && strings.EqualFold(owner.FQN, "Attribute") && strings.HasPrefix(owner.File, "stubs/") && ctx.PHP < phpversion.PHP85 {
					switch name.Value {
					case "IS_REPEATABLE":
						return 64, true
					case "TARGET_ALL":
						return 63, true
					case "TARGET_CONSTANT":
						return 0, false
					}
				}
				n, err := strconv.ParseInt(decl.Value, 10, 64)
				return n, err == nil
			}
		}
	}
	return 0, false
}

// NativeLocalValue recovers direct local assignments for lvalues and dynamic
// callees omitted by the general value graph. It rejects ambiguous writes,
// by-reference aliases and calls that may replace the tracked variable.
func NativeLocalValue(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	known := NativeValue(ctx, e)
	variable, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		if known != nil {
			return known
		}
		return e
	}
	scope := syntax.EnclosingVariableScope(e)
	var value syntax.Expr
	writes := 0
	scan := func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok {
			if v, ok := a.Var.(*syntax.Variable); ok && v.Name == variable.Name {
				if a.Value.Span().End >= e.Span().Start {
					return true
				}
				if a.ByRef || a.Op.Kind != syntax.TEqual || !NativeDominates(a, e) {
					value = nil
					writes++
					return true
				}
				value = a.Value
				writes++
			}
		}
		if call, ok := n.(*syntax.MethodCall); ok && call.Span().End < e.Span().Start {
			if receiver, ok := call.Var.(*syntax.Variable); ok && receiver.Name == variable.Name {
				classes := ctx.TypeOf(call.Var).Classes()
				builtin := false
				if len(classes) == 1 {
					if method, ok := call.Name.(*syntax.Identifier); ok {
						decl := ctx.Index().FindMethod(classes[0], method.Value, ctx.PHP)
						builtin = decl != nil && decl.Builtin
					}
				}
				if !builtin {
					value = nil
				}
			}
		}
		if call, ok := n.(*syntax.FuncCall); ok && call.Span().End < e.Span().Start && NativeBuiltinName(ctx, call) == "" {
			syntax.Inspect(call.Args, func(a syntax.Node) bool {
				if v, ok := a.(*syntax.Variable); ok && v.Name == variable.Name {
					value = nil
				}
				return true
			})
		}
		return true
	}
	for _, event := range nativeContractEvents(ctx, scope) {
		if event.Span().Start >= e.Span().Start {
			break
		}
		scan(event)
	}
	if writes == 0 {
		return known
	}
	if v, ok := value.(*syntax.Variable); ok {
		return NativeLocalValue(ctx, v)
	}
	return value
}

// NativeStateDominates permits ordinary lifecycle statement effects such as
// unset, in addition to expression statements.
func NativeStateDominates(before, at syntax.Node) bool {
	if NativeDominates(before, at) {
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
	if owner == nil {
		return syntax.EnclosingVariableScope(at) == nil
	}
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

// NativeLocalSentinelGuard recognizes lexical guards for a callable variable
// whose dynamic invocation is not included in the general value graph.
func NativeLocalSentinelGuard(ctx *analysis.Context, e syntax.Expr, sentinel string) bool {
	if NativeSentinelGuard(ctx, e, sentinel) {
		return true
	}
	return nativeGuard(ctx, e, func(cond syntax.Expr, truth bool) bool {
		b, ok := syntax.UnwrapParens(cond).(*syntax.Binary)
		if !ok || !((truth && b.Op.Kind == syntax.TIsNotIdentical) || (!truth && b.Op.Kind == syntax.TIsIdentical)) {
			return false
		}
		for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
			c, ok := pair[1].(*syntax.ConstFetch)
			if ok && c.Name != nil && strings.EqualFold(c.Name.Value, sentinel) && NativeSameValue(ctx, e, pair[0]) {
				return true
			}
		}
		return false
	})
}

type nativeContractEventKey struct{ scope syntax.Node }

// Cache lexical state events once per parsed scope; larger scopes deliberately
// establish no additional local facts rather than repeatedly scanning them.
func nativeContractEvents(ctx *analysis.Context, scope syntax.Node) []syntax.Node {
	return ctx.File.Memo(nativeContractEventKey{scope}, func() any {
		nodes := []syntax.Node{}
		visit := func(n syntax.Node) bool {
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
			switch n.(type) {
			case *syntax.Assign, *syntax.Unset, *syntax.FuncCall, *syntax.MethodCall:
				nodes = append(nodes, n)
			}
			return true
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		if len(nodes) > 512 {
			return []syntax.Node(nil)
		}
		return nodes
	}).([]syntax.Node)
}

// NativeMethodStateCalls returns direct local method effects on the same
// proven allocation, including factory allocations without graph identities.
func NativeMethodStateCalls(ctx *analysis.Context, at syntax.Node, e syntax.Expr, names ...string) []*syntax.MethodCall {
	var out []*syntax.MethodCall
	for _, n := range nativeContractEvents(ctx, syntax.EnclosingVariableScope(at)) {
		if n.Span().Start >= at.Span().Start {
			break
		}
		c, ok := n.(*syntax.MethodCall)
		if !ok {
			continue
		}
		id, known := c.Name.(*syntax.Identifier)
		if !known || !NativeDominates(c, at) || !NativeSameValue(ctx, e, c.Var) {
			continue
		}
		for _, name := range names {
			if strings.EqualFold(id.Value, name) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}
