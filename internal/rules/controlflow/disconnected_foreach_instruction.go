package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// disconnectedForeachInstruction reports foreach body statements that do not
// depend on the loop, and (optionally) per-iteration object creation.
type disconnectedForeachInstruction struct{}

func init() { register(disconnectedForeachInstruction{}) }

const (
	disconnectedMsg   = "Statement does not depend on the loop; move it out."
	disconnectedClone = "Create the object once before the loop and clone it here."
)

func (disconnectedForeachInstruction) ID() string { return "DisconnectedForeachInstruction" }

func (disconnectedForeachInstruction) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach}
}

type dfiClass uint8

const (
	dfiOther dfiClass = iota
	dfiControl
	dfiIncDec
	dfiNew
	dfiClone
	dfiDomCreate
	dfiAssign
	dfiAccumulate
)

func (disconnectedForeachInstruction) Check(ctx *analysis.Context, n syntax.Node) {
	loop := n.(*syntax.Foreach)
	body, ok := bracedBlock(ctx, loop.Body) // D1
	if !ok {
		return
	}
	var stmts []syntax.Stmt // D2/D3
	for _, s := range body.Stmts {
		if _, html := s.(*syntax.InlineHTML); html {
			return
		}
		if s.Span().Len() > 0 {
			stmts = append(stmts, s)
		}
	}
	if len(stmts) == 0 || dfiRepeatLoop(loop, body) {
		return
	}
	modified := map[string]bool{} // D4
	for p := syntax.Node(loop); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		fe, ok := p.(*syntax.Foreach)
		if !ok {
			continue
		}
		for _, e := range []syntax.Expr{fe.Key, fe.Value} {
			if e == nil {
				continue
			}
			syntax.Inspect(e, func(x syntax.Node) bool {
				if v, ok := x.(*syntax.Variable); ok && v.Name != "" {
					modified[v.Name] = true
				}
				return true
			})
		}
	}
	deps := make([]map[string]bool, len(stmts)) // pass 1
	bound := map[string]map[int]bool{}          // D6a: names bound by catch / inner foreach headers, per statement
	writers := map[string]map[int]bool{}        // custos: statements writing each name
	writes := make([]map[string]bool, len(stmts))
	for i, s := range stmts {
		w := map[string]bool{}
		writes[i] = w
		deps[i] = dfiCollect(ctx, s, w, func(name string) {
			if bound[name] == nil {
				bound[name] = map[int]bool{}
			}
			bound[name][i] = true
		})
		for name := range w {
			modified[name] = true
			if writers[name] == nil {
				writers[name] = map[int]bool{}
			}
			writers[name][i] = true
		}
	}
	suggestClone := ctx.Bool("SUGGEST_USING_CLONE")
	guarded := false          // custos: an earlier statement may leave the iteration
	for i, s := range stmts { // pass 2
		connected := guarded
		guarded = guarded || dfiMayLeave(s)
		for name := range deps[i] {
			if modified[name] || dfiBoundElsewhere(bound[name], i) {
				connected = true
				break
			}
		}
		// custos: a variable another statement of the body also writes (a
		// per-iteration reset such as `$level = 1;`) ties them together.
		for name := range writes[i] {
			if dfiBoundElsewhere(writers[name], i) {
				connected = true
				break
			}
		}
		if connected {
			continue
		}
		switch dfiClassify(ctx, s) {
		case dfiOther:
			if dfiDisconnected(s) && !dfiPerIteration(ctx, s) {
				ctx.ReportSeverity(dfiRange(ctx, s), meta.SeverityInfo, disconnectedMsg)
			}
		case dfiNew, dfiDomCreate:
			if suggestClone {
				ctx.ReportSeverity(s.Span(), meta.SeverityInfo, disconnectedClone)
			}
		}
	}
}

// dfiRepeatLoop reports whether the body never reads the loop's own key or
// value variable (`foreach (range(1, 5) as $attempt) { post(…); }`): the
// loop exists to repeat its body, so every statement belongs to it (custos).
func dfiRepeatLoop(loop *syntax.Foreach, body *syntax.Block) bool {
	dynamic := false // compact('v'), $$n, include… may read the variable
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch c := x.(type) {
		case *syntax.FuncCall:
			switch strings.ToLower(util.CallLastName(c)) {
			case "compact", "get_defined_vars", "extract":
				dynamic = true
			}
		case *syntax.Include, *syntax.Eval:
			dynamic = true
		case *syntax.Variable:
			dynamic = dynamic || c.Name == ""
		}
		return !dynamic
	})
	if dynamic {
		return false
	}
	for _, e := range []syntax.Expr{loop.Key, loop.Value} {
		found := false
		syntax.Inspect(e, func(x syntax.Node) bool {
			if v, ok := x.(*syntax.Variable); ok && (v.Name == "" || util.MentionsVariable(body, v.Name)) {
				found = true
			}
			return !found
		})
		if found {
			return false
		}
	}
	return true
}

// dfiParent returns the parent of n, seeing through call argument wrappers.
func dfiParent(n syntax.Node) syntax.Node {
	p := n.Parent()
	if a, ok := p.(*syntax.Arg); ok {
		return a.Parent()
	}
	return p
}

// dfiBoundElsewhere reports whether a statement other than i binds the name
// (catch variable or inner foreach header).
func dfiBoundElsewhere(binders map[int]bool, i int) bool {
	for j := range binders {
		if j != i {
			return true
		}
	}
	return false
}

// dfiCollect runs pass 1 on statement s: returns Dep(s), adds writes to m and
// reports names bound by catch clauses / inner foreach headers to bind.
func dfiCollect(ctx *analysis.Context, s syntax.Stmt, m map[string]bool, bind func(string)) map[string]bool {
	dep := map[string]bool{}
	syntax.Inspect(s, func(x syntax.Node) bool {
		if call, ok := x.(*syntax.FuncCall); ok { // D11
			if ctx.IsGlobalFunctionCall(call, "compact") {
				if args, ok := util.CallArgValues(call); ok {
					for _, a := range args {
						if c, _, ok := util.QuotedStringRaw(a); ok && c != "" {
							dep[c] = true
						}
					}
				}
			}
		}
		v, ok := x.(*syntax.Variable)
		if !ok || v.Name == "" || util.IsStaticPropName(v) {
			return true
		}
		dfiVariable(ctx, v, m, dep, bind)
		return true
	})
	return dep
}

func dfiVariable(ctx *analysis.Context, v *syntax.Variable, m, dep map[string]bool, bind func(string)) {
	name := v.Name
	var c syntax.Node = v
	for {
		pf, ok := c.Parent().(*syntax.PropertyFetch)
		if !ok || pf.Var != c {
			break
		}
		c = pf
	}
	p := dfiParent(c)
	if item, ok := p.(*syntax.ArrayItem); ok {
		if lit := item.Parent(); lit != nil {
			p = lit
			if pp := dfiParent(lit); pp != nil {
				p = pp
			}
		}
	}
	// D5
	if a, ok := p.(*syntax.Assign); ok {
		switch syntax.UnwrapParens(a.Var).(type) {
		case *syntax.List, *syntax.Array:
			if a.Value != syntax.Expr(v) {
				m[name], dep[name] = true, true
				return
			}
		}
		// D6
		if a.Var == c {
			m[name] = true
			_, isProp := c.(*syntax.PropertyFetch)
			_, inArg := a.Parent().(*syntax.Arg)
			if a.Op.Kind != syntax.TEqual || isProp || inArg {
				dep[name] = true
			}
			return
		}
	}
	// D6a: catch variables and inner foreach key/value targets are bindings.
	switch x := p.(type) {
	case *syntax.Catch:
		if x.Var == v {
			bind(name)
			return
		}
	case *syntax.Foreach:
		if x.Key != nil && x.Key.Span().Contains(v.Span()) || x.Value != nil && x.Value.Span().Contains(v.Span()) {
			bind(name)
			return
		}
	}
	// D7: only element writes modify the array
	if d, ok := p.(*syntax.ArrayDimFetch); ok && d.Var == c {
		dep[name] = true
		if dfiElementWrite(ctx, d) {
			m[name] = true
		}
	}
	// D8
	if list, ok := p.(*syntax.ArgList); ok && c == syntax.Node(v) {
		// D8b (custos: also resolved method, static and constructor calls)
		if params := dfiCalleeParams(ctx, list.Parent()); len(params) > 0 {
			idx := -1
			for i, a := range list.Args {
				if arg, ok := a.(*syntax.Arg); ok && arg.Value == syntax.Expr(v) {
					idx = i
					break
				}
			}
			if idx >= 0 {
				byRef := false
				if idx < len(params) {
					byRef = params[idx].ByRef
				} else if last := params[len(params)-1]; last.Variadic {
					byRef = last.ByRef
				}
				if byRef {
					m[name], dep[name] = true, true
					if _, isMethod := list.Parent().(*syntax.MethodCall); !isMethod {
						return
					}
				}
			}
		}
		if call, ok := list.Parent().(*syntax.MethodCall); ok && !call.NullSafe { // D8a
			obj := call.Var
			for {
				switch o := obj.(type) {
				case *syntax.Paren:
					obj = o.Expr
					continue
				case *syntax.PropertyFetch:
					obj = o.Var
					continue
				case *syntax.ArrayDimFetch:
					obj = o.Var
					continue
				case *syntax.MethodCall: // custos: a fluent chain modifies its root
					obj = o.Var
					continue
				}
				break
			}
			if ov, ok := obj.(*syntax.Variable); ok && ov.Name != "" {
				m[ov.Name] = true
				return
			}
		}
	}
	// D8c: a method call made for its side effect (an expression statement
	// whose result is discarded: `$bar->advance();`, `$stack->pop();`) may
	// change its receiver: the object counts as modified.
	if mc, ok := p.(*syntax.MethodCall); ok && mc.Var == c {
		// custos: also at the root of a discarded chain
		// (`$qb->where(…)->bind('k', $row);`, `$ctx->console()->tick();`).
		var top syntax.Node = mc
		for {
			q, ok := top.Parent().(*syntax.MethodCall)
			if !ok || q.Var != top {
				break
			}
			top = q
		}
		if _, stmt := top.Parent().(*syntax.ExprStmt); stmt {
			m[name], dep[name] = true, true
			return
		}
	}
	// D9
	if _, ok := p.(*syntax.IncDec); ok {
		m[name], dep[name] = true, true
		return
	}
	dep[name] = true // D10
}

func dfiClassify(ctx *analysis.Context, s syntax.Stmt) dfiClass {
	switch st := s.(type) {
	case *syntax.Break, *syntax.Continue, *syntax.Return:
		return dfiControl
	case *syntax.ExprStmt:
		switch e := st.Expr.(type) {
		case *syntax.IncDec:
			return dfiIncDec
		case *syntax.Assign:
			if d, ok := e.Var.(*syntax.ArrayDimFetch); ok && d.Dim == nil {
				return dfiAccumulate
			}
			if _, ok := ruleSimpleVar(e.Var); !ok {
				return dfiOther
			}
			if e.Op.Kind == syntax.TEqual {
				switch val := e.Value.(type) {
				case *syntax.New:
					return dfiNew
				case *syntax.Clone:
					return dfiClone
				case *syntax.MethodCall:
					if dfiIsDomCreate(ctx, val) {
						return dfiDomCreate
					}
				}
			}
			return dfiAssign
		}
	}
	return dfiOther
}

func dfiIsDomCreate(ctx *analysis.Context, call *syntax.MethodCall) bool {
	id, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "createElement") {
		return false
	}
	for _, cls := range ctx.TypeOf(call.Var).Classes() {
		if m := ctx.Index().FindMethod(cls, id.Value, ctx.PHP); m != nil &&
			strings.EqualFold(strings.TrimPrefix(m.Class, `\`), "DOMDocument") {
			return true
		}
	}
	return false
}

// dfiPerIterationFuncs are built-ins whose effect or result belongs to each
// iteration: output, stream writes, random numbers and clocks
// (lower-case).
var dfiPerIterationFuncs = func() map[string]bool {
	m := map[string]bool{}
	for _, f := range strings.Fields(`fwrite fputs fputcsv fprintf vfprintf fflush printf vprintf
		file_put_contents var_dump print_r var_export debug_zval_dump debug_print_backtrace
		rand mt_rand random_int random_bytes lcg_value uniqid microtime hrtime time array_rand
		shuffle str_shuffle`) {
		m[f] = true
	}
	return m
}()

// dfiPerIteration reports whether s outputs (echo, print) or calls one of
// the functions above: running it once instead of on every iteration would
// change the program (custos, see Divergences).
func dfiPerIteration(ctx *analysis.Context, s syntax.Stmt) bool {
	found := false
	syntax.Inspect(s, func(x syntax.Node) bool {
		switch c := x.(type) {
		case *syntax.FuncCall:
			found = dfiPerIterationFuncs[ctx.GlobalFunctionName(c)]
		case *syntax.Echo, *syntax.Print:
			found = true
		}
		return !found
	})
	return found
}

// dfiDisconnected checks the D12 shape conditions of an "other" statement.
func dfiDisconnected(s syntax.Stmt) bool {
	if e, ok := s.(*syntax.Echo); ok && e.Short {
		return false
	}
	exits, hasVar := false, false
	syntax.Inspect(s, func(x syntax.Node) bool {
		if x == syntax.Node(s) {
			return true
		}
		switch v := x.(type) {
		case *syntax.Break, *syntax.Continue, *syntax.Return, *syntax.Throw:
			exits = true
			return false
		case *syntax.Variable:
			if !util.IsStaticPropName(v) {
				hasVar = true
			}
		}
		return true
	})
	return !exits && hasVar
}

func dfiRange(ctx *analysis.Context, s syntax.Stmt) syntax.Span {
	switch s.(type) {
	case *syntax.If, *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach, *syntax.Switch, *syntax.Try:
		if t, ok := util.NextSignificant(ctx.File, s.Span().Start); ok {
			return syntax.Span{Start: t.Start, End: t.End}
		}
	}
	return s.Span()
}

// dfiElementWrite reports whether the element access d (outermost of its
// `[…]`/`->` chain) is written: assigned (any operator, or bound by
// reference), incremented/decremented, unset, used as the receiver of a
// method call, or passed to a by-reference parameter of a resolved function.
func dfiElementWrite(ctx *analysis.Context, d *syntax.ArrayDimFetch) bool {
	var o syntax.Node = d
	for {
		switch q := o.Parent().(type) {
		case *syntax.ArrayDimFetch:
			if q.Var == o {
				o = q
				continue
			}
		case *syntax.PropertyFetch:
			if q.Var == o {
				o = q
				continue
			}
		}
		break
	}
	switch q := dfiParent(o).(type) {
	case *syntax.Assign:
		return q.Var == o || q.ByRef && q.Value == o
	case *syntax.IncDec, *syntax.Unset:
		return true
	case *syntax.MethodCall:
		return q.Var == o
	case *syntax.ArgList:
		call, ok := q.Parent().(*syntax.FuncCall)
		if !ok {
			return false
		}
		f := ctx.Types().ResolveFunction(call)
		if f == nil || len(f.Params) == 0 {
			return false
		}
		for i, a := range q.Args {
			if arg, ok := a.(*syntax.Arg); ok && arg.Value == o {
				if i < len(f.Params) {
					return f.Params[i].ByRef
				}
				last := f.Params[len(f.Params)-1]
				return last.Variadic && last.ByRef
			}
		}
	}
	return false
}

// dfiMayLeave reports whether s contains a jump that may end the current
// iteration (continue or break not bound by a nested loop or switch,
// return, throw, exit) without being one itself: later statements then run
// on some iterations only, and moving them out of the loop changes
// behaviour. Nested functions and classes are skipped.
func dfiMayLeave(s syntax.Stmt) bool {
	switch s.(type) {
	case *syntax.Break, *syntax.Continue, *syntax.Return:
		return false // the rest of the body is dead code
	}
	found := false
	var walk func(n syntax.Node, inner bool)
	walk = func(n syntax.Node, inner bool) {
		syntax.Inspect(n, func(x syntax.Node) bool {
			if found {
				return false
			}
			switch x := x.(type) {
			case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
				return false
			case *syntax.Return, *syntax.Throw, *syntax.Exit:
				found = true
			case *syntax.Break, *syntax.Continue:
				if !inner || dfiJumpLevels(x) > 1 {
					found = true
				}
			default:
				if x != n && dfiIsLoopOrSwitch(x) {
					walk(x, true)
					return false
				}
			}
			return !found
		})
	}
	walk(s, dfiIsLoopOrSwitch(s))
	return found
}

func dfiIsLoopOrSwitch(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile, *syntax.Switch:
		return true
	}
	return false
}

// dfiJumpLevels is the level count of a break/continue (1 when omitted or
// not a literal).
func dfiJumpLevels(n syntax.Node) int {
	var e syntax.Expr
	switch j := n.(type) {
	case *syntax.Break:
		e = j.Num
	case *syntax.Continue:
		e = j.Num
	}
	if lit, ok := e.(*syntax.Literal); ok && lit.Raw != "" && lit.Raw[0] >= '2' && lit.Raw[0] <= '9' {
		return 2
	}
	return 1
}

// dfiCalleeParams returns the parameters of the resolved callee of call (a
// function, method, static method or constructor call), nil when unknown.
func dfiCalleeParams(ctx *analysis.Context, call syntax.Node) []index.Param {
	method := func(classes []string, name syntax.Node) []index.Param {
		id, ok := name.(*syntax.Identifier)
		if !ok {
			return nil
		}
		for _, cls := range classes {
			if m := ctx.Index().FindMethod(cls, id.Value, ctx.PHP); m != nil {
				return m.Params
			}
		}
		return nil
	}
	switch c := call.(type) {
	case *syntax.FuncCall:
		if f := ctx.Types().ResolveFunction(c); f != nil {
			return f.Params
		}
	case *syntax.MethodCall:
		return method(ctx.TypeOf(c.Var).Classes(), c.Name)
	case *syntax.StaticCall:
		if cls := ctx.Types().ClassRef(c.Class); cls != "" {
			return method([]string{cls}, c.Name)
		}
	case *syntax.New:
		if cls := ctx.Types().ClassRef(c.Class); cls != "" {
			if m := ctx.Index().FindMethod(cls, "__construct", ctx.PHP); m != nil {
				return m.Params
			}
		}
	}
	return nil
}
