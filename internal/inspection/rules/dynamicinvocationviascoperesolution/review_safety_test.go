package dynamicinvocationviascoperesolution

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestReviewSafety(t *testing.T) {
	New().(analysis.SemanticRule).Semantic()
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{
			"static private redeclared below",
			"<?php class A { private function m() {} function run() { static::m(); } } class B extends A { public function m() {} }",
			1, "",
		},
		{
			"static private not redeclared",
			"<?php class A { private function m() {} function run() { static::m(); } } class B extends A {}",
			1, "<?php class A { private function m() {} function run() { $this->m(); } } class B extends A {}",
		},
		{
			"static public redeclared below",
			"<?php class A { public function m() {} function run() { static::m(); } } class B extends A { public function m() {} }",
			1, "<?php class A { public function m() {} function run() { $this->m(); } } class B extends A { public function m() {} }",
		},
		{
			"other receiver of the same class",
			"<?php class N { function id() {} function d(N $o) { return $o::id(); } }",
			1, "",
		},
		{
			"other receiver in a nested closure",
			"<?php class N { function id() {} function d(N $o) { return fn () => $o::id(); } }",
			1, "",
		},
		{
			"other receiver from an ancestor",
			"<?php class P { function d(N $o) { return $o::id(); } } class N extends P { function id() {} }",
			1, "",
		},
		{
			"other receiver in a static method",
			"<?php class N { function id() {} static function d(N $o) { return $o::id(); } }",
			1, "<?php class N { function id() {} static function d(N $o) { return $o->id(); } }",
		},
		{
			"other receiver in an unrelated class",
			"<?php class N { function id() {} } class U { function d(N $o) { return $o::id(); } }",
			1, "<?php class N { function id() {} } class U { function d(N $o) { return $o->id(); } }",
		},
		{
			"this receiver",
			"<?php class N { function id() {} function d() { return $this::id(); } }",
			1, "<?php class N { function id() {} function d() { return $this->id(); } }",
		},
		{
			"enum case receiver",
			"<?php enum S { case A; function label() {} function t() { return self::A::label(); } }",
			1, "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			ix := index.New(nil)
			ix.Add(index.Extract(f))
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			e = e.WithIndex(ix)
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
