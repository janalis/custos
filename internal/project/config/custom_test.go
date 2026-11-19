package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/pattern"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCustomRules(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, FileName), `{
		"customRules": [{"id": "NoDump", "match": "\\dump(...)", "message": "Remove it"}],
		"customRuleFiles": ["rules/more.json"],
		"rules": {"NoDump": {"severity": "error"}, "More": {"enabled": false}}
	}`)
	writeFile(t, filepath.Join(dir, "rules/more.json"), `[{"id": "More", "match": "\\more()", "message": "m"}]`)
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Custom) != 2 || c.Custom[0].ID() != "NoDump" || c.Custom[1].ID() != "More" || len(c.Definitions) != 2 {
		t.Fatalf("custom: %v %v", c.Custom, c.Definitions)
	}
	if c.Rules["NoDump"].Severity != "error" || *c.Rules["More"].Enabled {
		t.Fatalf("overrides: %+v", c.Rules)
	}
	reg := c.Registry(nil)
	if len(reg) != 2 {
		t.Fatalf("registry %v", reg)
	}
	// The editor's settings keep the project's custom rules and may
	// configure them.
	over, err := Overlay(c, File{Rules: map[string]RuleSettings{"More": {Severity: "info"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(over.Custom) != 2 || over.Rules["More"].Severity != "info" || !only(over.CustomRuleFiles, "rules/more.json") {
		t.Fatalf("overlay: %+v", over)
	}
	// Captured sources resolve the rule files too.
	captured, err := ResolveCaptured(dir, map[string][]byte{
		FileName:          []byte(`{"customRuleFiles": ["rules/more.json"]}`),
		"rules/more.json": []byte(`[{"id": "More", "match": "\\more()", "message": "m"}]`),
	})
	if err != nil || len(captured.Custom) != 1 {
		t.Fatalf("captured: %v %v", captured, err)
	}
}

func only(list []string, want string) bool { return len(list) == 1 && list[0] == want }

func TestCustomRuleErrors(t *testing.T) {
	cases := []struct {
		file File
		want string
	}{
		{File{CustomRuleFiles: []string{"../x.json"}}, "leaves the project root"},
		{File{CustomRuleFiles: []string{"missing.json"}}, `customRuleFiles entry "missing.json"`},
		{File{CustomRuleFiles: []string{"bad.json"}}, "bad.json:"},
		{File{CustomRules: []pattern.Definition{{ID: "R", Message: "m"}}}, "custom rule R: match is required"},
		{File{CustomRules: []pattern.Definition{{ID: "R", Match: pattern.Patterns{"f()"}, Message: "m"}, {ID: "R", Match: pattern.Patterns{"g()"}, Message: "m"}}}, "declared twice"},
		{File{Rules: map[string]RuleSettings{"NotDeclared": {}}}, `unknown rule "NotDeclared"`},
		{File{CustomRuleTests: "../tests"}, "customRuleTests"},
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "bad.json"), `{`)
	for _, tc := range cases {
		if _, err := Resolve(dir, tc.file); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%+v: got %v, want %q", tc.file, err, tc.want)
		}
	}
}
