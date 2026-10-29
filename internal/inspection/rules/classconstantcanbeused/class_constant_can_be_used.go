package classconstantcanbeused

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// classConstantCanBeUsed reports class names written as strings and
// get_called_class()/get_parent_class() calls, suggesting `X::class`.
//
// It is a file rule: fixes may add `use` imports, and later fixes in the
// same file must see imports added by earlier ones, so all findings of a file
// share one fix state.
type classConstantCanBeUsed struct{}

func (classConstantCanBeUsed) ID() string                           { return "ClassConstantCanBeUsed" }
func (classConstantCanBeUsed) Kinds() []syntax.NodeKind             { return nil }
func (classConstantCanBeUsed) Semantic()                            {}
func (classConstantCanBeUsed) Check(*analysis.Context, syntax.Node) {}

// cccImport is one class `use` import item (existing or added by a fix).
type cccImport struct {
	fqn   string // with leading `\`
	local string
}

// cccBlock is one namespace block of the file: imports are only visible
// inside the block that declares them, so each block has its own fix state.
type cccBlock struct {
	name    string // namespace name without leading `\`; "" for global code
	global  bool   // file without any namespace declaration
	stmts   []syntax.Stmt
	imports []cccImport // class imports, document order, then added ones
	marker  syntax.Stmt // last existing `use` statement
	// Lower-cased class-likes declared in the block and first segments of
	// the names it references (including names written by earlier fixes):
	// a new import of such a name would redeclare or capture them.
	taken map[string]bool

	// Insertion point of the first import added by a fix, and whether it
	// was inserted after the marker (true) or before it.
	added      bool
	anchor     uint32
	chainAfter bool
	chained    bool
}

// cccState is shared by the fixes of one file.
type cccState struct {
	ctx *analysis.Context
	cur *cccBlock // block enclosing the node being visited
}

// firstSegment returns the lower-cased first segment of a written name.
func firstSegment(name string) string {
	if i := strings.IndexByte(name, '\\'); i >= 0 {
		name = name[:i]
	}
	return strings.ToLower(name)
}

// captured reports whether a class import of the block gives the first
// segment of the relative name rel another meaning.
func (b *cccBlock) captured(rel string) bool {
	seg := firstSegment(rel)
	for _, imp := range b.imports {
		if strings.ToLower(imp.local) == seg {
			return true
		}
	}
	return false
}

func (classConstantCanBeUsed) CheckFile(ctx *analysis.Context) {
	if ctx.PHP < phpversion.PHP55 { // E1
		return
	}
	st := &cccState{ctx: ctx, cur: &cccBlock{global: true, stmts: ctx.File.Stmts, taken: map[string]bool{}}}
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Namespace:
			st.cur = &cccBlock{stmts: n.Stmts, taken: map[string]bool{}}
			if n.Name != nil {
				st.cur.name = strings.TrimPrefix(n.Name.Value, `\`)
			}
		case *syntax.Use:
			b := st.cur
			b.marker = n
			for _, it := range n.Items {
				// Function and constant imports live in other symbol tables.
				if n.Type != syntax.UseNormal || it.Type != syntax.UseNormal {
					continue
				}
				name := it.Name.Value
				if n.Prefix != nil {
					name = strings.TrimSuffix(n.Prefix.Value, `\`) + `\` + strings.TrimPrefix(name, `\`)
				}
				name = `\` + strings.TrimPrefix(name, `\`)
				local := astquery.LastNamePart(name)
				if it.Alias != nil {
					local = it.Alias.Value
				}
				b.imports = append(b.imports, cccImport{fqn: name, local: local})
			}
			return false
		case *syntax.ClassLike:
			if n.Name != nil {
				st.cur.taken[strings.ToLower(n.Name.Value)] = true
			}
		case *syntax.Name:
			if n.NameKind == syntax.NameUnqualified || n.NameKind == syntax.NameQualified {
				st.cur.taken[firstSegment(n.Value)] = true
			}
		case *syntax.FuncCall:
			st.checkCall(n)
		case *syntax.Literal:
			st.checkLiteral(n)
		}
		return true
	})
}

func (st *cccState) checkCall(call *syntax.FuncCall) {
	if call.Args == nil || astquery.ArgCount(call) != 0 {
		return
	}
	var repl string
	cl := astquery.ClassScope(call)
	switch st.ctx.GlobalFunctionName(call) { // any case, global functions only
	case "get_called_class": // D1: static:: needs a class scope
		if cl == nil {
			return
		}
		repl = "static"
	case "get_parent_class": // D2: parent:: needs a class with a parent
		if cl == nil || cl.ClassKind != syntax.KindClass || len(cl.Extends) == 0 {
			return
		}
		repl = "parent"
	default:
		return
	}
	span := call.Span()
	st.ctx.Report(span, "Use "+repl+"::class instead.", diagnostic.Fix{
		Title: "Use " + repl + "::class",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl + "::class"}}
		},
	})
}

// isClassNameChar mirrors the accepted character set (D6).
func isClassNameChar(c byte) bool {
	return c >= 'A' && c <= 'z' || c >= '0' && c <= '9'
}

func (st *cccState) checkLiteral(lit *syntax.Literal) {
	ctx := st.ctx
	if lit.LitKind != syntax.LitString || len(lit.Raw) < 2 || (lit.Raw[0] != '\'' && lit.Raw[0] != '"') { // D3
		return
	}
	raw := lit.Raw[1 : len(lit.Raw)-1]
	if lit.Raw[0] == '"' && hasEscape(raw) {
		// custos: the runtime string is not the written text ("app\events"
		// with a single backslash holds an ESC byte).
		return
	}
	b := st.cur
	nsConcat := false
	report := syntax.Node(lit)
	switch p := lit.Parent().(type) { // D4
	case *syntax.Binary:
		mc, ok := p.Left.(*syntax.MagicConst)
		if p.Op.Kind != syntax.TDot || !ok || mc.Token.Kind != syntax.TNsC || p.Right != syntax.Expr(lit) {
			return
		}
		nsConcat = true
		report = p
	case *syntax.Assign:
		if p.Op.Kind != syntax.TEqual {
			return
		}
	case *syntax.Arg:
		if call := astquery.ParentFuncCall(lit); call != nil {
			if ctx.GlobalFunctionName(call) == "class_alias" && call.Args != nil &&
				len(call.Args.Args) == 2 && call.Args.Args[1] == syntax.Expr(p) {
				return
			}
		}
	}
	candidate := raw // D5
	if nsConcat {
		// custos: without the separating backslash the runtime value is
		// `NsName`, not a class of the namespace.
		if !strings.HasPrefix(raw, `\`) {
			return
		}
		if b.name != "" {
			candidate = `\` + b.name + raw
		}
	}
	if len(candidate) <= 3 { // D6
		return
	}
	for i := 0; i < len(candidate); i++ {
		if !isClassNameChar(candidate[i]) {
			return
		}
	}
	if !strings.Contains(candidate, `\`) && strings.ToLower(candidate) == candidate { // D7
		return
	}
	norm := strings.ReplaceAll(candidate, `\\`, `\`) // D8
	var fqn string
	switch {
	case strings.HasPrefix(norm, `\`):
		fqn = norm
	case strings.Contains(norm, `\`) || ctx.Bool("LOOK_ROOT_NS_UP"):
		fqn = `\` + norm
	default:
		return
	}
	var found []*index.Class // D9
	for _, c := range ctx.Index().ClassDecls(fqn, ctx.PHP) {
		if c.Kind != syntax.KindTrait {
			found = append(found, c)
		}
	}
	if len(found) != 1 || `\`+strings.TrimPrefix(found[0].FQN, `\`) != fqn {
		return
	}
	name := fqn
	if nsConcat {
		name = strings.TrimPrefix(strings.ReplaceAll(raw, `\\`, `\`), `\`)
	}
	span := report.Span()
	if strings.HasPrefix(raw, `\`) && (!nsConcat || b.name == "") {
		// custos: `::class` never has a leading backslash, so the fix
		// would change the string's value ('\A\B' !== 'A\B'). The same holds
		// for `__NAMESPACE__ . '\A'` when the namespace is empty.
		ctx.Report(span, "Use "+name+"::class instead of the class name string (::class has no leading backslash).")
		return
	}
	edits := func() []diagnostic.TextEdit { return st.edits(b, span, name) }
	if nsConcat {
		edits = func() []diagnostic.TextEdit { return b.relativeEdits(span, name) }
	}
	ctx.Report(span, "Use "+name+"::class instead of the class name string.", diagnostic.Fix{
		Title: "Use " + name + "::class",
		Edits: edits,
	})
}

// hasEscape reports whether the contents of a double-quoted literal hold an
// escape sequence that PHP decodes to something else than the written
// characters. `\\` is a plain backslash and is skipped as a pair.
func hasEscape(raw string) bool {
	for i := 0; i+1 < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		i++
		switch c := raw[i]; {
		case strings.IndexByte("ntrvef01234567", c) >= 0:
			return true
		case c == 'x' && i+1 < len(raw) && isHex(raw[i+1]):
			return true
		}
	}
	return false
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// relativeEdits rewrites `__NAMESPACE__ . '\rel'`. The name is never
// imported; when an import of the block would capture its first segment the
// explicit `namespace\` form is written.
func (b *cccBlock) relativeEdits(span syntax.Span, rel string) []diagnostic.TextEdit {
	if b.captured(rel) {
		rel = `namespace\` + rel
	} else {
		b.taken[firstSegment(rel)] = true
	}
	return []diagnostic.TextEdit{{Span: span, NewText: rel + "::class"}}
}

// edits implements F3 for the fully qualified class reference q of a literal
// in block b (see the spec).
func (st *cccState) edits(b *cccBlock, span syntax.Span, q string) []diagnostic.TextEdit {
	ctx := st.ctx
	short := astquery.LastNamePart(q)
	t := q
	canImport := ctx.Bool("IMPORT_CLASSES_ON_QF") && strings.Count(q, `\`) >= 2
	for _, imp := range b.imports { // step 1
		switch {
		case imp.fqn == q:
			t, canImport = imp.local, false
		case strings.EqualFold(imp.fqn, q), strings.EqualFold(imp.local, short):
			// `X::class` returns the spelling of the import, and import
			// aliases are case-insensitive: neither reuse nor add.
			canImport = false
		}
	}
	out := []diagnostic.TextEdit{}
	inNS := short
	if b.name != "" {
		inNS = b.name + `\` + short
	}
	if canImport && ctx.Index().Class(inNS, ctx.PHP) == nil { // step 3
		if ctx.Bool("USE_RELATIVE_QF") && b.name != "" && strings.HasPrefix(q, `\`+b.name+`\`) {
			if rel := q[len(b.name)+2:]; !b.captured(rel) {
				t = rel
				b.taken[firstSegment(rel)] = true
			}
		} else if !b.taken[strings.ToLower(short)] {
			if e, ok, kept := st.importEdit(b, q); ok {
				out = append(out, e)
				if kept {
					b.imports = append(b.imports, cccImport{fqn: q, local: short})
				}
				t = short
			}
		}
	}
	return append(out, diagnostic.TextEdit{Span: span, NewText: t + "::class"})
}

// importEdit inserts `use q;` into block b: after its last `use` statement,
// else before its first statement. A file without namespace declaration
// always gets it before its first statement. kept is false when the edit
// competes with an import inserted by a previous fix: applied alone it is
// correct, applied with the others it waits for the next fix pass, so later
// literals must not count on it.
func (st *cccState) importEdit(b *cccBlock, q string) (e diagnostic.TextEdit, ok, kept bool) {
	if b.added {
		// A previous fix already inserted an import: chain after it.
		if e, ok := st.chainImport(b, q); ok {
			return e, true, true
		}
		text := "use " + strings.TrimPrefix(q, `\`) + ";"
		if b.chainAfter {
			text = "\n" + text
		} else {
			text += "\n\n"
		}
		return diagnostic.TextEdit{Span: syntax.Span{Start: b.anchor, End: b.anchor}, NewText: text}, true, false
	}
	marker, after := b.marker, true
	if marker == nil || b.global {
		marker, after = firstStmt(b.stmts), false
		if d, ok := marker.(*syntax.Declare); ok {
			marker, _ = astquery.NextStmt(st.ctx.File, d)
		}
	}
	if marker == nil {
		return diagnostic.TextEdit{}, false, false
	}
	b.added = true
	text := "use " + strings.TrimPrefix(q, `\`) + ";"
	if after {
		b.anchor, b.chainAfter = marker.Span().End, true
		return diagnostic.TextEdit{Span: syntax.Span{Start: b.anchor, End: b.anchor}, NewText: "\n" + text}, true, true
	}
	b.anchor = marker.Span().Start
	return diagnostic.TextEdit{Span: syntax.Span{Start: b.anchor, End: b.anchor}, NewText: text + "\n\n"}, true, true
}

// chainImport adds a second import right after the one a previous fix
// inserted after the marker. Edits touching that insertion would conflict,
// so the import is inserted one byte further, after the character that
// follows the anchor.
func (st *cccState) chainImport(b *cccBlock, q string) (diagnostic.TextEdit, bool) {
	src := st.ctx.Src
	if !b.chainAfter || b.chained || int(b.anchor) >= len(src) {
		return diagnostic.TextEdit{}, false
	}
	b.chained = true
	text := "\nuse " + strings.TrimPrefix(q, `\`) + ";"
	if src[b.anchor] == '\n' {
		text = "use " + strings.TrimPrefix(q, `\`) + ";\n"
	}
	return diagnostic.TextEdit{Span: syntax.Span{Start: b.anchor + 1, End: b.anchor + 1}, NewText: text}, true
}

// firstStmt returns the first statement that is not inline HTML. The file
// holds the reported literal, so stmts is never empty and never only HTML;
// the last statement is returned as a fallback.
func firstStmt(stmts []syntax.Stmt) syntax.Stmt {
	i := 0
	for i < len(stmts)-1 {
		if _, ok := stmts[i].(*syntax.InlineHTML); !ok {
			break
		}
		i++
	}
	return stmts[i]
}
