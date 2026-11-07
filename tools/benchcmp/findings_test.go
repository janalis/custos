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
	} {
		got, err := parsers[tc.parser]([]byte(tc.out))
		if (err != nil) != tc.fails || (!tc.fails && got != tc.want) {
			t.Errorf("%s(%q) = %d, %v; want %d, fails=%v", tc.parser, tc.out, got, err, tc.want, tc.fails)
		}
	}
}
