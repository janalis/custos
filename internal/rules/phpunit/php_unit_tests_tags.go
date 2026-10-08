package phpunit

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

const (
	putProviderMsg = "The @dataProvider target cannot be resolved to a method."
	putNamedMsg    = "Give the provider's datasets string keys."
	putDependsMsg  = "The @depends target is missing or is not a test."
	putTestMsg     = "Remove '@test': the method name already marks it as a test."
)

// putDocToken returns the doc comment token attached to a declaration.
func putDocToken(f *syntax.File, n syntax.Node, name syntax.Node) (syntax.Token, bool) {
	i := util.TokenIndex(f, n.Span().Start)
	for j := i - 1; j >= 0; j-- {
		t := f.Tokens[j]
		switch t.Kind {
		case syntax.TWhitespace, syntax.TComment:
			continue
		case syntax.TDocComment:
			return t, true
		}
		break
	}
	// doc comment between attributes/modifiers and the name
	for k := i; k < len(f.Tokens) && f.Tokens[k].Start < name.Span().Start; k++ {
		if f.Tokens[k].Kind == syntax.TDocComment {
			return f.Tokens[k], true
		}
	}
	return syntax.Token{}, false
}

// putTag is one tag found in a doc comment.
type putTag struct {
	name       string // with '@'
	start, end uint32 // absolute span of the tag name
	value      string // first word after the tag name ("" when none)
	annotation bool   // starts a docblock line (D3)
}

func putIsTagChar(c byte) bool {
	return c == '_' || c == '-' || c == '\\' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// putTags lists the tags of a doc comment token.
func putTags(f *syntax.File, tok syntax.Token) []putTag {
	text := string(f.Src[tok.Start:tok.End])
	var out []putTag
	for i := 0; i < len(text); i++ {
		if text[i] != '@' || (i > 0 && putIsTagChar(text[i-1])) {
			continue
		}
		j := i + 1
		for j < len(text) && putIsTagChar(text[j]) {
			j++
		}
		if j == i+1 {
			continue
		}
		t := putTag{name: text[i:j], start: tok.Start + uint32(i), end: tok.Start + uint32(j)}
		// D3: annotation position
		k := i
		for k > 0 && (text[k-1] == ' ' || text[k-1] == '\t' || text[k-1] == '\r' || text[k-1] == '\n') {
			k--
		}
		if k == 3 && strings.HasPrefix(text, "/**") {
			t.annotation = true
		} else if k > 0 && text[k-1] == '*' {
			b := k - 1
			for b > 0 && (text[b-1] == ' ' || text[b-1] == '\t') {
				b--
			}
			t.annotation = b > 0 && text[b-1] == '\n'
		}
		// value: first word on the same line
		v := j
		for v < len(text) && (text[v] == ' ' || text[v] == '\t') {
			v++
		}
		w := v
		for w < len(text) && text[w] != ' ' && text[w] != '\t' && text[w] != '\n' && text[w] != '\r' &&
			!strings.HasPrefix(text[w:], "*/") {
			w++
		}
		t.value = text[v:w]
		out = append(out, t)
		i = j - 1
	}
	return out
}

// putRefLike reports whether a tag value is a reference-like word (D4).
func putRefLike(v string) bool {
	v = strings.TrimSuffix(v, "()")
	if v == "" {
		return false
	}
	cls, member, hasMember := strings.Cut(v, "::")
	if !hasMember {
		return putIsQualifiedName(cls)
	}
	if cls != "" && !putIsQualifiedName(cls) {
		return false
	}
	if member == "" {
		return cls != ""
	}
	if strings.HasPrefix(member, "<") {
		return cls != "" && strings.HasSuffix(member, ">")
	}
	return putIsIdent(member)
}

func putIsIdent(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80 || (i > 0 && c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func putIsQualifiedName(s string) bool {
	s = strings.TrimPrefix(s, `\`)
	if s == "" {
		return false
	}
	for _, p := range strings.Split(s, `\`) {
		if !putIsIdent(p) {
			return false
		}
	}
	return true
}

func putCheckTags(ctx *analysis.Context, m *syntax.Method) {
	cl := m.Parent().(*syntax.ClassLike) // methods only appear in class-like bodies
	if m.Name.Span().Len() == 0 {        // D1: recovered method without a name
		return
	}
	tok, ok := putDocToken(ctx.File, m, m.Name)
	if !ok {
		return
	}
	at := m.Span().Start
	classFQN := ctx.Types().ClassFQN(cl)
	isTestName := strings.HasPrefix(m.Name.Value, "test")
	for _, t := range putTags(ctx.File, tok) { // D2
		if !t.annotation { // D3
			continue
		}
		switch t.name {
		case "@test":
			if isTestName { // D11
				putReportTest(ctx, tok, t)
			}
			continue
		case "@dataProvider", "@depends", "@covers":
		default:
			continue
		}
		if !putRefLike(t.value) { // D4
			continue
		}
		switch t.name {
		case "@dataProvider":
			meth, decl := putResolveMember(ctx, t.value, classFQN, at)
			if meth == "" { // D5
				ctx.ReportSeverity(m.Name.Span(), meta.SeverityError, putProviderMsg)
				continue
			}
			if ctx.Bool("SUGGEST_TO_USE_NAMED_DATASETS") && decl != nil && putUnnamedDatasets(decl) { // D6
				ctx.ReportNode(m.Name, putNamedMsg)
			}
		case "@depends":
			meth, decl := putResolveMember(ctx, t.value, classFQN, at)
			if meth == "" { // D7
				ctx.ReportSeverity(m.Name.Span(), meta.SeverityError, putDependsMsg)
				continue
			}
			if !strings.HasPrefix(meth, "test") && decl != nil && !putHasTestTag(ctx, decl) { // D8
				ctx.ReportSeverity(m.Name.Span(), meta.SeverityError, putDependsMsg)
			}
		case "@covers":
			ref := t.value
			if strings.HasPrefix(ref, "::") { // D10a: ::m against @coversDefaultClass
				if def := putCoversDefaultClass(ctx, cl); def != "" && !strings.HasPrefix(ref, "::<") && putCoversResolves(ctx, def+ref, at) {
					continue
				}
			}
			if !putCoversResolves(ctx, ref, at) { // D9, D10
				ctx.ReportSeverity(m.Name.Span(), meta.SeverityError, "The @covers target '"+t.value+"' cannot be resolved.")
			}
		}
	}
}

// putResolveMember resolves a @dataProvider/@depends reference to a method:
// it returns the method's declared name ("" when unresolved) and its
// declaration when it lives in this file.
func putResolveMember(ctx *analysis.Context, ref, classFQN string, at uint32) (string, *syntax.Method) {
	ref = strings.TrimSuffix(ref, "()")
	cls, member, hasClass := strings.Cut(ref, "::")
	if !hasClass {
		member, cls = cls, ""
	}
	fqn := classFQN
	if hasClass {
		if cls == "" || member == "" || strings.HasPrefix(member, "<") {
			return "", nil
		}
		fqn = putClassRef(ctx, cls, at)
	}
	if fqn == "" {
		return "", nil
	}
	m := ctx.Index().FindMethod(fqn, member, ctx.PHP)
	if m == nil {
		return "", nil
	}
	return m.Name, putFindMethodDecl(ctx, m.Class, m.Name)
}

// putFindMethodDecl returns the declaration of class::method when the class
// is declared in this file.
func putFindMethodDecl(ctx *analysis.Context, class, method string) *syntax.Method {
	var found *syntax.Method
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		if found != nil {
			return false
		}
		cl, ok := n.(*syntax.ClassLike)
		if !ok {
			return true
		}
		if cl.Name != nil && strings.EqualFold(ctx.Types().ClassFQN(cl), strings.TrimPrefix(class, `\`)) {
			for _, mem := range cl.Members {
				if md, ok := mem.(*syntax.Method); ok && md.Name != nil && strings.EqualFold(md.Name.Value, method) {
					found = md
				}
			}
		}
		return found == nil
	})
	return found
}

// putHasTestTag reports whether a method's docblock carries @test anywhere.
func putHasTestTag(ctx *analysis.Context, m *syntax.Method) bool {
	tok, ok := putDocToken(ctx.File, m, m.Name)
	if !ok {
		return false
	}
	for _, t := range putTags(ctx.File, tok) {
		if t.name == "@test" {
			return true
		}
	}
	return false
}

// putUnnamedDatasets implements D6 on a provider declaration.
func putUnnamedDatasets(m *syntax.Method) bool {
	if m.Modifiers.Has(syntax.TAbstract) || m.Body == nil || len(m.Body.Stmts) == 0 {
		return false
	}
	ret, ok := m.Body.Stmts[len(m.Body.Stmts)-1].(*syntax.Return)
	if !ok || ret.Expr == nil {
		return false
	}
	arr, ok := ret.Expr.(*syntax.Array)
	if !ok || len(arr.Items) == 0 {
		return false
	}
	if first := arr.Items[0]; first != nil { // nil: an empty slot (invalid PHP)
		switch k := first.Key.(type) {
		case *syntax.Literal:
			return k.LitKind != syntax.LitString
		case *syntax.InterpolatedString:
			return false
		}
	}
	return true
}

// putCoversDefaultClass returns the value of the class docblock's
// @coversDefaultClass tag ("" when absent).
func putCoversDefaultClass(ctx *analysis.Context, cl *syntax.ClassLike) string {
	if cl.Name == nil {
		return ""
	}
	tok, ok := putDocToken(ctx.File, cl, cl.Name)
	if !ok {
		return ""
	}
	for _, t := range putTags(ctx.File, tok) {
		if t.annotation && t.name == "@coversDefaultClass" && putRefLike(t.value) {
			return t.value
		}
	}
	return ""
}

// putCoversResolves implements D9/D10.
func putCoversResolves(ctx *analysis.Context, ref string, at uint32) bool {
	needCallable := strings.Contains(ref, "::") && !strings.Contains(ref, "::<")
	endsColons := strings.HasSuffix(ref, "::")
	var classOK, callableOK bool
	cls, member, hasMember := strings.Cut(strings.TrimSuffix(ref, "()"), "::")
	switch {
	case hasMember && cls == "":
		callableOK = putFunctionExists(ctx, member, at)
	case hasMember:
		fqn := putClassRef(ctx, cls, at)
		if ctx.Index().Class(fqn, ctx.PHP) != nil {
			classOK = true
			if member != "" && !strings.HasPrefix(member, "<") && ctx.Index().FindMethod(fqn, member, ctx.PHP) != nil {
				callableOK = true
			}
			if endsColons {
				callableOK = true
			}
		}
	default:
		fqn := putClassRef(ctx, cls, at)
		if ctx.Index().Class(fqn, ctx.PHP) != nil {
			classOK = true
		} else if putFunctionExists(ctx, cls, at) {
			callableOK = true
		}
	}
	if needCallable {
		return callableOK
	}
	return classOK
}

// putClassRef resolves a class named in a tag. PHPUnit reads these names
// as fully qualified (`@covers Vendor\Pkg\Cls`, no leading backslash
// needed), so a name that exists as written wins over the namespace-
// relative reading (custos).
func putClassRef(ctx *analysis.Context, cls string, at uint32) string {
	if fq := strings.TrimPrefix(cls, `\`); strings.Contains(fq, `\`) && ctx.Index().Class(fq, ctx.PHP) != nil {
		return fq
	}
	return strings.TrimPrefix(ctx.Names().Class(cls, at), `\`)
}

// putFunctionExists resolves a function name (never empty: D4 rejects
// `::` and empty references).
func putFunctionExists(ctx *analysis.Context, name string, at uint32) bool {
	fqn, fb := ctx.Names().Function(name, at)
	return ctx.Index().ResolveFunction(fqn, fb, ctx.PHP) != nil
}

// putReportTest reports a redundant @test tag with its removal fix (F17).
func putReportTest(ctx *analysis.Context, tok syntax.Token, t putTag) {
	src := ctx.Src
	span := syntax.Span{Start: t.start, End: t.end}
	ctx.Report(span, putTestMsg, analysis.Fix{
		Title: "Remove @test",
		Edits: func() []analysis.TextEdit {
			end := t.end
			// value text on the same line, up to the newline or the closer
			e := end
			for e < tok.End && src[e] != '\n' && src[e] != '\r' && !(src[e] == '*' && e+1 < tok.End && src[e+1] == '/') {
				e++
			}
			for e > end && (src[e-1] == ' ' || src[e-1] == '\t') {
				e--
			}
			end = e
			start := t.start
			for start > tok.Start && (src[start-1] == ' ' || src[start-1] == '\t') {
				start--
			}
			if start > tok.Start+3 && src[start-1] == '*' {
				start--
			}
			return []analysis.TextEdit{{Span: syntax.Span{Start: start, End: end}}}
		},
	})
}
