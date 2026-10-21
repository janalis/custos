package semanticquery

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
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
	if c == nil || !c.Avail.In(ctx.PHP) || !strings.HasPrefix(c.File, "stubs/") || !strings.EqualFold(fqn, class) {
		return nil
	}
	return v
}

// NativeNewClass resolves the declaration instantiated by a lexical class name
// or a proven constant class string. Dynamic and ambiguous names stay unknown.
func NativeNewClass(ctx *analysis.Context, n *syntax.New) *index.Class {
	if name, ok := n.Class.(*syntax.Name); ok {
		if strings.EqualFold(name.Value, "self") {
			return nativeSelfClass(ctx, n)
		}
		return nativeAvailableClass(ctx, ctx.Names().Class(name.Value, name.Span().Start))
	}
	if name, ok := astquery.QuotedStringValue(NativeLocalValue(ctx, n.Class)); ok {
		return nativeAvailableClass(ctx, strings.TrimPrefix(name, `\`))
	}
	return nil
}

// nativeSelfClass resolves `self` written directly in a method body of a named
// class, interface or enum. A closure can be rebound to another class, a
// trait or anonymous class has no single declaration to instantiate, and a
// parameter default or attribute argument is not evaluated where it stands.
func nativeSelfClass(ctx *analysis.Context, at syntax.Node) *index.Class {
	for p := at.Parent(); p != nil; p = p.Parent() {
		switch p := p.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.Param, *syntax.Attribute:
			return nil
		case *syntax.ClassLike:
			if p.Name == nil || p.ClassKind == syntax.KindTrait {
				return nil
			}
			return nativeAvailableClass(ctx, ctx.Types().ClassFQN(p))
		}
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
	// The lexical scan names the producer; it does not see a call that may
	// have assigned a static local between the two reads.
	av, bv := NativeLocalValue(ctx, a), NativeLocalValue(ctx, b)
	return av != nil && av == bv && !nativeStaticReplaced(ctx, a, b, x, y)
}

// NativeConditionUse returns the truthiness expression containing a call.
// Integer sentinel comparisons are deliberately excluded.
// A first-class callable creates a closure and has no result to test; only a
// plain assignment of the result keeps its value as the tested operand.
func NativeConditionUse(ctx *analysis.Context, call *syntax.FuncCall, booleanComparisons bool) syntax.Expr {
	if call.Args != nil && len(call.Args.Args) == 1 {
		if _, ok := call.Args.Args[0].(*syntax.VariadicPlaceholder); ok {
			return nil
		}
	}
	var e syntax.Expr = call
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			if p.Op.Kind != syntax.TEqual || p.Value != e {
				return nil
			}
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
		case *syntax.If, *syntax.ElseIf, *syntax.While, *syntax.DoWhile:
			return e
		case *syntax.For:
			// Only the last condition expression decides the loop.
			if n := len(p.Cond); n == 0 || p.Cond[n-1] != e {
				return nil
			}
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
		arg, state := CallArgumentBinding(c.Args, 0, "size")
		if state == ArgumentMissing {
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
			if keep, state := CallArgumentBinding(c.Args, 1, "preserveKeys"); state != ArgumentMissing {
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
			keep := true
			if preserve, state := CallArgumentBinding(c.Args, 1, "preserveKeys"); state != ArgumentMissing {
				var known bool
				if keep, known = NativeTruth(ctx, preserve); !known {
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
func NativeStreamMode(ctx *analysis.Context, h, at syntax.Expr) (string, bool) {
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
		if strings.Contains(mode, "a") {
			mode = "a+"
		} else if strings.ContainsAny(mode, "w+") {
			mode = "w+"
		} else {
			mode = "r"
		}
	}
	if !ctx.Flow().Value(h).NonFalse && !NativeFailureExcluded(ctx, h, types.Of("resource", "false"), "false") {
		return "", false
	}
	if len(NativeStreamCalls(ctx, at, h, "fclose")) > 0 {
		return "", false
	}
	return mode, true
}

// NativeCallSucceeded reports whether call c (a function or method call) is
// known to have succeeded at at: the conditions on that very call guarding at
// rule its failure out. fseek fails with -1 rather than false: its result
// must be known to be 0.
func NativeCallSucceeded(ctx *analysis.Context, c, at syntax.Expr) bool {
	if NativeLabelled(ctx, at) {
		return false
	}
	if f, ok := c.(*syntax.FuncCall); ok && NativeBuiltinName(ctx, f) == "fseek" {
		return ctx.Types().Native().CallResultAsAt(c, at, types.Ints(types.IntRange{Lo: -1, Hi: 0})).IntsOnly(0, 0)
	}
	t := ctx.Types().Native().CallResultAt(c, at)
	return !t.IsUnknown() && !MayHoldAtom(t, "false")
}

// NativeStreamCalls includes guarded stream operations preceding the use.
func NativeStreamCalls(ctx *analysis.Context, at, h syntax.Expr, names ...string) []*syntax.FuncCall {
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
	return nativeLocalValue(ctx, e, "", nil)
}

// NativeLocalResult is NativeLocalValue for an inspection about results that
// fail with one value, "false" or "null": when several assignments reach the
// read, a call result counts as guarded only if that value is excluded on
// its path, whatever the declared return type of the callee says.
func NativeLocalResult(ctx *analysis.Context, e syntax.Expr, failure string) syntax.Expr {
	return nativeLocalValue(ctx, e, failure, nil)
}

// NativeLocalResultOf is NativeLocalResult for an inspection about the
// results of some calls only: when several assignments reach the read, the
// result of a call it does not ask about is passed over, so that an unguarded
// result of another API reaching the same read does not hide one it reports.
func NativeLocalResultOf(ctx *analysis.Context, e syntax.Expr, failure string, produced func(syntax.Expr) bool) syntax.Expr {
	if ctx.Flow().CompressedOrigins(e) {
		return nil
	}
	return nativeLocalValue(ctx, e, failure, produced)
}

// NativeLocalOnly is NativeLocalValue for an inspection that needs the one
// value a variable holds wherever the read runs: nil when the producers of
// several paths reach the read, or one producer and a path that left the
// variable unknown. NativeLocalValue answers such a read with a call result
// no guard covers, which is what a failure check asks, not what the variable
// is known to hold.
func NativeLocalOnly(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	if reaching, single := ctx.Flow().Reaching(e); !single && len(reaching) > 0 {
		return nil
	}
	return NativeLocalValue(ctx, e)
}

func nativeLocalValue(ctx *analysis.Context, e syntax.Expr, failure string, produced func(syntax.Expr) bool) syntax.Expr {
	if ctx.Flow().ProducerBudgetExceeded(e) || ctx.Flow().CompressedOrigins(e) {
		return nil
	}
	known := NativeValue(ctx, e)
	variable, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		if known != nil {
			return known
		}
		return e
	}
	// A resolved variable has exactly one producer. An unrelated known value
	// cannot hide a later producer, so do not scan its lexical write history.
	if produced != nil && known != nil {
		if produced(known) {
			return known
		}
		return nil
	}
	if result := nativeUnguardedResult(ctx, e, failure, produced); result != nil {
		return result
	}
	// Every result the inspection asks about is guarded on its own path:
	// what the other assignments left is not its concern.
	if reaching, single := ctx.Flow().Reaching(e); produced != nil && !single && len(reaching) > 0 {
		return nil
	}
	scope := syntax.EnclosingVariableScope(e)
	facts := nativeContractScope(ctx, scope)
	var value syntax.Expr
	writes := 0
	from := uint32(0)
	scan := func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok {
			if v, ok := a.Var.(*syntax.Variable); ok && v.Name == variable.Name {
				if a.Value.Span().End >= e.Span().Start {
					return true
				}
				from = a.Span().End
				// An assignment written inside a condition does not dominate
				// syntactically; it does when the flow resolves the read to it.
				if a.ByRef || a.Op.Kind != syntax.TEqual || !NativeDominates(a, e) && (known == nil || NativeUnsilenced(a.Value) != known) {
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

		return true
	}
	for _, event := range facts.nodes {
		if event.Span().Start >= e.Span().Start {
			break
		}
		scan(event)
	}
	if writes == 0 {
		return known
	}
	// The scan reads plain assignments only; any other write of the variable
	// after the selected one leaves its value unknown.
	// Element writes change an array or a string, not the object a variable holds.
	content := false
	switch value.(type) {
	case *syntax.Array, *syntax.Literal, *syntax.InterpolatedString:
		content = true
	}
	if value == nil || facts.localWrites(ctx, scope).rewritten(variable.Name, from, e.Span().Start, content) {
		return nil
	}
	if v, ok := value.(*syntax.Variable); ok {
		return nativeLocalValue(ctx, v, failure, produced)
	}
	return NativeUnsilenced(value)
}

// nativeUnguardedResult handles a variable several assignments reach: across
// a loop back-edge, or after a reassignment on some path only. It returns the
// first reaching call result that no guard on its own path separates from its
// failure value; a guard written before another assignment does not vouch for
// that assignment. nil when one assignment reaches the read, or when every
// reaching call result is guarded on its path. failure names the value the
// asking inspection is about; without one, a function result fails with false
// when its declared type can be false, with null otherwise, and a result of
// unknown type is guarded by either exclusion. produced, when given, names
// the call results the asking inspection is about: the others are passed over.
func nativeUnguardedResult(ctx *analysis.Context, e syntax.Expr, failure string, produced func(syntax.Expr) bool) syntax.Expr {
	reaching, single := ctx.Flow().Reaching(e)
	if single {
		return nil
	}
	for _, origin := range reaching {
		result := NativeUnsilenced(origin.Expr)
		if produced != nil && !produced(result) {
			continue
		}
		fails, null := failure == "false", failure == "null"
		switch call := result.(type) {
		case *syntax.FuncCall:
			if callee := ctx.Types().ResolveFunction(call); callee != nil && failure == "" {
				declared := "|" + strings.ToLower(callee.Return) + "|"
				fails = strings.Contains(declared, "|false|")
				null = !fails && (strings.Contains(declared, "|null|") || strings.HasPrefix(declared, "|?"))
			}
		case *syntax.MethodCall, *syntax.StaticCall:
		default:
			continue
		}
		if fails && !origin.NonFalse || null && !origin.NonNull || !fails && !null && !origin.NonFalse && !origin.NonNull {
			return result
		}
	}
	return nil
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

type nativeContractEventKey struct{ scope syntax.Node }

// nativeContractFacts are the lexical facts of one variable scope: its state
// events in source order and, once a query needs them, every place a local
// may be written.
type nativeContractFacts struct {
	nodes  []syntax.Node
	mu     sync.Mutex
	ready  atomic.Bool
	writes nativeLocalWrites
}

// localWrites indexes the scope's writes on first use. Most scopes never ask,
// and a later query reads the finished index without locking or allocating.
func (f *nativeContractFacts) localWrites(ctx *analysis.Context, scope syntax.Node) *nativeLocalWrites {
	if f.ready.Load() {
		return &f.writes
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.ready.Load() {
		f.writes.index(ctx, scope)
		f.ready.Store(true)
	}
	return &f.writes
}

// Cache lexical state events once per parsed scope; larger scopes deliberately
// establish no additional local facts rather than repeatedly scanning them.
func nativeContractEvents(ctx *analysis.Context, scope syntax.Node) []syntax.Node {
	return nativeContractScope(ctx, scope).nodes
}

func nativeContractScope(ctx *analysis.Context, scope syntax.Node) *nativeContractFacts {
	return ctx.File.Memo(nativeContractEventKey{scope}, func() any {
		facts := &nativeContractFacts{nodes: []syntax.Node{}}
		visit := func(n syntax.Node) bool {
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
			switch n.(type) {
			case *syntax.Assign, *syntax.Unset, *syntax.FuncCall, *syntax.MethodCall:
				facts.nodes = append(facts.nodes, n)
			}
			return true
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		if len(facts.nodes) > 512 {
			facts.nodes = nil
		}
		return facts
	}).(*nativeContractFacts)
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

// nativeAvailableClass keeps best-effort index candidates out of runtime
// contracts when the configured PHP version cannot provide the class.
func nativeAvailableClass(ctx *analysis.Context, name string) *index.Class {
	c := ctx.Index().Class(name, ctx.PHP)
	if c != nil && !c.Avail.In(ctx.PHP) {
		return nil
	}
	return c
}

// NativeFailureFlowOnly reports reads whose producer paths must be checked by
// flow instead of a lexical guard or the inferred type of an earlier write.
func NativeFailureFlowOnly(ctx *analysis.Context, e syntax.Expr) bool {
	if NativeLabelled(ctx, e) {
		return true
	}
	reaching, single := ctx.Flow().Reaching(e)
	return ctx.Flow().CompressedOrigins(e) || !single && len(reaching) > 0
}

// NativeLocalBuiltinResult follows an unguarded result of a finite resolved
// native callable contract (function name, or Class::method).
func NativeLocalBuiltinResult(ctx *analysis.Context, e syntax.Expr, failure string, names ...string) syntax.Expr {
	results := NativeLocalBuiltinResults(ctx, e, failure, names...)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// NativeLocalBuiltinResults returns all eligible native producer witnesses.
// Argument-sensitive contracts are individually retained, allowing consumers
// to examine every candidate before choosing one.
func NativeLocalBuiltinResults(ctx *analysis.Context, e syntax.Expr, failure string, names ...string) []syntax.Expr {
	if ctx.Flow().ProducerBudgetExceeded(e) {
		return nil
	}
	matches := func(result syntax.Expr) bool { return nativeBuiltinContract(ctx, result, names) }
	reaching, single := ctx.Flow().Reaching(e)
	if !single && len(reaching) > 0 {
		var results []syntax.Expr
		for _, origin := range reaching {
			result := NativeUnsilenced(origin.Expr)
			if matches(result) && ((failure == "false" && !origin.NonFalse) || (failure == "null" && !origin.NonNull)) {
				results = append(results, result)
			}
		}
		return results
	}
	result := nativeLocalValue(ctx, e, failure, matches)
	if result == nil || !matches(result) {
		return nil
	}
	return []syntax.Expr{result}
}

func nativeBuiltinContract(ctx *analysis.Context, result syntax.Expr, names []string) bool {
	if call, ok := result.(*syntax.FuncCall); ok {
		key := NativeBuiltinName(ctx, call)
		for _, name := range names {
			if strings.EqualFold(key, name) && key != "" {
				return true
			}
		}
		return false
	}
	for _, contract := range names {
		class, method, qualified := strings.Cut(contract, "::")
		if !qualified {
			continue
		}
		switch call := result.(type) {
		case *syntax.MethodCall:
			if id, ok := call.Name.(*syntax.Identifier); ok && strings.EqualFold(id.Value, method) && NativeMethod(ctx, call, class, method) {
				return true
			}
		case *syntax.StaticCall:
			name, named := call.Class.(*syntax.Name)
			id, methodNamed := call.Name.(*syntax.Identifier)
			if named && methodNamed && strings.EqualFold(id.Value, method) && strings.EqualFold(ctx.Names().Class(name.Value, name.Span().Start), class) {
				decl := ctx.Index().FindMethod(class, method, ctx.PHP)
				if decl != nil && decl.Builtin && decl.Avail.In(ctx.PHP) {
					return true
				}
			}
		}
	}
	return false
}
