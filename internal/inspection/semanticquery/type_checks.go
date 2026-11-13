package semanticquery

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
)

// Decided type checks: a comparison (`===`, `!==`, or loose where PHP's
// conversions decide it, see looseAtomsMeet) or a type test
// (`instanceof`, `is_*()`, a strict comparison with null) whose answer the
// native declarations of the code decide. ImpossibleTypeComparison reports
// the comparisons, RedundantTypeCheck the type tests.
//
// An operand counts only when the default (PHPDoc-aware) environment, the
// proven environment (see infer.Env.ProvenTypeOf) and the proven
// environment of the newest PHP version all decide the same answer: PHPDoc
// never proves code redundant, and neither does a builtin whose types
// changed between versions.

// DecidedCheck is a comparison or a type test whose answer the code decides.
type DecidedCheck struct {
	Node syntax.Expr
	// Comparison tells an identity comparison of operands that share no
	// value (ImpossibleTypeComparison); otherwise a type test whose answer
	// is always the same (RedundantTypeCheck).
	Comparison bool
	Message    string
}

type decidedCheckKey struct{ n syntax.Node }

// DecidedTypeCheck reports n when it is a decided comparison (see
// impossibleTypeComparison) or type test (see redundantTypeCheck); a strict
// comparison with the null literal is a type test. Cached per node.
func DecidedTypeCheck(ctx *analysis.Context, n syntax.Node) (DecidedCheck, bool) {
	r := ctx.Shared(decidedCheckKey{n}, func() any {
		if c, ok := impossibleTypeComparison(ctx, n); ok {
			return DecidedCheck{Node: c.node, Comparison: true, Message: c.message()}
		}
		if c, ok := redundantTypeCheck(ctx, n); ok {
			return DecidedCheck{Node: c.node, Message: c.message()}
		}
		return nil
	})
	c, ok := r.(DecidedCheck)
	return c, ok
}

// typeComparison is a comparison between operands that share no value, or
// for a loose one never compare equal.
type typeComparison struct {
	node        *syntax.Binary
	left, right types.Type
	// values tells that the operands' exact values decided it (both have
	// the same type).
	values bool
}

func (c typeComparison) message() string {
	k := c.node.Op.Kind
	tail := " share no value."
	if k == syntax.TIsEqual || k == syntax.TIsNotEqual {
		tail = " never compare equal."
	}
	return "This comparison is always " + boolWord(k == syntax.TIsNotIdentical || k == syntax.TIsNotEqual) + ": " +
		givenName(c.left, c.values) + " and " + givenName(c.right, c.values) + tail
}

// typeCheck is a type test whose answer is always the same.
type typeCheck struct {
	node   syntax.Expr
	result bool
	// typ is the tested value's proven type.
	typ types.Type
}

func (c typeCheck) message() string {
	return "This check is always " + boolWord(c.result) + ": the value is " + typeName(c.typ) + "."
}

func boolWord(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// impossibleTypeComparison reports `a === b` or `a !== b` when no value of
// a can be identical to a value of b, and `a == b` or `a != b` when PHP's
// loose comparison never finds a value of a equal to a value of b (see
// looseDisjoint), whatever their values (see operandType for the operands
// considered), unless one operand is the null literal.
func impossibleTypeComparison(ctx *analysis.Context, n syntax.Node) (typeComparison, bool) {
	b, ok := n.(*syntax.Binary)
	if !ok {
		return typeComparison{}, false
	}
	loose := false
	switch b.Op.Kind {
	case syntax.TIsIdentical, syntax.TIsNotIdentical:
	case syntax.TIsEqual, syntax.TIsNotEqual:
		loose = true
	default:
		return typeComparison{}, false
	}
	l, r := syntax.UnwrapParens(b.Left), syntax.UnwrapParens(b.Right)
	if syntax.IsNullConst(l) || syntax.IsNullConst(r) || literal(l) && literal(r) || uncheckedOperand(l) || uncheckedOperand(r) {
		return typeComparison{}, false
	}
	rel := typeRelations(ctx)
	out := typeComparison{node: b}
	exact := [2]int{-1, -1} // unknown until asked
	for env := range envCount(ctx) {
		lt, ok := operandType(ctx, l, env)
		if !ok {
			return typeComparison{}, false
		}
		rt, ok := operandType(ctx, r, env)
		if !ok {
			return typeComparison{}, false
		}
		var decided, byValues bool
		if loose {
			decided = looseDisjoint(lt, rt, rel)
		} else {
			decided, byValues = disjoint(coarse(lt), coarse(rt), rel)
		}
		if !decided && !loose && lt.HasValues() && rt.HasValues() {
			// Only the exact values of literals (or locals holding them)
			// may tell the operands apart.
			for i, x := range [2]syntax.Expr{l, r} {
				if exact[i] < 0 {
					exact[i] = 0
					if exactOperand(ctx, x) {
						exact[i] = 1
					}
				}
			}
			if exact == [2]int{1, 1} {
				decided, byValues = disjoint(lt, rt, rel)
			}
		}
		if !decided {
			return typeComparison{}, false
		}
		out.values = out.values || byValues
		if env == provenEnv {
			out.left, out.right = lt, rt
		}
	}
	if !checkedContext(ctx, b) || notOptimalReports(ctx, b) ||
		writtenLaterInLoop(ctx, l) || writtenLaterInLoop(ctx, r) || versionedOperand(ctx, l) || versionedOperand(ctx, r) {
		return typeComparison{}, false
	}
	if lit, ok := astquery.BoolConst(l); ok && ownedComparison(ctx, b, r, boolWord(lit)) {
		return typeComparison{}, false
	}
	if lit, ok := astquery.BoolConst(r); ok && ownedComparison(ctx, b, l, boolWord(lit)) {
		return typeComparison{}, false
	}
	return out, true
}

// typePredicates maps the global type predicates to the type they test.
var typePredicates = map[string]string{
	"is_null": "null", "is_int": "int", "is_integer": "int", "is_long": "int",
	"is_float": "float", "is_double": "float", "is_string": "string", "is_bool": "bool",
	"is_array": "array", "is_object": "object", "is_scalar": "scalar", "is_iterable": "iterable",
}

// parameterChecks are the predicates CallableParameterUseCaseInTypeContext
// reports on a parameter when they never succeed.
var parameterChecks = map[string]bool{
	"is_int": true, "is_float": true, "is_string": true, "is_bool": true, "is_array": true, "is_object": true,
}

// redundantTypeCheck reports a type test whose answer is always the same:
// `x instanceof C`, `is_*(x)` for the type predicates of typePredicates,
// `x === null` and `x !== null` (either operand order).
func redundantTypeCheck(ctx *analysis.Context, n syntax.Node) (typeCheck, bool) {
	var subject syntax.Expr
	var pred string
	negate := false
	fn := ""
	switch c := n.(type) {
	case *syntax.Instanceof:
		cls, ok := instanceofClass(ctx, c)
		if !ok {
			return typeCheck{}, false
		}
		subject, pred = c.Expr, cls
	case *syntax.Binary:
		if c.Op.Kind != syntax.TIsIdentical && c.Op.Kind != syntax.TIsNotIdentical {
			return typeCheck{}, false
		}
		l, r := syntax.UnwrapParens(c.Left), syntax.UnwrapParens(c.Right)
		switch {
		case syntax.IsNullConst(l) && !syntax.IsNullConst(r):
			subject = r
		case syntax.IsNullConst(r) && !syntax.IsNullConst(l):
			subject = l
		default:
			return typeCheck{}, false
		}
		pred, negate = "null", c.Op.Kind == syntax.TIsNotIdentical
	case *syntax.FuncCall:
		fn = ctx.GlobalFunctionName(c)
		p, ok := typePredicates[fn]
		if !ok {
			return typeCheck{}, false
		}
		arg, ok := singleArgument(c)
		if !ok {
			return typeCheck{}, false
		}
		subject, pred = arg, p
	default:
		return typeCheck{}, false
	}
	subject = syntax.UnwrapParens(subject)
	if uncheckedOperand(subject) {
		return typeCheck{}, false
	}
	rel := typeRelations(ctx)
	var verdict types.Verdict
	out := typeCheck{node: n.(syntax.Expr)}
	for env := range envCount(ctx) {
		t, ok := operandType(ctx, subject, env)
		if !ok {
			return typeCheck{}, false
		}
		v := predicateVerdict(pred, t, rel)
		if v == types.Maybe || env > defaultEnv && v != verdict {
			return typeCheck{}, false
		}
		verdict = v
		if env == provenEnv {
			out.typ = t
		}
	}
	out.result = (verdict == types.Yes) != negate
	if !checkedContext(ctx, n) || notOptimalReports(ctx, out.node) || writtenLaterInLoop(ctx, subject) ||
		lastBranchCondition(out.node, out.result) || versionedOperand(ctx, subject) {
		return typeCheck{}, false
	}
	if pred == "null" && ownedComparison(ctx, out.node, subject, "null") {
		return typeCheck{}, false
	}
	if verdict == types.No && parameterChecks[fn] && namedFunctionParameter(subject) &&
		ctx.RuleChecks("CallableParameterUseCaseInTypeContext", syntax.EnclosingFuncLike(subject)) {
		return typeCheck{}, false
	}
	return out, true
}

// lastBranchCondition reports whether check x, whose answer is result, is
// (through parentheses and `!`) the whole condition of the last elseif of
// a chain without else, and that condition always holds: the branch names
// the one case left (`elseif (is_array($v))` after `is_string($v)` on an
// array|string), which reads as documentation rather than dead code.
func lastBranchCondition(x syntax.Expr, result bool) bool {
	var n syntax.Node = x
	for {
		switch p := n.Parent().(type) {
		case *syntax.Paren:
			n = p
			continue
		case *syntax.Unary:
			if p.Op.Kind == syntax.TExclaim {
				result = !result
				n = p
				continue
			}
			return false
		case *syntax.ElseIf:
			chain, ok := p.Parent().(*syntax.If)
			return ok && result && p.Cond == n && chain.Else == nil && chain.ElseIfs[len(chain.ElseIfs)-1] == p
		case *syntax.If:
			// `else if (…)` written in two words: an if alone in an else.
			outer, ok := p.Parent().(*syntax.Else)
			return ok && result && p.Cond == n && p.Else == nil && len(p.ElseIfs) == 0 && outer.Body == syntax.Stmt(p)
		}
		return false
	}
}

// singleArgument returns the tested value of a one-argument call: written
// by position or named `value`, not unpacked.
func singleArgument(c *syntax.FuncCall) (syntax.Expr, bool) {
	if c.Args == nil || len(c.Args.Args) != 1 {
		return nil, false
	}
	a, ok := c.Args.Args[0].(*syntax.Arg)
	if !ok || a.Unpack || a.ByRef || a.Name != nil && a.Name.Value != "value" {
		return nil, false
	}
	return a.Value, true
}

// instanceofClass returns the class atom an instanceof tests: a written
// class name (not self, static or parent) resolving to a class or an
// interface whose whole hierarchy is known. A trait is left to
// SuspiciousBinaryOperation.
func instanceofClass(ctx *analysis.Context, c *syntax.Instanceof) (string, bool) {
	name, ok := c.Class.(*syntax.Name)
	if !ok {
		return "", false
	}
	switch strings.ToLower(name.Value) {
	case "self", "static", "parent":
		return "", false
	}
	atom := `\` + ctx.Names().Class(name.Value, name.Span().Start)
	rel := typeRelations(ctx)
	if !rel.Resolved(atom) || rel.isTrait(atom) {
		return "", false
	}
	return atom, true
}

// namedFunctionParameter reports whether x is a variable naming a
// parameter of the function or method directly enclosing it.
func namedFunctionParameter(x syntax.Expr) bool {
	v, ok := x.(*syntax.Variable)
	if !ok {
		return false
	}
	fn := syntax.EnclosingFuncLike(v)
	switch fn.(type) {
	case *syntax.Function, *syntax.Method:
	default:
		return false
	}
	for _, p := range syntax.FuncLikeParams(fn) {
		if p.Var != nil && p.Var.Name == v.Name {
			return true
		}
	}
	return false
}

// checkedContext reports whether a check at n may be reported: not in a
// test context and not inside an assert() call, which states what the
// author expects rather than testing it.
func checkedContext(ctx *analysis.Context, n syntax.Node) bool {
	if InTestContext(ctx, n) {
		return false
	}
	for p := n.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		if c, ok := p.(*syntax.FuncCall); ok && ctx.IsGlobalFunctionCall(c, "assert") {
			return false
		}
	}
	return true
}

// notOptimalReports reports whether NotOptimalIfConditions reports check
// n itself: n is an operand of the condition of an if or elseif, that rule
// reports instanceof flaws, n is one of them (see InstanceofFlaws), and
// its check of that if statement runs.
func notOptimalReports(ctx *analysis.Context, n syntax.Expr) bool {
	var stmt syntax.Node // the if statement that rule checks
	var cond syntax.Expr
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch c := p.(type) {
		case *syntax.Paren, *syntax.Unary:
			continue
		case *syntax.Binary:
			if c.Op.Kind == syntax.TBooleanAnd || c.Op.Kind == syntax.TBooleanOr {
				continue
			}
		case *syntax.If:
			stmt, cond = c, c.Cond
		case *syntax.ElseIf:
			stmt, cond = c.Parent(), c.Cond
		}
		break
	}
	return cond != nil && ctx.RuleBool("NotOptimalIfConditions", "REPORT_INSTANCE_OF_FLAWS") &&
		slices.Contains(InstanceofFlaws(ctx, cond), n) && ctx.RuleChecks("NotOptimalIfConditions", stmt)
}

// The environments whose answers must agree, cheapest first: the default
// one, the proven one and, below the newest PHP version, the proven one of
// the newest version.
const (
	defaultEnv = iota
	provenEnv
	newestEnv
)

type versionTwinKey struct{}

// envCount returns the number of environments to ask in ctx.
func envCount(ctx *analysis.Context) int {
	if ctx.PHP < phpversion.Max {
		return newestEnv + 1
	}
	return provenEnv + 1
}

// operandType types operand x of a check in environment env, or reports
// false when that type is unknown or mixed (an unreachable check, or the
// analysis gave up).
func operandType(ctx *analysis.Context, x syntax.Expr, env int) (types.Type, bool) {
	var t types.Type
	switch env {
	case defaultEnv:
		t = ctx.TypeOf(x)
	case provenEnv:
		t = ctx.Types().ProvenTypeOf(x)
	default:
		twin := ctx.Shared(versionTwinKey{}, func() any {
			return infer.NewEnv(ctx.File, ctx.Names(), ctx.Index(), phpversion.Max, ctx.InferOptions()).Proven()
		}).(*infer.Env)
		t = twin.TypeOf(x)
	}
	return t, checkable(t)
}

// uncheckedOperand reports an operand whose type does not decide a check:
// a constant (its value may be configuration) or `$this` in a closure
// (which may run bound to another object).
func uncheckedOperand(x syntax.Expr) bool {
	return constantRead(x) || reboundThis(x)
}

// checkable reports a type whose every member is known.
func checkable(t types.Type) bool {
	return !t.IsUnknown() && !t.Has("mixed")
}

// constantRead reports a global or class constant, or an element of one.
func constantRead(x syntax.Expr) bool {
	for {
		switch v := x.(type) {
		case *syntax.ConstFetch:
			_, isBool := astquery.BoolConst(v)
			return !isBool && !syntax.IsNullConst(v)
		case *syntax.ClassConstFetch:
			return true
		case *syntax.ArrayDimFetch:
			x = v.Var
		default:
			return false
		}
	}
}

// literal reports a scalar literal: a number (negated or not), a string
// without interpolation, true, false or null.
func literal(x syntax.Expr) bool {
	x = syntax.UnwrapParens(x)
	if c, ok := x.(*syntax.ConstFetch); ok {
		_, isBool := astquery.BoolConst(c)
		return isBool || syntax.IsNullConst(c)
	}
	_, isLiteral := x.(*syntax.Literal)
	return isLiteral || astquery.IsNumberLiteral(x)
}

// exactOperand reports a literal, or a local of a function whose every
// value reaching x is a literal assigned in that function. A write into a
// string offset reaching x needs no check here: inference then types the
// local without values (see infer's withElemWritesAt), so values never
// decide the comparison.
func exactOperand(ctx *analysis.Context, x syntax.Expr) bool {
	if literal(x) {
		return true
	}
	v, ok := x.(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return false
	}
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil {
		return false
	}
	if _, entry := flowquery.ReachingAssignmentsIn(ctx.File, scope, v, v.Name); entry {
		return false
	}
	vals, known := flowquery.PossibleValuesReaching(ctx.File, v)
	if !known || len(vals) == 0 {
		return false
	}
	for _, val := range vals {
		if !literal(val) {
			return false
		}
	}
	return true
}

// coarse drops the value facts of t that are not exact: int values and
// ranges, string values and lengths.
func coarse(t types.Type) types.Type { return t.WithoutValues() }

// writtenLaterInLoop reports whether a loop around x, within its function,
// writes x (or a value x reads through: its object, its array) after x: the
// value x sees may then come from a previous iteration.
func writtenLaterInLoop(ctx *analysis.Context, x syntax.Expr) bool {
	var loop syntax.Node
	for p := x.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			loop = p
		}
	}
	if loop == nil {
		return false
	}
	at := x.Span()
	found := false
	syntax.Inspect(loop, func(n syntax.Node) bool {
		if found || n != loop && (syntax.IsFuncLike(n) || isClassLike(n)) || n.Span().End <= at.Start {
			return false
		}
		switch w := n.(type) {
		case *syntax.Assign:
			found = (w.Var.Span().Start >= at.End || w.Value != nil && w.Value.Span().Contains(at)) && writesTo(ctx, w.Var, x)
		case *syntax.IncDec:
			found = w.Span().Start >= at.End && writesTo(ctx, w.Var, x)
		case *syntax.Foreach:
			found = w != loop && w.Span().Start >= at.End && (writesTo(ctx, w.Key, x) || writesTo(ctx, w.Value, x))
		case *syntax.Unset:
			for _, v := range w.Vars {
				found = found || w.Span().Start >= at.End && writesTo(ctx, v, x)
			}
		}
		return !found
	})
	return found
}

func isClassLike(n syntax.Node) bool {
	_, ok := n.(*syntax.ClassLike)
	return ok
}

// writesTo reports whether writing target changes x: target is x, a value
// x reads through, or a part of x; a destructuring target counts each of
// its slots.
func writesTo(ctx *analysis.Context, target, x syntax.Expr) bool {
	switch t := syntax.UnwrapParens(target).(type) {
	case nil:
		return false
	case *syntax.Array:
		return itemsWrite(ctx, t.Items, x)
	case *syntax.List:
		return itemsWrite(ctx, t.Items, x)
	}
	target = syntax.UnwrapParens(target)
	for b := x; b != nil; b = accessBase(b) {
		if astquery.Equivalent(ctx.File, b, target) {
			return true
		}
	}
	for b := target; b != nil; b = accessBase(b) {
		if astquery.Equivalent(ctx.File, b, x) {
			return true
		}
	}
	return false
}

func itemsWrite(ctx *analysis.Context, items []*syntax.ArrayItem, x syntax.Expr) bool {
	for _, it := range items {
		if it != nil && writesTo(ctx, it.Value, x) {
			return true
		}
	}
	return false
}

// accessBase returns the value access x reads through (nil for none).
func accessBase(x syntax.Expr) syntax.Expr {
	switch v := syntax.UnwrapParens(x).(type) {
	case *syntax.PropertyFetch:
		return v.Var
	case *syntax.MethodCall:
		return v.Var
	case *syntax.ArrayDimFetch:
		return v.Var
	}
	return nil
}

type typeRelationsKey struct{}

// typeChecks answers the class questions of the checks.
type typeChecks struct{ *IndexRelations }

func typeRelations(ctx *analysis.Context) typeChecks {
	return ctx.Shared(typeRelationsKey{}, func() any {
		return typeChecks{NewIndexRelations(ctx.Index(), ctx.PHP)}
	}).(typeChecks)
}

func (r typeChecks) isTrait(a string) bool {
	c := r.class(a)
	return c != nil && c.Kind == syntax.KindTrait
}

// classVerdict judges whether every object of class atom a is an instance
// of class atom c: Maybe for a trait (its $this is the using class), a
// hierarchy not wholly known, or one a dependency declares (see
// dependencyHierarchy).
func (r typeChecks) classVerdict(c, a string) types.Verdict {
	if r.isTrait(a) || !r.Resolved(a) || r.dependencyHierarchy(a) {
		return types.Maybe
	}
	return types.Accepts(types.Of(c), types.Of(a), types.AcceptContext{Classes: r.IndexRelations, Strict: true})
}

// dependencyHierarchy reports whether class atom a or one of its ancestors
// is declared in the project's vendor directory (see index.World.Vendor):
// another version of that package, one the project also supports, may
// extend or implement something else (an interface a later release adds to
// its base class).
func (r typeChecks) dependencyHierarchy(a string) bool {
	w := r.ix.World()
	for _, c := range r.ix.Ancestors(strings.TrimPrefix(a, `\`), r.ver) {
		if w.Dependency(c.File) {
			return true
		}
	}
	return false
}

// plainObject reports a class atom whose objects PHP compares and converts
// with its standard rules: a class or an enum (not an interface, which any
// object may implement) whose whole hierarchy is known, no dependency
// declares (see dependencyHierarchy), and no builtin class is an ancestor
// of (DateTime, SimpleXMLElement or GMP compare and convert their objects
// their own way). Its subclasses share that ancestry.
func (r typeChecks) plainObject(a string) bool {
	c := r.class(a)
	if c == nil || c.Kind != syntax.KindClass && c.Kind != syntax.KindEnum || !r.Resolved(a) || r.dependencyHierarchy(a) {
		return false
	}
	builtin := stubs.Index()
	for _, anc := range r.ix.Ancestors(c.FQN, r.ver) {
		if anc.Kind == syntax.KindClass && builtin.Class(anc.FQN, r.ver) == anc {
			return false
		}
	}
	return true
}

// atomKind groups atoms by what identity comparisons and type tests tell
// apart.
func atomKind(a string) string {
	switch a {
	case "null", "int", "float", "string", "bool", "true", "false", "array", "iterable", "callable", "object":
		return a
	}
	if strings.HasSuffix(a, "[]") {
		return "array"
	}
	if strings.HasPrefix(a, `\`) && !strings.HasPrefix(a, `\~`) {
		return "class"
	}
	return "" // mixed, void, never, resource, static, self, parent, templates
}

// predicateVerdict judges type test pred (an atom of typePredicates or a
// class atom) on every value of t: Yes when it always holds, No when it
// never does.
func predicateVerdict(pred string, t types.Type, rel typeChecks) types.Verdict {
	if t.Intersection() != nil {
		return types.Maybe
	}
	all, none := true, true
	for _, a := range t.Atoms() {
		v := atomPredicate(pred, a, rel)
		all, none = all && v == types.Yes, none && v == types.No
		if !all && !none {
			return types.Maybe
		}
	}
	if all { // t is known: it has atoms
		return types.Yes
	}
	return types.No
}

// atomPredicate judges type test pred on the values of atom a.
func atomPredicate(pred, a string, rel typeChecks) types.Verdict {
	k := atomKind(a)
	switch k {
	case "":
		return types.Maybe
	case "true", "false":
		k = "bool"
	}
	if strings.HasPrefix(pred, `\`) { // instanceof
		switch k {
		case "class":
			return rel.classVerdict(pred, a)
		case "object", "iterable", "callable":
			return types.Maybe
		}
		return types.No
	}
	switch pred {
	case "iterable":
		switch k {
		case "array", "iterable":
			return types.Yes
		case "class":
			return rel.classVerdict(`\Traversable`, a)
		case "object", "callable":
			return types.Maybe
		}
		return types.No
	case "object":
		switch k {
		case "object", "class":
			return types.Yes
		case "iterable", "callable":
			return types.Maybe
		}
		return types.No
	case "scalar":
		switch k {
		case "int", "float", "string", "bool":
			return types.Yes
		case "callable":
			return types.Maybe
		}
		return types.No
	case "string", "array":
		if k == "callable" || pred == "array" && k == "iterable" {
			return types.Maybe
		}
	}
	if k == pred {
		return types.Yes
	}
	return types.No
}

// disjoint reports whether no value of a is identical to a value of b, and
// whether exact values (rather than types) decided it. Both are known.
func disjoint(a, b types.Type, rel typeChecks) (decided, byValues bool) {
	if a.Intersection() != nil || b.Intersection() != nil {
		return false, false
	}
	for _, x := range a.Atoms() {
		for _, y := range b.Atoms() {
			switch meet, values := atomsMeet(x, y, a, b, rel); {
			case meet:
				return false, false
			case values:
				byValues = true
			}
		}
	}
	return true, byValues
}

// atomsMeet reports whether a value of atom x (of type a) may be identical
// to a value of atom y (of type b), and whether only their exact values
// tell them apart.
func atomsMeet(x, y string, a, b types.Type, rel typeChecks) (meet, values bool) {
	kx, ky := atomKind(x), atomKind(y)
	if kx == "" || ky == "" {
		return true, false
	}
	if kx > ky {
		kx, ky, x, y, a, b = ky, kx, y, x, b, a
	}
	switch {
	case kx == ky:
		switch kx {
		case "int":
			ra, okA := a.IntRanges()
			rb, okB := b.IntRanges()
			if okA && okB && singleValues(ra) && singleValues(rb) {
				return sharesInt(ra, rb), true
			}
		case "string":
			va, okA := a.StringValueList()
			vb, okB := b.StringValueList()
			if okA && okB {
				return sharesString(va, vb), true
			}
		case "class":
			if x == y || rel.isTrait(x) || rel.isTrait(y) {
				return true, false
			}
			// No object of one is an instance of the other (a final class
			// outside an interface's implementations, unrelated classes).
			return rel.classVerdict(x, y) != types.No && rel.classVerdict(y, x) != types.No, false
		}
		return true, false
	case kx == "bool" && (ky == "true" || ky == "false"), kx == "false" && ky == "true":
		return kx == "bool", false
	}
	pair := kx + "/" + ky
	switch pair {
	case "array/iterable", "array/callable", "callable/string", "callable/class", "callable/object",
		"callable/iterable", "class/iterable", "class/object", "iterable/object":
		return true, false
	}
	return false, false
}

// looseDisjoint reports whether `a == b` is false for every value of a and
// of b. Both are known; values never decide it (PHP compares a numeric
// string with a number by value).
func looseDisjoint(a, b types.Type, rel typeChecks) bool {
	if a.Intersection() != nil || b.Intersection() != nil {
		return false
	}
	for _, x := range a.Atoms() {
		for _, y := range b.Atoms() {
			if looseAtomsMeet(x, y, a, b, rel) {
				return false
			}
		}
	}
	return true
}

// looseAtomsMeet reports whether `x == y` may hold for values of atoms x (of
// type a) and y (of type b) under PHP 8's rules. It never does for:
//   - two plain objects (see plainObject) no object is an instance of both:
//     objects of different classes are never equal;
//   - a plain object and an array, null or false: the object converts to
//     no array and to true;
//   - an array and an int, a float or a string: arrays only equal arrays,
//     and null or a bool by truthiness;
//   - true and false, or true and null.
//
// Everything else may: an object equals 1 and true (and a string through
// __toString), null equals 0, ” and [], a number equals a numeric string.
func looseAtomsMeet(x, y string, a, b types.Type, rel typeChecks) bool {
	kx, ky := atomKind(x), atomKind(y)
	if kx > ky {
		kx, ky, x, y, a, b = ky, kx, y, x, b, a
	}
	switch kx + "/" + ky {
	case "class/class":
		if !rel.plainObject(x) || !rel.plainObject(y) {
			return true
		}
		meet, _ := atomsMeet(x, y, a, b, rel)
		return meet
	case "class/false", "class/null":
		return !rel.plainObject(x)
	case "array/class":
		return !rel.plainObject(y)
	case "array/int", "array/float", "array/string", "false/true", "null/true":
		return false
	}
	return true
}

func sharesInt(a, b []types.IntRange) bool {
	for _, x := range a {
		for _, y := range b {
			if x.Lo == y.Lo {
				return true
			}
		}
	}
	return false
}

func sharesString(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}

// comparisonOwner is a rule reporting a comparison of a builtin's result
// with a constant the builtin never returns. Its fields describe the
// comparisons the rule actually reports, so that only those are left to
// it.
type comparisonOwner struct {
	rule string
	// locals tells that the rule also follows the result through a local.
	locals bool
	// onComparison tells that the rule checks the comparison rather than
	// the call.
	onComparison bool
	// bare tells that the rule sees the call only as an operand of the
	// comparison itself, not wrapped in parentheses.
	bare bool
	// equality tells that the rule reports `==` and `===` but not their
	// negations (only for a "true" owner, whose check is a comparison).
	equality bool
}

// comparisonOwners lists the comparisonOwner of a strict comparison with a
// constant (true, false or null), by the constant and the lower-case
// function or `class::method` name.
var comparisonOwners = map[string]map[string]comparisonOwner{
	"true": {
		"array_search":      {rule: "ArraySearchUsedAsInArray", bare: true},
		"collator::compare": {rule: "CollatorComparisonTruthinessReversed", equality: true},
	},
	"false": {"xml_parse": {rule: "XmlParseZeroComparedWithFalse", locals: true, onComparison: true}},
	"null":  {"domelement::getattribute": {rule: "DomMissingAttributeCheckedAsNull", locals: true, onComparison: true}},
}

// ownedComparison reports whether check cmp comparing x with constant lit
// (true, false or null) is another rule's finding: x, or for a rule
// following locals a value assigned to local x, is a call comparisonOwners
// lists for lit, in a comparison of the shape that rule reports, and that
// rule's check of the comparison or of the call runs.
func ownedComparison(ctx *analysis.Context, cmp, x syntax.Expr, lit string) bool {
	owners := comparisonOwners[lit]
	for i, v := range origins(ctx, x) {
		for key, o := range owners {
			if key == "xml_parse" && (len(ctx.File.Errors) > 0 || !NativeCallbackReachable(ctx, cmp)) {
				continue
			}
			if key == "collator::compare" && !flowquery.Reachable(cmp, syntax.EnclosingFuncLike(cmp)) {
				continue
			}
			if (i == 0 || o.locals) && ownerShape(o, cmp, v) && ownerCall(ctx, v, key) && ctx.RuleChecks(o.rule, ownerNode(o, cmp, v)) {
				return true
			}
		}
	}
	return false
}

// ownerShape reports whether owner o reports check cmp of value v: v is
// not wrapped in parentheses when o needs a bare operand, and cmp is an
// equality when o reports only those.
func ownerShape(o comparisonOwner, cmp, v syntax.Expr) bool {
	if o.onComparison {
		if _, ok := cmp.(*syntax.Binary); !ok {
			return false
		}
	}
	if o.bare && v.Parent() != syntax.Node(cmp) {
		return false
	}
	if o.equality {
		k := cmp.(*syntax.Binary).Op.Kind
		return k == syntax.TIsIdentical || k == syntax.TIsEqual
	}
	return true
}

// ownerNode returns the node owner o checks for check cmp of call
// v.
func ownerNode(o comparisonOwner, cmp, v syntax.Expr) syntax.Node {
	if o.onComparison {
		return cmp
	}
	return v
}

// ownerCall reports whether x calls the builtin comparisonOwners key names:
// the global function, or for `class::method` that builtin method on a
// receiver natively typed as that class or a subclass, as the owning rules
// match it (NativeMethod).
func ownerCall(ctx *analysis.Context, x syntax.Expr, key string) bool {
	switch c := x.(type) {
	case *syntax.FuncCall:
		return ctx.GlobalFunctionName(c) == key
	case *syntax.MethodCall:
		class, method, ok := strings.Cut(key, "::")
		return ok && NativeMethod(ctx, c, class, method)
	}
	return false
}

// origins returns x and, for a local, every value a plain assignment in its
// scope gives it (parentheses stripped).
func origins(ctx *analysis.Context, x syntax.Expr) []syntax.Expr {
	vals := []syntax.Expr{x}
	if v, ok := x.(*syntax.Variable); ok {
		for _, a := range flowquery.VarAccesses(ctx.File, syntax.EnclosingFuncLike(v), v.Name) {
			if as, ok := a.By.(*syntax.Assign); ok && a.Write && as.Var == syntax.Expr(a.Var) {
				vals = append(vals, syntax.UnwrapParens(as.Value))
			}
		}
	}
	return vals
}

// laterVersions are the PHP versions a builtin's return type is compared
// at (see versionedOperand).
var laterVersions = []phpversion.Version{
	phpversion.PHP70, phpversion.PHP71, phpversion.PHP72, phpversion.PHP73, phpversion.PHP74,
	phpversion.PHP80, phpversion.PHP81, phpversion.PHP82, phpversion.PHP83, phpversion.PHP84, phpversion.PHP85,
}

// versionedOperand reports whether x, or a value assigned to local x, is a
// call to a builtin whose declared return type changes between the target
// PHP version and a later one (`sort()` returns bool before PHP 8.2, true
// after): the code may be meant for both.
func versionedOperand(ctx *analysis.Context, x syntax.Expr) bool {
	for _, v := range origins(ctx, x) {
		if versionedCall(ctx, v) {
			return true
		}
	}
	return false
}

func versionedCall(ctx *analysis.Context, x syntax.Expr) bool {
	ix := ctx.Index()
	switch c := x.(type) {
	case *syntax.FuncCall:
		f := ctx.Types().ResolveFunction(c)
		if f == nil || !f.Builtin {
			return false
		}
		return returnVaries(ctx.PHP, func(v phpversion.Version) string { return functionReturn(ix.Function(f.FQN, v)) })
	case *syntax.MethodCall:
		return methodVaries(ctx, ctx.TypeOf(c.Var).Classes(), identName(c.Name))
	case *syntax.StaticCall:
		return methodVaries(ctx, []string{ctx.Types().ClassRef(c.Class)}, identName(c.Name))
	}
	return false
}

// methodVaries reports whether a builtin method name of one of classes
// declares a return type that changes after the target version.
func methodVaries(ctx *analysis.Context, classes []string, name string) bool {
	ix := ctx.Index()
	for _, cls := range classes {
		cls = strings.TrimPrefix(cls, `\`)
		if m := ix.FindMethod(cls, name, ctx.PHP); m != nil && m.Builtin &&
			returnVaries(ctx.PHP, func(v phpversion.Version) string { return methodReturn(ix.FindMethod(cls, name, v)) }) {
			return true
		}
	}
	return false
}

// functionReturn and methodReturn print the declared return types of a
// declaration ("" for none).
func functionReturn(f *index.Function) string {
	if f == nil {
		return ""
	}
	return f.Return + ";" + f.DocReturn
}

func methodReturn(m *index.Method) string {
	if m == nil {
		return ""
	}
	return m.Return + ";" + m.DocReturn
}

// identName returns the name an identifier writes ("" for an expression).
func identName(n syntax.Expr) string {
	if id, ok := n.(*syntax.Identifier); ok {
		return id.Value
	}
	return ""
}

// returnVaries reports whether at gives another answer at a version after
// from than at from.
func returnVaries(from phpversion.Version, at func(phpversion.Version) string) bool {
	base := at(from)
	for _, v := range laterVersions {
		if v > from && at(v) != base {
			return true
		}
	}
	return false
}
