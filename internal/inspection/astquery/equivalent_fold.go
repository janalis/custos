package astquery

import (
	"bytes"
	"strings"

	"custos/internal/php/syntax"
)

// EquivalentFoldNames is Equivalent with PHP's case rules for names: keyword
// tokens and the names PHP resolves case-insensitively — function names,
// method names, and class names in calls, `new`, `instanceof` and `::`
// accesses — compare case-insensitively (`Foo\Bar::Make($x)` matches
// `foo\bar::make($x)`). Variables, properties, constants and named-argument
// labels stay case-sensitive, also when they are spelled like a keyword
// (`Foo::Array`, `f(array: 1)`).
func EquivalentFoldNames(f *syntax.File, a, b syntax.Node) bool {
	if a == nil || b == nil || a.Kind() != b.Kind() {
		return false
	}
	if va, ok := a.(*syntax.Variable); ok {
		vb := b.(*syntax.Variable)
		if va.NameExpr == nil && vb.NameExpr == nil {
			return va.Name == vb.Name
		}
	}
	sa, sb := a.Span(), b.Span()
	if bytes.Equal(f.Src[sa.Start:sa.End], f.Src[sb.Start:sb.End]) {
		return true
	}
	// Start offsets of name tokens, built on first need: true for a name PHP
	// resolves case-insensitively, false for one it compares exactly.
	var fold map[uint32]bool
	i, j := TokenIndex(f, sa.Start), TokenIndex(f, sb.Start)
	for {
		i = skipTrivia(f, i, sa.End)
		j = skipTrivia(f, j, sb.End)
		endA := i >= len(f.Tokens) || f.Tokens[i].End > sa.End
		endB := j >= len(f.Tokens) || f.Tokens[j].End > sb.End
		if endA || endB {
			return endA && endB
		}
		ta, tb := f.Tokens[i], f.Tokens[j]
		xa, xb := f.Src[ta.Start:ta.End], f.Src[tb.Start:tb.End]
		if ta.Kind != tb.Kind {
			return false
		}
		if !bytes.Equal(xa, xb) {
			if !bytes.EqualFold(xa, xb) {
				return false
			}
			// Casts are case-insensitive, and so are keywords unless one
			// spells a case-sensitive name; other tokens only as the names
			// PHP resolves case-insensitively.
			if !ta.Kind.IsCast() {
				if fold == nil {
					fold = map[uint32]bool{}
					foldNameSpans(a, fold)
					foldNameSpans(b, fold)
				}
				foldA, nameA := fold[ta.Start]
				foldB, nameB := fold[tb.Start]
				if ta.Kind.IsKeyword() {
					foldA, foldB = foldA || !nameA, foldB || !nameB
				}
				if !foldA || !foldB {
					return false
				}
			}
		}
		i++
		j++
	}
}

// foldNameSpans records the start offsets of the name tokens under root:
// true for those PHP compares case-insensitively, false for class constant
// names (other than the magic `class`) and named-argument labels, which it
// compares exactly even when the lexer reads them as a keyword.
func foldNameSpans(root syntax.Node, out map[uint32]bool) {
	add := func(e syntax.Expr) {
		switch e.(type) {
		case *syntax.Name, *syntax.Identifier:
			out[e.Span().Start] = true
		}
	}
	syntax.Inspect(root, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.FuncCall:
			add(x.Name)
		case *syntax.MethodCall:
			add(x.Name)
		case *syntax.StaticCall:
			add(x.Class)
			add(x.Name)
		case *syntax.New:
			add(x.Class)
		case *syntax.Instanceof:
			add(x.Class)
		case *syntax.ClassConstFetch:
			add(x.Class)
			if id, ok := x.Name.(*syntax.Identifier); ok && !strings.EqualFold(id.Value, "class") {
				out[id.Span().Start] = false
			}
		case *syntax.Arg:
			if x.Name != nil {
				out[x.Name.Span().Start] = false
			}
		case *syntax.StaticPropertyFetch:
			add(x.Class)
		}
		return true
	})
}
