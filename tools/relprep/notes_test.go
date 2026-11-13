package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatReleaseNotes(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{"prose", "First line\ncontinues here.\n\nAnother\nparagraph.", "First line continues here.\n\nAnother paragraph."},
		{"lists", "### Added\n\n- First\n  continuation.\n- Second\n  continuation.\n\n### Fixed\n\n1. Ordered\n   continuation.\n10. Longer\n    continuation.", "### Added\n\n- First continuation.\n- Second continuation.\n\n### Fixed\n\n1. Ordered continuation.\n10. Longer continuation."},
		{"nested", "- Parent\n  prose.\n  - Child\n    prose.\n- Next", "- Parent prose.\n  - Child prose.\n- Next"},
		{"fenced", "```php\none\ntwo\n```\n\nAfter\ncode.", "```php\none\ntwo\n```\n\nAfter code."},
		{"long fence", "````\n```\ninside\ncode\n````\nAfter\ncode.", "````\n```\ninside\ncode\n````\nAfter code."},
		{"tilde fence", "~~~php\none\n~~~not a closer\ntwo\n~~~\nAfter\ncode.", "~~~php\none\n~~~not a closer\ntwo\n~~~\nAfter code."},
		{"list fence", "- ```php\n  one\n  two\n  ```\n\n- Next\n  item.", "- ```php\n  one\n  two\n  ```\n\n- Next item."},
		{"indented", "    first\n    second\n\n\tthird\n\tfourth", "    first\n    second\n\n\tthird\n\tfourth"},
		{"hard breaks", "One  \nTwo\nthree\\\nFour\nfive", "One  \nTwo three\\\nFour five"},
		{"blocks", "# Heading\n> quote\n> next\n\n| A | B |\n| - | - |\n\n---\n***\n___\n===\n\n[ref]: https://example.test\n<div>\n</div>", "# Heading\n> quote\n> next\n\n| A | B |\n| - | - |\n\n---\n***\n___\n===\n\n[ref]: https://example.test\n<div>\n</div>"},
		{"inline ticks", "``inline``\ncontinues", "``inline`` continues"},
		{"empty list", "- \n  paragraph\n  continues", "- \n  paragraph continues"},
		{"CRLF", "\r\n- First\r\n  second.\r\n", "- First second."},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatReleaseNotes(tc.input)
			if got != tc.want {
				t.Fatalf("got %q; want %q", got, tc.want)
			}
			if formatReleaseNotes(got) != got {
				t.Fatal("formatting is not idempotent")
			}
			if strings.Join(strings.Fields(tc.input), " ") != strings.Join(strings.Fields(got), " ") {
				t.Fatal("formatting changed content")
			}
		})
	}
}

func TestFormatNotesCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.md")
	write(t, path, "- First\n  continuation.\n")
	var out, errs bytes.Buffer
	if code := run([]string{"-format-notes", path}, &out, &errs); code != 0 || out.String() != "- First continuation.\n" || errs.Len() != 0 {
		t.Fatalf("%d %q %q", code, out.String(), errs.String())
	}
	if read(t, path) != "- First\n  continuation.\n" {
		t.Fatal("input file was changed")
	}
	out.Reset()
	errs.Reset()
	if code := run([]string{"-format-notes", path + "missing"}, &out, &errs); code != 1 || !strings.Contains(errs.String(), "missing") {
		t.Fatalf("%d %s", code, &errs)
	}
	out.Reset()
	errs.Reset()
	if code := run([]string{"-format-notes", path, "-bump", "patch"}, &out, &errs); code != 2 {
		t.Fatalf("mixed modes accepted: %d", code)
	}
}

func TestReleaseNotesDoNotUnwrapChangelog(t *testing.T) {
	dir := repo(t)
	cl := "# Changelog\n\n## [Unreleased]\n\n### Fixed\n\n- First\n  continuation.\n"
	write(t, filepath.Join(dir, "CHANGELOG.md"), cl)
	path := filepath.Join(dir, "notes.md")
	if code, _, errs := runIn(dir, "-bump", "patch", "-notes", path); code != 0 {
		t.Fatal(errs)
	}
	if read(t, path) != "### Fixed\n\n- First continuation.\n" {
		t.Fatal(read(t, path))
	}
	if !strings.Contains(read(t, filepath.Join(dir, "CHANGELOG.md")), "- First\n  continuation.") {
		t.Fatal("changelog prose was unwrapped")
	}
}
