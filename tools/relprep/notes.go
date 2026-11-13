package main

import (
	"regexp"
	"strings"
)

var noteListPrefix = regexp.MustCompile(`^ *([-+*]|\d+[.)])[ \t]+`)

// formatReleaseNotes removes source-only prose wrapping. Markdown blocks,
// nested list boundaries, and explicit hard breaks retain their line breaks.
func formatReleaseNotes(body string) string {
	lines := strings.Split(strings.Trim(strings.ReplaceAll(body, "\r\n", "\n"), "\n"), "\n")
	out := make([]string, 0, len(lines))
	fence := ""
	indent := 0
	prose := false
	for _, line := range lines {
		text := strings.TrimSpace(line)
		if fence != "" {
			out = append(out, line)
			if strings.HasPrefix(text, fence) && strings.TrimSpace(strings.TrimLeft(text, fence[:1])) == "" {
				fence = ""
			}
			continue
		}
		if marker := noteFence(text); marker != "" {
			fence = marker
			prose = false
			out = append(out, line)
			continue
		}
		if text == "" || noteBlock(text) || strings.HasPrefix(line, "\t") {
			prose = false
			out = append(out, line)
			continue
		}
		if prefix := noteListPrefix.FindStringIndex(line); prefix != nil {
			indent = prefix[1]
			content := strings.TrimSpace(line[indent:])
			fence = noteFence(content)
			prose = content != "" && fence == "" && !noteBlock(content)
			out = append(out, line)
			continue
		}
		leading := len(line) - len(strings.TrimLeft(line, " "))
		if prose && leading == indent && !strings.HasSuffix(out[len(out)-1], "  ") && !strings.HasSuffix(out[len(out)-1], "\\") {
			out[len(out)-1] = strings.TrimRight(out[len(out)-1], " ") + " " + text
			continue
		}
		indent = leading
		prose = leading < 4
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func noteFence(text string) string {
	if len(text) < 3 || text[0] != '`' && text[0] != '~' {
		return ""
	}
	count := 1
	for count < len(text) && text[count] == text[0] {
		count++
	}
	if count < 3 {
		return ""
	}
	return text[:count]
}

func noteBlock(text string) bool {
	return strings.HasPrefix(text, "#") || strings.HasPrefix(text, ">") ||
		strings.HasPrefix(text, "|") || strings.HasPrefix(text, "<") ||
		strings.HasPrefix(text, "[") && strings.Contains(text, "]:") ||
		len(text) >= 3 && strings.Trim(text, " -*_=") == ""
}
