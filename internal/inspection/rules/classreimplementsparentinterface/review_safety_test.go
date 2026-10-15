package classreimplementsparentinterface

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	const decl = "<?php interface A {} interface B {} class P implements A, B {} "
	for _, tc := range []struct {
		name     string
		php      phpversion.Version
		source   string
		findings int
		fixed    string
	}{
		{"user parent future builtin before 8.0", phpversion.PHP74, "<?php class P implements \\Stringable {} class C extends P implements \\Stringable {}", 0, ""},
		{"user parent builtin from 8.0", phpversion.PHP80, "<?php class P implements \\Stringable {} class C extends P implements \\Stringable {}", 1, "<?php class P implements \\Stringable {} class C extends P {}"},
		{"user polyfill before 8.0", phpversion.PHP74, "<?php interface Stringable {} class P implements Stringable {} class C extends P implements Stringable {}", 1, "<?php interface Stringable {} class P implements Stringable {} class C extends P {}"},
		{"stringable before 8.0", phpversion.PHP74, "<?php class E extends \\Exception implements \\Stringable {}", 0, ""},
		{"stringable from 8.0", phpversion.PHP80, "<?php class E extends \\Exception implements \\Stringable {}", 1, "<?php class E extends \\Exception {}"},
		{"throwable before 8.0", phpversion.PHP74, "<?php class E extends \\Exception implements \\Throwable {}", 1, "<?php class E extends \\Exception {}"},
		{"user chain through builtin", phpversion.PHP74, "<?php class M extends \\Exception {} class E extends M implements \\Stringable, \\Throwable {}", 1, "<?php class M extends \\Exception {} class E extends M implements \\Stringable {}"},
		{"comment before implements", phpversion.PHP85, decl + "class C extends P /* k */ implements A {}", 1, ""},
		{"comment before comma", phpversion.PHP85, decl + "class C extends P implements A /* k */, \\Countable {}", 1, ""},
		{"comment after comma", phpversion.PHP85, decl + "class C extends P implements \\Countable, /* k */ A {}", 1, ""},
		{"comment between redundant entries", phpversion.PHP85, decl + "class C extends P implements A, /* k */ B {}", 2, ""},
		{"comment after first of three", phpversion.PHP85, decl + "class C extends P implements A /* k */, B, \\Countable {}", 2, decl + "class C extends P implements A /* k */,  \\Countable {}"},
		{"comment after the entry", phpversion.PHP85, decl + "class C extends P implements A/* k */{}", 1, decl + "class C extends P /* k */{}"},
		{"comment after the comma kept", phpversion.PHP85, decl + "class C extends P implements A, /* k */ \\Countable {}", 1, decl + "class C extends P implements  /* k */ \\Countable {}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte(tc.source)
			opt := syntax.Options{Version: tc.php}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
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
