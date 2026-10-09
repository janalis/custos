// Command architecture checks package ownership and inspection completeness.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("architecture", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	problems, err := check(*root)
	if err != nil {
		fmt.Fprintln(stderr, "architecture:", err)
		return 2
	}
	for _, problem := range problems {
		fmt.Fprintln(stdout, problem)
	}
	if len(problems) > 0 {
		return 1
	}
	fmt.Fprintln(stdout, "architecture: package boundaries and inspection catalogue verified")
	return 0
}

func check(root string) ([]string, error) {
	raw, err := os.ReadFile(filepath.Join(root, "internal/inspection/meta/rules.json"))
	if err != nil {
		return nil, err
	}
	var facts []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &facts); err != nil {
		return nil, err
	}
	expected := map[string]bool{}
	for _, fact := range facts {
		expected[fact.ID] = true
	}
	var problems []string
	report := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }
	ids := map[string]string{}
	constructors := map[string]bool{}
	ruleDirs := map[string]bool{}
	listed := map[string]int{}
	for _, base := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, base), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			from := filepath.ToSlash(filepath.Dir(rel))
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			imports := map[string]string{}
			for _, im := range file.Imports {
				target, _ := strconv.Unquote(im.Path.Value)
				alias := filepath.Base(target)
				if im.Name != nil {
					alias = im.Name.Name
				}
				imports[alias] = target
				target = strings.TrimPrefix(target, "custos/")
				if strings.HasPrefix(target, "internal/") && !allowed(from, target) {
					report("%s: forbidden dependency on %s", rel, target)
				}
			}
			isRule := strings.HasPrefix(from, "internal/inspection/rules/")
			if isRule {
				ruleDirs[from] = true
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if isRule {
					if fn.Name.Name == "init" {
						report("%s: inspections cannot register through init", rel)
					}
					if constructor(fn, imports) {
						constructors[from] = true
					}
					if fn.Name.Name == "ID" && fn.Recv != nil && fn.Body != nil {
						ast.Inspect(fn.Body, func(node ast.Node) bool {
							ret, ok := node.(*ast.ReturnStmt)
							if !ok || len(ret.Results) != 1 {
								return true
							}
							lit, ok := ret.Results[0].(*ast.BasicLit)
							if !ok || lit.Kind != token.STRING {
								return true
							}
							id, _ := strconv.Unquote(lit.Value)
							if previous := ids[id]; previous != "" {
								report("%s: duplicate inspection ID %s (also %s)", rel, id, previous)
							}
							ids[id] = from
							if filepath.Base(from) != strings.ToLower(id) {
								report("%s: directory must be lowercase ID %s", rel, id)
							}
							if !expected[id] {
								report("%s: unknown inspection ID %s", rel, id)
							}
							return true
						})
					}
				}
				if from == "internal/inspection/catalogue" && fn.Name.Name == "All" {
					ast.Inspect(fn.Body, func(node ast.Node) bool {
						call, ok := node.(*ast.CallExpr)
						if !ok {
							return true
						}
						sel, ok := call.Fun.(*ast.SelectorExpr)
						if !ok || sel.Sel.Name != "New" {
							return true
						}
						name, ok := sel.X.(*ast.Ident)
						if !ok {
							return true
						}
						target := strings.TrimPrefix(imports[name.Name], "custos/")
						if strings.HasPrefix(target, "internal/inspection/rules/") {
							listed[target]++
						}
						return true
					})
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for id := range expected {
		owner := ids[id]
		if owner == "" {
			report("%s: inspection implementation missing", id)
			continue
		}
		if !constructors[owner] {
			report("%s: New constructor missing", id)
		}
		if listed[owner] != 1 {
			report("%s: catalogue must construct inspection exactly once (got %d)", id, listed[owner])
		}
		spec, err := os.Stat(filepath.Join(root, "specs", id+".md"))
		if err != nil || spec.IsDir() {
			report("%s: specification missing", id)
		}
		fixtures, err := os.ReadDir(filepath.Join(root, "testdata/rules", id))
		hasPHP := false
		for _, fixture := range fixtures {
			if !fixture.IsDir() && strings.HasSuffix(fixture.Name(), ".php") {
				hasPHP = true
			}
		}
		if err != nil || !hasPHP {
			report("%s: own PHP fixtures missing", id)
		}
	}
	for owner := range ruleDirs {
		found := false
		for _, dir := range ids {
			if dir == owner {
				found = true
			}
		}
		if !found {
			report("%s: rule directory has no inspection ID", owner)
		}
	}
	for owner := range constructors {
		if listed[owner] != 1 {
			report("%s: constructor is absent or repeated in catalogue", owner)
		}
	}
	for owner := range listed {
		if !constructors[owner] {
			report("%s: catalogue references missing constructor", owner)
		}
	}
	sort.Strings(problems)
	return problems, nil
}

func allowed(from, to string) bool {
	within := func(prefix string) bool { return to == prefix || strings.HasPrefix(to, prefix+"/") }
	switch {
	case strings.HasPrefix(from, "internal/php/"):
		return within("internal/php")
	case strings.HasPrefix(from, "internal/semantic/"):
		return within("internal/php") || within("internal/semantic")
	case from == "internal/diagnostic":
		return within("internal/php")
	case from == "internal/fixing":
		return within("internal/php") || within("internal/diagnostic")
	case from == "internal/inspection/meta":
		return within("internal/diagnostic")
	case from == "internal/inspection/analysis":
		return within("internal/php") || within("internal/semantic") || within("internal/diagnostic") || within("internal/inspection/meta")
	case from == "internal/inspection/astquery":
		return within("internal/php") || within("internal/diagnostic")
	case from == "internal/inspection/flowquery":
		return within("internal/php") || within("internal/inspection/astquery")
	case strings.HasPrefix(from, "internal/inspection/rules/"):
		return within("internal/php") || within("internal/semantic") || within("internal/diagnostic") || within("internal/inspection/analysis") || within("internal/inspection/meta") || within("internal/inspection/astquery") || within("internal/inspection/flowquery") || within("internal/inspection/semanticquery") || within("internal/inspection/phpunit")
	case from == "internal/inspection/catalogue":
		return within("internal/inspection/analysis") || within("internal/inspection/rules")
	case strings.HasPrefix(from, "internal/inspection/"):
		return within("internal/php") || within("internal/semantic") || within("internal/diagnostic") || within("internal/inspection/analysis") || within("internal/inspection/meta") || within("internal/inspection/astquery") || within("internal/inspection/flowquery")
	case strings.HasPrefix(from, "internal/project"):
		return !within("internal/cli") && !within("internal/editor") && !within("internal/output")
	case strings.HasPrefix(from, "internal/output/"):
		return within("internal/diagnostic") || within("internal/php") || within("internal/inspection/meta")
	case strings.HasPrefix(from, "internal/platform/"):
		return within("internal/platform")
	case strings.HasPrefix(from, "cmd/"):
		return within("internal/cli")
	default:
		return true
	}
}

func constructor(fn *ast.FuncDecl, imports map[string]string) bool {
	if fn.Name.Name != "New" || fn.Recv != nil {
		return false
	}
	if len(fn.Type.Params.List) != 0 || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	result, ok := fn.Type.Results.List[0].Type.(*ast.SelectorExpr)
	if !ok || result.Sel.Name != "Rule" {
		return false
	}
	qualifier, ok := result.X.(*ast.Ident)
	return ok && imports[qualifier.Name] == "custos/internal/inspection/analysis"
}
