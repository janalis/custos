package flow

import "testing"

func TestNativeTransformContracts(t *testing.T) {
	const a = `$a=$_GET['a'];`
	for _, tc := range []struct {
		name, src string
		context   Context
		tainted   bool
	}{
		{"number format", a + `probe(number_format($a));`, HTML, false},
		{"number format with decimals", a + `probe(number_format($a, 2));`, HTMLDoubleQuoted, false},
		{"number format with separators", a + `probe(number_format($a, 2, $a, '.'));`, HTML, true},
		{"number format with a named separator", a + `probe(number_format($a, thousands_separator: $a));`, HTML, true},
		{"number format behind a spread", a + `probe(number_format($a, ...$o));`, HTML, true},
		{"validated integer", a + `probe(filter_var($a, FILTER_VALIDATE_INT));`, SQL, false},
		{"validated boolean", a + `probe(filter_var($a, \FILTER_VALIDATE_BOOLEAN));`, Shell, false},
		{"escaped special characters", a + `probe(filter_var($a, FILTER_SANITIZE_SPECIAL_CHARS));`, HTMLSingleQuoted, false},
		{"escaped special characters in SQL", a + `probe(filter_var($a, FILTER_SANITIZE_SPECIAL_CHARS));`, SQL, true},
		{"namespaced fallback", `namespace App; ` + a + `probe(filter_var($a, FILTER_VALIDATE_INT));`, HTML, false},
		{"default filter", a + `probe(filter_var($a));`, HTML, true},
		{"raw filter", a + `probe(filter_var($a, FILTER_UNSAFE_RAW));`, HTML, true},
		{"computed filter", a + `probe(filter_var($a, $f));`, HTML, true},
		{"project constant", `const FILTER_MINE = 257; ` + a + `probe(filter_var($a, FILTER_MINE));`, HTML, true},
		{"unknown constant", a + `probe(filter_var($a, FILTER_MISSING));`, HTML, true},
		{"line breaks kept escaped", a + `probe(nl2br(htmlspecialchars($a, ENT_QUOTES)));`, HTMLDoubleQuoted, false},
		{"joined escaped pieces", a + `probe(implode(',', [htmlspecialchars($a, ENT_QUOTES)]));`, HTMLSingleQuoted, false},
		{"formatted text adds quotes", a + `probe(sprintf('%s', htmlspecialchars($a, ENT_QUOTES)));`, HTMLDoubleQuoted, true},
		{"flags behind a spread", a + `probe(htmlspecialchars($a, ...$o));`, HTMLDoubleQuoted, true},
		{"flags behind a spread still escape markup", a + `probe(htmlspecialchars($a, ...$o));`, HTML, false},
		{"quoted for PDO", a + `$db = new PDO('sqlite::memory:'); probe($db->quote($a));`, SQL, false},
		{"JSON in a script", a + `probe(json_encode($a));`, HTMLScript, false},
		{"JSON with slashes left alone", a + `probe(json_encode($a, JSON_UNESCAPED_SLASHES));`, HTMLScript, true},
		{"JSON with tags escaped", a + `probe(json_encode($a, JSON_UNESCAPED_SLASHES | JSON_HEX_TAG));`, HTMLScript, false},
		{"JSON with computed flags", a + `probe(json_encode($a, $flags));`, HTMLScript, true},
		{"JSON flags behind a spread", a + `probe(json_encode($a, ...$o));`, HTMLScript, true},
		{"JSON is no HTML", a + `probe(json_encode($a));`, HTMLDoubleQuoted, true},
		{"JSON escaped for an attribute", a + `probe(htmlspecialchars(json_encode($a), ENT_QUOTES));`, HTMLScript | HTMLDoubleQuoted, false},
		{"escaped text is no script", a + `probe(htmlspecialchars($a, ENT_QUOTES));`, HTMLScript | HTMLDoubleQuoted, true},
		{"URL encoding starts no scheme", a + `probe(rawurlencode($a));`, HTMLURL | HTMLUnquoted, false},
		{"base64 may end an unquoted value", a + `probe(base64_encode($a));`, HTMLUnquoted, true},
		{"base64 starts no scheme", a + `probe(base64_encode($a));`, HTMLURL, false},
		{"numbers are safe for style", a + `probe((int) $a . '');`, HTMLRaw, false},
		{"digits checked", `function f(){` + a + `if (!ctype_digit($a)) { return; } probe($a);}`, HTMLDoubleQuoted, false},
		{"integer checked", `function f(){` + a + `if (is_int($a)) { probe($a); }}`, Header, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			calls := callsNamed(f, "probe")
			if got := e.Tainted(argument(calls[len(calls)-1]), tc.context); got != tc.tainted {
				t.Fatalf("tainted=%v value=%+v", got, e.Value(argument(calls[len(calls)-1])))
			}
		})
	}
}

func TestReshapes(t *testing.T) {
	if !Reshapes("trim") || !Reshapes("substr") || Reshapes("nl2br") || Reshapes("htmlspecialchars") || Reshapes("unknown") {
		t.Fatal("reshaping transforms")
	}
}
