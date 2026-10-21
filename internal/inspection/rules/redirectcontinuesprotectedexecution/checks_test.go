package redirectcontinuesprotectedexecution

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RedirectContinuesProtectedExecution"}})
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
		{"/** @custos-protected */ function protectedCall(){} function f(){if($auth){header('Location: /');}if(!isAuthorized()){header('Location: /');}protectedCall();}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

// A protected function declared in a file not analyzed has no body to read.
func TestForeignProtectedFunction(t *testing.T) {
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(syntax.Parse("external.php", []byte("<?php /** @custos-protected */ function externalProtected(){}"), syntax.Options{})))
	if got := engine(t).WithIndex(ix).Analyze(syntax.Parse("consumer.php", []byte("<?php externalProtected();"), syntax.Options{})); len(got) != 0 {
		t.Fatalf("foreign source unavailable: %+v", got)
	}
}
