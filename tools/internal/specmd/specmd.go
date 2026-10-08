// Package specmd reads the parts of a clean-room spec (specs/<ID>.md) that
// the documentation generators reuse: its "## " sections, the PHP version
// range of its front matter and the fenced code blocks of a section.
package specmd

import (
	"regexp"
	"strings"
)

// Section returns the trimmed body of the "## title" section of md, or "".
func Section(md, title string) string {
	i := strings.Index(md, "\n## "+title)
	if i < 0 {
		return ""
	}
	rest := md[i+len("\n## "+title):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	return strings.TrimSpace(rest)
}

var phpRe = regexp.MustCompile(`(?m)^php:\s*\{\s*min:\s*"([^"]*)",\s*max:\s*"([^"]*)"\s*\}`)

// PHP returns the target PHP versions the rule is active for, from the
// front matter line `php: { min: "7.1", max: "" }` ("" = unbounded).
func PHP(md string) (min, max string) {
	if m := phpRe.FindStringSubmatch(md); m != nil {
		return m[1], m[2]
	}
	return "", ""
}

// Block is a fenced code block of a section.
type Block struct {
	Lang string
	Body string // without the fences, ending with a newline
	// Prose is the text between the previous block (or, for the first
	// block, the section heading) and this one, trimmed.
	Prose string
}

// Blocks returns the ``` fenced blocks of a section body, in order.
func Blocks(section string) []Block {
	var out []Block
	var cur *Block
	var body, prose strings.Builder
	for _, line := range strings.Split(section, "\n") {
		switch {
		case cur == nil && strings.HasPrefix(line, "```"):
			cur = &Block{Lang: strings.TrimSpace(strings.TrimPrefix(line, "```")), Prose: strings.TrimSpace(prose.String())}
			body.Reset()
		case cur != nil && strings.TrimSpace(line) == "```":
			cur.Body = body.String()
			out = append(out, *cur)
			cur = nil
			prose.Reset()
		case cur != nil:
			body.WriteString(line + "\n")
		default:
			prose.WriteString(line + "\n")
		}
	}
	return out
}
