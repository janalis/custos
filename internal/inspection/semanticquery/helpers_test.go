package semanticquery

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func Parse(t *testing.T, src string) *syntax.File {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.Max})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %q: %v", src, f.Errors)
	}
	return f
}

// FirstExpr returns the expression of the first ExprStmt.
func FirstExpr(t *testing.T, f *syntax.File) syntax.Expr {
	t.Helper()
	for _, s := range f.Stmts {
		if es, ok := s.(*syntax.ExprStmt); ok {
			return es.Expr
		}
	}
	t.Fatal("no expression statement")
	return nil
}

func Text(f *syntax.File, n syntax.Node) string {
	s := n.Span()
	return string(f.Src[s.Start:s.End])
}
