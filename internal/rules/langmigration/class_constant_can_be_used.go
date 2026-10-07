package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// classConstantCanBeUsed reports class names written as strings and
// get_called_class()/get_parent_class() calls, suggesting `X::class`.
//
// It is a file rule: fixes may add `use` imports, and later fixes in the
// same file must see imports added by earlier ones, so all findings of a file
// share one fix state.
type classConstantCanBeUsed struct{}

func init() { register(classConstantCanBeUsed{}) }

func (classConstantCanBeUsed) ID() string { return "ClassConstantCanBeUsed" }

func (classConstantCanBeUsed) Kinds() []syntax.NodeKind { return nil }

func (classConstantCanBeUsed) Semantic() {}

func (classConstantCanBeUsed) Check(*analysis.Context, syntax.Node) {}

// cccImport is one `use` import item (existing or added by a fix).
type cccImport struct {
	fqn   string // with leading `\`
	local string
}

// cccState is shared by the fixes of one file.
type cccState struct {
	ctx     *analysis.Context
	imports []cccImport // existing imports, document order
	marker  syntax.Stmt // last existing `use` statement
	firstNS *syntax.Namespace
	added   []cccImport

	// Insertion point of the first import added by a fix, and whether it
	// was inserted after the marker (true) or before it.
	anchor     uint32
	chainAfter bool
	chained    bool
}

func (classConstantCanBeUsed) CheckFile(ctx *analysis.Context) {
	if ctx.PHP < phpver.PHP55 { // E1
		return
	}
	st := &cccState{ctx: ctx}
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Use:
			st.marker = n
			for _, it := range n.Items {
				name := it.Name.Value
				if n.Prefix != nil {
					name = strings.TrimSuffix(n.Prefix.Value, `\`) + `\` + strings.TrimPrefix(name, `\`)
				}
				name = `\` + strings.TrimPrefix(name, `\`)
				local := util.LastNamePart(name)
				if it.Alias != nil {
					local = it.Alias.Value
				}
				st.imports = append(st.imports, cccImport{fqn: name, local: local})
			}
		case *syntax.Namespace:
			if st.firstNS == nil && n.Name != nil {
				st.firstNS = n
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
	if call.Args == nil || util.ArgCount(call) != 0 {
		return
	}
	var repl string
	cl := classScope(call)
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
	st.ctx.Report(span, "Use "+repl+"::class instead.", analysis.Fix{
		Title: "Use " + repl + "::class",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl + "::class"}}
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
		if call := util.ParentFuncCall(lit); call != nil {
			if ctx.GlobalFunctionName(call) == "class_alias" && call.Args != nil &&
				len(call.Args.Args) == 2 && call.Args.Args[1] == syntax.Expr(p) {
				return
			}
		}
	}
	candidate := raw // D5
	if nsConcat {
		if cl := syntax.EnclosingClass(lit); cl != nil {
			ns := ctx.Names().Namespace(cl.Span().Start)
			candidate = `\` + ns + `\` + strings.TrimPrefix(raw, `\`)
			if ns == "" {
				candidate = `\` + strings.TrimPrefix(raw, `\`)
			}
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
	if !nsConcat && strings.HasPrefix(raw, `\`) {
		// custos: `::class` never has a leading backslash, so the fix
		// would change the string's value ('\A\B' !== 'A\B').
		ctx.Report(span, "Use "+name+"::class instead of the class name string (::class has no leading backslash).")
		return
	}
	ctx.Report(span, "Use "+name+"::class instead of the class name string.", analysis.Fix{
		Title: "Use " + name + "::class",
		Edits: func() []analysis.TextEdit { return st.edits(span, name) },
	})
}

// classScope returns the class-like whose scope n runs in: the nearest
// enclosing class-like, unless a named function declaration (which never has
// a class scope) comes first.
func classScope(n syntax.Node) *syntax.ClassLike {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p := p.(type) {
		case *syntax.Function:
			return nil
		case *syntax.ClassLike:
			return p
		}
	}
	return nil
}

// edits implements F3 for the class reference q (see the spec).
func (st *cccState) edits(span syntax.Span, q string) []analysis.TextEdit {
	ctx := st.ctx
	short := util.LastNamePart(q)
	t := q
	imported, collision := false, false
	for _, imp := range append(append([]cccImport(nil), st.imports...), st.added...) { // step 1
		// Class names and import aliases are case-insensitive in PHP.
		if strings.EqualFold(imp.fqn, q) {
			t, imported = imp.local, true
		} else if strings.EqualFold(imp.local, short) {
			collision = true
		}
	}
	out := []analysis.TextEdit{}
	if ctx.Bool("IMPORT_CLASSES_ON_QF") && !imported && !collision && strings.Count(q, `\`) >= 2 { // step 3
		var marker syntax.Stmt
		after := false
		if st.marker != nil {
			marker, after = st.marker, true
		}
		add := true
		if ns := st.firstNS; ns != nil {
			nsName := strings.TrimPrefix(ns.Name.Value, `\`)
			if c := ctx.Index().Class(nsName+`\`+short, ctx.PHP); c != nil && c.Kind == syntax.KindClass {
				add = false
			} else {
				if marker == nil && len(st.added) == 0 && len(ns.Stmts) > 0 {
					marker, after = ns.Stmts[0], false
				}
				if ctx.Bool("USE_RELATIVE_QF") && strings.HasPrefix(q, `\`+nsName+`\`) {
					t = q[len(nsName)+2:]
					add = false
				}
			}
		} else if len(st.added) == 0 {
			marker, after = firstStmt(ctx.File.Stmts), false
			st.marker = nil
		}
		if add && (marker != nil || len(st.added) > 0) {
			if len(st.added) > 0 {
				// A previous fix already inserted an import: chain after it.
				if e, ok := st.chainImport(q); ok {
					out = append(out, e)
				}
			} else {
				if d, ok := marker.(*syntax.Declare); ok && !after {
					marker, _ = util.NextStmt(ctx.File, d)
				}
				if marker != nil {
					text := "use " + strings.TrimPrefix(q, `\`) + ";"
					if after {
						st.anchor = marker.Span().End
						out = append(out, analysis.TextEdit{Span: syntax.Span{Start: st.anchor, End: st.anchor}, NewText: "\n" + text})
						st.chainAfter = true
					} else {
						st.anchor = marker.Span().Start
						out = append(out, analysis.TextEdit{Span: syntax.Span{Start: st.anchor, End: st.anchor}, NewText: text + "\n\n"})
						st.chainAfter = false
					}
				}
			}
			if len(out) > 0 {
				st.added = append(st.added, cccImport{fqn: q, local: short})
				t = short
			}
		}
	}
	return append(out, analysis.TextEdit{Span: span, NewText: t + "::class"})
}

// chainImport adds a second import right after the one a previous fix
// inserted after the marker. Edits touching that insertion would conflict,
// so the import is inserted one byte further, after the character that
// follows the anchor.
func (st *cccState) chainImport(q string) (analysis.TextEdit, bool) {
	src := st.ctx.Src
	if !st.chainAfter || st.chained || int(st.anchor) >= len(src) {
		return analysis.TextEdit{}, false
	}
	st.chained = true
	text := "\nuse " + strings.TrimPrefix(q, `\`) + ";"
	if src[st.anchor] == '\n' {
		text = "use " + strings.TrimPrefix(q, `\`) + ";\n"
	}
	return analysis.TextEdit{Span: syntax.Span{Start: st.anchor + 1, End: st.anchor + 1}, NewText: text}, true
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
