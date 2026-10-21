package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func auditNative(t *testing.T, id string, check func(*analysis.Context, syntax.Node, string), src string) []diagnostic.Finding {
	t.Helper()
	p := nativeCheckProbe{id, check}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{id}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("audit.php", []byte("<?php "+src), syntax.Options{}))
}

func TestNativeAuditIntentAndIdentity(t *testing.T) {
	for _, tc := range []struct {
		id, src string
		check   func(*analysis.Context, syntax.Node, string)
		want    int
	}{
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){$f=fn()=>access();}", CheckUntrustedForwardedClientAddress, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){ini_set('zlib.output_compression',true);header('Content-Length: '.mb_strlen($body));echo $body;}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen('é'));echo 'é';", CheckContentLengthUsesCharacterCount, 1},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen(1));echo 1;", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen(1.5));echo 1.5;", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen(true));echo true;", CheckContentLengthUsesCharacterCount, 0},
		{"UntrustedForwardedClientAddress", "if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){echo 'logged';}", CheckUntrustedForwardedClientAddress, 0},
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if($log){echo $_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1';access();}", CheckUntrustedForwardedClientAddress, 0},
		{"UntrustedForwardedClientAddress", "if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){grantAccess();}", CheckUntrustedForwardedClientAddress, 0},
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){access();}", CheckUntrustedForwardedClientAddress, 1},
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if(trustedProxy() && $_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){access();}", CheckUntrustedForwardedClientAddress, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body));echo $body;}", CheckContentLengthUsesCharacterCount, 1},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.grapheme_strlen($body));echo $body;}", CheckContentLengthUsesCharacterCount, 1},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body));echo 'different';}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body));$body='replacement';echo $body;}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){ob_start('compress');header('Content-Length: '.mb_strlen($body));echo $body;}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen('ascii'));echo 'ascii';", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body));echo $body,'tail';}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body));echo $body;echo 'tail';}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body));echo $body;transformOutput();}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body));print $body;}", CheckContentLengthUsesCharacterCount, 0},
	} {
		t.Run(tc.id+tc.src, func(t *testing.T) {
			got := auditNative(t, tc.id, tc.check, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func TestNativeAuditFixSyntax(t *testing.T) {
	for _, tc := range []struct {
		id, src, needle string
		check           func(*analysis.Context, syntax.Node, string)
		fixes           int
	}{
		{"RepeatedCookieHeaderReplacement", "header('Set-Cookie: a=1');header(header: 'Set-Cookie: b=2');", "replace: false", CheckRepeatedCookieHeaderReplacement, 1},
		{"RepeatedCookieHeaderReplacement", "header('Set-Cookie: a=1');header(header: 'Set-Cookie: b=2',);", "replace: false", CheckRepeatedCookieHeaderReplacement, 1},
		{"RepeatedCookieHeaderReplacement", "header('Set-Cookie: a=1');header('Set-Cookie: b=2',);", " false", CheckRepeatedCookieHeaderReplacement, 1},
		{"ContentLengthUsesCharacterCount", "namespace Site;function strlen($s){return 0;}function f($body){header('Content-Length: '.mb_strlen($body));echo $body;}", `\strlen($body)`, CheckContentLengthUsesCharacterCount, 1},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body,'invalid'));echo $body;}", "", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body,$encoding){header('Content-Length: '.mb_strlen($body,$encoding));echo $body;}", "", CheckContentLengthUsesCharacterCount, 0},
	} {
		t.Run(tc.id+tc.src, func(t *testing.T) {
			got := auditNative(t, tc.id, tc.check, tc.src)
			if len(got) != 1 || len(got[0].Fixes) != tc.fixes {
				t.Fatalf("got %+v, want one with %d fixes", got, tc.fixes)
			}
			if tc.fixes > 0 {
				edits := got[0].Fixes[0].Edits()
				found := false
				for _, edit := range edits {
					if strings.Contains(edit.NewText, tc.needle) {
						found = true
					}
				}
				if !found {
					t.Fatalf("edits %+v lack %q", edits, tc.needle)
				}
				if len(edits) != 1 {
					t.Fatalf("expected one reviewable edit, got %+v", edits)
				}
				edit := edits[0]
				original := "<?php " + tc.src
				fixed := original[:edit.Span.Start] + edit.NewText + original[edit.Span.End:]
				if parsed := syntax.Parse("fixed.php", []byte(fixed), syntax.Options{}); len(parsed.Errors) > 0 {
					t.Fatalf("fix does not parse: %s: %+v", fixed, parsed.Errors)
				}
			}
		})
	}
}

func BenchmarkNativeContentLengthBodyProof(b *testing.B) {
	p := nativeCheckProbe{"ContentLengthUsesCharacterCount", CheckContentLengthUsesCharacterCount}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("body.php", []byte("<?php function response($body){header('Content-Length: '.mb_strlen($body,'UTF-8'));echo $body;}"), syntax.Options{})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}

func TestNativeAuditSharedProofBoundaries(t *testing.T) {
	probeNative(t, "$x='same';probe($x,$x);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if !SameOperand(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("uniquely reaching assigned value lost")
		}
		if SameOperand(ctx, nil, CallArgument(c.Args, 0, "")) {
			t.Fatal("missing path must be unknown")
		}
	})
	probeNative(t, "probe($unknown,$unknown);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if SameOperand(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("unknown variables cannot prove identity")
		}
	})
	probeNative(t, "probe(mb_strlen($body,chooseEncoding()));", func(_ *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "").(*syntax.FuncCall)
		if ArgumentsRemovable(call) {
			t.Fatal("effectful optional argument must remain evaluated")
		}
	})
}
