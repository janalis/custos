// Package contentdispositionfilenameneedsquoting implements the native ContentDispositionFilenameNeedsQuoting inspection.
package contentdispositionfilenameneedsquoting

import (
	"regexp"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Quote the Content-Disposition filename."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ContentDispositionFilenameNeedsQuoting" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, _, known := semanticquery.ExpansionHeader(ctx, c)
	if !known || key != "content-disposition" {
		return
	}
	start, end, ok := filenameParameter(value)
	if !ok {
		return
	}
	text := strings.TrimSpace(value[start:end])
	if text == "" || strings.HasPrefix(text, `"`) {
		return
	}
	token := true
	for _, r := range text {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", r)) {
			token = false
		}
	}
	if token {
		return
	}
	var fixes []diagnostic.Fix
	arg := semanticquery.CallArgument(c.Args, 0, "header")
	if _, quote, ok := astquery.QuotedStringRaw(arg); ok && !strings.ContainsAny(text, "\r\n;") {
		full, _ := semanticquery.NativeString(ctx, arg)
		at := strings.Index(full, ":") + 1
		leading := len(full[at:]) - len(strings.TrimLeft(full[at:], " \t"))
		at += leading
		quoted := `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(text) + `"`
		replacement := full[:at+start] + quoted + full[at+end:]
		replacement = strings.ReplaceAll(replacement, `\`, `\\`)
		if quote == '\'' {
			replacement = "'" + strings.ReplaceAll(replacement, "'", `\'`) + "'"
		} else {
			replacement = `"` + strings.NewReplacer(`"`, `\"`, `$`, `\$`).Replace(replacement) + `"`
		}
		fixes = append(fixes, astquery.ReplaceFix(arg.Span(), replacement))
	}
	ctx.ReportNode(c, message, fixes...)
}

// filenameParameter returns the value range of the only filename parameter.
// Parameters are split on semicolons outside quoted-strings, so text inside
// another parameter's quoted value is never taken for a parameter; an
// unterminated quoted-string leaves the boundaries unknown.
func filenameParameter(value string) (start, end int, ok bool) {
	found, quoted, from := 0, false, 0
	for i := 0; i <= len(value); i++ {
		switch {
		case i < len(value) && quoted && value[i] == '\\':
			i++
			continue
		case i < len(value) && value[i] == '"':
			quoted = !quoted
			continue
		case i < len(value) && (quoted || value[i] != ';'):
			continue
		}
		if m := filename.FindStringIndex(value[from:i]); m != nil {
			start, end = from+m[1], i
			found++
		}
		from = i + 1
	}
	return start, end, found == 1 && !quoted
}

var filename = regexp.MustCompile(`(?i)^\s*filename\s*=\s*`)
