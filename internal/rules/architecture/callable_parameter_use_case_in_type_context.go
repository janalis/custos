package architecture

import (
	"sort"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/phpdoc"
	"custos/internal/syntax"
)

// callableParameterUseCaseInTypeContext reports is_*() checks that
// contradict a parameter's declared type and re-assignments of values of
// an unrelated type.
type callableParameterUseCaseInTypeContext struct{}

func init() { register(callableParameterUseCaseInTypeContext{}) }

func (callableParameterUseCaseInTypeContext) ID() string {
	return "CallableParameterUseCaseInTypeContext"
}

func (callableParameterUseCaseInTypeContext) Semantic() {}

func (callableParameterUseCaseInTypeContext) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFunction, syntax.KMethod}
}

const (
	cpMsgFalse  = "This check is always false for the declared parameter type; is the parameter being reused?"
	cpMsgTrue   = "This check is always true for the declared parameter type; is the parameter being reused?"
	cpMsgAssign = "Assigning a value of type %s does not match the parameter's declared type."
)

func (callableParameterUseCaseInTypeContext) Check(ctx *analysis.Context, n syntax.Node) {
	var params []*syntax.Param
	var body *syntax.Block
	switch fn := n.(type) {
	case *syntax.Function:
		params, body = fn.Params, fn.Body
	case *syntax.Method:
		params, body = fn.Params, fn.Body
	}
	if body == nil || len(params) == 0 {
		return
	}
	if util.InTestContext(ctx, n) { // D0
		return
	}
	tr := infer.NewTRules(ctx.Types())
	tr.SpecOnly = true
	tr.SoundArithmetic = true // `time() + $untyped` is not known to be a float
	cp := &cpState{ctx: ctx, fn: n, tr: tr}
	var accesses map[string][]*syntax.Variable
	for _, p := range params {
		if p.Var == nil || p.Var.Name == "" {
			continue
		}
		set, ok := cp.paramSet(p)
		if !ok {
			continue
		}
		if accesses == nil {
			accesses = cpReachableAccesses(body)
		}
		for _, v := range accesses[p.Var.Name] {
			cp.checkIs(v, set)
			cp.checkAssign(v, set)
		}
	}
}

type cpState struct {
	ctx *analysis.Context
	fn  syntax.Node
	tr  *infer.TRules
}

// cpNorm is the N normalisation of one type name.
func cpNorm(a string) string {
	if strings.Contains(a, "[]") {
		return "array"
	}
	low := strings.ToLower(strings.TrimPrefix(a, `\`))
	switch low {
	case "array", "iterable", "string", "float", "number", "null", "void", "mixed", "callable", "resource", "self", "object":
		return low
	case "bool", "boolean", "true", "false":
		return "bool"
	case "int", "integer":
		return "int"
	case "closure":
		return "callable"
	case "static", "$this":
		return "static"
	}
	if strings.HasPrefix(a, `\`) {
		return a
	}
	return `\` + a
}

// cpIsClass reports whether a P/R entry denotes a class name.
func cpIsClass(t string) bool {
	return (strings.HasPrefix(t, `\`) && t != `\Closure`) || t == "self" || t == "static"
}

// paramSet builds P (D1–D4); ok is false when the parameter is skipped.
func (cp *cpState) paramSet(p *syntax.Param) (map[string]bool, bool) {
	raw := []string{"array"} // D1: variadic
	if !p.Variadic {
		raw = cp.tr.ParamTypes(cp.fn, p).Atoms()
		// custos: without a declared or documented type the parameter
		// accepts anything; its default is just one possible value.
		if len(raw) == 0 {
			return nil, false
		}
		if p.Default != nil {
			raw = append(raw, cp.typeOf(p.Default)...)
		}
	}
	set := map[string]bool{}
	for _, r := range raw { // D2
		switch t := cpNorm(r); t {
		case "mixed", "object":
			return nil, false
		case "callable":
			set["callable"], set["array"], set["string"], set[`\Closure`] = true, true, true, true
		case "iterable":
			set["iterable"], set["array"], set[`\Traversable`] = true, true, true
		default:
			set[t] = true
		}
	}
	// D3 (empty P) cannot happen: raw is not empty (custos skip above).
	if len(set) == 1 && set["null"] && syntax.IsNullConst(p.Default) { // D4
		return nil, false
	}
	return set, true
}

// cpReachableAccesses collects variable accesses by name in the reachable
// statements of body, skipping nested function-likes and classes (D5).
func cpReachableAccesses(body *syntax.Block) map[string][]*syntax.Variable {
	out := map[string][]*syntax.Variable{}
	var visitNode func(n syntax.Node)
	var visitList func(list []syntax.Stmt) bool
	visitNode = func(n syntax.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike, *syntax.Function:
			return
		case *syntax.Variable:
			if x.Name != "" {
				out[x.Name] = append(out[x.Name], x)
			}
		case *syntax.Block:
			visitList(x.Stmts)
			return
		case *syntax.Case:
			visitNode(x.Cond)
			visitList(x.Stmts)
			return
		}
		syntax.Children(n, visitNode)
	}
	// visitList walks a statement list and reports whether it always
	// terminates; statements after a terminating one are unreachable.
	visitList = func(list []syntax.Stmt) bool {
		for _, s := range list {
			visitNode(s)
			if cpTerminates(s) {
				return true
			}
		}
		return false
	}
	visitList(body.Stmts)
	return out
}

// cpTerminates reports whether control never continues after s.
func cpTerminates(s syntax.Stmt) bool {
	switch x := s.(type) {
	case *syntax.Return, *syntax.Break, *syntax.Continue, *syntax.Goto:
		return true
	case *syntax.ExprStmt:
		switch x.Expr.(type) {
		case *syntax.Throw, *syntax.Exit:
			return true
		}
	case *syntax.Block:
		for _, st := range x.Stmts {
			if cpTerminates(st) {
				return true
			}
		}
	case *syntax.If:
		if x.Else == nil || !cpTerminates(x.Body) || !cpTerminates(x.Else.Body) {
			return false
		}
		for _, ei := range x.ElseIfs {
			if !cpTerminates(ei.Body) {
				return false
			}
		}
		return true
	}
	return false
}

// cpPlausible maps is_* functions to the P entries making them plausible.
var cpPlausible = map[string][]string{
	"is_array":    {"array", "iterable"},
	"is_string":   {"string"},
	"is_bool":     {"bool"},
	"is_int":      {"int", "number"},
	"is_float":    {"float", "number"},
	"is_resource": {"resource"},
	"is_numeric":  {"number", "float", "int"},
	"is_callable": {"callable", "array", "string", `\Closure`},
	"is_object":   {"object", "callable"},
	"is_a":        {"object", "string"},
}

// checkIs implements D6.
func (cp *cpState) checkIs(v *syntax.Variable, set map[string]bool) {
	arg, ok := v.Parent().(*syntax.Arg)
	if !ok || arg.Value != syntax.Expr(v) || arg.Unpack {
		return
	}
	call, ok := arg.Parent().Parent().(*syntax.FuncCall) // an Arg always sits in an ArgList
	if !ok {
		return
	}
	fn := cp.ctx.GlobalFunctionName(call) // any case; not a namespaced same-named function
	wanted, known := cpPlausible[fn]
	if !known {
		return
	}
	if fn == "is_numeric" && set["string"] {
		return
	}
	for _, w := range wanted {
		if set[w] {
			return
		}
	}
	if fn == "is_object" || fn == "is_a" {
		for t := range set {
			if cpIsClass(t) {
				return
			}
		}
	}
	msg := cpMsgFalse
	if u, ok := call.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		msg = cpMsgTrue
	}
	cp.ctx.ReportNode(call, msg)
}

// checkAssign implements D7 (at most one finding per assignment).
func (cp *cpState) checkAssign(v *syntax.Variable, set map[string]bool) {
	as, ok := v.Parent().(*syntax.Assign)
	if !ok || as.Var != syntax.Expr(v) || as.Op.Kind != syntax.TEqual || as.Value == nil {
		return
	}
	value := as.Value
	// custos: an inline `/** @var User $u */` on the assignment states the
	// value's type (a repository's `object|null` find result).
	if es, ok := as.Parent().(*syntax.ExprStmt); ok {
		if c := index.DocComment(cp.ctx.File, es); c != "" && phpdoc.Parse(c).VarType(v.Name) != "" {
			return
		}
	}
	r := map[string]bool{}
	for _, a := range cp.typeOf(value) { // D7b
		r[cpNorm(a)] = true
	}
	// custos: method, static and nullsafe calls carry the same failure
	// markers (`Yii::getAlias()` is string|false) as plain function calls.
	plainCall := cpAllCalls(value)
	others := 0
	for t := range r {
		if t != "bool" && t != "null" {
			others++
		}
	}
	if len(r) >= 2 { // D7c
		if plainCall && others > 0 {
			// custos: a call's false is a failure marker next to any other
			// type (`filemtime()` int|false, `fopen()` resource|false), and
			// next to string/array both false and null are (`array|bool|null`
			// lookups).
			delete(r, "bool")
			if r["string"] || r["array"] {
				delete(r, "null")
			}
		} else if r["null"] {
			for t := range set {
				if cpIsClass(t) {
					delete(r, "null")
					break
				}
			}
		}
	}
	if r["mixed"] { // D7d: mixed covers every type; custos skips (E3)
		return
	}
	for _, t := range cpSorted(r) { // D7e
		if t == "self" || t == "static" {
			t = cp.translateSelf(value, v)
			if t == "" {
				continue
			}
		}
		if cp.compatible(t, set) {
			continue
		}
		cp.ctx.ReportNode(value, strings.Replace(cpMsgAssign, "%s", t, 1))
		return
	}
}

func cpSorted(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// translateSelf maps a self/static value type to a class (D7e.1); "" when
// it cannot be translated.
func (cp *cpState) translateSelf(value syntax.Expr, target *syntax.Variable) string {
	x := syntax.UnwrapParens(value)
	if b, ok := x.(*syntax.Binary); ok && b.Op.Kind == syntax.TCoalesce {
		if lv, ok := syntax.UnwrapParens(b.Left).(*syntax.Variable); ok && lv.Name == target.Name {
			x = syntax.UnwrapParens(b.Right)
		}
	}
	var recv syntax.Expr
	switch c := x.(type) {
	case *syntax.MethodCall:
		recv = c.Var
	case *syntax.StaticCall:
		// A typed call on a class name implies the class resolves; an
		// unresolvable one falls back to the (unknown) receiver type.
		if nm, ok := c.Class.(*syntax.Name); ok {
			if cls := cp.ctx.Index().Class(cp.classRef(nm), cp.ctx.PHP); cls != nil {
				return `\` + strings.TrimPrefix(cls.FQN, `\`)
			}
		}
		recv = c.Class
	default:
		return ""
	}
	var classes []string
	for _, a := range cp.typeOf(recv) {
		if strings.HasPrefix(a, `\`) && !strings.Contains(a, "[]") {
			classes = append(classes, a)
		}
	}
	if len(classes) == 1 || len(classes) == 2 {
		return classes[0]
	}
	return ""
}

// cpAllCalls reports whether e is a call, or a match/ternary whose every
// result is one (custos: the failure markers of D7c then come from calls).
func cpAllCalls(e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		return true
	case *syntax.Unary: // `@iconv(…)`
		return x.Op.Kind == syntax.TAt && cpAllCalls(x.Expr)
	case *syntax.Match:
		for _, arm := range x.Arms {
			if arm.Body == nil || !cpAllCalls(arm.Body) {
				return false
			}
		}
		return len(x.Arms) > 0
	case *syntax.Ternary:
		return x.Then != nil && cpAllCalls(x.Then) && cpAllCalls(x.Else)
	}
	return false
}

// compatible implements D7e.2.
func (cp *cpState) compatible(t string, set map[string]bool) bool {
	if set[t] {
		return true
	}
	if t == "object" { // custos: an object may be the parameter's class
		for s := range set {
			if cpIsClass(s) {
				return true
			}
		}
	}
	if t == "int" && set["float"] { // custos: PHP accepts an int for float (even strict)
		return true
	}
	if set["callable"] && strings.HasPrefix(t, `\`) {
		// custos: an object with __invoke() is callable; unknown classes
		// may be.
		cls := strings.TrimPrefix(t, `\`)
		if cp.ctx.Index().Class(cls, cp.ctx.PHP) == nil || cp.ctx.Index().FindMethod(cls, "__invoke", cp.ctx.PHP) != nil {
			return true
		}
	}
	if !strings.HasPrefix(t, `\`) {
		return false
	}
	for s := range set {
		if strings.EqualFold(s, t) {
			return true
		}
	}
	mine, complete := cp.closureOf(t)
	if !complete {
		return true // unresolvable class or hierarchy: unknown, no report
	}
	for s := range set {
		if !cpIsClass(s) {
			continue
		}
		cls := s
		if s == "self" || s == "static" {
			// "" outside a class (a compile error in PHP): its closure
			// is {""}, which intersects nothing.
			cls = cp.ctx.Names().DeclFQN(syntax.EnclosingClass(cp.fn))
		}
		theirs := cp.closure(cls)
		if len(theirs) == 0 {
			theirs = map[string]bool{strings.ToLower(strings.TrimPrefix(cls, `\`)): true}
		}
		for k := range theirs {
			if mine[k] {
				return true
			}
		}
	}
	return false
}

// closure returns the lower-case FQNs of a class, its ancestors, interfaces
// and traits (empty when the class does not resolve).
func (cp *cpState) closure(fqn string) map[string]bool {
	out := map[string]bool{}
	for _, c := range cp.ctx.Index().Ancestors(strings.TrimPrefix(fqn, `\`), cp.ctx.PHP) {
		out[strings.ToLower(strings.TrimPrefix(c.FQN, `\`))] = true
	}
	return out
}

// closureOf is closure for a value class (D7e.2): it also reports whether
// the class and every class-like in its hierarchy resolve; an incomplete
// hierarchy makes compatibility unknown.
func (cp *cpState) closureOf(fqn string) (map[string]bool, bool) {
	out := map[string]bool{}
	complete := true
	queue := []string{strings.TrimPrefix(fqn, `\`)}
	for len(queue) > 0 {
		k := strings.ToLower(strings.TrimPrefix(queue[0], `\`))
		queue = queue[1:]
		if out[k] {
			continue
		}
		out[k] = true
		c := cp.ctx.Index().Class(k, cp.ctx.PHP)
		if c == nil {
			complete = false
			continue
		}
		queue = append(queue, c.Traits...)
		if c.Parent != "" {
			queue = append(queue, c.Parent)
		}
		queue = append(queue, c.Interfaces...)
	}
	return out, complete
}

// classRef resolves a class name reference (self/static/parent aware).
func (cp *cpState) classRef(nm *syntax.Name) string {
	switch strings.ToLower(nm.Value) {
	case "self", "static":
		return cp.ctx.Names().DeclFQN(syntax.EnclosingClass(nm))
	case "parent":
		return cp.ctx.Names().ParentFQN(syntax.EnclosingClass(nm))
	}
	return cp.ctx.Names().Class(nm.Value, nm.Span().Start)
}

// typeOf returns the raw (unnormalised) type atoms of e per the shared
// T-rules typer (specs/UnnecessaryCasting.md), nil when unknown.
func (cp *cpState) typeOf(e syntax.Expr) []string {
	return cp.tr.TypeOf(e).Atoms()
}
