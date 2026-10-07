package analysis

import "custos/internal/syntax"

// SuppressEdit returns the edit that suppresses rule for the finding at span:
// a `// @custos-ignore <rule>` line inserted before the innermost statement
// containing span that starts its own line (same indentation). ok is false
// when no such statement exists (e.g. code on the opening-tag line).
//
// The comment applies to the whole statement. A comment before the first
// statement of a file would apply to the whole file, so that statement is
// never annotated. Callers should re-analyse the edited source to confirm
// only the intended finding goes away.
func SuppressEdit(f *syntax.File, rule string, span syntax.Span) (TextEdit, bool) {
	var inner syntax.Stmt
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if !n.Span().Contains(span) {
			return false
		}
		if s, ok := n.(syntax.Stmt); ok {
			inner = s
		}
		return true
	})
	for n := syntax.Node(inner); n != nil; n = n.Parent() {
		s, ok := n.(syntax.Stmt)
		if !ok {
			continue
		}
		if len(f.Stmts) > 0 && s.Span().Start == f.Stmts[0].Span().Start {
			break // file-level position
		}
		if lineStart, indent, ok := ownLine(f.Src, s.Span().Start); ok {
			return TextEdit{
				Span:    syntax.Span{Start: lineStart, End: lineStart},
				NewText: indent + "// @custos-ignore " + rule + "\n",
			}, true
		}
	}
	return TextEdit{}, false
}

// ownLine reports whether only spaces and tabs precede offset on its line,
// returning the line start and that indentation.
func ownLine(src []byte, offset uint32) (lineStart uint32, indent string, ok bool) {
	i := offset
	for i > 0 && (src[i-1] == ' ' || src[i-1] == '\t') {
		i--
	}
	if i > 0 && src[i-1] != '\n' && src[i-1] != '\r' {
		return 0, "", false
	}
	return i, string(src[i:offset]), true
}
