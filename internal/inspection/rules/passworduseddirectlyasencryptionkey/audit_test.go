package passworduseddirectlyasencryptionkey

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestAuditRegressions exercises independent valid programs and comment-preservation boundaries.
func TestAuditRegressions(t *testing.T) {
	for _, tc := range []struct {
		src         string
		want, fixes int
	}{
		{"$_POST[\"password\"]=\"derived\";openssl_encrypt(\"x\",\"aes-128-cbc\",$_POST[\"password\"],0,$iv);", 0, -1},
		{"sanitizeRequest();openssl_encrypt(\"x\",\"aes-128-cbc\",$_POST[\"password\"],0,$iv);", 0, -1},
		{"$o->sanitize($_POST);openssl_encrypt(\"x\",\"aes-128-cbc\",$_POST[\"password\"],0,$iv);", 0, -1},
		{"C::sanitize($_POST);openssl_encrypt(\"x\",\"aes-128-cbc\",$_POST[\"password\"],0,$iv);", 0, -1},
		{"$GLOBALS['_POST']=['password'=>'derived'];openssl_encrypt(\"x\",\"aes-128-cbc\",$_POST[\"password\"],0,$iv);", 0, -1},
		{"$alias=&$GLOBALS['_POST'];openssl_encrypt(\"x\",\"aes-128-cbc\",$_POST[\"password\"],0,$iv);", 0, -1},
		{"new Sanitizer();openssl_encrypt(\"x\",\"aes-128-cbc\",$_POST[\"password\"],0,$iv);", 0, -1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("bad regression source: %+v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings want %d: %+v", len(got), tc.want, got)
			}
			if tc.fixes >= 0 {
				count := 0
				for _, d := range got {
					count += len(d.Fixes)
				}
				if count != tc.fixes {
					t.Fatalf("got %d fixes want %d", count, tc.fixes)
				}
			}
		})
	}
}

func BenchmarkRequestSourceProof(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte(`<?php openssl_encrypt("x","aes-128-cbc",$_POST["password"],0,$iv);`), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(f)
	}
}
