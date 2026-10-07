// Command cleanroom scans the repository for text that appears verbatim in
// the local EA checkout: long Java string literals (messages) and long
// fixture lines. Identifier-like hits (class names) are allowed.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	javaStr    = regexp.MustCompile(`"((?:[^"\\]|\\.){25,})"`)
	tag        = regexp.MustCompile(`<[^>]+>`)
	identifier = regexp.MustCompile(`^[\w\\]+$`)
	forbidden  = regexp.MustCompile(`\[EA\]|Psi[A-Z]\w+|Openapi\w+Util|ExpressionSemanticUtil|LocalQuickFix`)
)

func main() { os.Exit(run(".", os.Args[1:], os.Stdout, os.Stderr)) }

// run scans the repository at root against the EA checkout given by -ea;
// it returns 1 when verbatim upstream text or a forbidden token is found.
func run(root string, args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("cleanroom", flag.ContinueOnError)
	fl.SetOutput(stderr)
	ea := fl.String("ea", os.ExpandEnv("$HOME/Sites/phpinspectionsea"), "EA checkout")
	if err := fl.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if _, err := os.Stat(*ea); err != nil {
		fmt.Fprintln(stdout, "cleanroom: EA checkout not found, skipped")
		return 0
	}
	var needles []string
	walk(filepath.Join(*ea, "src/main/java"), ".java", func(_, b string) {
		for _, m := range javaStr.FindAllStringSubmatch(b, -1) {
			s := strings.TrimSpace(strings.NewReplacer(`\`, "", "[EA] ", "", "%s", "").Replace(m[1]))
			if len(s) >= 25 && !identifier.MatchString(s) {
				needles = append(needles, s)
			}
		}
	})
	walk(filepath.Join(*ea, "testData/fixtures"), "", func(_, b string) {
		for _, l := range strings.Split(b, "\n") {
			// Short lines are mostly generic PHP (signatures, braces); only
			// longer lines are distinctive enough to indicate copying.
			if l = strings.TrimSpace(tag.ReplaceAllString(l, "")); len(l) >= 40 {
				needles = append(needles, l)
			}
		}
	})
	hits := 0
	for _, dir := range []string{"specs", "internal", "cmd", "testdata", "docs"} {
		walk(filepath.Join(root, dir), "", func(p, text string) {
			p, _ = filepath.Rel(root, p)
			p = filepath.ToSlash(p)
			if strings.HasSuffix(p, "rules.json") {
				return
			}
			// The migration plan names upstream internals on purpose (facts,
			// not code); the conformance harness parses the "[EA]" prefix.
			exempt := p == "docs/migration.md" || strings.HasPrefix(p, "internal/conformance/")
			if m := forbidden.FindString(text); m != "" && !exempt {
				fmt.Fprintf(stdout, "%s: forbidden token %q\n", p, m)
				hits++
			}
			for _, n := range needles {
				if strings.Contains(text, n) {
					fmt.Fprintf(stdout, "%s: verbatim EA text %q\n", p, trunc(n))
					hits++
				}
			}
		})
	}
	if hits > 0 {
		fmt.Fprintf(stdout, "cleanroom: %d hit(s)\n", hits)
		return 1
	}
	fmt.Fprintln(stdout, "cleanroom: ok")
	return 0
}

// walk calls fn with the path and content of every readable file under root
// whose name ends with ext ("" for any).
func walk(root, ext string, fn func(path, text string)) {
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || (ext != "" && !strings.HasSuffix(p, ext)) {
			return nil
		}
		if b, err := os.ReadFile(p); err == nil {
			fn(p, string(b))
		}
		return nil
	})
}

// trunc shortens s to at most 70 bytes, on a rune boundary.
func trunc(s string) string {
	if len(s) <= 70 {
		return s
	}
	i := 70
	for !utf8.RuneStart(s[i]) {
		i--
	}
	return s[:i] + "…"
}
