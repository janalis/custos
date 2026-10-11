package mkdirracecondition

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"named temp var", "<?php mkdir(directory: dirname($f), recursive: true) || log();", 1, ""},
		{"named directory last", "<?php if (!mkdir(permissions: $m, directory: $d)) {}", 1, "<?php if (!mkdir(permissions: $m, directory: $d) && !is_dir($d)) {}"},
		{"named directory last, temp var", "<?php if (!mkdir(recursive: true, directory: dir())) {}", 1, "<?php if (!mkdir(recursive: true, directory: $concurrentDirectory = dir()) && !is_dir($concurrentDirectory)) {}"},
		{"named re-check", "<?php if (!mkdir(permissions: $m, directory: $d) && !is_dir($d)) {}", 0, ""},
		{"spread", "<?php mkdir(...$args);", 0, ""},
		{"spread after directory", "<?php if (!mkdir($d, ...$args)) {}", 0, ""},
		{"no directory", "<?php if (!mkdir(permissions: $m)) {}", 0, ""},
		{"loose false", "<?php if (mkdir($d) == false) {}", 1, "<?php if (!mkdir($d) && !is_dir($d)) {}"},
		{"loose yoda false", "<?php if (false == mkdir($d)) {}", 1, "<?php if (!mkdir($d) && !is_dir($d)) {}"},
		{"loose not true", "<?php if (mkdir($d) <> true) {}", 1, "<?php if (!mkdir($d) && !is_dir($d)) {}"},
		{"loose not false", "<?php if (mkdir($d) != false) {}", 1, "<?php if (mkdir($d) || is_dir($d)) {}"},
		{"xor", "<?php if (mkdir($d) xor $dry) {}", 0, ""},
		{"identical null", "<?php if (mkdir($d) === null) {}", 0, ""},
		{"bitwise and", "<?php if (!mkdir($d) & !mkdir($e)) {}", 0, ""},
		{"interpolated call", `<?php if (!mkdir("{$b}/{$s->next()}")) {}`, 1, `<?php if (!mkdir($concurrentDirectory = "{$b}/{$s->next()}") && !is_dir($concurrentDirectory)) {}`},
		{"interpolated property", `<?php if (!mkdir("$o->p/a")) {}`, 1, `<?php if (!mkdir($concurrentDirectory = "$o->p/a") && !is_dir($concurrentDirectory)) {}`},
		{"interpolated variables", `<?php if (!mkdir("{$b}/$n")) {}`, 1, `<?php if (!mkdir("{$b}/$n") && !is_dir("{$b}/$n")) {}`},
		{"heredoc", "<?php if (!mkdir(<<<P\n$d/sub\nP)) {}", 1, "<?php if (!mkdir($concurrentDirectory = <<<P\n$d/sub\nP) && !is_dir($concurrentDirectory)) {}"},
		{"nowdoc", "<?php if (!mkdir(<<<'P'\n/tmp/sub\nP)) {}", 1, "<?php if (!mkdir($concurrentDirectory = <<<'P'\n/tmp/sub\nP) && !is_dir($concurrentDirectory)) {}"},
		{"close tag", "<b><?php mkdir($d) ?></b>", 1, `<b><?php if (!mkdir($d) && !is_dir($d)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $d)); } ?></b>`},
		{"close tag after semicolon", "<b><?php mkdir($d); ?></b>", 1, `<b><?php if (!mkdir($d) && !is_dir($d)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $d)); } ?></b>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
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
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
