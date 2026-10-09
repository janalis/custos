package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReadSourceBounded(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "big.php")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	// sparse: 1 GB apparent size, nothing written
	if err := f.Truncate(1 << 30); err != nil {
		t.Skip("cannot create sparse file:", err)
	}
	f.Close()
	src, err := ReadSource(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(src) != syntax.MaxFileSize+1 {
		t.Fatalf("read %d bytes", len(src))
	}
	pf := syntax.ParseBest(p, src, syntax.Options{Version: phpversion.PHP84})
	if len(pf.Errors) != 1 || !strings.Contains(pf.Errors[0].Msg, "larger than") || len(pf.Stmts) != 0 {
		t.Fatalf("errors %v", pf.Errors)
	}
	e, err := analysis.NewEngine(nil, analysis.Config{})
	if err != nil {
		t.Fatal(err)
	}
	res := Run(e, []string{p}, syntax.Options{Version: phpversion.PHP84})
	if len(res) != 1 || len(res[0].Errors) != 1 {
		t.Fatalf("run: %+v", res[0].Errors)
	}
}
