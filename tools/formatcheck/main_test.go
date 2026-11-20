package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fake stands in for the tools: a formatted copy holds a .formatted marker,
// and each checker prints a captured sample, the samples' /work/copy root
// replaced by the directory it runs in.
type fake struct {
	t    *testing.T
	cmds []command
	// fail, when it returns a non-nil result or error, replaces a command.
	fail func(command) (result, error, bool)
	// pintConfig is the configuration of Pint's last fix.
	pintConfig []byte
}

func (f *fake) run(c command) (result, error) {
	f.cmds = append(f.cmds, c)
	if f.fail != nil {
		if r, err, ok := f.fail(c); ok {
			return r, err
		}
	}
	_, statErr := os.Stat(filepath.Join(c.dir, ".formatted"))
	formatted := statErr == nil
	sample := func(name string) result {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			f.t.Fatal(err)
		}
		root := strings.ReplaceAll(c.dir, "/", `\/`)
		return result{stdout: []byte(strings.ReplaceAll(string(data), `\/work\/copy`, root))}
	}
	switch filepath.Base(c.name) {
	case "custos":
		return result{}, os.WriteFile(filepath.Join(c.dir, ".formatted"), nil, 0o644)
	case "composer":
		bin := filepath.Join(c.dir, "vendor", "bin", "phpcs")
		return result{}, errors.Join(os.MkdirAll(filepath.Dir(bin), 0o755), os.WriteFile(bin, nil, 0o755))
	case "phpcs":
		switch {
		case formatted:
			return sample("phpcs-clean.json"), nil
		case slices.Contains(c.args, "--standard=Drupal"):
			return sample("drupal.json"), nil
		}
		return sample("phpcs.json"), nil
	case "php-cs-fixer":
		if c.args[0] == "list-rules" {
			return sample("csfixer-rules.json"), nil
		}
		if formatted && slices.ContainsFunc(c.args, func(a string) bool { return strings.Contains(a, "@PER-CS") }) {
			return sample("csfixer-clean.json"), nil
		}
		r := sample("csfixer.json")
		r.code = 8
		// A rule set given as JSON: its disabled fixers are not reported.
		if rules, ok := strings.CutPrefix(c.args[slices.IndexFunc(c.args, func(a string) bool { return strings.HasPrefix(a, "--rules=") })], "--rules={"); ok {
			var off map[string]bool
			var report map[string]any
			if err := errors.Join(json.Unmarshal([]byte("{"+rules), &off), json.Unmarshal(r.stdout, &report)); err != nil {
				f.t.Fatal(err)
			}
			file := report["files"].([]any)[0].(map[string]any)
			var kept []any
			for _, fixer := range file["appliedFixers"].([]any) {
				if on, listed := off[fixer.(string)]; on || !listed {
					kept = append(kept, fixer)
				}
			}
			file["appliedFixers"] = kept
			r.stdout, _ = json.Marshal(report)
		}
		return r, nil
	case "pint":
		if slices.Contains(c.args, "--test") {
			if formatted {
				return sample("pint-clean.json"), nil
			}
			return sample("pint-test.json"), nil
		}
		// The fix: change the code of one file and a docblock of another.
		f.pintConfig, _ = os.ReadFile(c.args[slices.Index(c.args, "--config")+1])
		code := filepath.Join(c.dir, "src", "A.php")
		doc := filepath.Join(c.dir, "src", "B.php")
		err := errors.Join(
			os.WriteFile(code, []byte("<?php\n$a = 2;\n"), 0o644),
			os.WriteFile(doc, []byte("<?php\n/**\n *\n */\n$b = 1;\n"), 0o644))
		return result{stdout: []byte(`{"tool":"pint","result":"fixed","files":[` +
			`{"path":"src\/A.php","fixers":["z_fixer","a_fixer"]},{"path":"src\/B.php","fixers":["phpdoc_trim"]}]}`)}, err
	}
	f.t.Fatalf("unexpected command %v", c)
	return result{}, nil
}

// env is a working directory with a benchmark configuration listing corpus
// "c", that corpus installed, the benchmark tools, drupal/coder and a custos
// binary in place.
type env struct {
	root string
	args []string
}

func newEnv(t *testing.T) env {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "competitors.json"), `{"corpora":[{"name":"c","paths":["src"]},{"name":"d","paths":["src"]}]}`)
	write(t, filepath.Join(root, "bench", "corpora", "c", "src", "A.php"), "<?php\n$a = 1;\n")
	write(t, filepath.Join(root, "bench", "corpora", "c", "src", "B.php"), "<?php\n/**\n */\n$b = 1;\n")
	for _, bin := range []string{"squizlabs-php_codesniffer/vendor/bin/phpcs", "friendsofphp-php-cs-fixer/vendor/bin/php-cs-fixer", "laravel-pint/vendor/bin/pint"} {
		write(t, filepath.Join(root, "bench", "tools", bin), "")
	}
	write(t, filepath.Join(root, "tools", "drupal-coder", "vendor", "bin", "phpcs"), "")
	write(t, filepath.Join(root, "tools", "drupal-coder.version"), coderVersion)
	write(t, filepath.Join(root, "custos"), "")
	divergences, err := os.ReadFile("../../internal/format/testdata/divergences.json")
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "divergences.json"), string(divergences))
	tmp := filepath.Join(root, "tmp")
	if err := os.Mkdir(tmp, 0o755); err != nil {
		t.Fatal(err)
	}
	return env{root, []string{
		"-config", filepath.Join(root, "competitors.json"), "-bench", filepath.Join(root, "bench"),
		"-tools", filepath.Join(root, "tools"), "-custos", filepath.Join(root, "custos"),
		"-divergences", filepath.Join(root, "divergences.json"), "-tmp", tmp, "-corpora", "c",
	}}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (e env) run(t *testing.T, f *fake, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(append(slices.Clone(e.args), args...), &stdout, &stderr, f.run)
	return code, stdout.String(), stderr.String()
}

func TestRunAllPresets(t *testing.T) {
	e := newEnv(t)
	f := &fake{t: t}
	code, out, errOut := e.run(t, f)
	if code != 1 {
		t.Fatalf("exit %d, want 1 (symfony and laravel findings remain)\n%s%s", code, out, errOut)
	}
	for _, want := range []string{
		"corpus  preset   checker", "c       psr12    phpcs --standard=PSR12         13      0      0          0",
		"c       symfony  php-cs-fixer --rules=@Symfony  15      15     3          12",
		"c       laravel  pint --preset laravel          15      0      0          1",
		"pint: remaining counts", "php-cs-fixer: remaining comes", "c symfony: 12 remaining", "  array_syntax: 1\n    src/Sample.php:2: first change: \"\" → \"\"",
		"c laravel: 1 remaining", `  a_fixer,z_fixer: 1` + "\n" + `    src/A.php:2: Pint changes "$a = 1;" → "$a = 2;"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "phpdoc_trim") {
		t.Errorf("a docblock-only fix is reported:\n%s", out)
	}
	if !strings.Contains(errOut, "formatcheck: c drupal") {
		t.Errorf("no progress on stderr: %s", errOut)
	}
	// The temporary copies are gone; the corpus is untouched.
	if entries, _ := os.ReadDir(filepath.Join(e.root, "tmp")); len(entries) != 0 {
		t.Errorf("copies left: %v", entries)
	}
	if _, err := os.Stat(filepath.Join(e.root, "bench", "corpora", "c", ".formatted")); err == nil {
		t.Error("the corpus was formatted")
	}
	// Pint's fix ran with the divergent fixers that are names off.
	var cfg struct {
		Preset string          `json:"preset"`
		Rules  map[string]bool `json:"rules"`
	}
	err := json.Unmarshal(f.pintConfig, &cfg)
	if _, on := cfg.Rules["yoda_style"]; err != nil || cfg.Preset != "laravel" || !on || cfg.Rules["yoda_style"] {
		t.Errorf("pint configuration %+v (%v)", cfg, err)
	}
}

func TestRunClean(t *testing.T) {
	e := newEnv(t)
	code, out, errOut := e.run(t, &fake{t: t}, "-presets", " psr12, per ,")
	if code != 0 || strings.Contains(out, "pint:") || strings.Contains(out, "\nc p") {
		t.Fatalf("exit %d\n%s%s", code, out, errOut)
	}
}

func TestRunJSON(t *testing.T) {
	e := newEnv(t)
	code, out, _ := e.run(t, &fake{t: t}, "-json", "-presets", "laravel")
	var got []outcome
	if err := json.Unmarshal([]byte(out), &got); err != nil || code != 1 {
		t.Fatalf("exit %d, %v:\n%s", code, err, out)
	}
	if len(got) != 1 || got[0].Remaining != 1 || got[0].Findings[0].File != "src/A.php" || got[0].Before != 15 {
		t.Errorf("%+v", got)
	}
}

func TestUsage(t *testing.T) {
	for _, args := range [][]string{{"-nope"}, {"extra"}} {
		var stderr bytes.Buffer
		if code := run(args, &bytes.Buffer{}, &stderr, (&fake{t: t}).run); code != 2 || !strings.Contains(stderr.String(), "usage:") {
			t.Errorf("%v: exit %d, %s", args, code, stderr.String())
		}
	}
}

func TestRunErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, e env)
		args  []string
		fail  func(command) (result, error, bool)
		want  string
	}{
		{name: "config", setup: func(t *testing.T, e env) { remove(t, e.root, "competitors.json") }, want: "competitors.json"},
		{name: "config syntax", setup: func(t *testing.T, e env) { write(t, filepath.Join(e.root, "competitors.json"), "{") }, want: "competitors.json: unexpected end"},
		{name: "corpus", args: []string{"-corpora", "x"}, want: `unknown corpus "x"`},
		{name: "preset", args: []string{"-presets", "psr1"}, want: `unknown preset "psr1"`},
		{name: "divergences", setup: func(t *testing.T, e env) { remove(t, e.root, "divergences.json") }, want: "divergences.json"},
		{name: "custos", setup: func(t *testing.T, e env) { remove(t, e.root, "custos") }, want: "run make build"},
		{name: "tool", setup: func(t *testing.T, e env) { remove(t, e.root, "bench/tools/laravel-pint") }, want: "pint --preset laravel: stat"},
		{name: "tmp", args: []string{"-tmp", "/nonexistent/formatcheck"}, want: "/nonexistent/formatcheck"},
		{name: "not installed", args: []string{"-corpora", "d"}, want: "corpus d is not installed"},
		{name: "copy", setup: func(t *testing.T, e env) { remove(t, e.root, "bench/corpora/c/src") }, want: "c: stat"},
		{
			name: "preset failure", args: []string{"-presets", "per"},
			fail: func(c command) (result, error, bool) {
				return result{}, errors.New("boom"), filepath.Base(c.name) == "custos"
			},
			want: "c per: boom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			if tt.setup != nil {
				tt.setup(t, e)
			}
			code, out, errOut := e.run(t, &fake{t: t, fail: tt.fail}, tt.args...)
			if code != 2 || out != "" || !strings.Contains(errOut, tt.want) {
				t.Errorf("exit %d, stdout %q, stderr %q; want %q", code, out, errOut, tt.want)
			}
		})
	}
}

func remove(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(root, rel)); err != nil {
		t.Fatal(err)
	}
}

// TestCheckPreset covers the failures of each step of a preset.
func TestCheckPreset(t *testing.T) {
	newSession := func(t *testing.T) (*session, string, *fake) {
		e := newEnv(t)
		f := &fake{t: t}
		s := &session{run: f.run, bench: filepath.Join(e.root, "bench"), tools: filepath.Join(e.root, "tools"), cpus: 2}
		return s, e.root, f
	}
	c := corpus{Name: "c", Paths: []string{"src"}}
	t.Run("copy", func(t *testing.T) {
		s, root, _ := newSession(t)
		if _, err := s.checkPreset("custos", c, "per", filepath.Join(root, "none"), filepath.Join(root, "out"), nil); err == nil {
			t.Error("no error")
		}
	})
	steps := []struct {
		name, preset string
		fail         func(n *int) func(command) (result, error, bool)
		want         string
	}{
		{"format status", "per", func(*int) func(command) (result, error, bool) {
			return func(c command) (result, error, bool) {
				return result{code: 3, stderr: []byte("bad")}, nil, filepath.Base(c.name) == "custos"
			}
		}, "custos (exit status 3): formatting failed: bad"},
		{"before", "per", func(*int) func(command) (result, error, bool) {
			return func(c command) (result, error, bool) {
				return result{}, errors.New("before"), filepath.Base(c.name) == "php-cs-fixer"
			}
		}, "before"},
		{"after", "per", func(n *int) func(command) (result, error, bool) {
			return func(c command) (result, error, bool) {
				if filepath.Base(c.name) != "php-cs-fixer" {
					return result{}, nil, false
				}
				*n++
				return result{}, errors.New("after"), *n == 2
			}
		}, "after"},
		{"gate", "laravel", func(*int) func(command) (result, error, bool) {
			return func(c command) (result, error, bool) {
				return result{}, errors.New("gate"), slices.Contains(c.args, "--config")
			}
		}, "gate"},
	}
	for _, st := range steps {
		t.Run(st.name, func(t *testing.T) {
			s, root, f := newSession(t)
			var n int
			f.fail = st.fail(&n)
			orig := filepath.Join(root, "bench", "corpora", "c")
			_, err := s.checkPreset("custos", c, st.preset, orig, filepath.Join(root, "out"), nil)
			if err == nil || !strings.Contains(err.Error(), st.want) {
				t.Errorf("error %v, want %q", err, st.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	if got := list(" a,,b ,"); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("%q", got)
	}
	if got := list(""); got != nil {
		t.Errorf("%q", got)
	}
}

// A file custos refuses to format remains, and the run goes on.
func TestCheckPresetRefusal(t *testing.T) {
	e := newEnv(t)
	f := &fake{t: t}
	s := &session{run: f.run, bench: filepath.Join(e.root, "bench"), tools: filepath.Join(e.root, "tools"), cpus: 2}
	f.fail = func(c command) (result, error, bool) {
		stderr := "custos: src/a.php: formatting would change the code; please report it, not formatted\ncustos: 1 file(s) could not be formatted\n"
		return result{code: 2, stderr: []byte(stderr)}, nil, filepath.Base(c.name) == "custos"
	}
	c := corpus{Name: "c", Paths: []string{"src"}}
	oc, err := s.checkPreset("custos", c, "psr12", filepath.Join(e.root, "bench", "corpora", "c"), filepath.Join(e.root, "out"), nil)
	if err != nil || oc.Remaining == 0 || oc.Findings[0] != (finding{Rule: "custos", File: "src/a.php", Message: "formatting would change the code; please report it, not formatted"}) {
		t.Fatalf("%v %+v", err, oc)
	}
}

func TestRefusals(t *testing.T) {
	got := refusals([]byte("custos: src/a.php: why\r\nnoise\ncustos: 2 file(s) could not be formatted\ncustos: dir: not php\n"))
	if len(got) != 1 || got[0].File != "src/a.php" || got[0].Message != "why" {
		t.Fatalf("%+v", got)
	}
}
