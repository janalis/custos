package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// ReturnMismatch is a returned value that the return type its function
// declares never accepts.
type ReturnMismatch struct {
	Value syntax.Expr
	// Function names the declaration (`f`, `Class::method`, `the closure`).
	Function string
	// Expected is the rejecting type, as declared.
	Expected string
	// Given is the value's type.
	Given types.Type
	// Proven tells that the native return type rejects the value's proven
	// type (see infer.Env.ProvenTypeOf); otherwise the `@return` type
	// rejects it, or the native type rejects only its PHPDoc-aware type.
	Proven bool
}

// Message is the finding's text: the type to return instead.
func (m ReturnMismatch) Message() string {
	return "Return " + declaredTypeName(m.Expected) + " from " + m.Function + ", not " + typeName(m.Given) + "."
}

type returnTypesKey struct{ n syntax.Node }

// ReturnTypeMismatch judges the value n returns: a `return` statement with
// a value, or the body of an arrow function. It reports the value when the
// return type its own function-like declares rejects every value of its
// type (see types.AllRejected): the native type, else the `@return` type.
// Generators, property hooks, top-level code and relative types in traits
// are not judged, nor `$this` in a closure. Cached per node for the file.
func ReturnTypeMismatch(ctx *analysis.Context, n syntax.Node) (ReturnMismatch, bool) {
	r := ctx.Shared(returnTypesKey{n}, func() any {
		m, ok := returnTypeMismatch(ctx, n)
		if !ok {
			return (*ReturnMismatch)(nil)
		}
		return &m
	}).(*ReturnMismatch)
	if r == nil {
		return ReturnMismatch{}, false
	}
	return *r, true
}

// returnContract is the return contract of one function-like, prepared for
// the checks of its returned values.
type returnContract struct {
	skip                 bool // a generator, or nothing to compare
	label                string
	accept               types.AcceptContext
	native, doc, wideDoc types.Type
	nativeText, docText  string
}

type returnContractKey struct{ fn syntax.Node }

func returnTypeMismatch(ctx *analysis.Context, n syntax.Node) (ReturnMismatch, bool) {
	var fn syntax.Node
	var value syntax.Expr
	switch r := n.(type) {
	case *syntax.Return:
		fn, value = syntax.EnclosingFuncLike(r), r.Expr
	case *syntax.ArrowFunction:
		fn, value = r, r.Expr
	}
	if value == nil {
		return ReturnMismatch{}, false
	}
	c := ctx.Shared(returnContractKey{fn}, func() any { return newReturnContract(ctx, fn) }).(*returnContract)
	if c.skip {
		return ReturnMismatch{}, false
	}
	given := ctx.TypeOf(value)
	if given.IsUnknown() {
		return ReturnMismatch{}, false
	}
	accept := c.accept
	m := ReturnMismatch{Value: value, Function: c.label, Given: given}
	switch {
	case types.AllRejected(c.native, given, accept):
		m.Expected = c.nativeText
		// Proven: the native type also rejects what the code alone shows.
		if proven := ctx.Types().ProvenTypeOf(value); types.AllRejected(c.native, proven, accept) {
			m.Proven, m.Given = true, proven
		}
	case !c.doc.IsUnknown() && types.AllRejected(c.docFor(given), given, accept):
		m.Expected = c.docText
	default:
		return ReturnMismatch{}, false
	}
	// The proven type of `$this` in a closure is unknown; the PHPDoc-aware
	// one still names the enclosing class.
	if !m.Proven && reboundThis(value) {
		return ReturnMismatch{}, false
	}
	return m, true
}

// newReturnContract prepares the return contract of fn (nil outside a
// function-like).
func newReturnContract(ctx *analysis.Context, fn syntax.Node) *returnContract {
	c := &returnContract{}
	var ret syntax.Expr
	switch f := fn.(type) {
	case *syntax.Function:
		ret, c.label = f.ReturnType, functionLabel(ctx, f)
	case *syntax.Method:
		ret, c.label = f.ReturnType, methodLabel(ctx, f)
	case *syntax.Closure:
		ret, c.label = f.ReturnType, "the closure"
	case *syntax.ArrowFunction:
		ret, c.label = f.ReturnType, "the arrow function"
	default:
		c.skip = true // top-level code, property hooks
		return c
	}
	contract := ctx.Types().ReturnContract(fn)
	c.native, c.doc, c.docText = contract.Native, contract.Doc, contract.DocText
	if ret != nil {
		c.nativeText = ctx.Text(ret)
	}
	// A @return type the native type rejects is stale; one that only
	// narrows values (ranges, lengths, literals) counts against exact
	// values only.
	check := fileTypeCheck(ctx)
	c.accept = check.weak
	if !c.doc.IsUnknown() && !c.native.IsUnknown() && types.AllRejected(c.native, c.doc, check.strict) {
		c.doc = types.Unknown
	}
	if c.native.IsUnknown() && c.doc.IsUnknown() || generatorBody(fn) {
		c.skip = true
	} else if !c.doc.IsUnknown() {
		c.wideDoc = types.FromDoc(c.doc.DocString(), nil)
	}
	return c
}

// docFor is the @return type to judge given against: value facts count
// only when given lists exact values.
func (c *returnContract) docFor(given types.Type) types.Type {
	if exactValues(given) {
		return c.doc
	}
	return c.wideDoc
}

// exactValues reports whether every int member of t lists its possible
// values and every string member its values or its byte lengths (a string
// literal).
func exactValues(t types.Type) bool {
	if t.Has("int") {
		rs, ok := t.IntRanges()
		if !ok || !singleValues(rs) {
			return false
		}
	}
	if _, listed := t.StringValueList(); t.Has("string") && !listed {
		rs, ok := t.StringLengthRanges()
		return ok && singleValues(rs)
	}
	return true
}

// singleValues reports whether every range of rs holds one value.
func singleValues(rs []types.IntRange) bool {
	for _, r := range rs {
		if r.Lo != r.Hi {
			return false
		}
	}
	return true
}

// generatorBody reports whether function-like fn yields (nested
// function-likes and classes aside).
func generatorBody(fn syntax.Node) bool {
	switch f := fn.(type) {
	case *syntax.Function:
		return NativeGeneratorBody(f.Body)
	case *syntax.Method:
		return NativeGeneratorBody(f.Body)
	case *syntax.Closure:
		return NativeGeneratorBody(f.Body)
	}
	return yields(fn.(*syntax.ArrowFunction).Expr)
}

// functionLabel names function f: its FQN followed by `()`.
func functionLabel(ctx *analysis.Context, f *syntax.Function) string {
	name := f.Name.Value
	if ns := ctx.Names().Namespace(f.Span().Start); ns != "" {
		name = ns + `\` + name
	}
	return name + "()"
}

// methodLabel names method m: `Class::method()`.
func methodLabel(ctx *analysis.Context, m *syntax.Method) string {
	return ctx.Types().ClassFQN(syntax.EnclosingClass(m)) + "::" + m.Name.Value + "()"
}
