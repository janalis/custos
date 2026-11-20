//go:build unix

package main

import "testing"

func TestParsers(t *testing.T) {
	for _, tc := range []struct {
		parser, out string
		want        int
		fails       bool
	}{
		{"custos", `{"findings":[{},{}]}`, 2, false},
		{"custos", `no json`, 0, true},
		{"custos-fix", "custos: would fix 2 file(s), 7 edit(s)\n", 7, false},
		{"custos-fix", "custos: nothing to fix\n", 0, false},
		{"custos-fix", "boom", 0, true},
		{"phpstan", "Note: banner\n" + `{"totals":{"errors":1,"file_errors":3}}`, 4, false},
		{"phpstan", `{}`, 0, true},
		{"phpstan", ``, 0, true},
		{"psalm", `[{},{},{}]`, 3, false},
		{"phan", `[]`, 0, false},
		{"mago", `{"issues":[{}]}`, 1, false},
		{"mago", `{}`, 0, true},
		{"mago", `x`, 0, true},
		{"phpcs", `{"totals":{"errors":2,"warnings":5}}`, 7, false},
		{"phpcs", `{}`, 0, true},
		{"phpcs", `x`, 0, true},
		{"php-cs-fixer", `{"files":[{},{}]}`, 2, false},
		{"php-cs-fixer", `{}`, 0, true},
		{"php-cs-fixer", `x`, 0, true},
		{"rector", `{"totals":{"changed_files":4,"errors":1}}`, 4, false},
		{"rector", `{}`, 0, true},
		{"rector", `x`, 0, true},
		{"custos-format", "custos: 12 file(s) would be reformatted, 60 unchanged, 4 skipped\n", 12, false},
		{"custos-format", "custos: 0 file(s) would be reformatted, 76 unchanged, 0 skipped\n", 0, false},
		{"custos-format", "custos: boom\n", 0, true},
		// mago fmt --check logs on standard error.
		{"mago-format", " INFO Overriding workspace directory with /w/guzzle.\n INFO Found 68 file(s) need formatting. Run the command without '--check' to format them.\n", 68, false},
		{"mago-format", " INFO Found 1 file(s) that need formatting.\n", 1, false},
		{"mago-format", " INFO Overriding workspace directory with /w/guzzle.\n INFO All files are already formatted.\n", 0, false},
		{"mago-format", " ERROR Failed to parse configuration file\n", 0, true},
		{"pint", `{"tool":"pint","result":"fail","files":[{"path":"src\/Client.php","fixers":["single_line_empty_body"]},{"path":"src\/Utils.php","fixers":["concat_space","trailing_comma_in_multiline"]}]}`, 2, false},
		{"pint", `{"tool":"pint","result":"passed"}`, 0, false},
		{"pint", "Deprecated: f(): {x} is deprecated\n" + `{"tool":"pint","result":"fail","files":[{"path":"a.php","fixers":[]}]}`, 1, false},
		{"pint", `{"tool":"pint"}`, 0, true},
		{"pint", "Deprecated: x\n" + `{"about":"PHP CS Fixer","files":[{"name":"a.php"},{"name":"b.php"}],"time":{"total":0}}`, 2, false},
		{"pint", `{"about":"PHP CS Fixer","files":[]}`, 0, false},
		{"pint", `x`, 0, true},
	} {
		got, err := parsers[tc.parser]([]byte(tc.out))
		if (err != nil) != tc.fails || (!tc.fails && got != tc.want) {
			t.Errorf("%s(%q) = %d, %v; want %d, fails=%v", tc.parser, tc.out, got, err, tc.want, tc.fails)
		}
	}
}
