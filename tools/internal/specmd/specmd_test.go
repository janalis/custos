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
