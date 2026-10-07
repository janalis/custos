// Command cleanroom scans the repository for text that appears verbatim in
// the local EA checkout: long Java string literals (messages) and long
// fixture lines. Identifier-like hits (class names) are allowed.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	javaStr    = regexp.MustCompile(`"((?:[^"\\]|\\.){25,})"`)
	tag        = regexp.MustCompile(`<[^>]+>`)
	identifier = regexp.MustCompile(`^[\w\\]+$`)
	forbidden  = regexp.MustCompile(`\[EA\]|Psi[A-Z]\w+|Openapi\w+Util|ExpressionSemanticUtil|LocalQuickFix`)
)

func main() {
	ea := flag.String("ea", os.ExpandEnv("$HOME/Sites/phpinspectionsea"), "EA checkout")
	flag.Parse()
	if _, err := os.Stat(*ea); err != nil {
		fmt.Println("cleanroom: EA checkout not found, skipped")
		return
	}
	var needles []string
	walk(filepath.Join(*ea, "src/main/java"), ".java", func(b string) {
		for _, m := range javaStr.FindAllStringSubmatch(b, -1) {
			s := strings.TrimSpace(strings.NewReplacer(`\`, "", "[EA] ", "", "%s", "").Replace(m[1]))
			if len(s) >= 25 && !identifier.MatchString(s) {
				needles = append(needles, s)
			}
		}
	})
	walk(filepath.Join(*ea, "testData/fixtures"), "", func(b string) {
		for _, l := range strings.Split(b, "\n") {
			// Short lines are mostly generic PHP (signatures, braces); only
			// longer lines are distinctive enough to indicate copying.
			if l = strings.TrimSpace(tag.ReplaceAllString(l, "")); len(l) >= 40 {
				needles = append(needles, l)
			}
		}
	})
	hits := 0
	for _, root := range []string{"specs", "internal", "cmd", "testdata", "docs"} {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || strings.HasSuffix(p, "rules.json") {
				return nil
			}
			raw, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			text := string(raw)
			// The migration plan names upstream internals on purpose (facts, not code).
			// The migration plan and the conformance harness name upstream
			// markers on purpose (facts / parsing the "[EA]" prefix).
			exempt := p == "docs/migration.md" || strings.HasPrefix(filepath.ToSlash(p), "internal/conformance/")
			if m := forbidden.FindString(text); m != "" && !exempt {
				fmt.Printf("%s: forbidden token %q\n", p, m)
				hits++
			}
			for _, n := range needles {
				if strings.Contains(text, n) {
					fmt.Printf("%s: verbatim EA text %q\n", p, trunc(n))
					hits++
				}
			}
			return nil
		})
	}
	if hits > 0 {
		fmt.Printf("cleanroom: %d hit(s)\n", hits)
		os.Exit(1)
	}
	fmt.Println("cleanroom: ok")
}

func walk(root, ext string, fn func(string)) {
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || (ext != "" && !strings.HasSuffix(p, ext)) {
			return nil
		}
		if b, err := os.ReadFile(p); err == nil {
			fn(string(b))
		}
		return nil
	})
}

func trunc(s string) string {
	if len(s) > 70 {
		return s[:70] + "…"
	}
	return s
}
