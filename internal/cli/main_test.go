package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/inspection/meta"
)

func writeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("composer.json", `{"require": {"php": ">=8.2"}}`)
	write("a.php", "<?php\n$x = !!$y;;\n")
	return dir
}

func runCLI(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := Run("dev", args, strings.NewReader(""), &out, &errb)
	return code, out.String(), errb.String()
}

func TestAnalyseJSON(t *testing.T) {
	dir := writeProject(t)
	code, out, stderr := runCLI(t, "analyse", "--format=json", "--rule", "NestedNotOperators,UnnecessarySemicolon", dir)
	if code != 0 { // both rules report info, below the default fail-on=warning
		t.Fatalf("exit %d, stderr %s", code, stderr)
	}
	var res struct {
		Files    int `json:"files"`
		Findings []struct {
			Rule string `json:"rule"`
			Line int    `json:"line"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if res.Files != 1 || len(res.Findings) != 2 {
		t.Fatalf("got %+v", res)
	}
	if code, _, _ := runCLI(t, "analyse", "--fail-on=info", "--rule", "UnnecessarySemicolon", dir); code != 1 {
		t.Fatalf("fail-on=info: exit %d", code)
	}
}

func TestFixDryRunAndApply(t *testing.T) {
	dir := writeProject(t)
	file := filepath.Join(dir, "a.php")
	code, out, _ := runCLI(t, "fix", "--dry-run", "--diff", "--rule", "NestedNotOperators", dir)
	if code != 0 || !strings.Contains(out, "+$x = (bool)$y;;") {
		t.Fatalf("dry-run diff: exit %d\n%s", code, out)
	}
	if b, _ := os.ReadFile(file); string(b) != "<?php\n$x = !!$y;;\n" {
		t.Fatalf("dry-run modified the file: %q", b)
	}
	if code, _, stderr := runCLI(t, "fix", "--rule", "NestedNotOperators,UnnecessarySemicolon", dir); code != 0 {
		t.Fatalf("fix: exit %d %s", code, stderr)
	}
	if b, _ := os.ReadFile(file); string(b) != "<?php\n$x = (bool)$y;\n" {
		t.Fatalf("fixed content: %q", b)
	}
}

func TestMiscCommands(t *testing.T) {
	if code, out, _ := runCLI(t, "explain", "UnnecessarySemicolonInspection"); code != 0 || !strings.Contains(out, "UnnecessarySemicolon") {
		t.Fatalf("explain: %d %s", code, out)
	}
	rules, err := meta.All()
	if err != nil {
		t.Fatal(err)
	}
	if code, out, _ := runCLI(t, "rules"); code != 0 || strings.Count(out, "\n") != len(rules) {
		t.Fatalf("rules: %d lines", strings.Count(out, "\n"))
	}
	if code, _, _ := runCLI(t, "analyse", "--rule", "Nope", "."); code != 2 {
		t.Fatalf("unknown rule: exit %d", code)
	}
	if code, _, _ := runCLI(t, "bogus"); code != 2 {
		t.Fatalf("unknown command: exit %d", code)
	}
}

func TestBaselineFlags(t *testing.T) {
	dir := writeProject(t)
	bl := filepath.Join(dir, "custos-baseline.json")
	if code, _, stderr := runCLI(t, "analyse", "--rule", "UnnecessarySemicolon", "--generate-baseline", bl, dir); code != 0 || !strings.Contains(stderr, "1 entries") {
		t.Fatalf("generate: %d %s", code, stderr)
	}
	code, out, _ := runCLI(t, "analyse", "--fail-on=info", "--format=json", "--rule", "UnnecessarySemicolon", "--baseline", bl, dir)
	if code != 0 || !strings.Contains(out, `"findings": []`) {
		t.Fatalf("baseline filter: %d %s", code, out)
	}
}

func TestConfiguredBaselineMayBeMissing(t *testing.T) {
	dir := writeProject(t)
	if err := os.WriteFile(filepath.Join(dir, "custos.json"), []byte(`{"baseline": "custos-baseline.json"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runCLI(t, "analyse", "--rule", "UnnecessarySemicolon", dir); code != 0 {
		t.Fatalf("missing configured baseline: exit %d %s", code, stderr)
	}
	if code, _, _ := runCLI(t, "analyse", "--baseline", filepath.Join(dir, "nope.json"), dir); code != 2 {
		t.Fatalf("missing explicit baseline must fail, got %d", code)
	}
}

// TestBaselineShowsRepeatsOfRecordedFindings: an access PossiblyNullAccess
// leaves out after a reported one on the same value is reported when the
// baseline records that one, and a generated baseline records it too, so
// that a later access added on the value is new.
func TestBaselineShowsRepeatsOfRecordedFindings(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.php")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(file, []byte("<?php\nfunction f(?Exception $a) {\n"+body+"}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	analyse := func(extra ...string) []int {
		t.Helper()
		_, out, _ := runCLI(t, append(append([]string{"analyse", "--fail-on=never", "--format=json", "--rule", "PossiblyNullAccess"}, extra...), dir)...)
		var res struct {
			Findings []struct {
				Line int `json:"line"`
			} `json:"findings"`
		}
		if err := json.Unmarshal([]byte(out), &res); err != nil {
			t.Fatalf("bad json: %v\n%s", err, out)
		}
		var got []int
		for _, f := range res.Findings {
			got = append(got, f.Line)
		}
		return got
	}
	write("$a->getMessage();\n$a->getCode();\n$a->getLine();\n")
	if got := analyse(); len(got) != 1 || got[0] != 3 {
		t.Fatalf("without baseline: lines %v, want [3]", got)
	}
	// A baseline recording the first access only shows the next one.
	bl := filepath.Join(dir, "bl.json")
	entry := `{"version": 1, "entries": [{"path": "a.php", "rule": "PossiblyNullAccess", "message": "Check this value for null before the access: it is Exception|null.", "line": "$a->getMessage();", "count": 1}]}`
	if err := os.WriteFile(bl, []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := analyse("--baseline", bl); len(got) != 1 || got[0] != 4 {
		t.Fatalf("first access recorded: lines %v, want [4]", got)
	}
	// A generated baseline records every access and hides them all.
	if code, _, stderr := runCLI(t, "analyse", "--rule", "PossiblyNullAccess", "--generate-baseline", bl, dir); code != 0 || !strings.Contains(stderr, "3 entries") {
		t.Fatalf("generate: %d %s", code, stderr)
	}
	if got := analyse("--baseline", bl); len(got) != 0 {
		t.Fatalf("generated baseline: lines %v, want none", got)
	}
	write("$a->getMessage();\n$a->getCode();\n$a->getLine();\n$a->getFile();\n")
	if got := analyse("--baseline", bl); len(got) != 1 || got[0] != 6 {
		t.Fatalf("access added after the baseline: lines %v, want [6]", got)
	}
}
