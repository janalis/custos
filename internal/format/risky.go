package format

import (
	"regexp"
	"slices"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// risky applies the risky settings that are on.
func (e *engine) risky() {
	s := e.s
	if s.Yoda != "keep" || s.IncrementStyle != "keep" || s.SingleQuote || s.HeredocToNowdoc || s.OrderedTraits {
		for _, st := range e.f.Stmts {
			e.riskyNode(st)
		}
	}
	if s.RemoveUnusedImports {
		e.unusedImports()
	}
	if s.DeclareStrictTypes {
		e.strictTypes()
	}
}

func (e *engine) riskyNode(n syntax.Node) {
	if isNil(n) || e.inString(n.Span().Start) {
		return
	}
	switch v := n.(type) {
	case *syntax.Binary:
		if e.s.Yoda != "keep" {
			e.yoda(v)
		}
	case *syntax.ExprStmt:
		if inc, ok := v.Expr.(*syntax.IncDec); ok {
			e.incrementStyle(inc)
		}
	case *syntax.For:
		for _, l := range v.Loop {
			if inc, ok := l.(*syntax.IncDec); ok {
				e.incrementStyle(inc)
			}
		}
	case *syntax.TraitUse:
		if e.s.OrderedTraits {
			e.orderTraits(v)
		}
	case *syntax.Literal:
		switch {
		case e.s.SingleQuote && v.LitKind == syntax.LitString && strings.HasPrefix(v.Raw, `"`):
			e.singleQuote(v)
		case e.s.HeredocToNowdoc && strings.HasPrefix(v.Raw, "<<<"):
			e.nowdoc(v)
		}
	}
	syntax.Children(n, e.riskyNode)
}

// constantOperand reports whether x is a literal or a constant.
func constantOperand(x syntax.Expr) bool {
	switch v := x.(type) {
	case *syntax.Literal, *syntax.ConstFetch, *syntax.MagicConst:
		return true
	case *syntax.ClassConstFetch:
		_, named := v.Class.(*syntax.Name)
		return named
	case *syntax.Unary:
		_, lit := v.Expr.(*syntax.Literal)
		return lit && (v.Op.Kind == syntax.TMinus || v.Op.Kind == syntax.TPlus)
	case *syntax.Array:
		return len(v.Items) == 0
	}
	return false
}

// yoda puts the constant operand of an equality comparison first.
func (e *engine) yoda(v *syntax.Binary) {
	defer e.newGroup()()
	switch v.Op.Kind {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
	default:
		return
	}
	moved := v.Left // the operand that changes sides
	if e.s.Yoda == "yoda" {
		if !constantOperand(v.Right) || constantOperand(v.Left) {
			return
		}
	} else {
		if !constantOperand(v.Left) || constantOperand(v.Right) {
			return
		}
		moved = v.Right
	}
	switch moved.(type) {
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch,
		*syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.Paren:
	default:
		return // only a plain term moves without changing how the comparison binds
	}
	l, r := v.Left.Span(), v.Right.Span()
	if e.multiline(l) || e.multiline(r) || e.commentBetween(l.Start, r.End) {
		return
	}
	a, b := e.unitAt(l.Start), e.unitBefore(l.End)
	c, d := e.unitAt(r.Start), e.unitBefore(r.End)
	if e.touched(a, d) {
		// Another rewrite holds an operand: try again on its output, unless
		// an ignore comment protects it.
		e.merged = e.merged || !e.frozenUnit(a) && !e.frozenUnit(d)
		return
	}
	left, right := e.unitsText(a, b), e.unitsText(c, d)
	e.replaceUnits(a, b, right, "yoda")
	e.replaceUnits(c, d, left, "yoda")
}

// unitsText renders units a to b with the whitespace written between them.
func (e *engine) unitsText(a, b int) string {
	var sb strings.Builder
	for k := a; k <= b; k++ {
		if k > a {
			sb.Write(e.src[e.gaps[k].start:e.gaps[k].end])
		}
		sb.WriteString(e.text[k])
	}
	return sb.String()
}

// multiline reports whether sp holds a line break.
func (e *engine) multiline(sp syntax.Span) bool {
	return strings.ContainsAny(string(e.src[sp.Start:sp.End]), "\n\r")
}

// touched reports whether a unit of [a, b] is already rewritten.
func (e *engine) touched(a, b int) bool {
	for k := a; k <= b; k++ {
		if e.deleted[k] || e.frozenUnit(k) {
			return true
		}
	}
	return false
}

// replaceUnits writes text in place of units a to b.
func (e *engine) replaceUnits(a, b int, text, rule string) {
	e.merged = true
	e.rewrite(a, text, rule)
	for k := a + 1; k <= b; k++ {
		e.remove(k, rule)
		e.gaps[k].drop = true
	}
	if b+1 < len(e.gaps) {
		e.gaps[b+1].drop = false
	}
}

// incrementStyle writes a standalone increment or decrement in the
// configured position.
func (e *engine) incrementStyle(v *syntax.IncDec) {
	defer e.newGroup()()
	want := e.s.IncrementStyle
	if want == "keep" || (want == "pre") == v.Prefix {
		return
	}
	switch v.Var.(type) {
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch:
	default:
		return // not a variable: the other position does not parse
	}
	sp := v.Var.Span()
	op := e.unitAt(v.Op.Span.Start)
	a, b := e.unitAt(sp.Start), e.unitBefore(sp.End)
	if op < 0 || a < 0 || b < a || e.touched(min(op, a), max(op, b)) || e.commentBetween(v.Span().Start, v.Span().End) {
		return
	}
	opText := e.text[op]
	e.merged = true
	if want == "pre" {
		e.remove(op, "increment_style")
		e.gaps[op].drop = true
		if op+1 < len(e.gaps) {
			e.gaps[op+1].drop = false
		}
		e.rewrite(a, opText+e.text[a], "increment_style")
		return
	}
	e.remove(op, "increment_style")
	e.rewrite(b, e.text[b]+opText, "increment_style")
}

// singleQuote writes a double-quoted string that needs no escape in
// single quotes.
func (e *engine) singleQuote(v *syntax.Literal) {
	raw := v.Raw
	if len(raw) < 2 || raw[0] != '"' || strings.ContainsAny(raw[1:len(raw)-1], `\'$`) {
		return
	}
	e.rewrite(e.unitAt(v.Span().Start), "'"+raw[1:len(raw)-1]+"'", "single_quote")
}

// nowdoc writes a heredoc without interpolation or escapes as a nowdoc.
func (e *engine) nowdoc(v *syntax.Literal) {
	k := e.unitAt(v.Span().Start)
	txt := e.text[k]
	nl := strings.IndexAny(txt, "\n\r") // a heredoc opener ends its line
	head, body := txt[:nl], txt[nl:]
	if strings.ContainsAny(body, `\$`) || strings.Contains(head, "'") {
		return
	}
	label := strings.Trim(strings.TrimSpace(strings.TrimPrefix(head, "<<<")), `"`)
	e.rewrite(k, "<<<'"+label+"'"+body, "heredoc_to_nowdoc")
}

// unusedImports removes the imports the file never mentions.
func (e *engine) unusedImports() {
	used := e.usedNames()
	for _, st := range e.allUses() {
		if st.Prefix != nil || len(st.Items) != 1 {
			continue
		}
		it := st.Items[0]
		alias := it.Name.Value[strings.LastIndex(it.Name.Value, `\`)+1:]
		if it.Alias != nil {
			alias = it.Alias.Value
		}
		key := alias
		if useKind(st) != syntax.UseConst {
			key = strings.ToLower(alias)
		}
		if used[key] {
			continue
		}
		sp := st.Span()
		a, b := e.unitAt(sp.Start), e.unitBefore(sp.End)
		if a < 1 || b < a || e.kind(b) != syntax.TSemicolon || e.touched(a, b) || e.commentBetween(sp.Start, sp.End) {
			continue // an import ended by ?> keeps the tag
		}
		keep := e.gaps[a].drop
		end := e.newGroup()
		for k := a; k <= b; k++ {
			e.remove(k, "remove_unused_imports")
		}
		end()
		after := &e.gaps[b+1] // b is never the last unit: EOF follows
		switch {
		case after.comment:
			e.gaps[a].drop = keep
			// A comment trailing the import takes its place.
			j := a // the gap written before it, past earlier removed imports
			for j > 1 && e.gaps[j].drop && e.deleted[j-1] {
				j--
			}
			e.gaps[j].toComment = true
			after.into = j
			if p0 := after.pieces[0]; !p0.comment {
				// The breaks before and after the import merge into one run.
				after.dropLead = true
				if n := lineBreaks(e.src[p0.start:p0.end]); n > 0 {
					before := &e.gaps[j]
					if before.comment {
						e.setTail(j, max(e.tailBreaks(before), n), "remove_unused_imports")
					} else {
						e.setNL(j, max(e.leadBreaks(before), n), "remove_unused_imports")
					}
				}
			}
		case e.leadBreaks(&e.gaps[e.written(a)]) > 0 || e.gaps[e.written(a)].comment || e.leadBreaks(&e.gaps[a]) > 0 || e.gaps[a].comment:
			e.gaps[a].drop = keep // the line break before the import (or an earlier removed one) stays
		default:
			e.gaps[a].drop = true
			after.drop = false // the space after it follows the spacing rules
		}
	}
}

// allUses returns the import statements of the file and its namespaces.
func (e *engine) allUses() []*syntax.Use {
	var out []*syntax.Use
	add := func(list []syntax.Stmt) {
		for _, st := range list {
			if u, ok := st.(*syntax.Use); ok {
				out = append(out, u)
			}
		}
	}
	add(e.f.Stmts)
	for _, st := range e.f.Stmts {
		if ns, ok := st.(*syntax.Namespace); ok {
			add(ns.Stmts)
		}
	}
	return out
}

var docWord = regexp.MustCompile(`[A-Za-z_\x80-\xff][A-Za-z0-9_\x80-\xff]*`)

// usedNames collects the first segment of every name written outside the
// imports, and every word of the doc comments.
func (e *engine) usedNames() map[string]bool {
	used := map[string]bool{}
	var inUse []syntax.Span
	for _, u := range e.allUses() {
		inUse = append(inUse, u.Span())
	}
	inside := func(off uint32) bool {
		for _, sp := range inUse {
			if sp.Start <= off && off < sp.End {
				return true
			}
		}
		return false
	}
	for _, t := range e.toks {
		switch t.Kind {
		case syntax.TString, syntax.TNameQualified:
			if inside(t.Start) {
				continue
			}
			w := string(e.src[t.Start:t.End])
			if i := strings.Index(w, `\`); i >= 0 {
				w = w[:i]
			}
			used[w] = true
			used[strings.ToLower(w)] = true
		case syntax.TDocComment, syntax.TComment:
			for _, w := range docWord.FindAll(e.src[t.Start:t.End], -1) {
				used[string(w)] = true
				used[strings.ToLower(string(w))] = true
			}
		}
	}
	return used
}

// strictTypes adds declare(strict_types=1) to a PHP-only file lacking it.
func (e *engine) strictTypes() {
	if e.template || !e.php.AtLeast(phpversion.PHP70) || len(e.sig) < 2 || e.kind(0) != syntax.TOpenTag {
		return
	}
	for _, st := range e.f.Stmts {
		if d, ok := st.(*syntax.Declare); ok {
			for _, it := range d.Items {
				if strings.EqualFold(it.Key.Value, "strict_types") {
					return
				}
			}
		}
	}
	txt := strings.TrimRight(e.text[0], " \t\r\n")
	e.rewrite(0, txt+e.eol+e.eol+"declare(strict_types=1);"+e.eol, "declare_strict_types")
	if e.kind(1) != syntax.TEOF && !e.gaps[1].comment {
		e.setNL(1, 1, "declare_strict_types")
	}
}

// orderTraits sorts the traits of a use statement, each name taking the
// place of another; the separators and the adaptation block stay.
func (e *engine) orderTraits(v *syntax.TraitUse) {
	defer e.newGroup()()
	sorted := slices.Clone(v.Traits)
	slices.SortStableFunc(sorted, func(a, b *syntax.Name) int {
		return strings.Compare(traitKey(a), traitKey(b))
	})
	if slices.Equal(sorted, v.Traits) {
		return
	}
	first, last := v.Traits[0].Span(), v.Traits[len(v.Traits)-1].Span()
	if e.commentBetween(first.Start, last.End) || e.touched(e.unitAt(first.Start), e.unitBefore(last.End)) {
		return
	}
	text := make([]string, len(sorted))
	for i, n := range sorted {
		text[i] = e.unitsText(e.unitAt(n.Span().Start), e.unitBefore(n.Span().End))
	}
	for i, n := range v.Traits {
		e.replaceUnits(e.unitAt(n.Span().Start), e.unitBefore(n.Span().End), text[i], "ordered_traits")
	}
}

// traitKey is the sort key of a trait name.
func traitKey(n *syntax.Name) string {
	return strings.ToLower(strings.TrimPrefix(n.Value, `\`))
}
