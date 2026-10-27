package autoloadingissues

import (
	"regexp"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// autoloadingIssues reports a file whose single class-like declaration is not
// named after the file (PSR-0/PSR-4 autoloading would not find it).
type autoloadingIssues struct{}

const autoloadingIssuesMsg = "File name does not match the class name; autoloading may fail."

var timestampedMigration = regexp.MustCompile(`^\d{4}_\d{2}_\d{2}_\d{6}_.+\.php$`)

func (autoloadingIssues) ID() string                           { return "AutoloadingIssues" }
func (autoloadingIssues) Kinds() []syntax.NodeKind             { return nil }
func (autoloadingIssues) Check(*analysis.Context, syntax.Node) {}
func (autoloadingIssues) CheckFile(ctx *analysis.Context) {
	p := strings.ReplaceAll(ctx.File.Path, "\\", "/")
	base := p[strings.LastIndexByte(p, '/')+1:]
	if !strings.HasSuffix(base, ".php") || base == "index.php" || base == "actions.class.php" ||
		timestampedMigration.MatchString(base) { // D1-D3
		return
	}

	// D4: top-level class-like declarations (file level or namespace body).
	var decl *syntax.ClassLike
	global := true
	count, funcs := 0, false
	collect := func(stmts []syntax.Stmt, inNamespace bool) {
		for _, s := range stmts {
			if c, ok := s.(*syntax.ClassLike); ok && c.Name != nil && c.Name.Span().Len() > 0 {
				count++
				decl = c
				global = !inNamespace
			}
			if _, ok := s.(*syntax.Function); ok {
				funcs = true
			}
		}
	}
	collect(ctx.File.Stmts, false)
	for _, s := range ctx.File.Stmts {
		if ns, ok := s.(*syntax.Namespace); ok {
			collect(ns.Stmts, ns.Name != nil)
		}
	}
	if count != 1 || funcs { // E6: functions are never autoloaded
		return
	}

	n := decl.Name.Value // D5
	e := base[:strings.IndexByte(base, '.')]
	x := n
	if global { // D6
		if i := strings.LastIndexByte(n, '_'); i >= 0 {
			x = n[i+1:]
		}
	}
	if e == x || e == n {
		return
	}
	if strings.HasSuffix(base, "class-"+strings.ReplaceAll(strings.ToLower(n), "_", "-")+".php") { // E1
		return
	}
	if usesOwnClass(ctx.File.Stmts, n) { // E5: a script declaring a class for its own use
		return
	}
	ctx.Report(decl.Name.Span(), autoloadingIssuesMsg)
}

// usesOwnClass reports whether code outside the file's declarations
// instantiates class n or calls one of its static methods (E5).
func usesOwnClass(stmts []syntax.Stmt, n string) bool {
	named := func(e syntax.Expr) bool {
		c, ok := e.(*syntax.Name)
		return ok && strings.EqualFold(c.Value[strings.LastIndexByte(c.Value, '\\')+1:], n)
	}
	found := false
	for _, st := range stmts {
		switch st := st.(type) {
		case *syntax.Namespace:
			found = found || usesOwnClass(st.Stmts, n)
			continue
		case *syntax.ClassLike:
			continue
		}
		syntax.Inspect(st, func(x syntax.Node) bool {
			switch x := x.(type) {
			case *syntax.ClassLike, *syntax.Function:
				return false
			case *syntax.New:
				found = found || named(x.Class)
			case *syntax.StaticCall:
				found = found || named(x.Class)
			}
			return !found
		})
	}
	return found
}
