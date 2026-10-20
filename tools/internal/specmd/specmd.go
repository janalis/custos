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

var (
	// itemLabel is a Detection or Exceptions item's own label ("- D1a. ").
	itemLabel = regexp.MustCompile(`(?m)^(\s*- )([DE]\d+[a-z]?)\. `)
	label     = regexp.MustCompile(`\b[DE]\d+[a-z]?\b`)
	// parenRef is a parenthetical reference such as " (D2)", " (see E1)"
	// or " (D10–D11)".
	parenRef = regexp.MustCompile(` \((?:see )?([DE]\d+[a-z]?)(?:(?:, |–|-|/| and )[DE]\d+[a-z]?)*\)`)
	// boilerplate is spec wording repeated across many rules that tells a
	// reader nothing about the rule at hand.
	boilerplate = []*regexp.Regexp{
		regexp.MustCompile(`(?m)^\s*- (?:D\d+[a-z]?\. )?Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments\. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments\.\n?`),
		regexp.MustCompile(` ?Cases outside [DE]\d+[a-z]? receive no finding\.`),
		regexp.MustCompile(` ?Apply only where the referenced language features and builtin/extension APIs exist\.`),
		regexp.MustCompile(` ?API-specific later syntax or behavior is gated as described in Detection\.`),
		regexp.MustCompile(" ?Both `[^`]*` and `[^`]*` suppress this inspection\\."),
		regexp.MustCompile(` ?Use ordinary rule configuration to enable or disable this inspection\.`),
	}
)

func dropBoilerplate(text string) string {
	for _, re := range boilerplate {
		text = re.ReplaceAllString(text, "")
	}
	return text
}

// References returns the item labels (D1, E2a…) defined in texts that
// texts also mention elsewhere, ignoring boilerplate.
func References(texts ...string) map[string]bool {
	defined, refs := map[string]bool{}, map[string]bool{}
	var mentions []string
	for _, t := range texts {
		t = dropBoilerplate(t)
		for _, m := range itemLabel.FindAllStringSubmatch(t, -1) {
			defined[m[2]] = true
		}
		mentions = append(mentions, label.FindAllString(itemLabel.ReplaceAllString(t, "$1"), -1)...)
	}
	for _, l := range mentions {
		if defined[l] {
			refs[l] = true
		}
	}
	return refs
}

// Readable prepares spec text for readers of the rule reference and
// `custos explain`: boilerplate is dropped, item labels are kept only when
// keep holds them (something on the page refers to them), parenthetical
// references to other labels are dropped, and "inspection" reads "rule".
func Readable(text string, keep map[string]bool) string {
	text = dropBoilerplate(text)
	text = itemLabel.ReplaceAllStringFunc(text, func(m string) string {
		sub := itemLabel.FindStringSubmatch(m)
		if keep[sub[2]] {
			return m
		}
		return sub[1]
	})
	text = parenRef.ReplaceAllStringFunc(text, func(m string) string {
		if keep[parenRef.FindStringSubmatch(m)[1]] {
			return m
		}
		return ""
	})
	return strings.TrimSpace(strings.ReplaceAll(text, "this inspection", "this rule"))
}
