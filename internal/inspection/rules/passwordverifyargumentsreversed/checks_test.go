package passwordverifyargumentsreversed

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PasswordVerifyArgumentsReversed"}})
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
		{"\n$hash = password_hash($password, PASSWORD_DEFAULT); password_verify($hash, $password);\n", 1},
		{"\n$hash = password_hash($password, PASSWORD_DEFAULT); password_verify($password, $hash);\n", 0},
		{"$p='secret';$h=password_hash($p,PASSWORD_DEFAULT);$p='new';password_verify($h,$p);", 0},
		{"function f($p){$h=password_hash($p,PASSWORD_DEFAULT);unknown($p);password_verify($h,$p);}", 0},
		{"password_verify('clear','hash');", 0},
		{"if(false){\n$hash = password_hash($password, PASSWORD_DEFAULT); password_verify($hash, $password);\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
