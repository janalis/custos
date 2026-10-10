package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

func TestNativeSQLLexicalFacts(t *testing.T) {
	for _, tc := range []struct {
		sql                         string
		known                       bool
		markers, positional, quoted int
		identifiers                 bool
	}{
		{"SELECT :id", true, 1, 0, 0, false},
		{"SELECT $$:quoted$$", false, 0, 0, 0, false},
		{"SELECT /* outer /* inner */ */ :id", false, 0, 0, 0, false},
		{"SELECT data ?? 'key' FROM items WHERE id=:id", false, 0, 0, 0, false},
		{"SELECT ? + ?", true, 2, 2, 0, false},
		{"SELECT ':name', :actual", true, 1, 0, 1, false},
		{"SELECT ':missing", false, 0, 0, 0, false},
		{"SELECT /* :ignored ? */ :id -- :ignore ?\n", true, 1, 0, 0, false},
		{"SELECT # :ignored ?\n :id", true, 1, 0, 0, false},
		{"SELECT /* unfinished", false, 0, 0, 0, false},
		{"SELECT 'it''s :literal', `:column`, ':bad name'", true, 0, 0, 0, false},
		{"SELECT 'slash\\quoted', \"double\"", true, 0, 0, 0, false},
		{"SELECT * FROM :table", true, 1, 0, 0, true},
		{"SELECT * FROM things JOIN :joined ON 1=1", true, 1, 0, 0, true},
		{"SELECT CAST(:id AS int)::text", true, 1, 0, 0, false},
		{"UPDATE :table SET label=:label", true, 2, 0, 0, true},
		{"SELECT ':', ':1bad', 'ordinary'", true, 0, 0, 0, false},
	} {
		t.Run(tc.sql, func(t *testing.T) {
			f, known := scanNativeSQL(tc.sql)
			if known != tc.known || len(f.markers) != tc.markers || f.positional != tc.positional || len(f.quoted) != tc.quoted || f.identifiers != tc.identifiers {
				t.Fatalf("got %+v known %v, want %+v", f, known, tc)
			}
		})
	}
	for _, tc := range []struct {
		s    string
		want bool
	}{{"", false}, {"name_2", true}, {"2name", false}, {"a-b", false}, {"_name", true}} {
		if got := nativeSQLName(tc.s); got != tc.want {
			t.Fatalf("name %q: got %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestNativeStatementDominance(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"earlier(); probe();", true},
		{"if($flag){earlier();} probe();", false},
		{"if(earlier()){probe();}", false},
		{"probe(); earlier();", false},
		{"if($flag){earlier(); probe();}", true},
		{"$flag && earlier(); probe();", false},
		{"$flag || earlier(); probe();", false},
		{"$flag ?? earlier(); probe();", false},
		{"$flag ? earlier() : 0; probe();", false},
		{"earlier() && $flag; probe();", true},
		{"$callback = function(){earlier();}; probe();", false},
		{"$callback = fn()=>earlier(); probe();", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			var earlier *syntax.FuncCall
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
					if call, ok := n.(*syntax.FuncCall); ok {
						if name, ok := call.Name.(*syntax.Name); ok && name.Value == "earlier" {
							earlier = call
						}
					}
					return true
				})
				if earlier == nil {
					t.Fatal("missing earlier call")
				}
				if got := NativeDominates(earlier, c); got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeDeterministicCleanupProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"function probe($value){} $h=fopen('path','rb');try{probe($h);}finally{fclose($h);}", true},
		{"$h=fopen('path','rb');try{probe($h);}finally{fclose($h);}", false},
		{"$h=fopen('path','rb');try{probe($h);}finally{echo 'done';}", false},
		{"$h=fopen('path','rb');try{probe($h);}finally{1+2;}", false},
		{"$h=fopen('path','rb');try{probe($h);}finally{fclose($other);}", false},
		{"$h=fopen('path','rb');probe($h);", false},
		{"probe();", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				if got := nativeFinally(ctx, c, CallArgument(c.Args, 0, "value"), "fclose"); got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeCommentPreservation(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{{"probe(1);", true}, {"probe(/*keep*/1);", false}, {"probe(1 //keep\n);", false}, {"probe(1 #keep\n);", false}} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				if got := nativeCleanEdit(ctx, c.Span()); got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

// nativeCheckProbe exercises the named query API on ordinary parser nodes.
// Checks accept generic nodes and must safely ignore unrelated syntax.
type nativeCheckProbe struct {
	id    string
	check func(*analysis.Context, syntax.Node, string)
}

func (p nativeCheckProbe) ID() string { return p.id }
func (nativeCheckProbe) Semantic()    {}
func (nativeCheckProbe) Flow()        {}
func (nativeCheckProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KEcho, syntax.KReturn, syntax.KForeach, syntax.KBinary, syntax.KAssign, syntax.KIncDec}
}

func (p nativeCheckProbe) Check(ctx *analysis.Context, n syntax.Node) {
	p.check(ctx, n, "Native API contract violation.")
}

func TestNativeChecksIgnoreUnrelatedSyntax(t *testing.T) {
	for _, tc := range []struct {
		id    string
		check func(*analysis.Context, syntax.Node, string)
	}{
		{"PdoPlaceholderBindingMismatch", CheckPdoPlaceholderBindingMismatch},
		{"PdoMixedPlaceholderStyles", CheckPdoMixedPlaceholderStyles},
		{"PdoQuotedPlaceholder", CheckPdoQuotedPlaceholder},
		{"PdoIdentifierPlaceholder", CheckPdoIdentifierPlaceholder},
		{"PdoSelectRowCountAssumption", CheckPdoSelectRowCountAssumption},
		{"PdoReferenceBindingVariableReuse", CheckPdoReferenceBindingVariableReuse},
		{"TransactionEarlyReturn", CheckTransactionEarlyReturn},
		{"TransactionExceptionWithoutRollback", CheckTransactionExceptionWithoutRollback},
		{"PdoFetchColumnFalsyValueLoss", CheckPdoFetchColumnFalsyValueLoss},
		{"PdoExecuteArrayReplacesBindings", CheckPdoExecuteArrayReplacesBindings},
		{"StreamUseAfterClose", CheckStreamUseAfterClose},
		{"OwnedStreamNotClosed", CheckOwnedStreamNotClosed},
		{"StreamOpenFailureUnchecked", CheckStreamOpenFailureUnchecked},
		{"FileLockFailureUnchecked", CheckFileLockFailureUnchecked},
		{"FileTruncatedBeforeLock", CheckFileTruncatedBeforeLock},
		{"ReadModifyWriteLockTooLate", CheckReadModifyWriteLockTooLate},
		{"TempnamDirectoryFallbackUnchecked", CheckTempnamDirectoryFallbackUnchecked},
		{"GlobFailureUnchecked", CheckGlobFailureUnchecked},
		{"DirectoryIteratorDotEntries", CheckDirectoryIteratorDotEntries},
		{"PathContainmentPrefixBoundary", CheckPathContainmentPrefixBoundary},
		{"HeadersAfterCommittedOutput", CheckHeadersAfterCommittedOutput},
		{"RedirectContinuesProtectedExecution", CheckRedirectContinuesProtectedExecution},
		{"SessionMutationAfterClose", CheckSessionMutationAfterClose},
		{"SessionLockHeldDuringBlockingCall", CheckSessionLockHeldDuringBlockingCall},
		{"SameSiteNoneWithoutSecure", CheckSameSiteNoneWithoutSecure},
		{"SessionCookieOptionsSetAfterStart", CheckSessionCookieOptionsSetAfterStart},
		{"RepeatedCookieHeaderReplacement", CheckRepeatedCookieHeaderReplacement},
		{"UntrustedForwardedClientAddress", CheckUntrustedForwardedClientAddress},
		{"ContentLengthUsesCharacterCount", CheckContentLengthUsesCharacterCount},
		{"NoContentResponseWithBody", CheckNoContentResponseWithBody},
		{"UntrustedSqlConstruction", CheckUntrustedSQLConstruction},
		{"UntrustedShellCommand", CheckUntrustedShellCommand},
		{"UnescapedHtmlOutput", CheckUnescapedHTMLOutput},
		{"UntrustedHeaderValue", CheckUntrustedHeaderValue},
		{"UnvalidatedRedirectTarget", CheckUnvalidatedRedirectTarget},
		{"UntrustedNetworkDestination", CheckUntrustedNetworkDestination},
		{"UntrustedFilesystemPath", CheckUntrustedFilesystemPath},
		{"UploadClientMimeTrusted", CheckUploadClientMimeTrusted},
		{"PasswordComparedWithFreshHash", CheckPasswordComparedWithFreshHash},
		{"FastDigestUsedForPasswordStorage", CheckFastDigestUsedForPasswordStorage},
	} {
		t.Run(tc.id, func(t *testing.T) {
			p := nativeCheckProbe{tc.id, tc.check}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{tc.id}})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte("<?php harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();")
			got := e.Analyze(syntax.Parse("unrelated.php", src, syntax.Options{}))
			if len(got) != 0 {
				t.Fatalf("unrelated nodes produced findings: %+v", got)
			}
		})
	}
}

func TestNativeWrapperSecurityContracts(t *testing.T) {
	for _, tc := range []struct {
		id, src string
		check   func(*analysis.Context, syntax.Node, string)
		want    int
	}{
		{"UntrustedFilesystemPath", "function openPath($p){readfile($p);} openPath($_GET['file']);", CheckUntrustedFilesystemPath, 1},
		{"UntrustedHeaderValue", "function emitHeader($s){header($s);} emitHeader($_GET['header']);", CheckUntrustedHeaderValue, 1},
		{"UntrustedNetworkDestination", "function redirectTo($u){header('Location: '.$u);} redirectTo($_GET['next']);", CheckUntrustedNetworkDestination, 0},
		{"UnvalidatedRedirectTarget", "function fetchURL($u){$h=curl_init($u);curl_exec($h);} fetchURL($_GET['url']);", CheckUnvalidatedRedirectTarget, 0},
		{"UntrustedNetworkDestination", "function fetchURL($u){$h=curl_init($u);curl_exec($h);} fetchURL($_GET['url']);", CheckUntrustedNetworkDestination, 1},
		{"UnvalidatedRedirectTarget", "function redirectTo($u){header('Location: '.$u);} redirectTo($_GET['next']);", CheckUnvalidatedRedirectTarget, 1},
		{"UntrustedSqlConstruction", "function selectSQL(PDO $p,$q){$p->query($q);} function action(PDO $p){selectSQL($p,$_GET['sql']);}", CheckUntrustedSQLConstruction, 1},
		{"UntrustedFilesystemPath", "function openPath($p){readfile($p);} openPath(unknownValidator($_GET['file']));", CheckUntrustedFilesystemPath, 0},
		{"UntrustedFilesystemPath", "function openPath($p){readfile($p);} function outer($p){openPath($p);}", CheckUntrustedFilesystemPath, 0},
		{"UntrustedSqlConstruction", "function action(PDO $p){$p->exec(statement:$_GET['sql']);}", CheckUntrustedSQLConstruction, 1},
		{"UntrustedNetworkDestination", "$h=curl_init('https://example.invalid');curl_setopt($h,CURLOPT_URL,$_GET['url']);curl_exec($h);", CheckUntrustedNetworkDestination, 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			p := nativeCheckProbe{tc.id, tc.check}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{tc.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("wrapper.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestNativePDOBindingProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		id, src string
		check   func(*analysis.Context, syntax.Node, string)
		want    int
	}{
		{"PdoPlaceholderBindingMismatch", "function f(PDO $p){$s=$p->prepare('SELECT :a,:b');$s->execute(['a'=>1]);}", CheckPdoPlaceholderBindingMismatch, 1},
		{"PdoPlaceholderBindingMismatch", "function f(PDO $p){$s=$p->prepare('SELECT ?');$s->bindValue(1,7);$s->execute();}", CheckPdoPlaceholderBindingMismatch, 0},
		{"PdoPlaceholderBindingMismatch", "function f(PDO $p,$name){$s=$p->prepare('SELECT :a');$s->bindValue($name,7);$s->execute();}", CheckPdoPlaceholderBindingMismatch, 0},
		{"PdoPlaceholderBindingMismatch", "function f(PDO $p){$s=$p->prepare('SELECT :a');retain($s);$s->execute();}", CheckPdoPlaceholderBindingMismatch, 0},
		{"PdoExecuteArrayReplacesBindings", "function f(PDO $p,$sql){$s=$p->prepare($sql);$s->bindValue(':a',7);$s->execute([]);}", CheckPdoExecuteArrayReplacesBindings, 0},
		{"PdoIdentifierPlaceholder", "function f(PDO $p,$sql){$p->prepare($sql);}", CheckPdoIdentifierPlaceholder, 0},
		{"PdoIdentifierPlaceholder", "function f(PDO $p){$p->prepare(\"SELECT ':unfinished\");}", CheckPdoIdentifierPlaceholder, 0},
		{"PdoMixedPlaceholderStyles", "function f(PDO $p,$sql){$p->prepare($sql);}", CheckPdoMixedPlaceholderStyles, 0},
		{"PdoMixedPlaceholderStyles", "function f(PDO $p){$p->prepare(\"SELECT ':unfinished\");}", CheckPdoMixedPlaceholderStyles, 0},
		{"PdoQuotedPlaceholder", "function f(PDO $p,$sql){$s=$p->prepare($sql);$s->execute(['a'=>7]);}", CheckPdoQuotedPlaceholder, 0},
		{"PdoQuotedPlaceholder", "function f(PDO $p,$args){$s=$p->prepare(\"SELECT ':a'\");$s->execute($args);}", CheckPdoQuotedPlaceholder, 0},
		{"PdoReferenceBindingVariableReuse", "function f(PDO $p,$sql){$s=$p->prepare($sql);$v=1;$s->bindParam(':a',$v);$v=2;$s->bindParam(':b',$v);$s->execute();}", CheckPdoReferenceBindingVariableReuse, 0},
		{"PdoReferenceBindingVariableReuse", "function f(PDO $p,$args){$s=$p->prepare('SELECT :a,:b');$v=1;$s->bindParam(':a',$v);$v=2;$s->bindParam(':b',$v);$s->execute($args);}", CheckPdoReferenceBindingVariableReuse, 0},
		{"PdoReferenceBindingVariableReuse", "function f(PDO $p,$name){$s=$p->prepare('SELECT :a,:b');$v=1;$s->bindParam($name,$v);$s->execute();}", CheckPdoReferenceBindingVariableReuse, 0},
		{"PdoReferenceBindingVariableReuse", "function f(PDO $p,$v){$s=$p->prepare('SELECT :a,:b');$s->bindParam(':a',$v);$s->execute();}", CheckPdoReferenceBindingVariableReuse, 0},
		{"PdoReferenceBindingVariableReuse", "function f(PDO $p){$s=$p->prepare('SELECT :a,:b');$v='a';$s->bindParam(':a',$v);$v='b';$s->bindParam(':b',$v);$s->execute();}", CheckPdoReferenceBindingVariableReuse, 1},
		{"PdoPlaceholderBindingMismatch", "function f(PDOStatement $s){$s->execute(['a'=>1]);}", CheckPdoPlaceholderBindingMismatch, 0},
		{"TransactionEarlyReturn", "function f(PDO $p){try{if($p->beginTransaction()){throw new Exception();}}catch(Exception $e){return;}}", CheckTransactionEarlyReturn, 0},
		{"PdoFetchColumnFalsyValueLoss", "function f(PDOStatement $s){$v=(bool)$s->fetchColumn();}", CheckPdoFetchColumnFalsyValueLoss, 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			p := nativeCheckProbe{tc.id, tc.check}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{tc.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("pdo.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestNativeLengthFixEvaluation(t *testing.T) {
	for _, tc := range []struct {
		src   string
		fixes int
		want  int
	}{
		{"header('Content-Length: '.mb_strlen($body,'UTF-8'));", 1, 1},
		{"header('Content-Length: '.mb_strlen($body,chooseEncoding()));", 0, 1},
		{"header('Content-Length: '.mb_strlen(...$args));", 0, 0},
		{"header('Content-Length: '.mb_strlen($body,&$encoding));", 0, 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			p := nativeCheckProbe{"ContentLengthUsesCharacterCount", CheckContentLengthUsesCharacterCount}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("length.php", []byte("<?php function f($body,$encoding,$args){"+tc.src+"echo $body;}"), syntax.Options{}))
			if len(got) != tc.want || (len(got) > 0 && len(got[0].Fixes) != tc.fixes) {
				t.Fatalf("findings %+v, want one with %d fixes", got, tc.fixes)
			}
		})
	}
}

func TestNativeConservativeOperationalBoundaries(t *testing.T) {
	for _, tc := range []struct {
		id, src string
		check   func(*analysis.Context, syntax.Node, string)
		want    int
	}{
		{"RepeatedCookieHeaderReplacement", "header('Set-Cookie: a=b');header('Set-Cookie: a=b');", CheckRepeatedCookieHeaderReplacement, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){if($other){continue;}unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 1},
		{"ReadModifyWriteLockTooLate", "file_put_contents('/tmp/f',function(){return 'x';},LOCK_EX);", CheckReadModifyWriteLockTooLate, 0},
		{"ContentLengthUsesCharacterCount", "header('Content-Length: 2'); header($prefix.mb_strlen($body));", CheckContentLengthUsesCharacterCount, 0},
		{"RepeatedCookieHeaderReplacement", "header($unknown);", CheckRepeatedCookieHeaderReplacement, 0},
		{"SameSiteNoneWithoutSecure", "setcookie('a','b',['samesite'=>'None']);", CheckSameSiteNoneWithoutSecure, 1},
		{"UntrustedForwardedClientAddress", "if($a===$b){}", CheckUntrustedForwardedClientAddress, 0},
		{"RedirectContinuesProtectedExecution", "/** @custos-protected */ function protectedCall(){} function f(){if($auth){header('Location: /');}if(!isAuthorized()){header('Location: /');}protectedCall();}", CheckRedirectContinuesProtectedExecution, 0},
		{"FastDigestUsedForPasswordStorage", "/** @custos-credential-store */ function store($v){} function f($p){store(md5(unknown($p)));}", CheckFastDigestUsedForPasswordStorage, 0},
		{"FastDigestUsedForPasswordStorage", "/** @custos-credential-store */ function store($v){} function f(){store(md5($_GET['p']));}", CheckFastDigestUsedForPasswordStorage, 0},
		{"FastDigestUsedForPasswordStorage", "/** @custos-credential-store */ function store($v){} store(md5($_GET['p']));", CheckFastDigestUsedForPasswordStorage, 0},
		{"FastDigestUsedForPasswordStorage", "/** @custos-credential-store */ function store($v){} function f($p){store(password_hash($p,PASSWORD_DEFAULT));}", CheckFastDigestUsedForPasswordStorage, 0},
		{"FastDigestUsedForPasswordStorage", "/** @custos-credential-store */ function store($v){} /** @param password-string $p */ function f($p){store(hash('sha256',$p));}", CheckFastDigestUsedForPasswordStorage, 1},
		{"PasswordComparedWithFreshHash", "$v=password_hash($p,PASSWORD_DEFAULT)+1;", CheckPasswordComparedWithFreshHash, 0},
		{"PasswordComparedWithFreshHash", "$v=password_hash()===$stored;", CheckPasswordComparedWithFreshHash, 0},
		{"UploadClientMimeTrusted", "if($_FILES['u']['type']==='image/png' && getimagesize($_FILES['u']['tmp_name'])){move_uploaded_file($_FILES['u']['tmp_name'],'/safe');}", CheckUploadClientMimeTrusted, 0},
		{"DirectoryIteratorDotEntries", "unlink('/safe'); unlink(makePath()->getPathname());", CheckDirectoryIteratorDotEntries, 0},
		{"DirectoryIteratorDotEntries", "foreach($iterator as $entry){unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 0},
		{"DirectoryIteratorDotEntries", "foreach(new ArrayIterator([]) as $entry){unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){if($enabled){unlink($entry->getPathname());}}", CheckDirectoryIteratorDotEntries, 1},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){if(!isAllowed()){unlink($entry->getPathname());}}", CheckDirectoryIteratorDotEntries, 1},
		{"FileLockFailureUnchecked", "fwrite($unknown,'x'); $h=fopen('/tmp/f','c+');fread($h,1);", CheckFileLockFailureUnchecked, 0},
		{"FileTruncatedBeforeLock", "fopen('/tmp/f','w');", CheckFileTruncatedBeforeLock, 0},
		{"OwnedStreamNotClosed", "function f(){fopen('/tmp/f','r');return 1;}", CheckOwnedStreamNotClosed, 0},
		{"OwnedStreamNotClosed", "function f(){$h=fopen('/tmp/f','r');return $h;}", CheckOwnedStreamNotClosed, 0},
		{"OwnedStreamNotClosed", "function f(){$h=fopen('/tmp/f','r');return fread($h,1);}", CheckOwnedStreamNotClosed, 0},
		{"ReadModifyWriteLockTooLate", "file_put_contents('/tmp/f','x',unknownFlags());file_put_contents('/tmp/f');", CheckReadModifyWriteLockTooLate, 0},
		{"StreamOpenFailureUnchecked", "fread($unknown,1);", CheckStreamOpenFailureUnchecked, 0},
		{"StreamUseAfterClose", "fread($unknown,1);", CheckStreamUseAfterClose, 0},
		{"DirectoryIteratorDotEntries", "foreach(new DirectoryIterator('/tmp') as $entry){if($other->isDot()){continue;}unlink($entry->getPathname());}", CheckDirectoryIteratorDotEntries, 1},
		{"TempnamDirectoryFallbackUnchecked", "function f(){$p=tempnam('/tmp','p');if($x!=='x'){return;} file_put_contents($p,'x');}", CheckTempnamDirectoryFallbackUnchecked, 1},
		{"TempnamDirectoryFallbackUnchecked", "function f(){$p=tempnam('/tmp','p');if(realpath($x)!=='x'){return;} file_put_contents($p,'x');}", CheckTempnamDirectoryFallbackUnchecked, 1},
		{"TempnamDirectoryFallbackUnchecked", "function f(){$p=tempnam('/tmp','p');if(realpath(dirname('/other'))!==realpath('/tmp')){return;} file_put_contents($p,'x');}", CheckTempnamDirectoryFallbackUnchecked, 1},
		{"TempnamDirectoryFallbackUnchecked", "file_put_contents(tempnam('/tmp','p'),'x');", CheckTempnamDirectoryFallbackUnchecked, 1},
		{"FastDigestUsedForPasswordStorage", "/** @custos-credential-store */ function store($v){} store('literal');", CheckFastDigestUsedForPasswordStorage, 0},
		{"ReadModifyWriteLockTooLate", "file_put_contents('/tmp/f','x',LOCK_EX);", CheckReadModifyWriteLockTooLate, 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			p := nativeCheckProbe{tc.id, tc.check}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("boundary.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func TestNativeProofAuxiliaryBoundaries(t *testing.T) {
	probeNative(t, "echo 'done'; probe('/tmp','/tmp');", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if nativeStateCall(ctx, c, "committed-output") != nil {
			t.Fatal("echo is not a call state")
		}
		if !nativeSameExpression(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("equal literal paths should match")
		}
	})
	p := nativeCheckProbe{"PasswordComparedWithFreshHash", CheckPasswordComparedWithFreshHash}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("old.php", []byte("<?php $v=password_hash($p,PASSWORD_DEFAULT)===$saved;"), syntax.Options{})
	ctx := analysis.NewTestContext(e, f)
	ctx.PHP = phpversion.PHP54
	syntax.InspectFile(f, func(n syntax.Node) bool {
		CheckPasswordComparedWithFreshHash(ctx, n, "Use password verification.")
		return true
	})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(syntax.Parse("external.php", []byte("<?php /** @custos-protected */ function externalProtected(){}"), syntax.Options{})))
	p = nativeCheckProbe{"RedirectContinuesProtectedExecution", CheckRedirectContinuesProtectedExecution}
	e, err = analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
	if err != nil {
		t.Fatal(err)
	}
	if got := e.WithIndex(ix).Analyze(syntax.Parse("consumer.php", []byte("<?php externalProtected();"), syntax.Options{})); len(got) != 0 {
		t.Fatalf("foreign source unavailable: %+v", got)
	}
}

func TestNativeNonCallGlobalState(t *testing.T) {
	checked := false
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if !ctx.IsGlobalFunctionCall(c, "header") {
			return
		}
		checked = true
		if _, known := ctx.Flow().GlobalStateBefore(c, "output"); !known {
			t.Fatal("literal echo must establish output state")
		}
		if nativeStateCall(ctx, c, "output") != nil {
			t.Fatal("echo state must not resolve to a function call")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("output.php", []byte("<?php echo 'body'; header('X-Value: example');"), syntax.Options{}))
	if !checked {
		t.Fatal("header was not dispatched")
	}
}
