package flow

import "testing"

func TestHTMLTransformSafety(t *testing.T) {
	for _, tc := range []struct {
		source string
		safe   bool
	}{
		{`probe(htmlspecialchars_decode(htmlspecialchars($_GET['v'])));`, false},
		{`probe(html_entity_decode(htmlentities($_GET['v'])));`, false},
		{`probe(urldecode(htmlspecialchars($_GET['v'])));`, false},
		{`probe(rawurldecode(htmlspecialchars($_GET['v'])));`, false},
		{`probe(base64_decode(htmlspecialchars($_GET['v'])));`, false},
		{`probe(hex2bin(htmlspecialchars($_GET['v'])));`, false},
		{`probe(str_replace('&lt;','<',htmlspecialchars($_GET['v'])));`, false},
		{`probe(strtoupper(htmlspecialchars($_GET['v'])));`, true},
		{`probe(trim(htmlspecialchars($_GET['v'])));`, true},
		{`probe(substr(htmlspecialchars($_GET['v']),0,20));`, true},
		{`probe(urlencode($_GET['v']));`, true},
		{`probe(rawurlencode($_GET['v']));`, true},
		{`probe(base64_encode($_GET['v']));`, true},
		{`probe(bin2hex($_GET['v']));`, true},
		{`probe(urldecode(rawurlencode($_GET['v'])));`, false},
		{`probe(base64_decode(base64_encode($_GET['v'])));`, false},
		{`probe(htmlspecialchars(html_entity_decode($_GET['v'])));`, true},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, f := environment(t, tc.source, nil)
			v := e.Value(argument(callsNamed(f, "probe")[0]))
			if got := v.Safe&HTML != 0; got != tc.safe {
				t.Fatalf("safe=%v want %v: %+v", got, tc.safe, v)
			}
		})
	}
}

func TestHTMLDecodedSummaries(t *testing.T) {
	for _, src := range []string{
		`function dec($x){return html_entity_decode($x);}probe(dec(htmlspecialchars($_GET['v'])));`,
		`function dec(&$x){$x=htmlspecialchars_decode($x);} $v=htmlspecialchars($_GET['v']);dec($v);probe($v);`,
		`function dec($x){return html_entity_decode($x);}function outer($x){return dec($x);}probe(outer(htmlspecialchars($_GET['v'])));`,
		`function dec($x,$flag){return $flag?$x:html_entity_decode($x);}probe(dec(htmlspecialchars($_GET['v']),$flag));`,
	} {
		t.Run(src, func(t *testing.T) {
			e, f := summarized(t, src)
			v := e.Value(argument(callsNamed(f, "probe")[0]))
			if v.Safe&(HTML|HTMLSingleQuoted|HTMLDoubleQuoted) != 0 || !v.Complete || len(v.Sources) == 0 {
				t.Fatalf("decoder summary restored escaping: %+v", v)
			}
		})
	}
}

func TestDecodedHTMLDependencies(t *testing.T) {
	sources := []Source{{Parameter: -1}, {Parameter: 0}, {Parameter: 1}}
	original := Value{Complete: true, Safe: HTML | HTMLSingleQuoted | HTMLDoubleQuoted | Shell, Sources: sources}
	got := decodedHTML(original)
	if got.Safe != Shell || sources[1].htmlDecoded || got.Sources[0].htmlDecoded || !got.Sources[1].htmlDecoded || !got.Sources[2].htmlDecoded {
		t.Fatalf("bad immutable dependency update: %+v; original=%+v", got, sources)
	}
	again := decodedHTML(got)
	if &again.Sources[0] != &got.Sources[0] {
		t.Fatal("already decoded sources were copied")
	}
}

func BenchmarkHTMLDecoding(b *testing.B) {
	for _, tc := range []struct {
		name string
		v    Value
	}{
		{"external", Value{Sources: []Source{{Parameter: -1}}, Safe: HTML}},
		{"symbolic", Value{Sources: []Source{{Parameter: 0}}, Safe: HTML}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				decodedHTML(tc.v)
			}
		})
	}
}
