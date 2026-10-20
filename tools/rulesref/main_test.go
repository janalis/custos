package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/meta"
)

func fake(t *testing.T, rules []meta.Rule, err error, descs map[string]meta.Description) {
	t.Helper()
	oc, od := catalogue, describe
	catalogue = func() ([]meta.Rule, error) { return rules, err }
	describe = func(id string) (meta.Description, bool) { d, ok := descs[id]; return d, ok }
	t.Cleanup(func() { catalogue, describe = oc, od })
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

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func contains(t *testing.T, name, doc string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(doc, w) {
			t.Errorf("%s: missing %q in:\n%s", name, w, doc)
		}
	}
}

const kettleSpec = "---\nid: Kettle\nphp: { min: \"7.1\", max: \"\" }\n---\n\n# Kettle\n\n## Examples\n\n" +
	"```php\n<?php\n<warning descr=\"Whistle once.\">whistle();\nwhistle();</warning>\n<warning descr=\"Whistle once.\">whistle();</warning><warning>x</warning>\nboil();\n```\n\n" +
	"```php\n<?php\nwhistle();\nboil();\n```\n"

func TestRun(t *testing.T) {
	fake(t, []meta.Rule{
		{
			ID: "Kettle", LegacyID: "KettleInspection", Group: "Hot Drinks", Severity: diagnostic.SeverityWarning, EnabledByDefault: true,
			Options: []meta.Option{{Name: "PITCH", Type: "int", Default: 3}, {Name: "TUNES", Type: "list"}, {Name: "MODE", Type: "enum"}},
		},
		{ID: "Toaster", LegacyID: "ToasterInspection", Group: "Hot Drinks", Severity: diagnostic.SeverityError, Experimental: true},
		{ID: "Lamp", LegacyID: "LampInspection", Group: "Bedroom", Severity: diagnostic.SeverityInfo},
		{ID: "Spoon", Native: true, Group: "bathroom", Severity: diagnostic.SeverityInfo, EnabledByDefault: true},
		{ID: "Clock", LegacyID: "ClockInspection", Group: "Bedroom", Severity: diagnostic.SeverityInfo, EnabledByDefault: true},
		{ID: "Rug", LegacyID: "RugInspection", Group: "Bedroom", Severity: diagnostic.SeverityInfo},
		{ID: "Bed", LegacyID: "BedInspection", Group: "Bedroom", Severity: diagnostic.SeverityInfo},
	}, nil, map[string]meta.Description{
		"Kettle":  {Summary: "Kettles should whistle once.", Options: "- `PITCH` (int, default 3)", Fix: true},
		"Toaster": {Summary: "Toast evenly.", Fix: true},
		"Lamp":    {Options: "- `DIM` (bool)"},
	})
	root := t.TempDir()
	write(t, filepath.Join(root, "specs/Kettle.md"), kettleSpec)
	write(t, filepath.Join(root, "specs/Lamp.md"), "php: { min: \"\", max: \"7.0\" }\n## Examples\nAt night:\n```php\n<info descr=\"Use &lt;?= {{x}} ?>.\">on();</info>\n```\n")
	write(t, filepath.Join(root, "specs/Toaster.md"), "php: { min: \"7.0\", max: \"8.0\" }\n## Examples\n```json\n{}\n```\n")
	write(t, filepath.Join(root, "specs/Rug.md"), "# Rug\n## Examples\n```php\n<warning>unclosed\n```\n")
	write(t, filepath.Join(root, "specs/Bed.md"), "# Bed\n## Examples\n```php\nsleep();\n```\n")
	// A page of a rule that no longer exists is removed.
	write(t, filepath.Join(root, outDir, "attic/Gone.md"), "stale")

	var stdout, stderr bytes.Buffer
	if code := run(root, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if got := stdout.String(); got != "docs/rules: 7 rules in 3 groups\n" {
		t.Errorf("stdout = %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, outDir, "attic")); !os.IsNotExist(err) {
		t.Errorf("stale page kept: %v", err)
	}

	index := read(t, filepath.Join(root, outDir, "index.md"))
	contains(t, "index", index,
		"title: Rules\neditLink: false\n",
		"custos ships **7 rules** in 3 groups.",
		"## Bedroom\n\n| Rule | Origin | Severity | Default | Quick-fix | PHP |\n| :--- | :--- | :--- | :---: | :---: | :--- |\n"+
			"| [Lamp](./bedroom/Lamp) | EA | info | off |  | PHP ≤ 7.0 |\n| [Clock](./bedroom/Clock) | EA | info | on |  | any |\n",
		"| [Kettle](./hot-drinks/Kettle) | EA | warning | on | ✓ | PHP ≥ 7.1 |\n| [Toaster](./hot-drinks/Toaster) | EA | error | off, experimental | ✓ | PHP 7.0–8.0 |\n",
		"| [Spoon](./bathroom/Spoon) | native | info | on |  | any |\n",
	)
	// Groups are sorted by display name, ignoring case.
	if i, j, k := strings.Index(index, "## bathroom"), strings.Index(index, "## Bedroom"), strings.Index(index, "## Hot Drinks"); i < 0 || i > j || j > k {
		t.Errorf("groups out of order: bathroom %d, Bedroom %d, Hot Drinks %d", i, j, k)
	}

	kettle := read(t, filepath.Join(root, outDir, "hot-drinks/Kettle.md"))
	contains(t, "Kettle", kettle,
		"# Kettle\n\n<Badge type=\"warning\" text=\"warning\" /> <Badge type=\"tip\" text=\"on by default\" /> <Badge type=\"tip\" text=\"quick-fix\" /> <Badge type=\"info\" text=\"PHP ≥ 7.1\" />",
		"Group: [Hot Drinks](/rules/#hot-drinks) · PhpStorm name: `KettleInspection`\n\nKettles should whistle once.\n",
		"::: code-group\n\n```php{2,3,4} [Before]\n<?php\nwhistle();\nwhistle();\nwhistle();x\nboil();\n```\n\n```php [After fix]\n<?php\nwhistle();\nboil();\n```\n\n:::\n",
		"Reported:\n\n<ul>\n<li>line 2: Whistle once.</li>\n<li>line 4: Whistle once.</li>\n</ul>\n\n## Options\n\n- `PITCH` (int, default 3)\n",
		"\"Kettle\": {\n      \"enabled\": false,\n      \"options\": {\n        \"PITCH\": 3,\n        \"TUNES\": []\n      }\n    }",
		"// @custos-ignore Kettle\n\n/**\n * @noinspection KettleInspection\n */",
	)

	toaster := read(t, filepath.Join(root, outDir, "hot-drinks/Toaster.md"))
	contains(t, "Toaster", toaster, `<Badge type="danger" text="error" /> <Badge type="info" text="off by default" /> <Badge type="warning" text="experimental" />`, "\"Toaster\": {\n      \"enabled\": true\n    }")
	if strings.Contains(toaster, "## Example") {
		t.Errorf("Toaster: non-PHP example rendered:\n%s", toaster)
	}

	lamp := read(t, filepath.Join(root, outDir, "bedroom/Lamp.md"))
	contains(t, "Lamp", lamp, "## Example\n\nAt night:\n\n```php{1}\non();\n```\n\nReported:\n\n<ul>\n<li>line 1: Use &lt;?= &#123;&#123;x&#125;&#125; ?&gt;.</li>\n</ul>\n")
	for _, id := range []string{"Clock", "Rug", "Bed"} {
		if doc := read(t, filepath.Join(root, outDir, "bedroom", id+".md")); strings.Contains(doc, "## Example") {
			t.Errorf("%s: unexpected example:\n%s", id, doc)
		}
	}

	contains(t, "sidebar", read(t, filepath.Join(root, sidebar)),
		"{\n    \"text\": \"Bedroom\",\n    \"collapsed\": true,\n    \"items\": [\n      {\n        \"text\": \"Lamp\",\n        \"link\": \"/rules/bedroom/Lamp\"\n      },")
}

func TestRunErrors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	fake(t, nil, errors.New("catalogue broken"), nil)
	if code := run(t.TempDir(), &stdout, &stderr); code != 1 || stderr.String() != "catalogue broken\n" {
		t.Fatalf("catalogue error: exit %d, %q", code, stderr.String())
	}
	fake(t, nil, nil, nil)

	// docs is a file: the old pages cannot be removed.
	root := t.TempDir()
	write(t, filepath.Join(root, "docs"), "")
	stderr.Reset()
	if code := run(root, &stdout, &stderr); code != 1 || stderr.Len() == 0 {
		t.Fatalf("remove error: exit %d, %q", code, stderr.String())
	}

	// docs/.vitepress is a file: the sidebar's directory cannot be created.
	root = t.TempDir()
	write(t, filepath.Join(root, "docs/.vitepress"), "")
	stderr.Reset()
	if code := run(root, &stdout, &stderr); code != 1 || stderr.Len() == 0 {
		t.Fatalf("mkdir error: exit %d, %q", code, stderr.String())
	}

	// The sidebar path is a directory: it cannot be written.
	root = t.TempDir()
	write(t, filepath.Join(root, sidebar, "x"), "")
	stderr.Reset()
	if code := run(root, &stdout, &stderr); code != 1 || stderr.Len() == 0 {
		t.Fatalf("write error: exit %d, %q", code, stderr.String())
	}
}

func TestNativePage(t *testing.T) {
	body := page(meta.Rule{ID: "NativeExample", Native: true, Group: "Correctness", Severity: diagnostic.SeverityWarning}, "correctness", meta.Description{}, "")
	contains(t, "native page", body, "@custos-ignore NativeExample", "Group: [Correctness]", `<Badge type="tip" text="native" />`)
	for _, unwanted := range []string{"PhpStorm name", "@noinspection"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("native page contains %q", unwanted)
		}
	}
	md := "# NativeExample\n\n## Detection\n\n- D1. Proven value.\n- D2. Traced value, unless E1 applies.\n\n## Exceptions (no report)\n\n- E1. Unknown values (D3).\n\n## Fix\n\nPreserve behavior.\n\n## PHP versions\n\nPHP 5.4 and later.\n"
	body = page(meta.Rule{ID: "NativeExample", Native: true}, "correctness", meta.Description{}, md)
	// Labels go unless referred to; a reference to a missing label goes too.
	contains(t, "native contracts", body, "- Proven value.", "- Traced value, unless E1 applies.", "- E1. Unknown values.", "Preserve behavior.", "PHP 5.4 and later.")
}

func TestByLetters(t *testing.T) {
	var items []sideItem
	for _, id := range []string{"Apple", "Avocado", "Banana", "Cherry", "Date"} {
		items = append(items, sideItem{Text: id})
	}
	collapsed := true
	got := byLetters(items, &collapsed)
	var texts []string
	for _, g := range got {
		texts = append(texts, fmt.Sprintf("%s:%d", g.Text, len(g.Items)))
	}
	if want := "A–D:5"; strings.Join(texts, ",") != want {
		t.Fatalf("got %v, want %s", texts, want)
	}
	many := make([]sideItem, bucketMax)
	for i := range many {
		many[i] = sideItem{Text: "B" + fmt.Sprint(i)}
	}
	got = byLetters(append(append(items[:1:1], many...), items[3:]...), &collapsed)
	texts = nil
	for _, g := range got {
		texts = append(texts, fmt.Sprintf("%s:%d", g.Text, len(g.Items)))
	}
	if want := fmt.Sprintf("A:1,B:%d,C–D:2", bucketMax); strings.Join(texts, ",") != want {
		t.Fatalf("got %v, want %s", texts, want)
	}
}

// TestRunSplitsLargeGroups: a group over splitOver rules gets letter
// sub-groups in the sidebar.
func TestRunSplitsLargeGroups(t *testing.T) {
	var rules []meta.Rule
	for i := 0; i <= splitOver; i++ {
		rules = append(rules, meta.Rule{ID: fmt.Sprintf("%c%02d", 'A'+i/bucketMax, i), Group: "Big", Severity: diagnostic.SeverityInfo})
	}
	fake(t, rules, nil, nil)
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run(root, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	contains(t, "sidebar", read(t, filepath.Join(root, sidebar)), `"text": "A"`, `"text": "B"`)
}
