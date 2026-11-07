//go:build unix

package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

func fakeEnv(kv ...string) []string {
	return append(os.Environ(), append([]string{"BENCHCMP_FAKE=1"}, kv...)...)
}

func TestExecute(t *testing.T) {
	exe, _ := os.Executable()
	s := execute([]string{exe}, t.TempDir(), fakeEnv("FAKE_STDOUT=out", "FAKE_STDERR=err", "FAKE_EXIT=3", "FAKE_SLEEP=100ms", "FAKE_CHILD=1"), time.Minute)
	if string(s.stdout) != "out" || string(s.stderr) != "err" || s.err == nil || s.timedOut || s.peak == 0 || s.wall < 100*time.Millisecond {
		t.Errorf("execute = %+v", s)
	}
	s = execute([]string{exe}, t.TempDir(), fakeEnv("FAKE_SLEEP=1m"), 100*time.Millisecond)
	if !s.timedOut || s.wall > testbudget.Of(30*time.Second) {
		t.Errorf("timeout: %+v", s)
	}
	if s := execute([]string{filepath.Join(t.TempDir(), "missing")}, "", nil, time.Minute); s.err == nil {
		t.Error("missing program: no error")
	}
}

// fixture makes a bench with an installed corpus "c" (two PHP files in
// src/) and a configuration template, and a tool running the fake.
func fixture(t *testing.T, runs, warmup int) (*bench, Corpus, Tool) {
	t.Helper()
	work, configs := t.TempDir(), t.TempDir()
	write(t, filepath.Join(work, "corpora", "c", "src", "a.php"), "<?php\n")
	write(t, filepath.Join(work, "corpora", "c", "src", "sub", "b.php"), "<?php\n")
	write(t, filepath.Join(work, "corpora", "c", "src", "README"), "")
	write(t, filepath.Join(configs, "tool.conf"), "root=@CORPUS@\ncache=@CACHE@\npath=@PATH@\n")
	exe, _ := os.Executable()
	b := &bench{
		cfg:     &Config{Runs: runs, Warmup: warmup, Packages: map[string]string{"v/t": "9.0.0"}},
		work:    work,
		configs: configs,
		custos:  filepath.Join(fakeBin(t, "custos"), "custos"),
		timeout: time.Minute,
		log:     func(string, ...any) {},
	}
	tool := Tool{
		ID: "t", Label: "T", Kind: "linter", Package: "v/t", Cmd: []string{exe, "{paths}"}, Findings: "psalm",
		Env: map[string]string{"BENCHCMP_FAKE": "1", "FAKE_STDOUT": `[{}, {}]`, "CONF": "{configs}/tool.conf"},
	}
	return b, Corpus{Name: "c", Repo: "https://github.com/v/c.git", Tag: "1.0.0", Paths: []string{"src", "lib"}}, tool
}

func TestMeasure(t *testing.T) {
	withPath(t, fakeBin(t, "php"), false)
	b, c, tool := fixture(t, 2, 1)
	c.Paths = []string{"src"}
	custos := Tool{
		ID: "custos", Label: "custos analyse", Kind: "analyzer", Cmd: []string{"{custos}"}, Findings: "custos",
		Env: map[string]string{"BENCHCMP_FAKE": "1", "FAKE_STDERR": "custos: would fix 1 file(s), 3 edit(s)"},
	}
	custos.Findings, custos.Stream = "custos-fix", "stderr"
	r, err := b.measure([]Tool{tool, custos}, []Corpus{c})
	if err != nil {
		t.Fatal(err)
	}
	if r.PHP != "8.4.1" || r.Runs != 2 || len(r.Corpora) != 1 || r.Corpora[0].Files != 2 {
		t.Fatalf("results = %+v", r)
	}
	rows := r.Corpora[0].Rows
	if rows[0].Status != "ok" || rows[0].Version != "9.0.0" || rows[0].Findings != 2 || rows[0].Min > rows[0].Median || rows[0].Median > rows[0].Max {
		t.Errorf("tool row = %+v", rows[0])
	}
	if rows[1].Status != "ok" || rows[1].Version != "1.2.3" || rows[1].Findings != 3 {
		t.Errorf("custos row = %+v", rows[1])
	}

	if _, err := b.measure([]Tool{tool}, []Corpus{{Name: "missing", Paths: []string{"src"}}}); err == nil {
		t.Error("missing corpus: no error")
	}
}

func TestMeasureTool(t *testing.T) {
	b, c, tool := fixture(t, 3, 0)
	c.Paths = []string{"src"}
	if row := b.measureTool(tool, c); row.Status != "ok" || row.Min > row.Median || row.Findings != 2 {
		t.Errorf("odd runs: %+v", row)
	}

	tool.Env["FAKE_STDOUT"] = "nope"
	tool.Env["FAKE_EXIT"] = "255"
	tool.Env["FAKE_STDERR"] = "\n  Fatal: " + strings.Repeat("x", 200) + "\nmore"
	row := b.measureTool(tool, c)
	if !strings.HasPrefix(row.Status, "error: no JSON in output (exit status 255): Fatal: xxx") || !strings.HasSuffix(row.Status, "x…") {
		t.Errorf("failing tool: %q", row.Status)
	}
	tool.Env["FAKE_EXIT"], tool.Env["FAKE_STDERR"] = "0", ""
	if row := b.measureTool(tool, c); row.Status != "error: no JSON in output" {
		t.Errorf("bad output: %q", row.Status)
	}

	tool.Env["FAKE_SLEEP"] = "1m"
	b.timeout = 100 * time.Millisecond
	if row := b.measureTool(tool, c); row.Status != "timeout" {
		t.Errorf("slow tool: %+v", row)
	}

	b.configs = filepath.Join(t.TempDir(), "missing")
	if row := b.measureTool(tool, c); !strings.Contains(row.Status, "missing") {
		t.Errorf("missing configs: %+v", row)
	}
	b.work = filepath.Join(t.TempDir(), "missing")
	if row := b.measureTool(tool, c); row.Status == "ok" {
		t.Errorf("missing work: %+v", row)
	}
}

func TestEvenRunsMedian(t *testing.T) {
	b, c, tool := fixture(t, 2, 0)
	c.Paths = []string{"src"}
	row := b.measureTool(tool, c)
	if row.Median != (row.Min+row.Max)/2 {
		t.Errorf("median of 2 runs: %+v", row)
	}
}

func TestRenderConfigs(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	write(t, filepath.Join(src, "a.conf"), "root=@CORPUS@\n  - @CORPUS@/@PATH@\nend @CACHE@")
	if err := renderConfigs(src, dst, "/c", []string{"src", "lib"}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, filepath.Join(dst, "a.conf")); got != "root=/c\n  - /c/src\n  - /c/lib\nend "+dst {
		t.Errorf("rendered %q", got)
	}
	if err := os.Mkdir(filepath.Join(src, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := renderConfigs(src, dst, "/c", nil); err == nil {
		t.Error("directory template: no error")
	}
}

func TestExpand(t *testing.T) {
	got := expand([]string{"{bin}/x", "--a={bin}", "{paths}", "y"}, map[string]string{"{bin}": "/b"}, []string{"p", "q"})
	if want := []string{"/b/x", "--a=/b", "p", "q", "y"}; !slices.Equal(got, want) {
		t.Errorf("expand = %q", got)
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine([]byte("  \n ")); got != "" {
		t.Errorf("empty: %q", got)
	}
}

func TestVersions(t *testing.T) {
	withPath(t, t.TempDir(), true)
	b := &bench{custos: filepath.Join(t.TempDir(), "custos")}
	if b.custosVersion() != "unknown" || phpVersion() != "unknown" {
		t.Error("missing programs: version known")
	}
}

func TestSelectByName(t *testing.T) {
	items := []string{"a", "b", "c"}
	id := func(s string) string { return s }
	if got, _ := selectByName(items, id, ""); !slices.Equal(got, items) {
		t.Errorf("no filter: %q", got)
	}
	if got, _ := selectByName(items, id, "c,a"); !slices.Equal(got, []string{"a", "c"}) {
		t.Errorf("filter: %q", got)
	}
	if _, err := selectByName(items, id, "z"); !errors.Is(err, errNoMatch) {
		t.Errorf("no match: %v", err)
	}
}
