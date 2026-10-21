package fastdigestusedforpasswordstorage

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FastDigestUsedForPasswordStorage"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"/** @custos-credential-store */ function store($v){} function f($p){store(md5(unknown($p)));}", 0},
		{"/** @custos-credential-store */ function store($v){} function f(){store(md5($_GET['p']));}", 0},
		{"/** @custos-credential-store */ function store($v){} store(md5($_GET['p']));", 0},
		{"/** @custos-credential-store */ function store($v){} function f($p){store(password_hash($p,PASSWORD_DEFAULT));}", 0},
		{"/** @custos-credential-store */ function store($v){} /** @param password-string $p */ function f($p){store(hash('sha256',$p));}", 1},
		{"/** @custos-credential-store */ function store($v){} store('literal');", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
