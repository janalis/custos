package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/format"
)

// The committed page is what the registry renders (make rules-doc).
func TestPageIsCurrent(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", out))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != render() {
		t.Errorf("%s is stale: run make rules-doc", out)
	}
}

func TestPresetOrder(t *testing.T) {
	if got := slices.Sorted(slices.Values(presetOrder)); !slices.Equal(got, format.Presets()) {
		t.Errorf("presetOrder %v, want the presets %v", presetOrder, format.Presets())
	}
}

func TestRender(t *testing.T) {
	page := render()
	for _, want := range []string{
		"## File\n",
		"## Risky\n",
		"| Setting | Type | Values | Description |\n",
		"| Setting | `psr1` | `psr12` | `per` | `symfony` | `laravel` | `drupal` |\n",
		" (risky) |",
		", risky: ",
		"- [`per`](" + format.PresetReference("per") + ")",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	if strings.Contains(page, "{{") {
		t.Error("page holds {{, which VitePress would compile")
	}
}

func TestAllowedAndValue(t *testing.T) {
	for _, c := range []struct {
		d    *format.Setting
		want string
	}{
		{&format.Setting{Kind: format.KindBool}, "`true`, `false`"},
		{&format.Setting{Kind: format.KindInt}, "integer, 0 or more"},
		{&format.Setting{Kind: format.KindEnum, Values: []string{"a", "b"}}, "`a`, `b`"},
		{&format.Setting{Kind: format.KindList, Values: []string{"a", "b"}}, "list of `a`, `b`"},
	} {
		if got := allowed(c.d); got != c.want {
			t.Errorf("allowed(%v) = %q, want %q", c.d.Kind, got, c.want)
		}
	}
	for v, want := range map[any]string{true: "`true`", 120: "`120`", "keep": "`keep`"} {
		if got := value(v); got != want {
			t.Errorf("value(%v) = %q, want %q", v, got, want)
		}
	}
	if got := value([]string{}); got != "`[]`" {
		t.Errorf("empty list: %q", got)
	}
	if got := value([]string{"a", "b"}); got != "`a`, `b`" {
		t.Errorf("list: %q", got)
	}
	if got := cell("a | b"); got != `a \| b` {
		t.Errorf("cell: %q", got)
	}
}

func TestRun(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs", "guide"), 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(root, &stdout, &stderr); code != 0 || !strings.HasPrefix(stdout.String(), "format settings: ") {
		t.Fatalf("run = %d, stdout %q, stderr %q", code, &stdout, &stderr)
	}
	if b, _ := os.ReadFile(filepath.Join(root, out)); string(b) != render() {
		t.Error("written page differs from render")
	}
	stdout.Reset()
	if code := run(filepath.Join(root, "missing"), &stdout, &stderr); code != 1 || stderr.Len() == 0 {
		t.Errorf("run without docs/guide = %d, stderr %q", code, &stderr)
	}
}
