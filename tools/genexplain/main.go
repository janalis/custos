// Command genexplain extracts the "Summary" and "Options" sections of every
// spec (specs/<ID>.md, our own clean-room text) into
// internal/inspection/meta/descriptions.json for `custos explain`.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"custos/tools/internal/specmd"
)

type desc struct {
	Summary string `json:"summary"`
	Options string `json:"options,omitempty"`
	Fix     bool   `json:"fix,omitempty"` // custos offers a quick-fix (from own fixtures)
}

func main() { os.Exit(run(".", os.Stdout, os.Stderr)) }

// run reads root/specs and root/testdata/rules and writes
// root/internal/inspection/meta/descriptions.json.
func run(root string, stdout, stderr io.Writer) int {
	files, _ := filepath.Glob(filepath.Join(root, "specs", "*.md"))
	out := map[string]desc{}
	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".md")
		if strings.HasPrefix(id, "_") || id == "README" {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		md := string(b)
		d := desc{Summary: specmd.Section(md, "Summary"), Options: specmd.Section(md, "Options"), Fix: hasRealFix(root, id)}
		if strings.EqualFold(strings.Trim(d.Options, ". "), "none") {
			d.Options = ""
		}
		out[id] = d
	}
	// encoding/json sorts map keys: the output is deterministic.
	b, _ := json.MarshalIndent(out, "", "  ")
	if err := os.WriteFile(filepath.Join(root, "internal", "inspection", "meta", "descriptions.json"), append(b, '\n'), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "descriptions: %d rules\n", len(out))
	return 0
}

var tagRe = regexp.MustCompile(`</?(?:error|warning|weak_warning|info)(?:\s[^>]*)?>|<caret>`)

// hasRealFix reports whether some own fixture's expected fix output
// (*.fixed.*) differs from its source once markup and whitespace are
// ignored — "no change" expectations do not count as a quick-fixing.
func hasRealFix(root, id string) bool {
	var fixed []string
	for _, pat := range []string{"*.fixed.*", "*/*.fixed.*"} {
		m, _ := filepath.Glob(filepath.Join(root, "testdata", "rules", id, pat))
		fixed = append(fixed, m...)
	}
	norm := func(b []byte) string { return strings.Join(strings.Fields(tagRe.ReplaceAllString(string(b), "")), " ") }
	for _, f := range fixed {
		src := filepath.Join(filepath.Dir(f), strings.Replace(filepath.Base(f), ".fixed.", ".", 1))
		a, err1 := os.ReadFile(src)
		b, err2 := os.ReadFile(f)
		if err1 != nil || err2 != nil || norm(a) != norm(b) {
			return true
		}
	}
	return false
}
