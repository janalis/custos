package ambiguousreplacementbackreference_test

import (
	"strings"
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/ambiguousreplacementbackreference"
	"custos/internal/php/syntax"
)

func TestCompleteLiteralFix(t *testing.T) {
	for _, tc := range []struct{ name, replacement, want string }{
		{"multiple", "'$11 and $12'", "'${1}1 and ${1}2'"},
		{"double quoted", `"\$11"`, `"\${1}1"`},
		{"double literal dollar", `"$11"`, `"\${1}1"`},
		{"quote and escapes preserved", `'it\'s $11\n'`, `'it\'s ${1}1\n'`},
		{"double escape preservation", `"line\n\$11\t\$12"`, `"line\n\${1}1\t\${1}2"`},
		{"escaped trailing dollar", `"\$11\$x"`, `"\${1}1\$x"`},
		{"escaped PCRE dollar preserved", `'\$11 and $12'`, `'\$11 and ${1}2'`},
		{"double escaped PCRE dollar preserved", `"\\$11 and \$12"`, `"\\$11 and \${1}2"`},
		{"double even PCRE slashes", `"\\\\$11"`, `"\\\\\${1}1"`},
		{"binary prefix", `b'$11'`, `b'${1}1'`},
		{"eleven references", "'" + strings.Repeat("$11 ", 11) + "'", "'" + strings.Repeat("${1}1 ", 11) + "'"},
		{"encoded dollar withheld", `"\x2411"`, `"\x2411"`},
		{"encoded digit withheld", `"\$1\x31"`, `"\$1\x31"`},
		{"nowdoc withheld", "<<<'TEXT'\n$11\nTEXT", "<<<'TEXT'\n$11\nTEXT"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			prefix := "<?php $result=preg_replace('/(r)/',"
			suffix := ",'r');"
			result := fixing.FixSource(e, "test.php", []byte(prefix+tc.replacement+suffix), fixing.Options{SinglePass: true})
			if string(result.Source) != prefix+tc.want+suffix {
				t.Fatalf("got %s want %s", result.Source, prefix+tc.want+suffix)
			}
			if tc.replacement != tc.want {
				if result.Applied != 1 {
					t.Fatalf("applied %d", result.Applied)
				}
				parsed := syntax.Parse("test.php", result.Source, syntax.Options{})
				if len(parsed.Errors) != 0 {
					t.Fatalf("parse errors: %+v", parsed.Errors)
				}
				if findings := e.Analyze(parsed); len(findings) != 0 {
					t.Fatalf("remaining findings: %+v", findings)
				}
			}
		})
	}
}

func BenchmarkCompleteLiteralFix(b *testing.B) {
	e, _ := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
	f := syntax.Parse("bench.php", []byte("<?php preg_replace('/(r)/','$11 and $12','r');"), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(f)
	}
}
