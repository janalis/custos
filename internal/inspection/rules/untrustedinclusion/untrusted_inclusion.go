package untrustedinclusion

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// untrustedInclusion reports include/require of relative paths, which are
// resolved through include_path.
type untrustedInclusion struct{}

func (untrustedInclusion) ID() string               { return "UntrustedInclusion" }
func (untrustedInclusion) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KInclude} }

// inclusionLiteral returns the raw content of a string literal; ok is false
// for non-literals. Double-quoted strings starting with an interpolation are
// not considered (their path is unknown).
func inclusionLiteral(e syntax.Expr) (string, bool) {
	if raw, _, ok := astquery.QuotedStringRaw(e); ok {
		return raw, true
	}
	is, ok := e.(*syntax.InterpolatedString)
	if !ok || is.Heredoc || is.Backtick || len(is.Parts) == 0 {
		return "", false
	}
	first, ok := is.Parts[0].(*syntax.StringPart)
	if !ok {
		return "", false
	}
	return first.Raw, true
}

func (untrustedInclusion) Check(ctx *analysis.Context, n syntax.Node) {
	inc := n.(*syntax.Include)
	if inc.Expr == nil || inc.Expr.Span().Len() == 0 {
		return
	}
	content, ok := inclusionLiteral(inc.Expr) // D2
	if !ok {
		count := 0
		for _, v := range semanticquery.DiscoverValues(ctx.Types(), inc.Expr) {
			if c, isLit := inclusionLiteral(v); isLit {
				content = c
				count++
			}
		}
		if count != 1 {
			return
		}
	}
	if content == "" || content[0] == '/' || content[0] == '\\' { // D3, E1, E3
		return
	}
	if uiStreamWrapper(content) { // E4
		return
	}
	if len(content) >= 2 && content[1] == ':' && (content[0]|0x20 >= 'a' && content[0]|0x20 <= 'z') {
		return
	}
	ctx.ReportNode(inc, "Relative include depends on include_path; anchor it with __DIR__.")
}

// uiStreamWrapper reports whether path starts with a stream-wrapper scheme
// (`phar://`, `file://`, …): a letter, then at least one more letter, digit,
// `+`, `-` or `.`, then `://`. Single letters are drive letters (D3).
func uiStreamWrapper(path string) bool {
	i := 0
	for i < len(path) {
		c := path[i] | 0x20
		switch {
		case c >= 'a' && c <= 'z':
		case i > 0 && (path[i] >= '0' && path[i] <= '9' || path[i] == '+' || path[i] == '-' || path[i] == '.'):
		default:
			return i >= 2 && len(path) >= i+3 && path[i:i+3] == "://"
		}
		i++
	}
	return false
}
