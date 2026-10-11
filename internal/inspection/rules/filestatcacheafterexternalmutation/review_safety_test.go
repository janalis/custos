package filestatcacheafterexternalmutation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// exec and system clear the stat cache themselves from PHP 8.3 on, so the
// second read is only stale on older targets.
func TestReviewVersionThreshold(t *testing.T) {
	for _, tc := range []struct {
		name   string
		php    phpversion.Version
		source string
		want   int
	}{
		{"exec 5.3", phpversion.PHP53, `$a=filesize("f");exec("truncate -s 0 f");$b=filesize("f");`, 1},
		{"exec 8.2", phpversion.PHP82, `$a=filesize("f");exec("truncate -s 0 f");$b=filesize("f");`, 1},
		{"system 8.2", phpversion.PHP82, `$a=filemtime("f");system("touch f");$b=filemtime("f");`, 1},
		{"exec 8.3", phpversion.PHP83, `$a=filesize("f");exec("truncate -s 0 f");$b=filesize("f");`, 0},
		{"system 8.3", phpversion.PHP83, `$a=filemtime("f");system("touch f");$b=filemtime("f");`, 0},
		{"exec rm 8.5", phpversion.PHP85, `$a=filesize("f");exec("rm f");$b=filesize("f");`, 0},
		{"default target", 0, `$a=filesize("f");exec("truncate -s 0 f");$b=filesize("f");`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FileStatCacheAfterExternalMutation"}, PHP: tc.php})
			if err != nil {
				t.Fatal(err)
			}
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
