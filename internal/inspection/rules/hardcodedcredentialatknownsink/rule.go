// Package hardcodedcredentialatknownsink implements the native HardcodedCredentialAtKnownSink inspection.
package hardcodedcredentialatknownsink

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Load credentials from external configuration."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HardcodedCredentialAtKnownSink" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew, syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var password syntax.Expr
	switch c := n.(type) {
	case *syntax.New:
		if !semanticquery.NamesGlobalClass(ctx, c.Class, "PDO") && !semanticquery.NamesGlobalClass(ctx, c.Class, "mysqli") {
			return
		}
		name := c.Class.(*syntax.Name)
		constructor := ctx.Index().FindMethod(ctx.Names().Class(name.Value, name.Span().Start), "__construct", ctx.PHP)
		if constructor != nil && constructor.Builtin {
			password = semanticquery.CallArgument(c.Args, 2, "password")
		}
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "mysqli_connect":
			password = semanticquery.CallArgument(c.Args, 2, "password")
		case "mysqli_real_connect":
			password = semanticquery.CallArgument(c.Args, 3, "password")
		default:
			return
		}
	}
	s, known := semanticquery.NativeString(ctx, password)
	if !known || s == "" {
		return
	}
	// The path policy touches the filesystem, so it runs only for a proven
	// literal credential.
	patterns := ctx.List("excludedPaths")
	if len(patterns) > 0 {
		parts := strings.Split(projectPath(ctx.File.Path), "/")
		for _, pattern := range patterns {
			if glob(strings.Split(pattern, "/"), parts) {
				return
			}
		}
	}
	ctx.ReportNode(password, message)
}

// projectPath returns file relative to its project root with forward
// slashes, so the path policy does not depend on the working directory or on
// directory names above the project. The root is the nearest ancestor holding
// custos.json, else the outermost one holding composer.json (a package or
// fixture application nested in the project stays part of it), else the
// working directory. A file outside that root keeps its absolute path.
func projectPath(file string) string {
	abs, _ := filepath.Abs(file)
	root, _ := filepath.Abs(".")
	for d := filepath.Dir(abs); ; d = filepath.Dir(d) {
		if exists(filepath.Join(d, "custos.json")) {
			root = d
			break
		}
		if exists(filepath.Join(d, "composer.json")) {
			root = d
		}
		if filepath.Dir(d) == d {
			break
		}
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		rel = abs
	}
	return strings.ReplaceAll(filepath.ToSlash(rel), `\`, "/")
}

func exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func glob(pattern, parts []string) bool {
	if len(pattern) > 128 || len(parts) > 256 {
		return false
	}
	row := make([]bool, len(parts)+1)
	next := make([]bool, len(row))
	row[0] = true
	for _, segment := range pattern {
		clear(next)
		for j := range next {
			if segment == "**" {
				next[j] = row[j] || (j > 0 && next[j-1])
			} else if j > 0 && row[j-1] {
				matched, err := path.Match(segment, parts[j-1])
				next[j] = err == nil && matched
			}
		}
		row, next = next, row
	}
	return row[len(parts)]
}
