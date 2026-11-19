package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeCustomProject writes a project declaring custom rules and changes
// into it.
func writeCustomProject(t *testing.T, config string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range map[string]string{
		"custos.json": config,
		"src/a.php":   "<?php\ndump($x);\n$d = new \\DateTimeImmutable();\n",
	} {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	return dir
}

const customConfig = `{
	"customRules": [
		{"id": "NoDump", "match": "\\dump(...)", "message": "Remove the dump", "severity": "error", "fix": "", "description": "Debug calls stay out of src."},
		{"id": "UseClock", "match": "new \\DateTimeImmutable()", "message": "Use the clock", "fix": "$this->clock->now()"}
	],
	"rules": {"UseClock": {"severity": "info"}}
}`

func TestCustomRulesCommands(t *testing.T) {
	dir := writeCustomProject(t, customConfig)
	code, out, stderr := runCLI(t, "analyse", "--no-cache", "--format=sarif", "--rule", "NoDump,UseClock")
	if code != 1 || !strings.Contains(out, `"text": "Debug calls stay out of src."`) || !strings.Contains(out, `"Custom"`) || !strings.Contains(out, `"ruleId": "UseClock"`) {
		t.Fatalf("analyse: %d %s %s", code, out, stderr)
	}
	code, out, _ = runCLI(t, "explain", "NoDump")
	if code != 0 || !strings.Contains(out, "group: Custom, severity: error") || !strings.Contains(out, "has a quick-fix") || !strings.Contains(out, "Debug calls stay out of src.") {
		t.Fatalf("explain: %d %s", code, out)
	}
	code, out, _ = runCLI(t, "rules", "--json")
	var rows []struct {
		ID          string `json:"id"`
		Implemented bool   `json:"implemented"`
		HasFix      bool   `json:"hasFix"`
	}
	if err := json.Unmarshal([]byte(out), &rows); code != 0 || err != nil {
		t.Fatalf("rules: %d %v", code, err)
	}
	if last := rows[len(rows)-1]; last.ID != "UseClock" || !last.Implemented || !last.HasFix {
		t.Fatalf("rules: %+v", last)
	}
	if code, _, stderr := runCLI(t, "fix", "--no-cache", "--rule", "NoDump,UseClock"); code != 0 {
		t.Fatalf("fix: %d %s", code, stderr)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "src/a.php"))
	if string(got) != "<?php\n$d = $this->clock->now();\n" {
		t.Fatalf("fixed: %q", got)
	}
}

func TestCustomRulesInvalidConfig(t *testing.T) {
	writeCustomProject(t, `{"customRules": [{"id": "Bad", "match": "f(", "message": "m"}]}`)
	for _, args := range [][]string{{"analyse"}, {"rules"}, {"explain", "Bad"}} {
		if code, _, stderr := runCLI(t, args...); code != 2 || !strings.Contains(stderr, "custom rule Bad") {
			t.Errorf("%v: %d %s", args, code, stderr)
		}
	}
	// A catalogue rule is explained without reading the configuration.
	if code, _, _ := runCLI(t, "explain", "NestedNotOperators"); code != 0 {
		t.Fatalf("explain catalogue rule: %d", code)
	}
}
