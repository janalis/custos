package csvescapedefaultdependency

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestReviewSafety pins when the backslash default may be written out for an
// SplFileObject call: only when no other code can have configured the escape.
func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"CsvEscapeDefaultDependency"}})
	if err != nil {
		t.Fatal(err)
	}
	const create = "$o = new SplFileObject('x'); "
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"control escape with separator override", create + `$o->setCsvControl(',', '"', ''); $o->fgetcsv(';');`, 0, ""},
		{"control escape with enclosure override", create + `$o->setCsvControl(escape: ''); $o->fputcsv([], ',', "'");`, 0, ""},
		{"later control resets escape", create + `$o->setCsvControl(',', '"', ''); $o->setCsvControl(';', escape: "\\"); $o->fgetcsv();`, 0, ""},
		{"later control drops escape", create + `$o->setCsvControl(',', '"', ''); $o->setCsvControl(';'); $o->fgetcsv();`, 2, create + `$o->setCsvControl(',', '"', ''); $o->setCsvControl(';', escape: "\\"); $o->fgetcsv();`},
		{"control without escape keeps default", create + `$o->setCsvControl(';'); $o->rewind(); $o->fgetcsv();`, 2, create + `$o->setCsvControl(';', escape: "\\"); $o->rewind(); $o->fgetcsv(escape: "\\");`},
		{"local default", create + `$o->rewind(); $o->fgetcsv(';');`, 1, create + `$o->rewind(); $o->fgetcsv(';', escape: "\\");`},
		{"temp file object", `$o = new \SplTempFileObject(); $o?->fputcsv([]);`, 1, `$o = new \SplTempFileObject(); $o?->fputcsv([], escape: "\\");`},
		{"inside function", `function f() { $o = new SplFileObject('x'); return $o->fgetcsv(); }`, 1, `function f() { $o = new SplFileObject('x'); return $o->fgetcsv(escape: "\\"); }`},
		{"parameter", `function f(SplFileObject $o) { return $o->fgetcsv(); }`, 1, ""},
		{"this", `class C extends SplFileObject { function f() { return $this->fputcsv([]); } }`, 1, ""},
		{"property", `class C { public SplFileObject $o; function f() { return $this->o->fgetcsv(); } }`, 1, ""},
		{"subclass", `class C extends SplFileObject {} $o = new C('x'); $o->fgetcsv();`, 1, ""},
		{"dynamic class", `function f(SplFileObject $p, $c) { $o = $p; $o = new $c('x'); $o->fgetcsv(); }`, 0, ""},
		{"passed to call", create + `configure($o); $o->fgetcsv();`, 1, ""},
		{"aliased", create + `$p = $o; $o->fgetcsv();`, 1, ""},
		{"captured", create + `$g = function () use ($o) {}; $o->fgetcsv();`, 1, ""},
		{"conditional escape", create + `if ($c) { $o->setCsvControl(',', '"', ''); } $o->fgetcsv();`, 1, ""},
		{"later escape in loop", `function f() { $o = new SplFileObject('x'); while (g()) { $o->fgetcsv(); $o->setCsvControl(escape: ''); } }`, 1, ""},
		{"spread control", create + `$o->setCsvControl(...$a); $o->fgetcsv();`, 2, ""},
		{"dynamic method", create + `$o->$m(); $o->fgetcsv();`, 1, ""},
		{"reassigned", create + `$o = new SplFileObject('y'); $o->fgetcsv();`, 1, ""},
		{"conditional creation", `function f(SplFileObject $o, $c) { if ($c) { $o = new SplFileObject('x'); } return $o->fgetcsv(); }`, 1, ""},
		{"by reference", `function f(SplFileObject $p) { $o = &$p; $o->fgetcsv(); }`, 1, ""},
		{"coalescing assignment", `function f(?SplFileObject $o) { $o ??= new SplFileObject('x'); $o->fgetcsv(); }`, 1, ""},
		{"variable variable", create + `$$n->setCsvControl(',', '"', ''); $o->fgetcsv();`, 1, ""},
		{"include", create + `include 'configure.php'; $o->fgetcsv();`, 0, ""},
		{"include in function", `function f() { $o = new SplFileObject('x'); include 'configure.php'; $o->fgetcsv(); }`, 0, ""},
		{"include before", `function f() { include 'configure.php'; $o = new SplFileObject('x'); $o->fgetcsv(); }`, 1, ""},
		{"eval before", `function f() { eval($code); $o = new SplFileObject('x'); $o->fgetcsv(); }`, 1, ""},
		{"eval", create + `eval($code); $o->fgetcsv();`, 1, ""},
		{"spread arguments", create + `$o->fputcsv(...$a);`, 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte("<?php " + tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
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
			if string(res.Source) != "<?php "+want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
