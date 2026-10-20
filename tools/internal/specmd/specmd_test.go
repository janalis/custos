package specmd

import (
	"reflect"
	"testing"
)

func TestSection(t *testing.T) {
	md := "# Kettle\n\n## Summary\n\nWhistles once.\n\n## Options\nNone.\n"
	if got := Section(md, "Summary"); got != "Whistles once." {
		t.Errorf("Summary = %q", got)
	}
	if got := Section(md, "Options"); got != "None." {
		t.Errorf("Options = %q", got)
	}
	if got := Section(md, "Examples"); got != "" {
		t.Errorf("Examples = %q", got)
	}
}

func TestPHP(t *testing.T) {
	if min, max := PHP("---\nphp: { min: \"7.1\", max: \"\" }\n---\n"); min != "7.1" || max != "" {
		t.Errorf("got %q, %q", min, max)
	}
	if min, max := PHP("# no front matter\n"); min != "" || max != "" {
		t.Errorf("got %q, %q", min, max)
	}
}

func TestBlocks(t *testing.T) {
	sec := "```php\n<?php\nf();\n```\n\n```php\n<?php\n```\nAfter a note:\n```json\n{}\n```\n```php\nunclosed\n"
	want := []Block{
		{Lang: "php", Body: "<?php\nf();\n"},
		{Lang: "php", Body: "<?php\n"},
		{Lang: "json", Body: "{}\n", Prose: "After a note:"},
	}
	if got := Blocks(sec); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v\nwant %#v", got, want)
	}
	if got := Blocks("Intro.\n```php\nx\n```"); len(got) != 1 || got[0].Prose != "Intro." {
		t.Errorf("leading prose: %#v", got)
	}
}

func TestReadable(t *testing.T) {
	sec := "- D1. Report a kettle left on (see E1).\n" +
		"- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.\n" +
		"  - D2a. Report it twice (D1, D3), unless E2 applies.\n" +
		"- E1. Cold kettles are fine. Cases outside D1 receive no finding.\n" +
		"- E2. Both `@custos-ignore K` and `@noinspection K` suppress this inspection. Empty kettles too."
	keep := References(sec)
	if !reflect.DeepEqual(keep, map[string]bool{"E1": true, "D1": true, "E2": true}) {
		t.Fatalf("references: %v", keep)
	}
	// Labels referred to stay; a reference to a dropped label goes.
	delete(keep, "D1")
	want := "- Report a kettle left on (see E1).\n  - Report it twice, unless E2 applies.\n- E1. Cold kettles are fine.\n- E2. Empty kettles too."
	if got := Readable(sec, keep); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if got := Readable("None. Use ordinary rule configuration to enable or disable this inspection.", nil); got != "None." {
		t.Errorf("options: %q", got)
	}
	if got := Readable("Checks this inspection's input (D2).", nil); got != "Checks this rule's input." {
		t.Errorf("rename: %q", got)
	}
}
