package typeunsafecomparison

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// typeUnsafeComparison reports loose equality comparisons.
type typeUnsafeComparison struct{}

func (typeUnsafeComparison) ID() string               { return "TypeUnsafeComparison" }
func (typeUnsafeComparison) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (typeUnsafeComparison) Semantic()                {}

// comparableClasses are classes whose instances PHP compares meaningfully.
var comparableClasses = []string{
	`DateTime`, `DateTimeImmutable`, `IntlBreakIterator`, `IntlTimeZone`,
	`PDO`, `PDOStatement`, `ArrayObject`, `SplObjectStorage`, `Closure`,
}

var tucKeywordTypes = map[string]string{
	"int": "int", "integer": "int", "string": "string", "bool": "bool", "boolean": "bool",
	"true": "true", "false": "false", "float": "float", "null": "null", "void": "void",
	"mixed": "mixed", "callable": "callable", "resource": "resource", "iterable": "iterable",
	"object": "object", "static": "static", "self": "self", "$this": "$this",
}

func (typeUnsafeComparison) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsEqual && b.Op.Kind != syntax.TIsNotEqual {
		return
	}
	if b.Left == nil || b.Right == nil || b.Left.Span().Len() == 0 || b.Right.Span().Len() == 0 { // D4
		return
	}
	strict := "!=="
	if b.Op.Kind == syntax.TIsEqual {
		strict = "==="
	}
	span := b.Span()

	// Step 1.
	lit, other := b.Right, b.Left
	content, isLit := tucStringContent(ctx, lit)
	if !isLit {
		lit, other = b.Left, b.Right
		content, isLit = tucStringContent(ctx, lit)
	}
	if isLit {
		other = syntax.UnwrapParens(other)
		if cls, objectOnly := tucObjectWithoutToString(ctx, other); objectOnly { // D2
			// A nullable object compared with '' is a null test
			// (null == '' holds), not a string comparison.
			if val, known := tucLiteralValue(lit, content); cls != "" && known && val == "" && ctx.TypeOf(other).Has("null") {
				return
			}
			if cls != "" {
				ctx.ReportSeverity(span, diagnostic.SeverityError, cls+" has no __toString(), so it cannot be compared to a string.")
			}
			return
		}
		if val, known := tucLiteralValue(lit, content); known && val != "" && !astquery.IsComparisonNumericString(val) { // D3
			if !tucStrictSafe(ctx, other, val) {
				ctx.ReportSeverity(span, diagnostic.SeverityWarning, "Use '"+strict+"' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.")
				return
			}
			op := b.Op.Span
			ctx.ReportSeverity(span, diagnostic.SeverityWarning, "Use '"+strict+"' here; the string is not numeric, so strict comparison is safe.", diagnostic.Fix{
				Title: "Use " + strict,
				Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: op, NewText: strict}} },
			})
			return
		}
	}

	// Step 2.
	if tucComparableObject(ctx, b.Left) || tucComparableObject(ctx, b.Right) { // D5
		return
	}
	// Two objects: == compares their properties, === their identity — a
	// different question, not type juggling (custos diverges).
	if tucObjectsOnly(ctx, b.Left) && tucObjectsOnly(ctx, b.Right) {
		return
	}
	ctx.ReportSeverity(span, diagnostic.SeverityInfo, "Prefer '"+strict+"' to avoid implicit type juggling.") // D6
}

// tucStringContent returns the raw contents of a string literal operand
// (quoted, interpolated, heredoc or nowdoc; not parenthesised).
func tucStringContent(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	switch x := e.(type) {
	case *syntax.Literal:
		if x.LitKind != syntax.LitString {
			return "", false
		}
	case *syntax.InterpolatedString:
		if x.Backtick {
			return "", false
		}
	default:
		return "", false
	}
	t := ctx.Text(e)
	if t != "" && (t[0] == 'b' || t[0] == 'B') {
		t = t[1:]
	}
	if strings.HasPrefix(t, "<<<") {
		nl := strings.IndexByte(t, '\n') // a heredoc always has an opening newline
		last := strings.LastIndexByte(t, '\n')
		if last == nl { // empty body
			return "", true
		}
		return t[nl+1 : last], true
	}
	if len(t) >= 2 {
		return t[1 : len(t)-1], true
	}
	return "", true
}

// tucLiteralValue returns the runtime value of the string literal lit
// (raw contents given); known is false when the value cannot be determined
// statically (interpolation, escapes in a heredoc).
func tucLiteralValue(lit syntax.Expr, content string) (string, bool) {
	l, ok := lit.(*syntax.Literal)
	if !ok {
		return "", false // interpolated
	}
	raw := l.Raw
	if raw != "" && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	if v, ok := astquery.StringLiteralValue(raw); ok {
		return v, true
	}
	if strings.HasPrefix(raw, "<<<") {
		head := strings.TrimLeft(raw[3:], " \t")
		if strings.HasPrefix(head, "'") || !strings.Contains(content, "\\") {
			return content, true // nowdoc, or heredoc without escapes
		}
	}
	return "", false
}

// tucNormalizedParts returns the normalised type parts of e (nil when unknown).
func tucNormalizedParts(ctx *analysis.Context, e syntax.Expr) []string {
	return tucParts(ctx.TypeOf(e))
}

// tucParts returns the normalised parts of t (nil when unknown).
func tucParts(t types.Type) []string {
	if t.IsUnknown() {
		return nil
	}
	var out []string
	for _, a := range t.Atoms() {
		if strings.HasSuffix(a, "[]") {
			out = append(out, "array")
			continue
		}
		if k, ok := tucKeywordTypes[strings.ToLower(strings.TrimPrefix(a, `\`))]; ok {
			out = append(out, k)
			continue
		}
		out = append(out, a)
	}
	return out
}

// tucObjectWithoutToString implements D2. objectOnly reports whether every
// non-null part is a class; cls is the first resolved class lacking
// __toString() ("" when all have one).
func tucObjectWithoutToString(ctx *analysis.Context, e syntax.Expr) (cls string, objectOnly bool) {
	parts := tucNormalizedParts(ctx, e)
	var classes []string
	for _, p := range parts {
		if p == "null" {
			continue
		}
		if !strings.HasPrefix(p, `\`) {
			return "", false
		}
		classes = append(classes, p)
	}
	if len(classes) == 0 {
		return "", false
	}
	ix := ctx.Index()
	for _, c := range classes {
		decl := ix.Class(c, ctx.PHP)
		if decl == nil || !tucHierarchyKnown(ctx, decl.FQN) {
			continue
		}
		if ix.FindMethod(decl.FQN, "__toString", ctx.PHP) == nil {
			return `\` + strings.TrimPrefix(decl.FQN, `\`), true
		}
	}
	return "", true
}

// tucObjectsOnly reports whether e is known to hold only objects.
func tucObjectsOnly(ctx *analysis.Context, e syntax.Expr) bool {
	parts := tucNormalizedParts(ctx, e)
	for _, p := range parts {
		if p != "object" && !strings.HasPrefix(p, `\`) {
			return false
		}
	}
	return len(parts) > 0
}

// tucComparableObject implements D5.
func tucComparableObject(ctx *analysis.Context, e syntax.Expr) bool {
	ix := ctx.Index()
	for _, p := range tucNormalizedParts(ctx, e) {
		if !strings.HasPrefix(p, `\`) || ix.Class(p, ctx.PHP) == nil {
			continue
		}
		for _, c := range comparableClasses {
			if ix.IsSubtype(p, c, ctx.PHP) {
				return true
			}
		}
	}
	return false
}

// tucHierarchyKnown reports whether every parent, interface and trait of
// class (transitively) resolves: otherwise an unresolved ancestor may
// declare __toString() and D2 cannot decide.
func tucHierarchyKnown(ctx *analysis.Context, class string) bool {
	ix := ctx.Index()
	for _, c := range ix.Ancestors(class, ctx.PHP) {
		refs := append(append([]string{c.Parent}, c.Interfaces...), c.Traits...)
		for _, r := range refs {
			if r != "" && ix.Class(r, ctx.PHP) == nil {
				return false
			}
		}
	}
	return true
}

// tucStrictSafe reports whether `e == 'text'` (a non-empty, non-numeric
// string) always equals `e === 'text'`, from native types of stable variables
// only (spec F2, F3 and "Kind S fix only when equivalent"). text is the
// compared string, "" when e only flows into the compared variable.
func tucStrictSafe(ctx *analysis.Context, e syntax.Expr, text string) bool {
	parts := tucParts(ctx.Types().Native().TypeOf(e))
	if len(parts) == 0 {
		return false
	}
	infinity := text == "INF" || text == "-INF"
	for _, p := range parts {
		if !tucSafePart(ctx, p) || infinity && p == "float" {
			return false
		}
	}
	return tucProven(ctx, e, false)
}

// tucSafePart reports whether a value of the keyword type p compares the
// same loosely and strictly with a non-numeric string.
func tucSafePart(ctx *analysis.Context, p string) bool {
	switch p {
	case "string", "null", "array":
		return true
	case "int", "float":
		return ctx.PHP >= phpversion.PHP80
	}
	return false
}

// tucSafeDecl is tucSafePart over a native return type declaration.
func tucSafeDecl(ctx *analysis.Context, decl string) bool {
	for p := range strings.SplitSeq(strings.TrimPrefix(decl, "?"), "|") {
		if !tucSafePart(ctx, strings.ToLower(strings.TrimPrefix(p, `\`))) {
			return false
		}
	}
	return true
}

// tucProven reports whether the native type inferred for e can be relied
// on: e is built only from forms whose type PHP guarantees, and every
// variable it rests on is stable. via is set when e is the receiver of the
// compared member rather than the compared value itself.
func tucProven(ctx *analysis.Context, e syntax.Expr, via bool) bool {
	switch x := e.(type) {
	case *syntax.Paren:
		return tucProven(ctx, x.Expr, via)
	case *syntax.Literal, *syntax.InterpolatedString, *syntax.Array, *syntax.ConstFetch, *syntax.MagicConst, *syntax.New:
		return true
	case *syntax.Variable:
		return tucStableVar(ctx, x, via)
	case *syntax.Ternary:
		then := x.Then
		if then == nil {
			then = x.Cond
		}
		return tucBranch(ctx, then, via) && tucBranch(ctx, x.Else, via)
	case *syntax.Match:
		for _, arm := range x.Arms {
			if !tucBranch(ctx, arm.Body, via) {
				return false
			}
		}
		return true
	case *syntax.Unary:
		switch x.Op.Kind {
		case syntax.TStringCast, syntax.TIntCast, syntax.TDoubleCast, syntax.TArrayCast:
			return true // the result type does not depend on the operand
		}
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TDot:
			return true
		case syntax.TCoalesce:
			return tucBranch(ctx, x.Left, via) && tucBranch(ctx, x.Right, via)
		}
	case *syntax.FuncCall:
		// A natively declared return type holds whatever the arguments are.
		f := ctx.Types().ResolveFunction(x)
		return f != nil && tucSafeDecl(ctx, f.Return)
	case *syntax.StaticCall:
		return tucDeclared(ctx, []string{semanticquery.StaticCallClass(ctx, x.Class)}, x.Name, true, via)
	case *syntax.MethodCall:
		return tucDeclared(ctx, tucNativeClasses(ctx, x.Var), x.Name, true, via) && tucProven(ctx, x.Var, true)
	case *syntax.PropertyFetch:
		return tucDeclared(ctx, tucNativeClasses(ctx, x.Var), x.Name, false, via) && tucProven(ctx, x.Var, true)
	}
	return false
}

// tucBranch is tucProven for one alternative of a ternary, `??` or match:
// its own type must be known, since a union silently drops unknown parts.
func tucBranch(ctx *analysis.Context, e syntax.Expr, via bool) bool {
	return !ctx.Types().Native().TypeOf(e).IsUnknown() && tucProven(ctx, e, via)
}

func tucNativeClasses(ctx *analysis.Context, e syntax.Expr) []string {
	return ctx.Types().Native().TypeOf(e).Classes()
}

// tucDeclared reports whether the member `name` (a method, else a
// property) of every class in classes carries a native type declaration,
// which PHP enforces on every write or return; unless via, that type must
// itself be safe for the comparison. Members known from `@method` or
// `@property` tags only do not count.
func tucDeclared(ctx *analysis.Context, classes []string, name syntax.Expr, method, via bool) bool {
	id, _ := name.(*syntax.Identifier)
	ok := false
	for _, c := range classes {
		c = strings.TrimPrefix(c, `\`)
		decl := ""
		switch {
		case id == nil: // dynamic member name
		case method:
			if m := ctx.Index().FindMethod(c, id.Value, ctx.PHP); m != nil && !m.Magic {
				decl = m.Return
			}
		default:
			if p := ctx.Index().FindProperty(c, id.Value, ctx.PHP); p != nil && !p.Magic {
				decl = p.Type
			}
		}
		if decl == "" || !via && !tucSafeDecl(ctx, decl) {
			return false
		}
		ok = true
	}
	return ok
}

// tucVarKey identifies one stability question of a scope.
type tucVarKey struct {
	name string
	via  bool
}

// tucScope holds what one function-like scope does to its variables.
type tucScope struct {
	dynamic  bool            // variable-variables (none is ever stable), extract(), include, eval
	byRef    map[string]bool // variables a reference may have been taken of
	accesses map[string][]flowquery.VarAccess
	stable   map[tucVarKey]bool // tucStableVar answers; false while computing
}

// tucStableVar reports whether the value of v, anywhere in its function,
// is one its declarations account for: v is `$this`, or a local of a
// function, method or closure that is a natively typed by-value parameter
// or starts undefined, is only written by plain assignments (`=`, `??=`,
// `.=`) of proven values and is never exposed to a reference, a by-reference
// argument, `global`, `static`, variable-variables or included code. The
// check is flow-insensitive, so loop back-edges are covered. With via, v is
// the receiver of the compared member: it must then have a single definition.
func tucStableVar(ctx *analysis.Context, v *syntax.Variable, via bool) bool {
	if v.Name == "this" {
		return true
	}
	scope := syntax.EnclosingVariableScope(v)
	switch scope.(type) {
	case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
	default:
		return false // top-level code (globals) or a property hook
	}
	facts := tucScopeFacts(ctx, scope)
	key := tucVarKey{v.Name, via}
	if done, seen := facts.stable[key]; seen {
		return done // false while computing: a cycle of assignments
	}
	facts.stable[key] = false
	ok := tucComputeStable(ctx, scope, facts, v.Name, via)
	facts.stable[key] = ok
	return ok
}

func tucComputeStable(ctx *analysis.Context, scope syntax.Node, facts *tucScope, name string, via bool) bool {
	if facts.dynamic || facts.byRef[name] || syntax.Superglobal(name) {
		return false
	}
	defs := 0
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.Var == nil || p.Var.Name != name {
			continue
		}
		if p.Type == nil || p.ByRef {
			return false
		}
		defs++
	}
	if defs == 0 {
		// Not a parameter: closures import it, arrow functions capture it.
		switch s := scope.(type) {
		case *syntax.ArrowFunction:
			return false
		case *syntax.Closure:
			for _, u := range s.Uses {
				if u.Var != nil && u.Var.Name == name {
					return false
				}
			}
		}
	}
	for _, a := range facts.accesses[name] {
		if u, isUse := a.By.(*syntax.ClosureUse); isUse && u.ByRef {
			return false
		}
		if !a.Write {
			continue
		}
		as, isAssign := a.By.(*syntax.Assign)
		if !isAssign || as.ByRef || as.Var != syntax.Expr(a.Var) {
			return false // foreach, global, static, catch, unset, ++, destructuring
		}
		defs++
		switch as.Op.Kind {
		case syntax.TConcatEqual:
		case syntax.TEqual, syntax.TCoalesceEqual:
			if via && !tucBranch(ctx, as.Value, true) || !via && !tucStrictSafe(ctx, as.Value, "") {
				return false
			}
		default:
			return false
		}
	}
	return !via || defs <= 1
}

// tucScopeFacts summarises scope in one walk, cached per scope.
func tucScopeFacts(ctx *analysis.Context, scope syntax.Node) *tucScope {
	return ctx.Memo("scope:"+strconv.FormatUint(uint64(scope.Span().Start), 10), func() any {
		s := &tucScope{
			byRef:    map[string]bool{},
			accesses: flowquery.VarAccessesByName(ctx.File, scope),
			stable:   map[tucVarKey]bool{},
		}
		syntax.Inspect(syntax.VariableScopeBody(scope), func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.Function, *syntax.Closure:
				return false // a scope of its own; by-reference imports are accesses
			case *syntax.ClassLike:
				s.markArgs(x.Args, nil, false)
				return false
			case *syntax.Variable:
				s.dynamic = s.dynamic || x.NameExpr != nil
			case *syntax.Include, *syntax.Eval:
				s.dynamic = true
			case *syntax.Assign:
				if x.ByRef {
					s.mark(x.Value)
				}
			case *syntax.ArrayItem:
				if x.ByRef {
					s.mark(x.Value)
				}
			case *syntax.Foreach:
				if x.ByRef {
					s.mark(x.Expr)
				}
			case *syntax.FuncCall:
				switch ctx.GlobalFunctionName(x) {
				case "extract", "parse_str", "mb_parse_str":
					s.dynamic = true
				}
				callee, resolved := semanticquery.ResolveCallee(ctx, x)
				s.markArgs(x.Args, callee.Params, resolved)
			case *syntax.MethodCall:
				callee, resolved := semanticquery.ResolveCallee(ctx, x)
				s.markArgs(x.Args, callee.Params, resolved)
			case *syntax.StaticCall:
				callee, resolved := semanticquery.ResolveCallee(ctx, x)
				s.markArgs(x.Args, callee.Params, resolved)
			case *syntax.New:
				params, resolved := tucCtorParams(ctx, x)
				s.markArgs(x.Args, params, resolved)
			}
			return true
		})
		return s
	}).(*tucScope)
}

// tucCtorParams returns the constructor parameters of a `new` on a named
// class whose hierarchy resolves (none when it declares no constructor).
func tucCtorParams(ctx *analysis.Context, n *syntax.New) ([]index.Param, bool) {
	cls := semanticquery.StaticCallClass(ctx, n.Class)
	if cls == "" || ctx.Index().Class(cls, ctx.PHP) == nil || !tucHierarchyKnown(ctx, cls) {
		return nil, false
	}
	if m := ctx.Index().FindMethod(cls, "__construct", ctx.PHP); m != nil {
		return m.Params, true
	}
	return nil, true
}

// mark records that a reference to the variable e is rooted in may exist.
func (s *tucScope) mark(e syntax.Expr) {
	for {
		switch x := e.(type) {
		case *syntax.ArrayDimFetch:
			e = x.Var
			continue
		case *syntax.PropertyFetch:
			e = x.Var
			continue
		case *syntax.Variable:
			s.byRef[x.Name] = true
		}
		return
	}
}

// markArgs marks the arguments of a call that may bind by reference: all
// of them when the callee is unresolved.
func (s *tucScope) markArgs(list *syntax.ArgList, params []index.Param, resolved bool) {
	if list == nil {
		return
	}
	pos := 0
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			continue // first-class callable syntax
		}
		hazard := !resolved || arg.ByRef || arg.Unpack
		if arg.Name != nil {
			for _, p := range params {
				hazard = hazard || p.ByRef && strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value)
			}
		} else {
			// Arguments past the last parameter bind to it when variadic.
			if i := min(pos, len(params)-1); i >= 0 && params[i].ByRef {
				hazard = true
			}
			pos++
		}
		if hazard {
			s.mark(arg.Value)
		}
	}
}
