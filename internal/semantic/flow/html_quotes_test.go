package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"

	phpversion "custos/internal/php/version"
)

func TestHTMLQuoteSafety(t *testing.T) {
	for _, tc := range []struct {
		source  string
		version phpversion.Version
		safe    Context
	}{
		{`$v=htmlspecialchars($_GET['v']);probe($v);`, phpversion.PHP53, HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v']);probe($v);`, phpversion.PHP74, HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v']);probe($v);`, phpversion.PHP81, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$v=htmlentities($_GET['v']);probe($v);`, 0, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v'],ENT_COMPAT);probe($v);`, phpversion.PHP85, HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v'],ENT_QUOTES|ENT_SUBSTITUTE);probe($v);`, phpversion.PHP85, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v'],ENT_NOQUOTES);probe($v);`, phpversion.PHP85, 0},
		{`$flags=ENT_QUOTES;$v=htmlspecialchars($_GET['v'],0|$flags);probe($v);`, phpversion.PHP85, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$flags=$c?ENT_COMPAT:ENT_QUOTES;$v=htmlspecialchars($_GET['v'],$flags);probe($v);`, phpversion.PHP85, 0},
		{`$flags=3;$v=htmlspecialchars($_GET['v'],$flags);probe($v);`, phpversion.PHP85, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$v=htmlspecialchars(flags:3,string:$_GET['v']);probe($v);`, phpversion.PHP85, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v'],$flags);probe($v);`, phpversion.PHP85, 0},
		{`$v=htmlspecialchars($_GET['v'],MISSING_FLAGS);probe($v);`, phpversion.PHP85, 0},
		{`const FLAGS='unknown';$v=htmlspecialchars($_GET['v'],FLAGS);probe($v);`, phpversion.PHP85, 0},
		{`const FLAGS=3;$v=htmlspecialchars($_GET['v'],FLAGS);probe($v);`, phpversion.PHP85, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v'],3 & $flags);probe($v);`, phpversion.PHP85, 0},
		{`$v=htmlspecialchars($_GET['v'],3 | $flags);probe($v);`, phpversion.PHP85, 0},
		{`namespace App;$v=htmlspecialchars($_GET['v'],ENT_QUOTES);probe($v);`, phpversion.PHP85, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`namespace App;const ENT_QUOTES=0;$v=htmlspecialchars($_GET['v'],ENT_QUOTES);probe($v);`, phpversion.PHP85, 0},
		{`$v=htmlspecialchars($_GET['v'],...$args);probe($v);`, phpversion.PHP85, 0},
		{`$v=$c?htmlspecialchars($_GET['v'],ENT_QUOTES):htmlspecialchars($_GET['v'],ENT_COMPAT);probe($v);`, phpversion.PHP85, HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v'],` + strings.Repeat("(", 17) + `ENT_QUOTES` + strings.Repeat(")", 17) + `);probe($v);`, phpversion.PHP85, HTMLSingleQuoted | HTMLDoubleQuoted},
		{`$v=htmlspecialchars($_GET['v'],` + strings.Repeat("0|", 17) + `ENT_QUOTES);probe($v);`, phpversion.PHP85, 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, f := environment(t, tc.source, nil)
			e.types.PHP = tc.version
			v := e.Value(argument(callsNamed(f, "probe")[0]))
			if got := v.Safe & (HTMLSingleQuoted | HTMLDoubleQuoted); got != tc.safe {
				t.Fatalf("safe=%v want=%v value=%+v", got, tc.safe, v)
			}
			if v.Safe&HTML == 0 || v.Safe&URL != 0 {
				t.Fatalf("mixed text and URL contracts: %+v", v)
			}
		})
	}
}

func TestHTMLQuoteSafetySummaries(t *testing.T) {
	for _, src := range []string{
		`function escapeLabel($x){return htmlspecialchars($x,ENT_QUOTES);} probe(escapeLabel($_GET['v']));`,
		`function escapeLabel(&$x){$x=htmlspecialchars($x,ENT_QUOTES);} $x=$_GET['v'];escapeLabel($x);probe($x);`,
	} {
		e, f := summarized(t, src)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		if v.Safe&(HTMLSingleQuoted|HTMLDoubleQuoted) != (HTMLSingleQuoted | HTMLDoubleQuoted) {
			t.Fatalf("summary dropped quote proof: %+v", v)
		}
	}
}

func BenchmarkHTMLQuoteSafety(b *testing.B) {
	e, f := environment(b, `probe(htmlspecialchars($_GET['v'],ENT_QUOTES|ENT_SUBSTITUTE));`, nil)
	arg := argument(callsNamed(f, "probe")[0])
	e.Value(arg)
	r := e.scope(nil)
	flags := r.values[callsNamed(f, "htmlspecialchars")[0].Args.Args[1].(*syntax.Arg).Value]
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.htmlQuoteSafety(flags, false, r)
	}
}

func TestHTMLQuoteTransformContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		safe   bool
	}{
		{`probe(strtoupper(htmlspecialchars($_GET['v'],ENT_QUOTES)));`, true},
		{`probe(htmlspecialchars_decode(htmlspecialchars($_GET['v'],ENT_QUOTES),ENT_QUOTES));`, false},
		{`probe(urldecode(htmlspecialchars(rawurlencode($_GET['v']),ENT_QUOTES)));`, false},
		{`probe(bin2hex($_GET['v']));`, true},
		{`probe(base64_encode($_GET['v']));`, true},
		{`probe(urlencode($_GET['v']));`, true},
		{`probe(rawurlencode($_GET['v']));`, true},
	} {
		e, f := environment(t, tc.source, nil)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		if got := v.Safe&(HTMLSingleQuoted|HTMLDoubleQuoted) == HTMLSingleQuoted|HTMLDoubleQuoted; got != tc.safe {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
}

func TestHTMLFlagBudgetStatus(t *testing.T) {
	e, f := environment(t, `$v=htmlspecialchars($_GET['v'],`+strings.Repeat("0|", 17)+`ENT_QUOTES);probe($v);`, nil)
	e.Value(argument(callsNamed(f, "probe")[0]))
	for _, gap := range e.Gaps() {
		if gap.Reason == "flow-budget" && gap.Span.Len() > 0 {
			return
		}
	}
	t.Fatal("missing bounded HTML flag analysis status")
}
