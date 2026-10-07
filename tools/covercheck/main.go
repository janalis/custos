// Command covercheck fails unless every statement in a Go coverage profile
// was executed by at least one test binary. The only statements it exempts
// are the body of a one-statement `func main()` in package main (the
// `os.Exit(run(...))` wrapper of each command, which no test can run): the
// exemption is derived from the source, so it never goes stale.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("covercheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	what := fs.String("what", "packages", "name of the measured code, for messages")
	module := fs.String("module", "custos", "module path (profile entries are module-relative)")
	root := fs.String("root", ".", "module root directory")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: covercheck [-what name] [-module path] [-root dir] profile")
		return 2
	}
	uncovered, err := check(fs.Arg(0), *module, *root)
	if err != nil {
		fmt.Fprintln(stderr, "covercheck:", err)
		return 2
	}
	for _, u := range uncovered {
		fmt.Fprintln(stdout, "uncovered:", u)
	}
	if len(uncovered) > 0 {
		fmt.Fprintf(stdout, "%d uncovered block(s) in %s\n", len(uncovered), *what)
		return 1
	}
	fmt.Fprintf(stdout, "coverage: %s 100%%\n", *what)
	return 0
}

// block is one profile entry: file and line.col range.
type block struct {
	file                     string
	line0, col0, line1, col1 int
}

// check returns the blocks (as written in the profile) that no test binary
// executed and that are not an exempt main wrapper, sorted.
func check(profile, module, root string) ([]string, error) {
	f, err := os.Open(profile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stmts := map[string]int{}
	hit := map[string]bool{}
	sc := bufio.NewScanner(f)
	for first := true; sc.Scan(); first = false {
		line := sc.Text()
		if first && strings.HasPrefix(line, "mode:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("malformed profile line %q", line)
		}
		n, err1 := strconv.Atoi(fields[1])
		c, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("malformed profile line %q", line)
		}
		stmts[fields[0]] = n
		if c > 0 {
			hit[fields[0]] = true
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	mains := map[string][2]token.Position{} // file -> main body braces
	var out []string
	for k, n := range stmts {
		if n == 0 || hit[k] {
			continue
		}
		b, err := parseBlock(k)
		if err != nil {
			return nil, err
		}
		body, ok := mains[b.file]
		if !ok {
			if body, err = mainBody(filepath.Join(root, strings.TrimPrefix(b.file, module+"/"))); err != nil {
				return nil, err
			}
			mains[b.file] = body
		}
		if body[0].Line != 0 && within(b, body) {
			continue
		}
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

// parseBlock parses "file.go:L0.C0,L1.C1".
func parseBlock(k string) (block, error) {
	colon := strings.LastIndexByte(k, ':')
	var b block
	if colon < 0 {
		return b, fmt.Errorf("malformed block %q", k)
	}
	b.file = k[:colon]
	if _, err := fmt.Sscanf(k[colon+1:], "%d.%d,%d.%d", &b.line0, &b.col0, &b.line1, &b.col1); err != nil {
		return b, fmt.Errorf("malformed block %q", k)
	}
	return b, nil
}

// mainBody returns the brace positions of `func main()`'s body when path is
// in package main and that body has exactly one statement; zero otherwise.
func mainBody(path string) ([2]token.Position, error) {
	var none [2]token.Position
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return none, err
	}
	if f.Name.Name != "main" {
		return none, nil
	}
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok && fn.Recv == nil && fn.Name.Name == "main" && fn.Body != nil && len(fn.Body.List) == 1 {
			return [2]token.Position{fset.Position(fn.Body.Lbrace), fset.Position(fn.Body.Rbrace)}, nil
		}
	}
	return none, nil
}

// within reports whether b lies inside the braces (columns 1-based).
func within(b block, body [2]token.Position) bool {
	after := b.line0 > body[0].Line || (b.line0 == body[0].Line && b.col0 >= body[0].Column)
	before := b.line1 < body[1].Line || (b.line1 == body[1].Line && b.col1 <= body[1].Column+1)
	return after && before
}
