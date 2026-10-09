package infer

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// Conditional return types (`@return ($x is string ? int : float)`,
// index.Function.CondReturn): a call takes the branch its argument
// decides, recursively through nested conditionals; a test the argument
// does not decide keeps both branches. A call where no test is decided is
// typed as before (the union of every branch, from the flattened doc type).

// condCall types a call to a function or method whose documented return
// type is the conditional cond (canonical form), given its parameters, the
// call's arguments, its declared return type and its @return type (doc;
// the flattened conditional when @return holds it). ok is false when no
// test is decided and doc is set (the call is typed as before), or the
// result disagrees with the declared type. With no @return (the
// conditional is a @phpstan-return), an undecided call gets the union of
// the branches.
func (e *Env) condCall(cond string, params []index.Param, args *syntax.ArgList, declared, doc string) (types.Type, bool) {
	c := e.parsedCond(cond)
	if c == nil {
		return types.Unknown, false
	}
	decided := false
	var t types.Type
	if args == nil || isFirstClassCallable(args) {
		t = e.condType(c, nil, nil, &decided)
	} else {
		t = e.condType(c, params, args, &decided)
	}
	if (!decided && doc != "") || t.IsUnknown() {
		return types.Unknown, false
	}
	return e.withDeclared(t, declared)
}

// parsedCond parses a canonical conditional once per Env (nil when it does
// not parse, e.g. a stale or hand-edited index).
func (e *Env) parsedCond(cond string) *types.Cond {
	if c, ok := e.conds[cond]; ok {
		return c
	}
	c, ok := types.ParseCond(cond, nil, nil)
	if !ok {
		c = nil
	}
	if e.conds == nil {
		e.conds = map[string]*types.Cond{}
	}
	e.conds[cond] = c
	return c
}

// condType is the type of conditional c for the call: the decided branch,
// else the union of both (decided records that a test was decided).
func (e *Env) condType(c *types.Cond, params []index.Param, args *syntax.ArgList, decided *bool) types.Type {
	branch := func(b types.CondBranch) types.Type {
		if b.Cond != nil {
			return e.condType(b.Cond, params, args, decided)
		}
		return types.FromDoc(b.Type, nil)
	}
	switch e.condTest(c, params, args) {
	case condThen:
		*decided = true
		return branch(c.Then)
	case condElse:
		*decided = true
		return branch(c.Else)
	}
	return types.Union(branch(c.Then), branch(c.Else))
}

type condOutcome int8

const (
	condUnknown condOutcome = iota
	condThen
	condElse
)

// argValue describes the value bound to a parameter: its type and, for
// literals and class constants, the value itself.
type argValue struct {
	typ        types.Type
	str        string // string literal value (isStr)
	isStr      bool
	num        string // number literal text
	cls, cname string // class constant
}

// condTest decides the test of c for the call (condUnknown when the
// argument does not decide it).
func (e *Env) condTest(c *types.Cond, params []index.Param, args *syntax.ArgList) condOutcome {
	if args == nil {
		return condUnknown
	}
	i := -1
	for j, p := range params {
		if p.Name == c.Param {
			i = j
		}
	}
	if i < 0 {
		return condUnknown
	}
	var v argValue
	if x := tplArg(args, params, i); x != nil {
		v = e.exprValue(x)
	} else {
		for _, a := range args.Args {
			if a, ok := a.(*syntax.Arg); !ok || a.Unpack {
				return condUnknown // `...$args` may bind it
			}
		}
		if params[i].Variadic || params[i].Default == "" {
			return condUnknown
		}
		v = e.textValue(params[i].Default)
	}
	r := e.testValue(c, v)
	if r != condUnknown && c.Negated {
		r = condThen + condElse - r
	}
	return r
}

// exprValue describes argument expression x.
func (e *Env) exprValue(x syntax.Expr) argValue {
	v := argValue{typ: e.TypeOf(x)}
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal:
		switch n.LitKind {
		case syntax.LitString:
			v.str, v.isStr = plainString(n.Raw)
		default:
			v.num = n.Raw
		}
	case *syntax.Unary:
		if lit, ok := syntax.UnwrapParens(n.Expr).(*syntax.Literal); ok && n.Op.Kind == syntax.TMinus && lit.LitKind != syntax.LitString {
			v.num = "-" + lit.Raw
		}
	case *syntax.ClassConstFetch:
		id, ok := n.Name.(*syntax.Identifier)
		if nm, named := n.Class.(*syntax.Name); ok && named && !strings.EqualFold(id.Value, "class") {
			v.cls, v.cname = e.classRef(nm), id.Value
		}
	}
	return v
}

// textValue describes a parameter's default value from its source text.
func (e *Env) textValue(s string) argValue {
	s = strings.TrimSpace(s)
	v := argValue{typ: literalTextType(s)}
	switch strings.ToLower(s) {
	case "true", "false", "null":
		v.typ = types.Of(strings.ToLower(s))
	}
	switch {
	case len(s) >= 2 && (s[0] == '\'' || s[0] == '"'):
		v.str, v.isStr = plainString(s)
	case v.typ.OnlyOf("int") || v.typ.OnlyOf("float"):
		v.num = s
	default:
		if cls, name, ok := strings.Cut(s, "::"); ok && !strings.EqualFold(name, "class") {
			v.cls, v.cname = strings.TrimPrefix(cls, `\`), name
		}
	}
	return v
}

// constValue returns the initialiser text of class constant cls::name
// ("" when unknown).
func (e *Env) constValue(cls, name string) string {
	if k := e.Index.FindConst(cls, name, e.PHP); k != nil && !k.Case {
		return strings.TrimSpace(k.Value)
	}
	return ""
}

// literalOf returns a comparable form of v's value: `s:<value>` for a
// string, `n:<value>` for a number ("" when not a literal), following a
// class constant to its initialiser.
func (e *Env) literalOf(v argValue) string {
	switch {
	case v.isStr:
		return "s:" + v.str
	case v.num != "":
		return numKey(v.num)
	case v.cls != "":
		return e.textLiteral(e.constValue(v.cls, v.cname))
	}
	return ""
}

// textLiteral is literalOf for a literal's source text.
func (e *Env) textLiteral(s string) string {
	if s == "" {
		return ""
	}
	tv := e.textValue(s)
	if tv.cls != "" {
		return "" // a constant defined by another constant: not followed
	}
	return e.literalOf(tv)
}

// numKey normalises a number literal ("" when it does not parse).
func numKey(s string) string {
	f, err := strconv.ParseFloat(strings.ReplaceAll(s, "_", ""), 64)
	if err != nil {
		return "" // hexadecimal, octal, binary: not compared
	}
	return "n:" + strconv.FormatFloat(f, 'g', -1, 64)
}

// testValue decides whether value v passes the (non-negated) test of c.
func (e *Env) testValue(c *types.Cond, v argValue) condOutcome {
	switch c.TargetKind() {
	case types.TargetString, types.TargetNumber, types.TargetConst:
		var want string
		var wantType types.Type
		if c.TargetKind() == types.TargetConst {
			cls, name, _ := strings.Cut(strings.TrimPrefix(c.Target, `\`), "::")
			if v.cls != "" && strings.EqualFold(v.cls, cls) && v.cname == name {
				return condThen
			}
			val := e.constValue(cls, name)
			want, wantType = e.textLiteral(val), literalTextType(val)
		} else {
			want, wantType = e.textLiteral(c.Target), literalTextType(c.Target)
		}
		if got := e.literalOf(v); want != "" && got != "" {
			if got == want {
				return condThen
			}
			return condElse
		}
		if !wantType.IsUnknown() && !v.typ.IsUnknown() && !overlaps(v.typ, wantType) {
			return condElse
		}
		return condUnknown
	}
	if v.typ.IsUnknown() {
		return condUnknown
	}
	target := types.FromDoc(c.Target, nil)
	if c.Exact && e.within(v.typ, target) {
		return condThen
	}
	if !overlaps(v.typ, target) {
		return condElse
	}
	return condUnknown
}

// within reports whether every atom of t belongs to type target.
func (e *Env) within(t, target types.Type) bool {
	if target.Has("mixed") {
		return true
	}
	for _, a := range t.Atoms() {
		if !e.atomIn(a, target) {
			return false
		}
	}
	return true
}

// Value categories of atoms (see atomCats): two types may share a value
// only when their categories meet.
const (
	catNull = 1 << iota
	catTrue
	catFalse
	catInt
	catFloat
	catString
	catArray
	catObject
	catResource
	catAll = 1<<iota - 1
)

// atomCats returns the categories of values atom a may hold.
func atomCats(a string) int {
	switch a {
	case "null", "void":
		return catNull
	case "true":
		return catTrue
	case "false":
		return catFalse
	case "bool":
		return catTrue | catFalse
	case "int":
		return catInt
	case "float":
		return catFloat
	case "string":
		return catString
	case "resource":
		return catResource
	case "never":
		return 0
	case "callable":
		return catString | catArray | catObject
	case "iterable":
		return catArray | catObject
	case "object", "static", "self", "parent":
		return catObject
	case "array":
		return catArray
	}
	switch {
	case strings.HasSuffix(a, "[]"):
		return catArray
	case strings.HasPrefix(a, `\`):
		return catObject
	}
	return catAll // mixed
}

// overlaps reports whether types a and b may share a value.
func overlaps(a, b types.Type) bool {
	ca, cb := 0, 0
	for _, x := range a.Atoms() {
		ca |= atomCats(x)
	}
	for _, x := range b.Atoms() {
		cb |= atomCats(x)
	}
	return ca&cb != 0
}
