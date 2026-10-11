package regexreplacementcallbackfallsthrough

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	const decls = `<?php
function fail(string $m): never { throw new InvalidArgumentException($m); }
function note(string $m): void { echo $m; }
class T {
    public static function stop(): never { exit(1); }
    public function halt(): never { exit(1); }
    public function log(): void {}
    public function run($s, T $t) {
        %s
    }
}
`
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"never function after guarded return", "if ($m[0] !== 'q') { return 'x'; } fail('u');", 0},
		{"never function in else", "if ($m[0] !== 'q') { return 'x'; } else { fail('u'); }", 0},
		{"never in every branch", "if ($m[0] !== 'q') { return 'x'; } elseif ($m[0] === 'z') { fail('u'); } else { (fail('u')); }", 0},
		{"elseif branch completes", "if ($m[0] !== 'q') { return 'x'; } elseif ($m[0] === 'z') { note('u'); } else { fail('u'); }", 1},
		{"if branch completes", "if ($m[0] !== 'q') { note('x'); } else { fail('u'); }", 1},
		{"else branch completes", "if ($m[0] !== 'q') { fail('x'); } else { note('u'); }", 1},
		{"no else", "if ($m[0] !== 'q') { fail('x'); }", 1},
		{"literal true never", "if (true) { fail('u'); }", 0},
		{"literal false never", "if (false) { fail('u'); }", 1},
		{"literal false else never", "if (false) { note('u'); } else { fail('u'); }", 0},
		{"literal false else completes", "if (false) { fail('u'); } else { note('u'); }", 1},
		{"void function", "note('u');", 1},
		{"static never", "self::stop();", 0},
		{"this never", "$this->halt();", 0},
		{"typed receiver never", "$t->halt();", 0},
		{"typed receiver void", "$t->log();", 1},
		{"this void", "$this->log();", 1},
		{"unresolved function", "unknown_fn('u');", 0},
		{"dynamic function", "$m[0]();", 0},
		{"unknown receiver", "$o->anything();", 0},
		{"plain expression", "$x = strtoupper($m[0]);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			call := "preg_replace_callback('~a~', function ($m) use ($t, $o) { " + tc.body + " }, $s);"
			src := []byte(strings.Replace(decls, "%s", call, 1))
			f := syntax.ParseBest("review.php", src, syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
