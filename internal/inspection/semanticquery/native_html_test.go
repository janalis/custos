package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeHTMLTextPrefix(t *testing.T) {
	for _, tc := range []struct {
		prefix string
		want   bool
	}{
		{"<p>", true},
		{"<p title=\"a>b\">", true},
		{"<p title='a>b'>", true},
		{"</p> text", true},
		{"<P> text", true},
		{"<p> <a href=\"unfinished >", false},
		{"<p title='unfinished >", false},
		{"<p> <a href=", false},
		{"plain >", false},
		{"<script><p>", false},
		{"<style><p>", false},
		{"<!-- >", false},
		{"<", false},
		{"</", false},
		{"<1>", false},
		{"<p <b>", false},
	} {
		t.Run(tc.prefix, func(t *testing.T) {
			if got := nativeHTMLTextPrefix(tc.prefix); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNativeHTMLTextInterpolationContext(t *testing.T) {
	for _, tc := range []struct {
		expr string
		want bool
	}{
		{`'<p>'.$_GET['x'].'</p>'`, true},
		{`'<p title="a>b">'.$_GET['x']`, true},
		{`'<p title="a>'.$_GET['x'].'">'`, false},
		{`'<p>'.'<a title="a>'.$_GET['x'].'">'`, false},
		{`'<script>'.('<p>'.$_GET['x'])`, false},
		{`'<sty'.'le>'.('<p>'.$_GET['x'])`, false},
		{`$_GET['prefix'].('<p>'.$_GET['x'])`, false},
		{`'<p>'.'trusted'.'</p>'`, false},
		{`'<p>'.htmlspecialchars($_GET['x'],ENT_QUOTES|ENT_SUBSTITUTE,'UTF-8')`, false},
		{`$_GET['x']`, false},
	} {
		t.Run(tc.expr, func(t *testing.T) {
			probeNative(t, "probe("+tc.expr+");", func(ctx *analysis.Context, call *syntax.FuncCall) {
				if got := nativeHTMLText(ctx, CallArgument(call.Args, 0, "value")); got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func BenchmarkNativeHTMLTextPrefix(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		nativeHTMLTextPrefix(`<p title="a>b">text`)
	}
}
