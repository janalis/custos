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
		{"UploadClientMimeTrusted", "function f(){move_uploaded_file($_FILES['a']['tmp_name'],'upload');}", CheckUploadClientMimeTrusted, 0},
		{"UploadClientMimeTrusted", "move_uploaded_file($_FILES['a']['name'],'upload');", CheckUploadClientMimeTrusted, 0},
		{"UploadClientMimeTrusted", "move_uploaded_file($_FILES['tmp_name'],'upload');", CheckUploadClientMimeTrusted, 0},
		{"UploadClientMimeTrusted", "move_uploaded_file($other['a']['tmp_name'],'upload');", CheckUploadClientMimeTrusted, 0},
		{"UploadClientMimeTrusted", "if('image/png'===$_FILES['a']['type']){move_uploaded_file($_FILES['a']['tmp_name'],'upload');}", CheckUploadClientMimeTrusted, 1},
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){$f=fn()=>access();}", CheckUntrustedForwardedClientAddress, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){ini_set('zlib.output_compression',true);header('Content-Length: '.mb_strlen($body));echo $body;}", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen('é'));echo 'é';", CheckContentLengthUsesCharacterCount, 1},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen(1));echo 1;", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen(1.5));echo 1.5;", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: '.mb_strlen(true));echo true;", CheckContentLengthUsesCharacterCount, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){$entry=new SplFileInfo('/safe/file');unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){if($flag){$entry=new SplFileInfo('/safe/file');}unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){$f=function(){$entry=1;};unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 1},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){unset($entry);unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){$entry++;unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){foreach($other as $entry){}unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 0},
		{"PdoPlaceholderBindingMismatch", "function f(PDO $p,$flag){$s=$p->prepare('SELECT :id');$flag && $s->bindValue(':id',1);$s->execute();}", CheckPdoPlaceholderBindingMismatch, 0},
		{"PdoPlaceholderBindingMismatch", "function f(PDO $p,$flag){$s=$p->prepare('SELECT :id');$flag ? $s->bindValue(':id',1) : 0;$s->execute();}", CheckPdoPlaceholderBindingMismatch, 0},
		{"ReadModifyWriteLockTooLate", "file_put_contents(nextPath(),file_get_contents(nextPath()).'x',LOCK_EX);", CheckReadModifyWriteLockTooLate, 0},
		{"ReadModifyWriteLockTooLate", "file_put_contents('file',file_get_contents('file').'x',LOCK_EX);", CheckReadModifyWriteLockTooLate, 1},
		{"UntrustedForwardedClientAddress", "if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){echo 'logged';}", CheckUntrustedForwardedClientAddress, 0},
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if($log){echo $_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1';access();}", CheckUntrustedForwardedClientAddress, 0},
		{"UntrustedForwardedClientAddress", "if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){grantAccess();}", CheckUntrustedForwardedClientAddress, 0},
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){access();}", CheckUntrustedForwardedClientAddress, 1},
		{"UntrustedForwardedClientAddress", "/** @custos-protected */ function access(){} if(trustedProxy() && $_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){access();}", CheckUntrustedForwardedClientAddress, 0},
		{"UploadClientMimeTrusted", "if($_FILES['a']['type']==='image/png'){move_uploaded_file($_FILES['a']['tmp_name'],'upload');}", CheckUploadClientMimeTrusted, 1},
		{"UploadClientMimeTrusted", "if($_FILES['a']['type']==='image/png'){move_uploaded_file($_FILES['b']['tmp_name'],'upload');}", CheckUploadClientMimeTrusted, 0},
		{"UploadClientMimeTrusted", "if($_FILES['a']['type']==='image/png'){echo 'rejected';}else{move_uploaded_file($_FILES['a']['tmp_name'],'upload');}", CheckUploadClientMimeTrusted, 0},
		{"UploadClientMimeTrusted", "if($log){echo $_FILES['a']['type'];move_uploaded_file($_FILES['a']['tmp_name'],'upload');}", CheckUploadClientMimeTrusted, 0},
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
		{"PasswordComparedWithFreshHash", "namespace Site;function password_verify($p,$h){return false;}function f($password,$stored){return $password===password_hash($stored,PASSWORD_DEFAULT);}", `\password_verify(`, CheckPasswordComparedWithFreshHash, 1},
		{"PdoExecuteArrayReplacesBindings", "function f(PDO $p){$s=$p->prepare('SELECT :a');$s->bindValue(':a',1);$s->execute([],sideEffect());}", "", CheckPdoExecuteArrayReplacesBindings, 0},
		{"ContentLengthUsesCharacterCount", "function f($body){header('Content-Length: '.mb_strlen($body,'invalid'));echo $body;}", "", CheckContentLengthUsesCharacterCount, 0},
		{"ContentLengthUsesCharacterCount", "function f($body,$encoding){header('Content-Length: '.mb_strlen($body,$encoding));echo $body;}", "", CheckContentLengthUsesCharacterCount, 0},
		{"PasswordComparedWithFreshHash", "function f($p,$stored){return $stored===password_hash($p,9999);}", "", CheckPasswordComparedWithFreshHash, 0},
		{"PasswordComparedWithFreshHash", "function f($p,$stored){return $stored===password_hash($p,PASSWORD_DEFAULT,['cost'=>-1]);}", "", CheckPasswordComparedWithFreshHash, 0},
		{"PasswordComparedWithFreshHash", "function f($p,$stored){return $stored===password_hash($p,PASSWORD_DEFAULT,[],1);}", "", CheckPasswordComparedWithFreshHash, 0},
		{"PasswordComparedWithFreshHash", "use function strlen as password_verify; function f($p,$stored){return $stored===password_hash($p,PASSWORD_DEFAULT);}", `\password_verify(`, CheckPasswordComparedWithFreshHash, 1},
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
		if !nativeSameExpression(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("uniquely reaching assigned value lost")
		}
		if nativeSameExpression(ctx, nil, CallArgument(c.Args, 0, "")) {
			t.Fatal("missing path must be unknown")
		}
	})
	probeNative(t, "probe($unknown,$unknown);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if nativeSameExpression(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("unknown variables cannot prove identity")
		}
	})
	probeNative(t, "probe(mb_strlen($body,chooseEncoding()));", func(_ *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "").(*syntax.FuncCall)
		if nativeLengthFixSafe(call) {
			t.Fatal("effectful optional argument must remain evaluated")
		}
	})
}

func BenchmarkNativeIteratorBindingProof(b *testing.B) {
	p := nativeCheckProbe{"DirectoryIteratorDotEntries", CheckDirectoryIteratorDotEntries}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("iterator.php", []byte("<?php foreach(new DirectoryIterator('/tmp') as $entry){if($entry->isDot()){continue;}unlink($entry->getPathname());}"), syntax.Options{})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
