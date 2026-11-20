package format

import (
	"strconv"
	"strings"
	"testing"
)

func TestSettingSetAndGet(t *testing.T) {
	var s Settings
	cases := []struct {
		id    string
		value any
		want  any
	}{
		{"final_newline", true, true},
		{"final_newline", "false", false},
		{"indent_size", 2, 2},
		{"indent_size", float64(3), 3},
		{"indent_size", "8", 8},
		{"concat_space", "none", "none"},
		{"trailing_comma", "arrays,match", []string{"arrays", "match"}},
		{"trailing_comma", "", []string{}},
		{"trailing_comma", []string{"parameters"}, []string{"parameters"}},
		{"trailing_comma", []any{"arguments", "arrays"}, []string{"arguments", "arrays"}},
	}
	for _, c := range cases {
		d, ok := Lookup(c.id)
		if !ok {
			t.Fatalf("no setting %s", c.id)
		}
		if err := d.Set(&s, c.value); err != nil {
			t.Fatalf("%s=%v: %v", c.id, c.value, err)
		}
		if got := d.Get(&s); strings.Join(strings.Fields(strings.Trim(stringOf(got), "[]")), ",") != strings.Join(strings.Fields(strings.Trim(stringOf(c.want), "[]")), ",") {
			t.Errorf("%s=%v: got %v, want %v", c.id, c.value, got, c.want)
		}
	}
}

func stringOf(v any) string {
	switch x := v.(type) {
	case []string:
		return "[" + strings.Join(x, " ") + "]"
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(x)
	case string:
		return x
	}
	return "?"
}

func TestSettingSetErrors(t *testing.T) {
	var s Settings
	bad := []struct {
		id    string
		value any
	}{
		{"final_newline", "maybe"},
		{"final_newline", 3},
		{"indent_size", -1},
		{"indent_size", 17},
		{"indent_size", 1.5},
		{"indent_size", "x"},
		{"indent_size", true},
		{"concat_space", "two"},
		{"concat_space", 1},
		{"trailing_comma", "arrays,nope"},
		{"trailing_comma", []any{1}},
		{"trailing_comma", 7},
	}
	for _, c := range bad {
		d, _ := Lookup(c.id)
		if err := d.Set(&s, c.value); err == nil || !strings.Contains(err.Error(), c.id) {
			t.Errorf("%s=%v: error %v", c.id, c.value, err)
		}
	}
}

func TestRegistry(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range Registry() {
		if seen[d.ID] || d.Description == "" || d.Group == "" {
			t.Errorf("setting %q: duplicate or undocumented", d.ID)
		}
		seen[d.ID] = true
		if d.Kind.String() == "" || (d.Kind == KindEnum || d.Kind == KindList) && len(d.Values) == 0 {
			t.Errorf("setting %q: no values", d.ID)
		}
		if _, ok := messages[d.ID]; !ok && d.ID != "spacing" && d.ID != "indent_style" && d.ID != "indent_size" && d.ID != "tab_width" && d.ID != "continuation_indent" && d.ID != "single_line_closures" && d.ID != "report_long_lines" {
			t.Errorf("setting %q: no change message", d.ID)
		}
	}
	for _, kind := range []Kind{KindBool, KindInt, KindEnum, KindList} {
		if kind.String() == "" {
			t.Error("kind without name")
		}
	}
	if _, ok := Lookup("nope"); ok {
		t.Error("unknown setting found")
	}
}

func TestPresets(t *testing.T) {
	names := Presets()
	if strings.Join(names, " ") != "drupal laravel per psr1 psr12 symfony" {
		t.Fatalf("presets %v", names)
	}
	for _, n := range names {
		if !strings.HasPrefix(PresetReference(n), "https://") {
			t.Errorf("%s: no reference", n)
		}
		r, err := Resolve(n, true)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range Registry() {
			if r.Source[d.ID] == "" {
				t.Errorf("%s: %s has no source", n, d.ID)
			}
		}
	}
	if _, err := Resolve("nope", false); err == nil {
		t.Error("unknown preset accepted")
	}
	if _, err := Resolve("per", false, Layer{Source: "x", Values: map[string]any{"nope": 1}}); err == nil || !strings.Contains(err.Error(), "unknown format setting") {
		t.Errorf("unknown setting: %v", err)
	}
	if _, err := Resolve("per", false, Layer{Source: "x", Values: map[string]any{"indent_size": "big"}}); err == nil || !strings.HasPrefix(err.Error(), "x: ") {
		t.Errorf("bad value: %v", err)
	}
	r, _ := Resolve("symfony", false, Layer{Source: "custos.json", Values: map[string]any{"concat_space": "one"}})
	if r.ConcatSpace != "one" || r.Source["concat_space"] != "custos.json" || r.Source["yoda"] != "default" {
		t.Errorf("layers: %+v", r.Source)
	}
	r, _ = Resolve("symfony", true)
	if r.Yoda != "yoda" || r.Source["yoda"] != "preset symfony (risky)" {
		t.Errorf("risky: %v %v", r.Yoda, r.Source["yoda"])
	}
}

func TestChangeMessage(t *testing.T) {
	if (Change{Rule: "concat_space"}).Message() == "" || (Change{Rule: "nope"}).Message() != "Fix the layout" {
		t.Error("messages")
	}
	if (Result{}).Changed() || !(Result{Changes: []Change{{}}}).Changed() {
		t.Error("Changed")
	}
}

func TestSourceDefaults(t *testing.T) {
	res := Source("a.php", []byte("<?php\necho  1;\n"), Options{})
	if string(res.Output) != "<?php\n\necho 1;\n" {
		t.Errorf("default settings: %q", res.Output)
	}
	if plural(2, "syntax error") != "2 syntax errors" {
		t.Error("plural")
	}
}
