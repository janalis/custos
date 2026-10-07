package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"custos/internal/analysis"
	"custos/internal/fix"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// FuzzRules runs every rule (and every fix) on fuzzed sources and fails on
// any rule panic (reported by the engine as an "internal" finding) or on a
// fix that cannot be applied. Seeded with the own fixtures.
func FuzzRules(f *testing.F) {
	seeds, _ := filepath.Glob("../../testdata/rules/*/*.php")
	for i, p := range seeds {
		if i%3 != 0 { // keep the seed corpus small
			continue
		}
		if b, err := os.ReadFile(p); err == nil && len(b) < 4096 {
			f.Add(string(b))
		}
	}
	f.Add("<?php class A { function f($x) { return $x?->y ?? throw new E(); } }")
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpver.PHP84, EnableAll: true})
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, src string) {
		for _, v := range []phpver.Version{phpver.PHP56, phpver.PHP84} {
			file := syntax.ParseBest("fuzz.php", []byte(src), syntax.Options{Version: v, Permissive: true})
			for _, fd := range e.Analyze(file) {
				if fd.Rule == "internal" && strings.Contains(fd.Message, "crashed") {
					t.Fatalf("PHP %s: %s", v, fd.Message)
				}
				for _, fx := range fd.Fixes {
					fix.Apply(file.Src, fx.Edits())
				}
			}
		}
	})
}

// stringSinks are calls whose string literal arguments rules parse (regex,
// printf/scanf formats, dates, callables, class names, ini names, JSON,
// doc tags, suppression comments, ...); FuzzStringLiterals substitutes the
// fuzzed text for S in each.
var stringSinks = []string{
	"preg_match('S', $s); preg_replace('S', 'S', $s); preg_split('S', $s); preg_quote($s, 'S');",
	"printf('S', 1); sprintf('S', 1, 2); sscanf($s, 'S'); fprintf($f, 'S');",
	"new DateInterval('S'); date('S'); $d->format('S'); DateTime::createFromFormat('S', $s);",
	"ini_get('S'); ini_set('S', '1'); json_decode('S'); unserialize('S');",
	"call_user_func('S'); call_user_func(['S', 'S']); is_callable('S'); array_map('S', $a);",
	"class_exists('S'); constant('S'); define('S', 1); $x = 'S'; new $x();",
	"str_replace('S', 'S', $s); strpos($s, 'S') === 0; substr($s, 0, 2) === 'S'; hash('S', $s);",
	"include 'S'; fopen('S', 'S'); compact('S'); version_compare($v, 'S', '>=');",
	"$a = ['S' => 1, 'S' => 2]; in_array('S', $a); mb_substr($s, 0, 1, 'S'); htmlspecialchars($s, ENT_QUOTES, 'S');",
	"$this->assertEquals('S', $s); assert('S'); uniqid('S'); curl_setopt($c, CURLOPT_URL, 'S');",
}

// FuzzStringLiterals runs every rule on code passing the fuzzed text as a
// single-quoted literal (and inside a double-quoted one, a doc comment and
// a suppression comment) to the string-consuming functions rules parse
// arguments of: no rule may panic, and analysis must stay fast.
func FuzzStringLiterals(f *testing.F) {
	for _, s := range []string{"/a+(b|c)*/i", "%1$s %d %%", "P1Y2M", "Y-m-d H:i", "A\\B::c", "{\"a\": [1]}", "\\u{1F600}", "[[[(", "@noinspection ALL"} {
		f.Add(s)
	}
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpver.PHP84, EnableAll: true})
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, s string) {
		single := strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s)
		double := strings.NewReplacer(`\`, `\\`, `"`, `\"`, `$`, `\$`).Replace(s)
		comment := strings.ReplaceAll(s, "*/", "")
		var b strings.Builder
		b.WriteString("<?php\nclass C extends T {\n/** @param S $s\n * @return S\n * @noinspection S\n */\nfunction m($s, $a, $d, $f, $c, $v) {\n")
		b.WriteString("// @custos-ignore S\n")
		for _, sink := range stringSinks {
			b.WriteString(strings.ReplaceAll(sink, "S", single) + "\n")
		}
		b.WriteString(`$q = "` + double + `"; preg_match("` + double + `", $s); printf("` + double + "\", 1);\n}\n}\n")
		src := strings.ReplaceAll(b.String(), "@param S", "@param "+comment)
		src = strings.ReplaceAll(src, "@return S", "@return "+comment)
		src = strings.ReplaceAll(src, "@noinspection S", "@noinspection "+comment)
		src = strings.ReplaceAll(src, "@custos-ignore S", "@custos-ignore "+strings.ReplaceAll(comment, "\n", " "))
		start := time.Now()
		file := syntax.ParseBest("fuzz.php", []byte(src), syntax.Options{Version: phpver.PHP84, Permissive: true})
		for _, fd := range e.Analyze(file) {
			if fd.Rule == "internal" && strings.Contains(fd.Message, "crashed") {
				t.Fatalf("%q: %s", s, fd.Message)
			}
		}
		if d := time.Since(start); d > 10*time.Second {
			t.Fatalf("%q: analysis took %v", s, d)
		}
	})
}
