// Package regexutf8subjectknowninvalid implements the native RegexUtf8SubjectKnownInvalid inspection.
package regexutf8subjectknowninvalid

import (
	"strings"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply valid UTF-8 to a Unicode pattern."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RegexUtf8SubjectKnownInvalid" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	position := 1
	switch name {
	case "preg_match", "preg_match_all", "preg_split":
	case "preg_replace":
		position = 2
	default:
		return
	}
	p, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "pattern"))
	if !k {
		return
	}
	if !unicodeModifier(p) {
		return
	}
	subject, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, position, "subject"))
	if known && !utf8.ValidString(subject) {
		ctx.ReportNode(c, message)
	}
}

// unicodeModifier reports whether pattern compiles with the u modifier: the
// body ends at the first unescaped closing delimiter (the matching one for
// bracket-style delimiters) and only modifier letters may follow it.
func unicodeModifier(pattern string) bool {
	p := strings.TrimLeft(pattern, " \t\n\r\v\f")
	if p == "" || p[0] == '\\' || p[0] == 0 || p[0] >= 0x80 || isAlnum(p[0]) {
		return false
	}
	open, closing := p[0], p[0]
	if i := strings.IndexByte("([{<", open); i >= 0 {
		closing = ")]}>"[i]
	}
	end := -1
	for i, depth := 1, 1; i < len(p) && end < 0; i++ {
		switch p[i] {
		case '\\':
			i++
		case closing:
			if depth--; depth == 0 {
				end = i
			}
		case open:
			depth++
		}
	}
	if end < 0 {
		return false
	}
	unicode := false
	for _, c := range []byte(p[end+1:]) {
		switch {
		case c == 'u':
			unicode = true
		case !strings.ContainsRune("imsxnrADSUXJ \n", rune(c)):
			return false
		}
	}
	return unicode
}

func isAlnum(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
