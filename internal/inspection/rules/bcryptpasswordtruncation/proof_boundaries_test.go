package bcryptpasswordtruncation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "BcryptPasswordTruncation" || len(r.Kinds()) == 0 {
		t.Fatal("invalid rule contract")
	}
	if s, ok := r.(interface{ Semantic() }); ok {
		s.Semantic()
	}
	if f, ok := r.(interface{ Flow() }); ok {
		f.Flow()
	}
}

func TestUnavailableVersion(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: 503, Only: []string{"BcryptPasswordTruncation"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte("<?php $h=password_hash(str_repeat(\"p\",80),PASSWORD_BCRYPT);"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("unavailable API reported: %+v", got)
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"strlen(\"x\");", 0}, {"password_hash($x,PASSWORD_BCRYPT);", 0}, {"password_hash(unknown(),PASSWORD_BCRYPT);", 0}, {"password_hash(random_bytes($unknown),PASSWORD_BCRYPT);", 0}, {"password_hash(str_repeat($x,80),PASSWORD_BCRYPT);", 0}, {"password_hash(str_repeat(\"x\",-1),PASSWORD_BCRYPT);", 0}, {"password_hash(\"x\",$algorithm);", 0}, {"password_hash(\"x\",PASSWORD_BCRYPT);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"BcryptPasswordTruncation"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
