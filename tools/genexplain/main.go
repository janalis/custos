// Command genexplain extracts the "Summary" and "Options" sections of every
// spec (specs/<ID>.md, our own clean-room text) into
// internal/meta/descriptions.json for `custos explain`.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type desc struct {
	Summary string `json:"summary"`
	Options string `json:"options,omitempty"`
	Fix     bool   `json:"fix,omitempty"` // custos offers a quick-fix (from own fixtures)
}

func section(md, title string) string {
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

func main() {
	files, _ := filepath.Glob("specs/*.md")
	out := map[string]desc{}
	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".md")
		if strings.HasPrefix(id, "_") || id == "README" {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		md := string(b)
		d := desc{Summary: section(md, "Summary"), Options: section(md, "Options"), Fix: hasRealFix(id)}
		if strings.EqualFold(strings.Trim(d.Options, ". "), "none") {
			d.Options = ""
		}
		out[id] = d
	}
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b, _ := json.MarshalIndent(out, "", "  ")
	if err := os.WriteFile("internal/meta/descriptions.json", append(b, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("descriptions: %d rules\n", len(keys))
}

var tagRe = regexp.MustCompile(`</?(?:error|warning|weak_warning|info)(?:\s[^>]*)?>|<caret>`)

// hasRealFix reports whether some own fixture's expected fix output
// (*.fixed.*) differs from its source once markup and whitespace are
// ignored — "no change" expectations do not count as a quick-fix.
func hasRealFix(id string) bool {
	var fixed []string
	for _, pat := range []string{"*.fixed.*", "*/*.fixed.*"} {
		m, _ := filepath.Glob(filepath.Join("testdata/rules", id, pat))
		fixed = append(fixed, m...)
	}
	norm := func(b []byte) string { return strings.Join(strings.Fields(tagRe.ReplaceAllString(string(b), "")), " ") }
	for _, f := range fixed {
		src := strings.Replace(f, ".fixed.", ".", 1)
		a, err1 := os.ReadFile(src)
		b, err2 := os.ReadFile(f)
		if err1 != nil || err2 != nil || norm(a) != norm(b) {
			return true
		}
	}
	return false
}
