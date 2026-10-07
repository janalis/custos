package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/index"
	"custos/internal/phpdoc"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

// returnTypeCanBeDeclared suggests a native return type for methods whose
// returned values and @return documentation agree on one declarable type.
type returnTypeCanBeDeclared struct{}

func init() { register(returnTypeCanBeDeclared{}) }

func (returnTypeCanBeDeclared) ID() string { return "ReturnTypeCanBeDeclared" }

func (returnTypeCanBeDeclared) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

func (returnTypeCanBeDeclared) Semantic() {}

// rtdMagic holds the magic method names, lower-case (PHP method names are
// case-insensitive).
var rtdMagic = map[string]bool{
	"__construct": true, "__destruct": true, "__call": true, "__callstatic": true, "__get": true,
	"__set": true, "__isset": true, "__unset": true, "__sleep": true, "__wakeup": true,
	"__tostring": true, "__invoke": true, "__set_state": true, "__clone": true, "__debuginfo": true,
}

var rtdBuiltin = map[string]bool{
	"array": true, "iterable": true, "string": true, "bool": true, "int": true, "float": true,
	"number": true, "null": true, "void": true, "mixed": true, "callable": true, "resource": true,
	"static": true, "self": true, "object": true,
}

var rtdScalarOK = map[string]bool{
	"self": true, "array": true, "callable": true, "bool": true, "float": true, "int": true, "string": true,
}

// rtdNormalize implements D7.
func rtdNormalize(a string) string {
	low := strings.ToLower(a)
	switch {
	case strings.Contains(a, "[]"):
		return "array"
	case low == "boolean" || low == "true" || low == "false":
		return "bool"
	case low == "integer":
		return "int"
	case low == `\closure` || low == "closure":
		return "callable"
	case low == "$this":
		return "static"
	}
	if rtdBuiltin[strings.TrimPrefix(low, `\`)] {
		return strings.TrimPrefix(low, `\`)
	}
	return `\` + strings.TrimPrefix(a, `\`)
}

// rtdWalkOwn visits the nodes of a method body that belong to the method
// itself (nested functions, closures and classes are skipped).
func rtdWalkOwn(body syntax.Node, fn func(syntax.Node)) {
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		}
		fn(n)
		return true
	})
}

// rtdFirstReturn returns the first return statement in document order at any
// depth (nested closures included).
func rtdFirstReturn(body syntax.Node) *syntax.Return {
	var first *syntax.Return
	syntax.Inspect(body, func(n syntax.Node) bool {
		if first != nil {
			return false
		}
		if r, ok := n.(*syntax.Return); ok {
			first = r
			return false
		}
		return true
	})
	return first
}

func rtdIsNullLit(e syntax.Expr) bool {
	c, ok := e.(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "null")
}

// rtdExprType is the inferred type of a returned expression, with `$this`
// and `new static` typed as static.
func rtdExprType(ctx *analysis.Context, e syntax.Expr) types.Type {
	switch x := e.(type) {
	case *syntax.Paren:
		return rtdExprType(ctx, x.Expr)
	case *syntax.Variable:
		if x.Name == "this" {
			return types.Of("static")
		}
	case *syntax.New:
		if n, ok := x.Class.(*syntax.Name); ok && strings.EqualFold(n.Value, "static") {
			return types.Of("static")
		}
	}
	return ctx.TypeOf(e)
}

func (r returnTypeCanBeDeclared) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP70 { // E1
		return
	}
	m := n.(*syntax.Method)
	class, ok := m.Parent().(*syntax.ClassLike)
	if !ok || m.Name == nil || m.ReturnType != nil || rtdMagic[strings.ToLower(m.Name.Value)] { // D1-D3
		return
	}
	abstract := m.Body == nil
	var doc *phpdoc.Doc
	if c := index.DocComment(ctx.File, m); c != "" {
		doc = phpdoc.Parse(c)
	}
	hasReturnTag := false
	if doc != nil {
		_, hasReturnTag = doc.Tag("return")
	}
	if abstract && !hasReturnTag { // D4
		return
	}

	// D5: R0.
	known := map[string]bool{}
	unknown := false
	add := func(t types.Type) {
		if t.IsUnknown() {
			unknown = true
			return
		}
		for _, a := range t.Atoms() {
			known[a] = true
		}
	}
	at := m.Span().Start
	resolve := func(w string) string { return ctx.Names().Class(w, at) }
	docRet := ""
	if hasReturnTag {
		docRet = doc.ReturnType()
		add(types.FromDoc(docRet, resolve))
	}
	hasYield := false
	bareReturn, valueReturn := false, false // own `return;` / `return expr;`
	if !abstract {
		rtdWalkOwn(m.Body, func(x syntax.Node) {
			switch x := x.(type) {
			case *syntax.Return:
				if x.Expr == nil {
					bareReturn = true
					add(types.Void)
				} else {
					valueReturn = true
					t := rtdExprType(ctx, x.Expr)
					if t.IsUnknown() {
						t = rtdInheritedParamType(ctx, class, m, x.Expr)
					}
					add(t)
				}
			case *syntax.Yield, *syntax.YieldFrom:
				hasYield = true
			}
		})
	}
	if unknown && len(known) != 1 { // D6
		return
	}
	set := map[string]bool{} // D7
	for a := range known {
		set[rtdNormalize(a)] = true
	}
	if hasYield && !set[`\Generator`] { // D8
		set[`\Generator`] = true
		if rtdFirstReturn(m.Body) == nil {
			delete(set, "null")
		}
	}
	if len(set) > 0 && !abstract { // D9
		if !set["null"] && !set["void"] {
			last := syntax.Stmt(nil)
			if len(m.Body.Stmts) > 0 {
				last = m.Body.Stmts[len(m.Body.Stmts)-1]
			}
			isExit := false
			switch l := last.(type) {
			case *syntax.Return:
				isExit = true
			case *syntax.ExprStmt:
				_, isExit = l.Expr.(*syntax.Throw)
			}
			if !isExit {
				set["null"] = true
			}
		}
		if len(set) == 1 && set["null"] {
			if fr := rtdFirstReturn(m.Body); fr != nil && fr.Expr != nil && !rtdIsNullLit(fr.Expr) {
				delete(set, "null")
			}
		}
	}

	l71 := ctx.PHP >= phpver.PHP71
	suggestion := ""
	switch len(set) {
	case 0: // D10
		if !l71 {
			return
		}
		if !abstract {
			fr := rtdFirstReturn(m.Body)
			if fr != nil && rtdEnclosingFunc(fr) != syntax.Node(m) {
				fr = nil
			}
			if fr != nil {
				return
			}
		}
		suggestion = "void"
	case 1: // D11
		var t string
		for a := range set {
			t = a
		}
		s := ""
		if l71 && (t == "null" || t == "void") {
			s = "void"
		} else {
			s = r.compact(ctx, class, doc, docRet, t)
		}
		switch {
		case strings.HasPrefix(t, `\`) || rtdScalarOK[t] || s == "self" || s == "static":
		case l71 && s == "void":
		default:
			return
		}
		if hasReturnTag {
			if tag, _ := doc.Tag("return"); strings.TrimSpace(tag.Text) == "static" { // static guard
				if ctx.PHP < phpver.PHP80 {
					return
				}
				s = "static"
			}
		}
		if s == "static" && ctx.PHP < phpver.PHP80 {
			return // see Divergences: never suggest static below 8.0
		}
		suggestion = s
	case 2: // D12
		if !l71 {
			return
		}
		if set["void"] {
			delete(set, "void")
		} else if set["null"] {
			delete(set, "null")
		}
		if len(set) != 1 {
			return
		}
		var t string
		for a := range set {
			t = a
		}
		s := ""
		if t == "null" || t == "void" {
			s = "void"
		} else {
			s = r.compact(ctx, class, doc, docRet, t)
		}
		switch {
		case strings.HasPrefix(t, `\`) || rtdScalarOK[t] || s == "self":
			suggestion = "?" + s
		case s == "void":
			suggestion = "void"
		default:
			return
		}
	default:
		return
	}

	// D14: the declaration must accept the method's own return statements:
	// `return null;` is a compile error under `: void`, and a bare `return;`
	// under any other type (generators excepted).
	if !hasYield && (suggestion == "void" && valueReturn || suggestion != "void" && bareReturn) {
		return
	}
	span := m.Name.Span()
	if rtdOverridden(ctx, class, m) { // D13
		ctx.Report(span, "Declare ': "+suggestion+"' as the return type (update the whole hierarchy with a signature refactoring).")
		return
	}
	pos, ok := rtdParamsEnd(ctx, m)
	if !ok {
		ctx.Report(span, "Declare ': "+suggestion+"' as the return type.")
		return
	}
	text := ": " + suggestion
	if int(pos) < len(ctx.Src) && ctx.Src[pos] == '{' {
		text += " "
	}
	ctx.Report(span, "Declare ': "+suggestion+"' as the return type.", analysis.Fix{
		Title: "Declare the return type",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: syntax.Span{Start: pos, End: pos}, NewText: text}}
		},
	})
}

func rtdEnclosingFunc(n syntax.Node) syntax.Node {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			return p
		}
	}
	return nil
}

// compact implements compact(t).
func (returnTypeCanBeDeclared) compact(ctx *analysis.Context, class *syntax.ClassLike, doc *phpdoc.Doc, docRet string, t string) string {
	if !strings.HasPrefix(t, `\`) && t != "static" {
		return t
	}
	if docRet != "" && ctx.Bool("LOOKUP_PHPDOC_RETURN_DECLARATIONS") { // 1
		for _, p := range strings.Split(docRet, "|") {
			if p = strings.TrimSpace(p); p == "self" || p == "$this" {
				return "self"
			}
		}
	}
	if t == "static" { // 2
		return "static"
	}
	// 3: imports in the enclosing statement lists.
	for p := syntax.Node(class); ; {
		parent := p.Parent()
		var list []syntax.Stmt
		switch x := parent.(type) {
		case *syntax.Namespace:
			list = x.Stmts
		case *syntax.Block:
			list = x.Stmts
		case nil:
			list = ctx.File.Stmts
		}
		for _, s := range list {
			u, ok := s.(*syntax.Use)
			if !ok {
				continue
			}
			for _, it := range u.Items {
				kind := u.Type
				if u.Prefix != nil {
					kind = it.Type
				}
				if kind != syntax.UseNormal {
					continue
				}
				name := strings.TrimPrefix(it.Name.Value, `\`)
				if u.Prefix != nil {
					name = strings.TrimPrefix(strings.TrimSuffix(u.Prefix.Value, `\`), `\`) + `\` + name
				}
				if `\`+name == t {
					if it.Alias != nil {
						return it.Alias.Value
					}
					return name[strings.LastIndexByte(name, '\\')+1:]
				}
			}
		}
		if parent == nil {
			break
		}
		p = parent
	}
	// 4: same namespace prefix.
	if ns := ctx.Names().Namespace(class.Span().Start); ns != "" && strings.HasPrefix(t, `\`+ns+`\`) {
		return t[len(ns)+2:]
	}
	return t
}

// rtdOverridden implements D13.
func rtdOverridden(ctx *analysis.Context, class *syntax.ClassLike, m *syntax.Method) bool {
	if class.Modifiers.Has(syntax.TFinal) || m.Modifiers.Has(syntax.TFinal) || m.Modifiers.Has(syntax.TPrivate) {
		return false
	}
	fqn := ctx.Types().ClassFQN(class)
	if fqn == "" {
		return false
	}
	ix := ctx.Index()
	name := strings.ToLower(m.Name.Value)
	self := strings.ToLower(strings.TrimPrefix(fqn, `\`))
	for _, c := range ix.Ancestors(fqn, ctx.PHP) {
		if strings.ToLower(strings.TrimPrefix(c.FQN, `\`)) == self {
			continue
		}
		if _, ok := c.Methods[name]; ok {
			return true
		}
	}
	seen := map[string]bool{self: true}
	queue := []string{fqn}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, ch := range ix.ChildrenAll(cur) {
			k := strings.ToLower(strings.TrimPrefix(ch, `\`))
			if seen[k] {
				continue
			}
			seen[k] = true
			if c := ix.Class(ch, ctx.PHP); c != nil {
				if _, ok := c.Methods[name]; ok {
					return true
				}
			}
			queue = append(queue, ch)
		}
	}
	return false
}

// rtdParamsEnd returns the offset right after the `)` closing the parameter
// list of m.
func rtdParamsEnd(ctx *analysis.Context, m *syntax.Method) (uint32, bool) {
	end := m.Span().End
	if m.Body != nil {
		end = m.Body.Span().Start
	}
	toks := ctx.File.Tokens
	lo, hi := 0, len(toks)
	for lo < hi {
		mid := (lo + hi) / 2
		if toks[mid].Start < end {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	for i := lo - 1; i >= 0 && toks[i].Start >= m.Span().Start; i-- {
		if toks[i].Kind == syntax.TRParen {
			return toks[i].End, true
		}
	}
	return 0, false
}

// rtdInheritedParamType types a returned, untyped and undocumented parameter
// of m from the same parameter of the method it overrides (declared type,
// else doc type), as PhpStorm inherits parameter types.
func rtdInheritedParamType(ctx *analysis.Context, class *syntax.ClassLike, m *syntax.Method, e syntax.Expr) types.Type {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return types.Unknown
	}
	idx := -1
	for i, p := range m.Params {
		if p.Var != nil && p.Var.Name == v.Name && p.Type == nil {
			idx = i
		}
	}
	if idx < 0 {
		return types.Unknown
	}
	// Only when the variable is never reassigned in the method.
	reassigned := false
	rtdWalkOwn(m.Body, func(n syntax.Node) {
		if a, ok := n.(*syntax.Assign); ok {
			if t, ok := a.Var.(*syntax.Variable); ok && t.Name == v.Name {
				reassigned = true
			}
		}
	})
	if reassigned {
		return types.Unknown
	}
	fqn := ctx.Types().ClassFQN(class)
	if fqn == "" {
		return types.Unknown
	}
	self := strings.ToLower(strings.TrimPrefix(fqn, `\`))
	name := strings.ToLower(m.Name.Value)
	for _, c := range ctx.Index().Ancestors(fqn, ctx.PHP) {
		if strings.ToLower(strings.TrimPrefix(c.FQN, `\`)) == self {
			continue
		}
		pm, ok := c.Methods[name]
		if !ok || idx >= len(pm.Params) {
			continue
		}
		p := pm.Params[idx]
		if p.Type != "" {
			return types.FromDoc(p.Type, nil)
		}
		if p.DocType != "" {
			return types.FromDoc(p.DocType, nil)
		}
		return types.Unknown
	}
	return types.Unknown
}
