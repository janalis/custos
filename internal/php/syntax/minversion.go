package syntax

import (
	"sort"
	"strings"

	phpversion "github.com/janalis/custos/internal/php/version"
)

// MinVersion returns the lowest PHP version whose grammar or meaning supports
// the syntax that node n itself introduces (its children are not included).
// Nodes using only syntax every supported version accepts return
// phpversion.Min. It is metadata only: it never changes parsing.
//
// The table is deliberately conservative: when a construct may or may not
// need a newer version (e.g. string keys in an array spread), the lower
// version is returned, so callers comparing two trees never reject code
// because of a guess.
func MinVersion(n Node) phpversion.Version {
	if n == nil {
		return phpversion.Min
	}
	return minVersion(nil, n)
}

// RequiredVersion returns the highest MinVersion over every node of f, and
// the first node (in source order) that requires it, or nil when the file
// only uses syntax available in phpversion.Min.
func RequiredVersion(f *File) (phpversion.Version, Node) {
	best, at := phpversion.Min, Node(nil)
	InspectFile(f, func(n Node) bool {
		if v := minVersion(f, n); v > best {
			best, at = v, n
		}
		return true
	})
	return best, at
}

func minVersion(f *File, n Node) phpversion.Version {
	v := phpversion.Min
	at := func(o phpversion.Version) {
		if o > v {
			v = o
		}
	}
	switch x := n.(type) {
	case *PropertyHook:
		at(phpversion.PHP84)
	case *ClassConst:
		at(classConstVersion(x))
	case *Property:
		at(propertyVersion(x))
	case *Param:
		at(paramVersion(x))
	case *New:
		at(newVersion(x))
	case *Function:
		at(signatureVersion(f, x, x.Params, x.ReturnType, nil))
	case *Method:
		at(signatureVersion(f, x, x.Params, x.ReturnType, nil))
	case *Closure:
		if x.Static {
			at(phpversion.PHP54)
		}
		at(signatureVersion(f, x, x.Params, x.ReturnType, x.Uses))
	case *ArrowFunction:
		at(phpversion.PHP74)
		at(signatureVersion(f, x, x.Params, x.ReturnType, nil))
	case *Array:
		at(arrayVersion(x))
	case *List:
		at(listItemsVersion(x.Items))
	case *ArrayItem:
		if x.Unpack && !IsDestructuring(x.Parent()) {
			at(phpversion.PHP74)
		}
	case *Arg:
		if x.Unpack {
			at(phpversion.PHP56)
		}
		if x.Name != nil {
			at(phpversion.PHP80)
		}
	case *VariadicPlaceholder:
		at(phpversion.PHP81)
	case *ArgList:
		if len(x.Args) > 0 && trailingComma(f, x, x.Args[len(x.Args)-1]) {
			at(phpversion.PHP73)
		}
	case *Isset:
		at(trailingListVersion(f, x, x.Vars))
	case *Unset:
		at(trailingListVersion(f, x, x.Vars))
	case *Binary:
		at(operatorVersion(x.Op.Kind))
	case *Assign:
		at(operatorVersion(x.Op.Kind))
	case *Unary:
		if x.Op.Kind == TVoidCast {
			at(phpversion.PHP85)
		}
	case *PropertyFetch:
		at(nullsafeVersion(x.NullSafe))
		at(dereferenceVersion(x.Var))
	case *MethodCall:
		at(nullsafeVersion(x.NullSafe))
		at(dereferenceVersion(x.Var))
	case *ArrayDimFetch:
		at(dimFetchVersion(x))
	case *FuncCall:
		at(callableVersion(x.Name))
	case *ClassConstFetch:
		at(classConstFetchVersion(x))
	case *Empty:
		if !isWritable(UnwrapParens(x.Expr)) {
			at(phpversion.PHP55)
		}
	case *Yield:
		at(phpversion.PHP55)
	case *YieldFrom:
		at(phpversion.PHP70)
	case *Foreach:
		at(foreachVersion(x.Value))
	case *Try:
		if x.Finally != nil {
			at(phpversion.PHP55)
		}
	case *Catch:
		if len(x.Types) > 1 {
			at(phpversion.PHP71)
		}
		if x.Var == nil {
			at(phpversion.PHP80)
		}
	case *Use:
		at(useVersion(f, x))
	case *ConstStmt:
		for _, c := range x.Consts {
			at(constValueVersion(c.Value))
		}
	case *StaticVar:
		at(initializerVersion(x.Default))
	case *ClassLike:
		at(classLikeVersion(x))
	case *TraitUse:
		at(phpversion.PHP54)
	case *EnumCase:
		at(phpversion.PHP81)
	case *Match:
		at(phpversion.PHP80)
	case *Throw:
		if _, stmt := x.Parent().(*ExprStmt); !stmt {
			at(phpversion.PHP80)
		}
	case *AttributeGroup:
		at(phpversion.PHP80)
	case *Clone:
		if x.Args != nil && len(x.Args.Args) != 1 {
			at(phpversion.PHP85)
		}
	case *MagicConst:
		at(magicConstVersion(x.Token.Kind))
	case *Literal:
		at(literalVersion(f, x))
	case *InterpolatedString:
		at(interpolatedVersion(f, x))
	}
	return v
}

func asymmetricVisibility(mods Modifiers) bool {
	return mods.Has(TPrivateSet) || mods.Has(TProtectedSet) || mods.Has(TPublicSet)
}

func maxVersion(a, b phpversion.Version) phpversion.Version {
	if a > b {
		return a
	}
	return b
}

// ---- declarations ------------------------------------------------------------------

func classConstVersion(x *ClassConst) phpversion.Version {
	v := phpversion.Min
	if x.Type != nil {
		v = phpversion.PHP83
	}
	if x.Modifiers.Has(TPublic) || x.Modifiers.Has(TProtected) || x.Modifiers.Has(TPrivate) {
		v = maxVersion(v, phpversion.PHP71)
	}
	if x.Modifiers.Has(TFinal) {
		v = maxVersion(v, phpversion.PHP81)
	}
	if c, ok := x.Parent().(*ClassLike); ok && c.ClassKind == KindTrait {
		v = maxVersion(v, phpversion.PHP82)
	}
	for _, c := range x.Consts {
		v = maxVersion(v, constValueVersion(c.Value))
	}
	return v
}

func propertyVersion(x *Property) phpversion.Version {
	v := phpversion.Min
	if asymmetricVisibility(x.Modifiers) {
		v = phpversion.PHP84
	}
	if x.Type != nil {
		v = maxVersion(v, maxVersion(phpversion.PHP74, typeVersion(x.Type)))
	}
	if x.Modifiers.Has(TReadonly) {
		v = maxVersion(v, phpversion.PHP81)
	}
	for _, p := range x.Props {
		v = maxVersion(v, initializerVersion(p.Default))
	}
	return v
}

func paramVersion(x *Param) phpversion.Version {
	v := phpversion.Min
	if asymmetricVisibility(x.Modifiers) {
		v = phpversion.PHP84
	}
	if x.Type != nil {
		v = maxVersion(v, typeVersion(x.Type))
	}
	if x.Variadic {
		v = maxVersion(v, phpversion.PHP56)
	}
	if x.Modifiers.Has(TReadonly) {
		v = maxVersion(v, phpversion.PHP81)
	}
	if len(x.Modifiers) > 0 {
		v = maxVersion(v, phpversion.PHP80)
	}
	return maxVersion(v, initializerVersion(x.Default))
}

// signatureVersion covers what a function-like adds beyond its parameters:
// a return type and trailing commas after the parameters or closure uses.
func signatureVersion(f *File, n Node, params []*Param, ret Expr, uses []*ClosureUse) phpversion.Version {
	v := phpversion.Min
	if ret != nil {
		v = maxVersion(phpversion.PHP70, typeVersion(ret))
	}
	if len(params) > 0 && trailingComma(f, n, params[len(params)-1]) {
		v = maxVersion(v, phpversion.PHP80)
	}
	if len(uses) > 0 && trailingComma(f, n, uses[len(uses)-1]) {
		v = maxVersion(v, phpversion.PHP80)
	}
	return v
}

// typeVersion is the version a parameter, property or return type needs.
func typeVersion(t Expr) phpversion.Version {
	switch x := t.(type) {
	case *NullableType:
		return maxVersion(phpversion.PHP71, typeVersion(x.Type))
	case *UnionType:
		v := phpversion.PHP80
		for _, m := range x.Types {
			if _, ok := m.(*IntersectionType); ok {
				v = maxVersion(v, phpversion.PHP82)
			}
			v = maxVersion(v, memberTypeVersion(m))
		}
		return v
	case *IntersectionType:
		return phpversion.PHP81
	case *Name:
		switch strings.ToLower(x.Value) {
		case "null", "false", "true":
			return phpversion.PHP82
		}
		return memberTypeVersion(x)
	}
	return phpversion.Min
}

// memberTypeVersion is typeVersion for one member of a union: null and false
// are allowed there since 8.0, true since 8.2.
func memberTypeVersion(t Expr) phpversion.Version {
	x, ok := t.(*Name)
	if !ok {
		return typeVersion(t)
	}
	switch strings.ToLower(x.Value) {
	case "callable":
		return phpversion.PHP54
	case "int", "float", "string", "bool":
		return phpversion.PHP70
	case "iterable", "void":
		return phpversion.PHP71
	case "object":
		return phpversion.PHP72
	case "mixed", "static", "null", "false":
		return phpversion.PHP80
	case "never":
		return phpversion.PHP81
	case "true":
		return phpversion.PHP82
	}
	return phpversion.Min
}

func classLikeVersion(x *ClassLike) phpversion.Version {
	v := phpversion.Min
	switch x.ClassKind {
	case KindTrait:
		v = phpversion.PHP54
	case KindEnum:
		v = phpversion.PHP81
	}
	if x.Name == nil {
		v = maxVersion(v, phpversion.PHP70)
	}
	if x.Modifiers.Has(TReadonly) {
		v = maxVersion(v, phpversion.PHP82)
	}
	return v
}

func useVersion(f *File, x *Use) phpversion.Version {
	v := phpversion.Min
	if x.Type != UseNormal {
		v = phpversion.PHP56
	}
	if x.Prefix == nil {
		return v
	}
	v = phpversion.PHP70
	if len(x.Items) > 0 && trailingComma(f, x, x.Items[len(x.Items)-1]) {
		v = maxVersion(v, phpversion.PHP72)
	}
	return v
}

// ---- initializers ------------------------------------------------------------------

// constValueVersion is initializerVersion for `const` values, where arrays
// are only allowed since 5.6.
func constValueVersion(e Expr) phpversion.Version {
	v := initializerVersion(e)
	if _, ok := e.(*Array); ok {
		v = maxVersion(v, phpversion.PHP56)
	}
	return v
}

// initializerVersion is the version a compile-time initializer (constant,
// property, parameter or static variable default) needs: anything beyond a
// static scalar needs 5.6 constant expressions, `new` needs 8.1.
func initializerVersion(e Expr) phpversion.Version {
	v := phpversion.Min
	if e == nil {
		return v
	}
	Inspect(e, func(n Node) bool {
		switch x := n.(type) {
		case *New:
			v = maxVersion(v, phpversion.PHP81)
		case *Binary, *Ternary, *Paren, *ArrayDimFetch:
			v = maxVersion(v, phpversion.PHP56)
		case *Unary:
			if x.Op.Kind != TMinus && x.Op.Kind != TPlus {
				v = maxVersion(v, phpversion.PHP56)
			}
		}
		return true
	})
	return v
}

// ---- expressions -------------------------------------------------------------------

func operatorVersion(k TokenKind) phpversion.Version {
	switch k {
	case TPow, TPowEqual:
		return phpversion.PHP56
	case TCoalesce, TSpaceship:
		return phpversion.PHP70
	case TCoalesceEqual:
		return phpversion.PHP74
	case TPipe:
		return phpversion.PHP85
	}
	return phpversion.Min
}

func nullsafeVersion(nullsafe bool) phpversion.Version {
	if nullsafe {
		return phpversion.PHP80
	}
	return phpversion.Min
}

// dereferenceVersion covers `(new A)->m()` (5.4) as the object of a member
// access; the unparenthesized form is classified on the New node.
func dereferenceVersion(obj Expr) phpversion.Version {
	if p, ok := obj.(*Paren); ok {
		if _, ok := p.Expr.(*New); ok {
			return phpversion.PHP54
		}
	}
	return phpversion.Min
}

func dimFetchVersion(x *ArrayDimFetch) phpversion.Version {
	switch o := x.Var.(type) {
	case *FuncCall, *MethodCall, *StaticCall:
		return phpversion.PHP54
	case *Array:
		return phpversion.PHP55
	case *Literal:
		if o.LitKind == LitString {
			return phpversion.PHP55
		}
	}
	return dereferenceVersion(x.Var)
}

// callableVersion classifies what is called in `<expr>(...)`: names,
// variables, offsets and static properties work everywhere, an array
// literal since 5.4, any other expression since 7.0 (uniform variable
// syntax).
func callableVersion(name Expr) phpversion.Version {
	switch name.(type) {
	case *Name, *Variable, *ArrayDimFetch:
		return phpversion.Min
	case *Array:
		return phpversion.PHP54
	}
	return phpversion.PHP70
}

func classConstFetchVersion(x *ClassConstFetch) phpversion.Version {
	id, ok := x.Name.(*Identifier)
	if !ok {
		return phpversion.PHP83
	}
	if !strings.EqualFold(id.Value, "class") {
		return phpversion.Min
	}
	if _, ok := x.Class.(*Name); ok {
		return phpversion.PHP55
	}
	return phpversion.PHP80
}

func newVersion(x *New) phpversion.Version {
	switch p := x.Parent().(type) {
	case *FuncCall:
		if p.Name == x {
			return phpversion.PHP84
		}
	case *MethodCall:
		if p.Var == x {
			return phpversion.PHP84
		}
	case *PropertyFetch:
		if p.Var == x {
			return phpversion.PHP84
		}
	case *ArrayDimFetch:
		if p.Var == x {
			return phpversion.PHP84
		}
	case *StaticCall:
		if p.Class == x {
			return phpversion.PHP84
		}
	case *StaticPropertyFetch:
		if p.Class == x {
			return phpversion.PHP84
		}
	case *ClassConstFetch:
		if p.Class == x {
			return phpversion.PHP84
		}
	}
	return phpversion.Min
}

// ---- arrays and destructuring ----------------------------------------------------------

func arrayVersion(x *Array) phpversion.Version {
	v := phpversion.Min
	if x.Short {
		v = phpversion.PHP54
	}
	if !IsDestructuring(x) {
		return v
	}
	if x.Short {
		v = phpversion.PHP71
	}
	return maxVersion(v, listItemsVersion(x.Items))
}

// listItemsVersion covers keyed (7.1) and by-reference (7.3) destructuring.
func listItemsVersion(items []*ArrayItem) phpversion.Version {
	v := phpversion.Min
	for _, it := range items {
		if it == nil {
			continue
		}
		if it.Key != nil {
			v = maxVersion(v, phpversion.PHP71)
		}
		if it.ByRef {
			v = maxVersion(v, phpversion.PHP73)
		}
	}
	return v
}

// IsDestructuring reports whether n is an array pattern being assigned to:
// the target of `=`, the value of a foreach, or nested in such a pattern.
func IsDestructuring(n Node) bool {
	for {
		switch p := n.Parent().(type) {
		case *Assign:
			return p.Var == n && p.Op.Kind == TEqual
		case *Foreach:
			return p.Value == n
		case *ArrayItem:
			n = p.Parent()
		default:
			return false
		}
	}
}

func foreachVersion(value Expr) phpversion.Version {
	if _, ok := value.(*List); ok {
		return phpversion.PHP55
	}
	return phpversion.Min
}

// ---- tokens and literals ---------------------------------------------------------------

func magicConstVersion(k TokenKind) phpversion.Version {
	switch k {
	case TTraitC:
		return phpversion.PHP54
	case TPropertyC:
		return phpversion.PHP84
	}
	return phpversion.Min
}

func literalVersion(f *File, x *Literal) phpversion.Version {
	switch x.LitKind {
	case LitInt, LitFloat:
		return numberVersion(x.Raw)
	}
	v := phpversion.Min
	if strings.HasPrefix(x.Raw, "<<<") {
		v = heredocVersion(f, x)
	}
	if !singleQuoted(x.Raw) {
		v = maxVersion(v, escapeVersion(x.Raw))
	}
	return v
}

func interpolatedVersion(f *File, x *InterpolatedString) phpversion.Version {
	v := phpversion.Min
	if x.Heredoc {
		v = heredocVersion(f, x)
	}
	for _, p := range x.Parts {
		if s, ok := p.(*StringPart); ok {
			v = maxVersion(v, escapeVersion(s.Raw))
		}
	}
	return v
}

// singleQuoted reports whether a string literal is single-quoted or a
// nowdoc, where escapes are not interpreted.
func singleQuoted(raw string) bool {
	raw = strings.TrimLeft(raw, "bB")
	if rest, ok := strings.CutPrefix(raw, "<<<"); ok {
		raw = strings.TrimLeft(rest, " \t")
	}
	return strings.HasPrefix(raw, "'")
}

func numberVersion(raw string) phpversion.Version {
	v := phpversion.Min
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "0b"):
		v = phpversion.PHP54
	case strings.HasPrefix(lower, "0o"):
		v = phpversion.PHP81
	}
	if strings.Contains(raw, "_") {
		v = maxVersion(v, phpversion.PHP74)
	}
	return v
}

// escapeVersion detects `\u{...}` escapes (7.0) in double-quoted text.
func escapeVersion(raw string) phpversion.Version {
	for i := 0; i+2 < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		if raw[i+1] == 'u' && raw[i+2] == '{' {
			return phpversion.PHP70
		}
		i++ // skip the escaped character, so `\\u{` stays literal
	}
	return phpversion.Min
}

// heredocVersion detects flexible heredoc/nowdoc (7.3): an indented closing
// marker, or one followed by anything other than an optional `;` and a
// line break.
func heredocVersion(f *File, n Node) phpversion.Version {
	f = fileOf(f, n)
	sp := n.Span()
	body := f.Src[sp.Start:sp.End]
	last := body[strings.LastIndexByte(string(body), '\n')+1:]
	if len(last) > 0 && (last[0] == ' ' || last[0] == '\t') {
		return phpversion.PHP73
	}
	rest := f.Src[sp.End:]
	if len(rest) > 0 && rest[0] == ';' {
		rest = rest[1:]
	}
	if len(rest) == 0 || rest[0] == '\n' || rest[0] == '\r' {
		return phpversion.Min
	}
	return phpversion.PHP73
}

func trailingListVersion(f *File, n Node, xs []Expr) phpversion.Version {
	if len(xs) > 0 && trailingComma(f, n, xs[len(xs)-1]) {
		return phpversion.PHP73
	}
	return phpversion.Min
}

// trailingComma reports whether the first significant token after last
// (inside its list owner n) is a comma.
func trailingComma(f *File, n, last Node) bool {
	f = fileOf(f, n)
	toks := f.Tokens
	end := last.Span().End
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= end })
	for ; i < len(toks); i++ {
		if !toks[i].Kind.IsTrivia() {
			return toks[i].Kind == TComma
		}
	}
	return false
}

// fileOf returns f, or the file n belongs to when f is nil.
func fileOf(f *File, n Node) *File {
	for f == nil {
		if x, ok := n.(*File); ok {
			return x
		}
		n = n.Parent()
	}
	return f
}
