package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func sample(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestExecRunner(t *testing.T) {
	r, err := execRunner(command{dir: t.TempDir(), env: []string{"FORMATCHECK_X=x"}, name: "sh", args: []string{"-c", "echo $FORMATCHECK_X; echo e >&2; exit 3"}})
	if err != nil || string(r.stdout) != "x\n" || string(r.stderr) != "e\n" || r.code != 3 {
		t.Errorf("%+v %v", r, err)
	}
	if _, err := execRunner(command{name: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Error("no error for a missing program")
	}
}

func TestToolError(t *testing.T) {
	err := toolError("/x/phpcs", result{code: 2, stderr: []byte(strings.Repeat("a", 2001))}, errors.New("bad"))
	if !strings.HasPrefix(err.Error(), "phpcs (exit status 2): bad: aaa") || !strings.HasSuffix(err.Error(), "a…") {
		t.Errorf("%v", err)
	}
}

// Parsers read outputs captured from the pinned tools (testdata/*.json).
func TestParsePHPCS(t *testing.T) {
	tests := []struct {
		file  string
		count int
		first finding
	}{
		{"phpcs.json", 13, finding{"PSR12.Files.FileHeader.SpacingAfterTagBlock", "src/Sample.php", 1, "Header blocks must be separated by a single blank line"}},
		{"drupal.json", 37, finding{"Drupal.WhiteSpace.OpenTagNewline.BlankLine", "src/Sample.php", 1, "The PHP open tag must be followed by exactly one blank line"}},
		{"phpcs-clean.json", 0, finding{}},
	}
	for _, tt := range tests {
		fs, err := parsePHPCS(sample(t, tt.file), "/work/copy")
		if err != nil || len(fs) != tt.count || (tt.count > 0 && fs[0] != tt.first) {
			t.Errorf("%s: %d findings, %+v, %v", tt.file, len(fs), fs, err)
		}
	}
	for _, bad := range []string{"x", `{"totals":{}}`} {
		if _, err := parsePHPCS([]byte(bad), "/"); err == nil {
			t.Errorf("%s: no error", bad)
		}
	}
}

func TestParseFixer(t *testing.T) {
	fs, err := parseFixer(sample(t, "csfixer.json"))
	if err != nil || len(fs) != 15 {
		t.Fatalf("%d findings, %v", len(fs), err)
	}
	if want := (finding{"array_syntax", "src/Sample.php", 2, `first change: "" → ""`}); fs[0] != want {
		t.Errorf("%+v", fs[0])
	}
	if fs, err := parseFixer(sample(t, "csfixer-clean.json")); err != nil || len(fs) != 0 {
		t.Errorf("%+v %v", fs, err)
	}
	for _, bad := range []string{"x", `{"about":"x"}`} {
		if _, err := parseFixer([]byte(bad)); err == nil {
			t.Errorf("%s: no error", bad)
		}
	}
}

func TestFirstChange(t *testing.T) {
	tests := []struct {
		diff      string
		line      int
		old, repl string
	}{
		{"--- a\n+++ b\n@@ -3,3 +3,3 @@\n a\n-b\n+c\n d\n", 4, "b", "c"},
		{"@@ -1 +1 @@\n-  removed\n context\n", 1, "  removed", ""},
		{"@@ -7,2 +7,3 @@\n x\n+ added\n", 8, "", " added"},
		{"@@ -1,2 +1,2 @@\n--- double\n+++ double\n", 1, "-- double", "++ double"},
		{"no hunk\n-x\n", 0, "", ""},
		{"@@ -1 +1 @@\n x\n", 0, "", ""},
	}
	for _, tt := range tests {
		line, old, repl := firstChange(tt.diff)
		if line != tt.line || old != tt.old || repl != tt.repl {
			t.Errorf("%q: %d %q %q", tt.diff, line, old, repl)
		}
	}
}

func TestParsePint(t *testing.T) {
	tests := []struct {
		file   string
		files  int
		fixers int
	}{
		{"pint-test.json", 1, 15},
		{"pint-fix.json", 1, 15},
		{"pint-clean.json", 0, 0},
	}
	for _, tt := range tests {
		files, err := parsePint(sample(t, tt.file))
		if err != nil || len(files) != tt.files || (tt.files > 0 && len(files[0].Fixers) != tt.fixers) {
			t.Errorf("%s: %+v %v", tt.file, files, err)
		}
	}
	for _, bad := range []string{"x", `{"tool":"pint"}`} {
		if _, err := parsePint([]byte(bad)); err == nil {
			t.Errorf("%s: no error", bad)
		}
	}
	// Outside an AI agent, Pint prints PHP-CS-Fixer's report.
	out := "Deprecated: x\n" + `{"about":"PHP CS Fixer","files":[{"name":"src\/a.php","appliedFixers":["braces_position","concat_space"],"diff":"-"}],"time":{"total":0}}`
	if files, err := parsePint([]byte(out)); err != nil || len(files) != 1 || files[0].Path != "src/a.php" || len(files[0].Fixers) != 2 {
		t.Errorf("%+v %v", files, err)
	}
}

func TestCodeChange(t *testing.T) {
	tests := []struct {
		name, before, after string
		line                int
		old, repl           string
		changed             bool
	}{
		{"same", "<?php\n$a;\n", "<?php\n$a;\n", 0, "", "", false},
		{"docblock", "<?php\n/**\n * x\n */\n$a;\n", "<?php\n/**\n * y\n *\n */\n$a;\n", 0, "", "", false},
		{"one-line docblocks", "/** a */\n/**/\n$a;\n", "/** b */\n/**/\n$a;\n", 0, "", "", false},
		{"changed", "<?php\n/** x */\n  $a;\n", "<?php\n/** x */\n  $b;\n", 3, "$a;", "$b;", true},
		{"removed", "<?php\n$a;\n$b;", "<?php\n$a;", 3, "$b;", "", true},
		{"added", "<?php\n$a;", "<?php\n$a;\n$b;", 2, "", "$b;", true},
		{"added after docblock", "/**", "/**\n */\n$b;", 1, "", "$b;", true},
	}
	for _, tt := range tests {
		line, old, repl, changed := codeChange([]byte(tt.before), []byte(tt.after))
		if line != tt.line || old != tt.old || repl != tt.repl || changed != tt.changed {
			t.Errorf("%s: %d %q %q %v", tt.name, line, old, repl, changed)
		}
	}
}

// failing is a session whose runner returns r and err for every command.
func failing(t *testing.T, r result, err error) *session {
	t.Helper()
	return &session{
		run:   func(command) (result, error) { return r, err },
		bench: t.TempDir(), tools: t.TempDir(), log: &strings.Builder{}, cpus: 1,
	}
}

func TestCheckerFailures(t *testing.T) {
	boom := errors.New("boom")
	root := t.TempDir()
	paths := []string{"src"}
	for preset, ck := range checkers {
		if _, err := ck.check(failing(t, result{}, boom), root, paths); !errors.Is(err, boom) {
			t.Errorf("%s: run error %v", preset, err)
		}
		if _, err := ck.check(failing(t, result{stdout: []byte("x"), stderr: []byte("trace")}, nil), root, paths); err == nil || !strings.Contains(err.Error(), "trace") {
			t.Errorf("%s: unreadable output gives %v", preset, err)
		}
	}
	if _, err := failing(t, result{code: 4, stderr: []byte("parse error")}, nil).fixer("@PER-CS", root, paths); err == nil || !strings.Contains(err.Error(), "check failed: parse error") {
		t.Errorf("failed check gives %v", err)
	}
	// A file erroring inside PHP-CS-Fixer only warns: the report holds the rest.
	s := failing(t, result{code: 72, stdout: []byte(`{"files": []}`)}, nil)
	if fs, err := s.fixer("@PER-CS", root, paths); err != nil || len(fs) != 0 || !strings.Contains(s.log.(*strings.Builder).String(), "errored inside PHP-CS-Fixer") {
		t.Errorf("exit 72 gives %v %v", fs, err)
	}
}

func TestPintGate(t *testing.T) {
	setup := func(t *testing.T) string {
		root := filepath.Join(t.TempDir(), "laravel")
		write(t, filepath.Join(root, "src", "A.php"), "<?php\n$a;\n")
		return root
	}
	div := []divergence{{Rule: "yoda_style", Reason: "r"}, {Rule: "phpdoc_*", Reason: "r"}}
	// gate returns a session whose PHP-CS-Fixer rule list is known.
	gate := func(t *testing.T, r result, err error) *session {
		s := failing(t, r, err)
		s.fixerRules = []string{}
		return s
	}
	t.Run("config", func(t *testing.T) {
		root := setup(t)
		var config string
		s := failing(t, result{}, nil)
		s.run = func(c command) (result, error) {
			if c.args[0] == "list-rules" {
				return result{stdout: sample(t, "csfixer-rules.json")}, nil
			}
			data, err := os.ReadFile(c.args[slices.Index(c.args, "--config")+1])
			config = string(data)
			return result{stdout: []byte("Deprecated: noise\n" + `{"tool":"pint","result":"passed"}`)}, err
		}
		fs, err := s.pintGate(root, []string{"src"}, div)
		// Patterns are turned off as the PHP-CS-Fixer rules they match.
		if err != nil || len(fs) != 0 || !strings.Contains(config, `"yoda_style":false`) || !strings.Contains(config, `"phpdoc_trim":false`) {
			t.Errorf("%+v %v %s", fs, err, config)
		}
	})
	t.Run("rule list", func(t *testing.T) {
		if _, err := failing(t, result{}, errors.New("boom")).pintGate(setup(t), []string{"src"}, nil); err == nil {
			t.Error("no error")
		}
	})
	t.Run("copy", func(t *testing.T) {
		if _, err := failing(t, result{}, nil).pintGate(filepath.Join(t.TempDir(), "none"), []string{"src"}, nil); err == nil {
			t.Error("no error")
		}
	})
	t.Run("config write", func(t *testing.T) {
		root := setup(t)
		if err := os.MkdirAll(root+"-pint.pint.json", 0o755); err != nil {
			t.Fatal(err)
		}
		s := failing(t, result{}, nil)
		s.fixerRules = []string{}
		if _, err := s.pintGate(root, []string{"src"}, nil); err == nil {
			t.Error("no error")
		}
	})
	t.Run("pint", func(t *testing.T) {
		s := failing(t, result{}, errors.New("boom"))
		s.fixerRules = []string{}
		if _, err := s.pintGate(setup(t), []string{"src"}, nil); err == nil {
			t.Error("no error")
		}
	})
	t.Run("missing before", func(t *testing.T) {
		out := result{stdout: []byte(`{"result":"fixed","files":[{"path":"src/none.php","fixers":["x"]}]}`)}
		if _, err := gate(t, out, nil).pintGate(setup(t), []string{"src"}, nil); err == nil {
			t.Error("no error")
		}
	})
	t.Run("missing after", func(t *testing.T) {
		root := setup(t)
		s := gate(t, result{}, nil)
		s.run = func(c command) (result, error) {
			return result{stdout: []byte(`{"result":"fixed","files":[{"path":"src/A.php","fixers":["x"]}]}`)},
				os.Remove(filepath.Join(c.dir, "src", "A.php"))
		}
		if _, err := s.pintGate(root, []string{"src"}, nil); err == nil {
			t.Error("no error")
		}
	})
}

func TestReady(t *testing.T) {
	s := failing(t, result{}, nil)
	for preset, ck := range checkers {
		if ck.install != nil {
			continue
		}
		if err := ck.ready(s); err == nil || !strings.Contains(err.Error(), "benchcmp install") {
			t.Errorf("%s: %v", preset, err)
		}
	}
}

func TestInstallCoder(t *testing.T) {
	t.Run("installs, then keeps", func(t *testing.T) {
		f := &fake{t: t}
		s := failing(t, result{}, nil)
		s.run = f.run
		for range 2 {
			if err := checkers["drupal"].ready(s); err != nil {
				t.Fatal(err)
			}
		}
		if len(f.cmds) != 1 || !slices.Contains(f.cmds[0].args, "drupal/coder:"+coderVersion) {
			t.Errorf("%+v", f.cmds)
		}
		if !strings.Contains(s.log.(*strings.Builder).String(), "installing drupal/coder") {
			t.Error("no log")
		}
	})
	t.Run("composer error", func(t *testing.T) {
		if err := failing(t, result{}, errors.New("boom")).installCoder(); err == nil {
			t.Error("no error")
		}
	})
	t.Run("composer status", func(t *testing.T) {
		if err := failing(t, result{code: 1}, nil).installCoder(); err == nil || !strings.Contains(err.Error(), "installing drupal/coder failed") {
			t.Errorf("%v", err)
		}
	})
	t.Run("directory", func(t *testing.T) {
		s := failing(t, result{}, nil)
		s.tools = filepath.Join(s.tools, "file")
		write(t, s.tools, "")
		if err := checkers["drupal"].ready(s); err == nil {
			t.Error("no error")
		}
	})
}

func TestFixerGate(t *testing.T) {
	var rules []string
	s := failing(t, result{}, nil)
	s.run = func(c command) (result, error) {
		if c.args[0] == "list-rules" {
			return result{stdout: sample(t, "csfixer-rules.json")}, nil
		}
		rules = append(rules, strings.TrimPrefix(c.args[slices.IndexFunc(c.args, func(a string) bool { return strings.HasPrefix(a, "--rules=") })], "--rules="))
		return result{stdout: sample(t, "csfixer-clean.json")}, nil
	}
	div := []divergence{{Rule: "phpdoc_[at]*", Reason: "r"}, {Rule: "yoda_style", Reason: "r"}, {Rule: "nothing_*", Reason: "r"}}
	for range 2 {
		if fs, err := fixerGate("@Symfony")(s, t.TempDir(), []string{"src"}, div); err != nil || len(fs) != 0 {
			t.Fatalf("%+v %v", fs, err)
		}
	}
	want := `{"@Symfony":true,"phpdoc_align":false,"phpdoc_trim":false,"yoda_style":false}`
	if len(rules) != 2 || rules[0] != want {
		t.Errorf("rules %q, want %s", rules, want)
	}
	if len(s.fixerRules) != 5 {
		t.Errorf("rules listed %q", s.fixerRules)
	}
	// The rule list is read once.
	s.run = func(command) (result, error) { return result{}, errors.New("listed again") }
	if _, err := s.listFixerRules(); err != nil {
		t.Error(err)
	}
}

func TestListFixerRulesErrors(t *testing.T) {
	boom := errors.New("boom")
	if _, err := fixerGate("@PER-CS")(failing(t, result{}, boom), t.TempDir(), nil, nil); !errors.Is(err, boom) {
		t.Errorf("run error gives %v", err)
	}
	for _, out := range []string{"x", `{"rules":{}}`} {
		if _, err := failing(t, result{stdout: []byte(out)}, nil).listFixerRules(); err == nil || !strings.Contains(err.Error(), "no rule list") {
			t.Errorf("%s: %v", out, err)
		}
	}
}
