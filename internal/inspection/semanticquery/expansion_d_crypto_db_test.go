package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type expansionCryptoDBProbe struct {
	id    string
	kinds []syntax.NodeKind
	check func(*analysis.Context, syntax.Node)
}

func (p expansionCryptoDBProbe) ID() string                                 { return p.id }
func (p expansionCryptoDBProbe) Semantic()                                  {}
func (p expansionCryptoDBProbe) Flow()                                      {}
func (p expansionCryptoDBProbe) Kinds() []syntax.NodeKind                   { return p.kinds }
func (p expansionCryptoDBProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n) }

func TestExpansionCryptoDBContracts(t *testing.T) {
	cases := []struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{
		{"HashContextUsedAfterFinalization", "<?php\n$h = hash_init('sha256'); hash_final($h); hash_update($h, 'next');\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"HashContextUsedAfterFinalization", "<?php\n$h = hash_init('sha256'); hash_update($h, 'next'); hash_final($h);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"HashContextUsedAfterFinalization", "<?php $h=hash_init('sha256');hash_final($h);hash_copy($h);", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"HashContextUsedAfterFinalization", "<?php $h=hash_init('sha256');hash_final($h);$h=hash_init('sha256');hash_update($h,'x');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"HashContextUsedAfterFinalization", "<?php $h=hash_init('sha256');if($flag){hash_final($h);}hash_update($h,'x');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"HashContextUsedAfterFinalization", "<?php $h=hash_init('sha256');hash_final($h);unknown($h);hash_update($h,'x');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"HashContextUsedAfterFinalization", "<?php namespace Local; function hash_final($h){} hash_final($h); hash_update($h,'x');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"HashFileFailureUsedAsDigest", "<?php\nfile_put_contents($manifest, hash_file('sha256', $path));\n", 1, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest},
		{"HashFileFailureUsedAsDigest", "<?php\n$h = hash_file('sha256', $path); if ($h !== false) { file_put_contents($manifest, $h); }\n", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest},
		{"HashFileFailureUsedAsDigest", "<?php $h=hash_file('sha256',$p);if($h===false){throw new RuntimeException();}file_put_contents($p,$h);", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest},
		{"HashFileFailureUsedAsDigest", "<?php $h=hash_file('sha256',$p);if($h==='abc'){return true;}", 1, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest},
		{"HashFileFailureUsedAsDigest", "<?php return hash_file('sha256',$p);", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest},
		{"HashFileFailureUsedAsDigest", "<?php file_put_contents($p,'x');", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest},
		{"HashFileFailureUsedAsDigest", "<?php $h=hash_file('sha256',$p); if('abc'!==$h){return false;}", 1, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest},
		{"OpenSslSigningFailureIgnored", "<?php\nopenssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256); echo base64_encode($signature);\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored},
		{"OpenSslSigningFailureIgnored", "<?php\nif (openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256)) { echo base64_encode($signature); }\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored},
		{"OpenSslSigningFailureIgnored", "<?php openssl_sign($d,$s,$k);echo $s;", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored},
		{"OpenSslSigningFailureIgnored", "<?php openssl_sign($d,$s,$k);file_put_contents($p,$s);", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored},
		{"OpenSslSigningFailureIgnored", "<?php openssl_sign($d,$s,$k);$s='changed';echo base64_encode($s);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored},
		{"OpenSslSigningFailureIgnored", "<?php $ok=openssl_sign($d,$s,$k);echo base64_encode($s);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored},
		{"OpenSslSigningFailureIgnored", "<?php openssl_sign($d,$s,$k);unrelated();", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored},
		{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php\n$cipher = openssl_public_encrypt($data, $out, $key); echo base64_encode($cipher);\n", 1, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext},
		{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php\nif (openssl_public_encrypt($data, $out, $key)) { echo base64_encode($out); }\n", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext},
		{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php $ok=openssl_private_encrypt($d,$out,$k);echo $ok;", 1, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext},
		{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php $out=openssl_public_encrypt($d,$out,$k);echo $out;", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext},
		{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php $ok=openssl_public_encrypt($d,$out,$k);file_put_contents($p,$ok);", 1, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext},
		{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php $ok=openssl_public_encrypt($d,$out,$k);$ok='x';echo base64_encode($ok);", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext},
		{"PasswordVerifyArgumentsReversed", "<?php\n$hash = password_hash($password, PASSWORD_DEFAULT); password_verify($hash, $password);\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckPasswordVerifyArgumentsReversed},
		{"PasswordVerifyArgumentsReversed", "<?php\n$hash = password_hash($password, PASSWORD_DEFAULT); password_verify($password, $hash);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPasswordVerifyArgumentsReversed},
		{"PasswordVerifyArgumentsReversed", "<?php $p='secret';$h=password_hash($p,PASSWORD_DEFAULT);$p='new';password_verify($h,$p);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPasswordVerifyArgumentsReversed},
		{"PasswordVerifyArgumentsReversed", "<?php function f($p){$h=password_hash($p,PASSWORD_DEFAULT);unknown($p);password_verify($h,$p);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPasswordVerifyArgumentsReversed},
		{"PasswordVerifyArgumentsReversed", "<?php password_verify('clear','hash');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPasswordVerifyArgumentsReversed},
		{"SodiumSecretboxKeyLengthMismatch", "<?php\nsodium_crypto_secretbox($message, $nonce, random_bytes(16));\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch},
		{"SodiumSecretboxKeyLengthMismatch", "<?php\nsodium_crypto_secretbox($message, $nonce, sodium_crypto_secretbox_keygen());\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch},
		{"SodiumSecretboxKeyLengthMismatch", "<?php sodium_crypto_secretbox_open($m,$n,str_repeat('x',16));", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch},
		{"SodiumSecretboxKeyLengthMismatch", "<?php sodium_crypto_secretbox($m,$n,str_repeat('xx',16));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch},
		{"SodiumSecretboxKeyLengthMismatch", "<?php sodium_crypto_secretbox($m,$n,$k);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch},
		{"SodiumSecretboxKeyLengthMismatch", "<?php sodium_crypto_secretbox($m,$n,str_repeat($s,16));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch},
		{"SodiumSecretboxKeyLengthMismatch", "<?php sodium_crypto_secretbox($m,$n,random_bytes(0));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch},
		{"SodiumSecretboxNonceReused", "<?php\n$n = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES); sodium_crypto_secretbox('alpha', $n, $k); sodium_crypto_secretbox('beta', $n, $k);\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused},
		{"SodiumSecretboxNonceReused", "<?php\nsodium_crypto_secretbox('alpha', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k); sodium_crypto_secretbox('beta', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused},
		{"SodiumSecretboxNonceReused", "<?php function f($k,$n){sodium_crypto_secretbox('a',$n,$k);sodium_crypto_secretbox('a',$n,$k);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused},
		{"SodiumSecretboxNonceReused", "<?php function f($k,$n){sodium_crypto_secretbox('a',$n,$k);sodium_crypto_secretbox($m,$n,$k);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused},
		{"SodiumSecretboxNonceReused", "<?php function f($k,$n){sodium_crypto_secretbox('a',$n,$k);$n=random_bytes(24);sodium_crypto_secretbox('b',$n,$k);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused},
		{"SodiumSecretboxNonceReused", "<?php function f($k,$n){sodium_crypto_secretbox('a',$n,$k);$k=random_bytes(32);sodium_crypto_secretbox('b',$n,$k);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused},
		{"DetachedSignaturePassedToCombinedVerifier", "<?php\n$sig = sodium_crypto_sign_detached($message, $secret); sodium_crypto_sign_open($sig, $public);\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckDetachedSignaturePassedToCombinedVerifier},
		{"DetachedSignaturePassedToCombinedVerifier", "<?php\n$sig = sodium_crypto_sign_detached($message, $secret); sodium_crypto_sign_verify_detached($sig, $message, $public);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckDetachedSignaturePassedToCombinedVerifier},
		{"DetachedSignaturePassedToCombinedVerifier", "<?php $s=sodium_crypto_sign_detached($m,$k);sodium_crypto_sign_open($s.$m,$p);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckDetachedSignaturePassedToCombinedVerifier},
		{"DetachedSignaturePassedToCombinedVerifier", "<?php sodium_crypto_sign_open($s,$p);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckDetachedSignaturePassedToCombinedVerifier},
		{"SodiumKdfContextLengthMismatch", "<?php\nsodium_crypto_kdf_derive_from_key(32, 2, 'short', $key);\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumKdfContextLengthMismatch},
		{"SodiumKdfContextLengthMismatch", "<?php\nsodium_crypto_kdf_derive_from_key(32, 2, 'APPKEY01', $key);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumKdfContextLengthMismatch},
		{"SodiumKdfContextLengthMismatch", "<?php sodium_crypto_kdf_derive_from_key(32,1,str_repeat('x',7),$k);", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumKdfContextLengthMismatch},
		{"SodiumKdfContextLengthMismatch", "<?php sodium_crypto_kdf_derive_from_key(32,1,$c,$k);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumKdfContextLengthMismatch},
		{"SecretstreamPushAfterFinalTag", "<?php\nsodium_crypto_secretstream_xchacha20poly1305_push($s, 'end', '', SODIUM_CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_TAG_FINAL); sodium_crypto_secretstream_xchacha20poly1305_push($s, 'more');\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSecretstreamPushAfterFinalTag},
		{"SecretstreamPushAfterFinalTag", "<?php\nsodium_crypto_secretstream_xchacha20poly1305_push($s, 'more'); sodium_crypto_secretstream_xchacha20poly1305_push($s, 'end', '', SODIUM_CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_TAG_FINAL);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSecretstreamPushAfterFinalTag},
		{"SecretstreamPushAfterFinalTag", "<?php function f($s){sodium_crypto_secretstream_xchacha20poly1305_push($s,'a','',0);sodium_crypto_secretstream_xchacha20poly1305_push($s,'b');}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSecretstreamPushAfterFinalTag},
		{"SecretstreamPushAfterFinalTag", "<?php function f($s){if($flag){sodium_crypto_secretstream_xchacha20poly1305_push($s,'a','',3);}sodium_crypto_secretstream_xchacha20poly1305_push($s,'b');}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSecretstreamPushAfterFinalTag},
		{"CurlHeaderOptionRequiresArray", "<?php\ncurl_setopt($ch, CURLOPT_HTTPHEADER, 'Accept: text/plain');\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlHeaderOptionRequiresArray", "<?php\ncurl_setopt($ch, CURLOPT_HTTPHEADER, ['Accept: text/plain']);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlHeaderOptionRequiresArray", "<?php curl_setopt_array($ch,[CURLOPT_HTTPHEADER=>'Accept: x']);", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlHeaderOptionRequiresArray", "<?php curl_setopt_array($ch,[CURLOPT_HTTPHEADER=>['Accept: x']]);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlHeaderOptionRequiresArray", "<?php curl_setopt_array($ch,$options);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlHeaderOptionRequiresArray", "<?php curl_setopt_array($ch,[CURLOPT_TIMEOUT=>1]);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlHeaderOptionRequiresArray", "<?php curl_setopt($ch,$option,'x');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlHeaderOptionRequiresArray", "<?php curl_setopt($ch,CURLOPT_HTTPHEADER,$unknown);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray},
		{"CurlUploadDeclaredSizeMismatch", "<?php\n$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); curl_setopt($ch, CURLOPT_INFILESIZE, 99);\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php\n$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); curl_setopt($ch, CURLOPT_INFILESIZE, 6);\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php $h=fopen('data://text/plain,abcdef','r');curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,-1);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php $h=fopen('data://text/plain,abcdef','r');curl_setopt($c,CURLOPT_UPLOAD,false);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php $h=fopen('/tmp/x','r');curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php $h=fopen('data://text/plain,abcdef','w');curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php $h=fopen('data://text/plain,abcdef','r');fread($h,1);curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php curl_setopt($c,CURLOPT_INFILESIZE,99);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlUploadDeclaredSizeMismatch", "<?php curl_setopt($c,CURLOPT_INFILESIZE,$size);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch},
		{"CurlReadCallbackExceedsRequestedSize", "<?php\ncurl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit + 2));\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php\ncurl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit));\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php curl_setopt($c,CURLOPT_READFUNCTION,fn($c)=>'x');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php curl_setopt($c,CURLOPT_READFUNCTION,function($c,$h,$n){});", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>'x');", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>str_repeat('xx',$n+1));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>str_repeat('x',$n-1));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>str_repeat('x',$unknown+1));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlReadCallbackExceedsRequestedSize", "<?php curl_setopt($c,CURLOPT_READFUNCTION,function($c,$h,$n){$n=1;return str_repeat('x',$n+1);});", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"CurlHeaderCallbackMissingByteCount", "<?php\ncurl_setopt($ch, CURLOPT_HEADERFUNCTION, function($ch, $line) { echo $line; });\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderCallbackMissingByteCount},
		{"CurlHeaderCallbackMissingByteCount", "<?php\ncurl_setopt($ch, CURLOPT_HEADERFUNCTION, function($ch, $line) { return strlen($line); });\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderCallbackMissingByteCount},
		{"CurlHeaderCallbackMissingByteCount", "<?php curl_setopt($c,CURLOPT_HEADERFUNCTION,function($c,$line){return;});", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderCallbackMissingByteCount},
		{"CurlHeaderCallbackMissingByteCount", "<?php curl_setopt($c,CURLOPT_HEADERFUNCTION,fn($c,$line)=>strlen($line));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderCallbackMissingByteCount},
		{"CurlHeaderCallbackMissingByteCount", "<?php curl_setopt($c,CURLOPT_HEADERFUNCTION,function(){yield 1;});", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderCallbackMissingByteCount},
		{"CurlResponseZeroRejected", "<?php\ncurl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (!curl_exec($ch)) { throw new RuntimeException(); }\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlResponseZeroRejected", "<?php\ncurl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (curl_exec($ch) === false) { throw new RuntimeException(); }\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(curl_exec($c)==false){return false;}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,false);if(!curl_exec($c)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(!curl_exec($c)){echo 'empty';}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlResponseZeroRejected", "<?php if(!curl_exec($c)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);unknown($c);if(!curl_exec($c)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(curl_exec($c)===''){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected},
		{"CurlWholeUrlEscapedAsComponent", "<?php\ncurl_setopt($ch, CURLOPT_URL, curl_escape($ch, 'https://example.net/a'));\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlWholeURLEscapedAsComponent},
		{"CurlWholeUrlEscapedAsComponent", "<?php\ncurl_setopt($ch, CURLOPT_URL, 'https://example.net/' . curl_escape($ch, 'a b'));\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlWholeURLEscapedAsComponent},
		{"CurlWholeUrlEscapedAsComponent", "<?php $e=curl_escape($c,'https://example.com/a');curl_setopt($c,CURLOPT_URL,$e);", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlWholeURLEscapedAsComponent},
		{"CurlWholeUrlEscapedAsComponent", "<?php curl_setopt($c,CURLOPT_URL,curl_escape($c,'relative/path'));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlWholeURLEscapedAsComponent},
		{"CurlWholeUrlEscapedAsComponent", "<?php curl_setopt($c,CURLOPT_URL,curl_escape($c,$url));", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlWholeURLEscapedAsComponent},
		{"NonblockingSocketConnectPendingRejected", "<?php\nif (socket_set_nonblock($s)) { if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); } }\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckNonblockingSocketConnectPendingRejected},
		{"NonblockingSocketConnectPendingRejected", "<?php\nsocket_set_block($s); if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); }\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckNonblockingSocketConnectPendingRejected},
		{"NonblockingSocketConnectPendingRejected", "<?php if(socket_set_nonblock($s)){if(!socket_connect($s,$h,80)){socket_last_error($s);return false;}}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckNonblockingSocketConnectPendingRejected},
		{"NonblockingSocketConnectPendingRejected", "<?php socket_set_nonblock($s);if(!socket_connect($s,$h,80)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckNonblockingSocketConnectPendingRejected},
		{"NonblockingSocketConnectPendingRejected", "<?php if(socket_set_nonblock($s)){socket_set_block($s);if(!socket_connect($s,$h,80)){return false;}}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckNonblockingSocketConnectPendingRejected},
		{"SocketReadZeroRejected", "<?php\nif (!$data = socket_read($s, 1024, PHP_BINARY_READ)) { throw new RuntimeException(); }\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSocketReadZeroRejected},
		{"SocketReadZeroRejected", "<?php\n$data = socket_read($s, 1024, PHP_BINARY_READ); if ($data === false) { throw new RuntimeException(); }\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSocketReadZeroRejected},
		{"SocketReadZeroRejected", "<?php if(socket_read($s,10,PHP_BINARY_READ)==false){return false;}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckSocketReadZeroRejected},
		{"SocketReadZeroRejected", "<?php if(!socket_read($s,10,PHP_NORMAL_READ)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSocketReadZeroRejected},
		{"SocketReadZeroRejected", "<?php if(!socket_read($s,10,PHP_BINARY_READ)){echo 'empty';}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSocketReadZeroRejected},
		{"TlsHandshakePendingAccepted", "<?php\nif (stream_set_blocking($s, false)) { if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) { fwrite($s, $secret); } }\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckTLSHandshakePendingAccepted},
		{"TlsHandshakePendingAccepted", "<?php\nif (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) === true) { fwrite($s, $secret); }\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckTLSHandshakePendingAccepted},
		{"TlsHandshakePendingAccepted", "<?php if(stream_set_blocking($s,false)){if(stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false){echo 'ok';}}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckTLSHandshakePendingAccepted},
		{"TlsHandshakePendingAccepted", "<?php stream_set_blocking($s,false);if(stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false){fwrite($s,'x');}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckTLSHandshakePendingAccepted},
		{"TlsHandshakePendingAccepted", "<?php if(stream_set_blocking($s,true)){if(stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false){fwrite($s,'x');}}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckTLSHandshakePendingAccepted},
		{"FtpPendingTransferAcceptedAsComplete", "<?php\nif (ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) { return true; }\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete},
		{"FtpPendingTransferAcceptedAsComplete", "<?php\nif (ftp_nb_get($ftp, $local, $remote, FTP_BINARY) === FTP_FINISHED) { return true; }\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete},
		{"FtpPendingTransferAcceptedAsComplete", "<?php if(ftp_nb_continue($f)){echo 'complete';}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete},
		{"FtpPendingTransferAcceptedAsComplete", "<?php if(!ftp_nb_continue($f)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete},
		{"FtpPendingTransferAcceptedAsComplete", "<?php if(ftp_nb_continue($f)){echo $text;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete},
		{"FtpPendingTransferAcceptedAsComplete", "<?php if(ftp_nb_continue($f)===FTP_MOREDATA){return true;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete},
		{"PdoBoundValueAssumedLive", "<?php\n$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindValue(':n', $n); $n = 11; $s->execute();\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoBoundValueAssumedLive},
		{"PdoBoundValueAssumedLive", "<?php\n$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindParam(':n', $n); $n = 11; $s->execute();\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoBoundValueAssumedLive},
		{"PdoBoundValueAssumedLive", "<?php function f(PDOStatement $s){$n=1;$s->bindValue(':n',$n);$n=2;$s->bindValue(':n',$n);$s->execute();}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoBoundValueAssumedLive},
		{"PdoBoundValueAssumedLive", "<?php function f(PDOStatement $s){$n=1;$s->bindValue(':n',$n);$n=2;$s->execute(['n'=>2]);}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoBoundValueAssumedLive},
		{"PdoBoundValueAssumedLive", "<?php function f(PDOStatement $s){$s->bindValue(':n',1);$s->execute();}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoBoundValueAssumedLive},
		{"PdoBoundValueAssumedLive", "<?php function f(PDOStatement $s){$n=new stdClass();$s->bindValue(':n',$n);$n=new stdClass();$s->execute();}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoBoundValueAssumedLive},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php\n$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :n + :n');\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php\n$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :a + :b');\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('sqlite::memory:');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT :n+:n');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,true);$p->prepare('SELECT :n+:n');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x');$p->prepare('SELECT :n+:n');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x',null,null,[PDO::ATTR_EMULATE_PREPARES=>false]);$p->prepare('SELECT :n+:n');", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare(\"SELECT ':n', :n\");", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT :n+:N');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
		{"PdoGroupedFetchReadsRemovedColumn", "<?php\n$pdo = new PDO($dsn); $rows = $pdo->query(\"SELECT 'x' AS category, 7 AS value\")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['category'];\n", 1, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn},
		{"PdoGroupedFetchReadsRemovedColumn", "<?php\n$pdo = new PDO($dsn); $rows = $pdo->query(\"SELECT 'x' AS category, 7 AS value\")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['value'];\n", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn},
		{"PdoGroupedFetchReadsRemovedColumn", "<?php function f(PDO $p){$r=$p->query('SELECT * FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn},
		{"PdoGroupedFetchReadsRemovedColumn", "<?php function f(PDO $p){$r=$p->query('SELECT x AS a, y AS a FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['a'];}", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn},
		{"PdoGroupedFetchReadsRemovedColumn", "<?php function f(PDO $p){$r=$p->query('SELECT x AS a,y AS b FROM x')->fetchAll(PDO::FETCH_ASSOC);echo $r[0]['a'];}", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn},
		{"PdoGroupedFetchReadsRemovedColumn", "<?php function f(PDO $p){$r=$p->query('SELECT x AS a,y AS b FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['a'];}", 1, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn},
		{"PdoFetchIntoRetainsSharedRows", "<?php\n$object=new stdClass();$pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { $rows[] = $row; }\n", 1, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows},
		{"PdoFetchIntoRetainsSharedRows", "<?php\n$object=new stdClass();$pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { $rows[] = clone $row; }\n", 0, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows},
		{"PdoFetchIntoRetainsSharedRows", "<?php function f(PDOStatement $s){$o=new stdClass();$s->setFetchMode(PDO::FETCH_INTO,$o);while($r=$s->fetch()){$r=new stdClass();$rows[]=$r;}}", 0, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows},
		{"PdoFetchIntoRetainsSharedRows", "<?php function f(PDOStatement $s){$o=new stdClass();$s->setFetchMode(PDO::FETCH_OBJ);while($r=$s->fetch()){$rows[]=$r;}}", 0, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows},
		{"PdoFetchIntoRetainsSharedRows", "<?php function f(PDOStatement $s){$o=new stdClass();$s->setFetchMode(PDO::FETCH_INTO,$o);$rows[]=$s->fetch();}", 0, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows},
		{"PdoExecZeroRejectedAsFailure", "<?php\n$pdo = new PDO($dsn); if (!$pdo->exec('DELETE FROM jobs WHERE done = 1')) { throw new RuntimeException(); }\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoExecZeroRejectedAsFailure},
		{"PdoExecZeroRejectedAsFailure", "<?php\n$pdo = new PDO($dsn); if ($pdo->exec('DELETE FROM jobs WHERE done = 1') === false) { throw new RuntimeException(); }\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoExecZeroRejectedAsFailure},
		{"PdoExecZeroRejectedAsFailure", "<?php function f(PDO $p){if($p->exec('DELETE FROM x')==false){return false;}}", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoExecZeroRejectedAsFailure},
		{"PdoExecZeroRejectedAsFailure", "<?php function f(PDO $p){if(!$p->exec('DELETE FROM x')){echo 'none';}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoExecZeroRejectedAsFailure},
		{"MysqliEscapedValueSurvivesCharsetChange", "<?php\n$db = new mysqli(); $db->set_charset('utf8mb4'); $v = $db->real_escape_string($input); if ($db->set_charset('gbk')) { $db->query(\"SELECT '$v'\"); }\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange},
		{"MysqliEscapedValueSurvivesCharsetChange", "<?php\n$db = new mysqli(); $db->set_charset('gbk'); $v = $db->real_escape_string($input); $db->query(\"SELECT '$v'\");\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange},
		{"MysqliEscapedValueSurvivesCharsetChange", "<?php function f(mysqli $db){$db->set_charset('utf8mb4');$v=$db->real_escape_string($x);if($db->set_charset('utf8mb4')){$db->query(\"SELECT '$v'\");}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange},
		{"MysqliEscapedValueSurvivesCharsetChange", "<?php function f(mysqli $db){$db->set_charset('utf8mb4');$v=$db->real_escape_string($x);$db->set_charset('gbk');$db->query(\"SELECT '$v'\");}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange},
		{"MysqliMultiQueryResultsNotDrained", "<?php\n$db = new mysqli(); if ($db->multi_query('SELECT 7; SELECT 11')) { $db->query('SELECT 14'); }\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained},
		{"MysqliMultiQueryResultsNotDrained", "<?php\n$db = new mysqli(); $db->query('SELECT 7'); $db->query('SELECT 14');\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained},
		{"MysqliMultiQueryResultsNotDrained", "<?php function f(mysqli $db){if($db->multi_query('SELECT 1;')){$db->query('SELECT 2');}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained},
		{"MysqliMultiQueryResultsNotDrained", "<?php function f(mysqli $db){if($db->multi_query('SELECT 1;SELECT 2')){while($db->more_results()){$db->next_result();}$db->query('SELECT 3');}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained},
		{"MysqliMultiQueryResultsNotDrained", "<?php function f(mysqli $db){if($db->multi_query('SELECT 1;SELECT 2;SELECT 3')){$db->next_result();$db->query('SELECT 4');}}", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained},
		{"PgReturnedRowsUsedAsAffectedRows", "<?php\n$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo pg_num_rows($r); }\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows},
		{"PgReturnedRowsUsedAsAffectedRows", "<?php\n$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo pg_affected_rows($r); }\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows},
		{"PgReturnedRowsUsedAsAffectedRows", "<?php $r=pg_query($db,'UPDATE x SET y=1 RETURNING y');if($r!==false){pg_num_rows($r);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows},
		{"PgReturnedRowsUsedAsAffectedRows", "<?php $r=pg_query($db,\"UPDATE x SET y='RETURNING'\");if($r!==false){pg_num_rows($r);}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows},
		{"PgReturnedRowsUsedAsAffectedRows", "<?php $r=pg_query($db,'SELECT 1');if($r!==false){pg_num_rows($r);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows},
		{"PgReturnedRowsUsedAsAffectedRows", "<?php $r=pg_query($db,'UPDATE x SET y=1');pg_num_rows($r);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows},
		{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php\n$db = new SQLite3(':memory:'); $db->createFunction('dice', fn() => random_int(1, 6), 0, SQLITE3_DETERMINISTIC);\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic},
		{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php\n$db = new SQLite3(':memory:'); $db->createFunction('twice', fn($n) => $n * 2, 1, SQLITE3_DETERMINISTIC);\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic},
		{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php function f(SQLite3 $d){$d->createFunction('x',fn()=>time()+1,0,SQLITE3_DETERMINISTIC);}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic},
		{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php function f(SQLite3 $d){$d->createFunction('x',fn()=>time(),0,0);}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic},
		{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php function f(SQLite3 $d){$d->createFunction('x',$callback,0,SQLITE3_DETERMINISTIC);}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic},
		{"SqliteForeignKeySettingInsideTransaction", "<?php\n$db = new SQLite3(':memory:'); $db->exec('BEGIN; PRAGMA foreign_keys=ON;');\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction},
		{"SqliteForeignKeySettingInsideTransaction", "<?php\n$db = new SQLite3(':memory:'); $db->exec('PRAGMA foreign_keys=ON; BEGIN;');\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction},
		{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new SQLite3(':memory:');$d->exec('BEGIN');$d->exec('PRAGMA foreign_keys=ON');", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction},
		{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new SQLite3(':memory:');$d->exec('BEGIN;COMMIT;PRAGMA foreign_keys=ON');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction},
		{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new PDO('sqlite::memory:');$d->beginTransaction();$d->exec('PRAGMA foreign_keys=ON');", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction},
		{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new PDO('mysql:host=x');$d->beginTransaction();$d->exec('PRAGMA foreign_keys=ON');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction},
	}
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"HashContextUsedAfterFinalization", "<?php if(false){\n$h = hash_init('sha256'); hash_final($h); hash_update($h, 'next');\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"HashFileFailureUsedAsDigest", "<?php if(false){\nfile_put_contents($manifest, hash_file('sha256', $path));\n}", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"OpenSslSigningFailureIgnored", "<?php if(false){\nopenssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256); echo base64_encode($signature);\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php if(false){\n$cipher = openssl_public_encrypt($data, $out, $key); echo base64_encode($cipher);\n}", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PasswordVerifyArgumentsReversed", "<?php if(false){\n$hash = password_hash($password, PASSWORD_DEFAULT); password_verify($hash, $password);\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPasswordVerifyArgumentsReversed})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SodiumSecretboxKeyLengthMismatch", "<?php if(false){\nsodium_crypto_secretbox($message, $nonce, random_bytes(16));\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxKeyLengthMismatch})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SodiumSecretboxNonceReused", "<?php if(false){\n$n = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES); sodium_crypto_secretbox('alpha', $n, $k); sodium_crypto_secretbox('beta', $n, $k);\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"DetachedSignaturePassedToCombinedVerifier", "<?php if(false){\n$sig = sodium_crypto_sign_detached($message, $secret); sodium_crypto_sign_open($sig, $public);\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckDetachedSignaturePassedToCombinedVerifier})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SodiumKdfContextLengthMismatch", "<?php if(false){\nsodium_crypto_kdf_derive_from_key(32, 2, 'short', $key);\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumKdfContextLengthMismatch})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SecretstreamPushAfterFinalTag", "<?php if(false){\nsodium_crypto_secretstream_xchacha20poly1305_push($s, 'end', '', SODIUM_CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_TAG_FINAL); sodium_crypto_secretstream_xchacha20poly1305_push($s, 'more');\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSecretstreamPushAfterFinalTag})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlHeaderOptionRequiresArray", "<?php if(false){\ncurl_setopt($ch, CURLOPT_HTTPHEADER, 'Accept: text/plain');\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderOptionRequiresArray})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlUploadDeclaredSizeMismatch", "<?php if(false){\n$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); curl_setopt($ch, CURLOPT_INFILESIZE, 99);\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlReadCallbackExceedsRequestedSize", "<?php if(false){\ncurl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit + 2));\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlHeaderCallbackMissingByteCount", "<?php if(false){\ncurl_setopt($ch, CURLOPT_HEADERFUNCTION, function($ch, $line) { echo $line; });\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderCallbackMissingByteCount})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlResponseZeroRejected", "<?php if(false){\ncurl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (!curl_exec($ch)) { throw new RuntimeException(); }\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlWholeUrlEscapedAsComponent", "<?php if(false){\ncurl_setopt($ch, CURLOPT_URL, curl_escape($ch, 'https://example.net/a'));\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlWholeURLEscapedAsComponent})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonblockingSocketConnectPendingRejected", "<?php if(false){\nif (socket_set_nonblock($s)) { if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); } }\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckNonblockingSocketConnectPendingRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SocketReadZeroRejected", "<?php if(false){\nif (!$data = socket_read($s, 1024, PHP_BINARY_READ)) { throw new RuntimeException(); }\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSocketReadZeroRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"TlsHandshakePendingAccepted", "<?php if(false){\nif (stream_set_blocking($s, false)) { if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) { fwrite($s, $secret); } }\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckTLSHandshakePendingAccepted})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"FtpPendingTransferAcceptedAsComplete", "<?php if(false){\nif (ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) { return true; }\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoBoundValueAssumedLive", "<?php if(false){\n$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindValue(':n', $n); $n = 11; $s->execute();\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoBoundValueAssumedLive})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php if(false){\n$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :n + :n');\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoGroupedFetchReadsRemovedColumn", "<?php if(false){\n$pdo = new PDO($dsn); $rows = $pdo->query(\"SELECT 'x' AS category, 7 AS value\")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['category'];\n}", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoFetchIntoRetainsSharedRows", "<?php if(false){\n$pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $object = new stdClass(); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { $rows[] = $row; }\n}", 0, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoExecZeroRejectedAsFailure", "<?php if(false){\n$pdo = new PDO($dsn); if (!$pdo->exec('DELETE FROM jobs WHERE done = 1')) { throw new RuntimeException(); }\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoExecZeroRejectedAsFailure})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliEscapedValueSurvivesCharsetChange", "<?php if(false){\n$db = new mysqli(); $db->set_charset('utf8mb4'); $v = $db->real_escape_string($input); if ($db->set_charset('gbk')) { $db->query(\"SELECT '$v'\"); }\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliMultiQueryResultsNotDrained", "<?php if(false){\n$db = new mysqli(); if ($db->multi_query('SELECT 7; SELECT 11')) { $db->query('SELECT 14'); }\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PgReturnedRowsUsedAsAffectedRows", "<?php if(false){\n$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo pg_num_rows($r); }\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php if(false){\n$db = new SQLite3(':memory:'); $db->createFunction('dice', fn() => random_int(1, 6), 0, SQLITE3_DETERMINISTIC);\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteForeignKeySettingInsideTransaction", "<?php if(false){\n$db = new SQLite3(':memory:'); $db->exec('BEGIN; PRAGMA foreign_keys=ON;');\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction})

	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"HashFileFailureUsedAsDigest", "<?php hash_file('sha256',$p)=='abc';", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary}, CheckHashFileFailureUsedAsDigest})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"OpenSslSigningFailureIgnored", "<?php function f($d,$k){$unused=function(){};openssl_sign($d,$s,$k);echo base64_encode($s);}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckOpenSslSigningFailureIgnored})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"OpenSslEncryptionStatusUsedAsCiphertext", "<?php echo openssl_public_encrypt($d,$out[0],$k);", 0, []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho}, CheckOpenSslEncryptionStatusUsedAsCiphertext})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SecretstreamPushAfterFinalTag", "<?php function f($s){sodium_crypto_secretstream_xchacha20poly1305_push($s,'a','',$tag);sodium_crypto_secretstream_xchacha20poly1305_push($s,'b');}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckSecretstreamPushAfterFinalTag})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlResponseZeroRejected", "<?php if($f){curl_setopt($c,CURLOPT_RETURNTRANSFER,true);}if(!curl_exec($c)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(!(curl_exec($c))){return false;}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(+curl_exec($c)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(curl_exec($c)){return false;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlResponseZeroRejected", "<?php curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(!curl_exec($c)){$f=function(){return false;};}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlResponseZeroRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlUploadDeclaredSizeMismatch", "<?php curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlUploadDeclaredSizeMismatch})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CurlHeaderCallbackMissingByteCount", "<?php curl_setopt($c,CURLOPT_HEADERFUNCTION,function($c,$l){$f=function(){};return strlen($l);});", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckCurlHeaderCallbackMissingByteCount})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonblockingSocketConnectPendingRejected", "<?php socket_connect($s,$h,80);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckNonblockingSocketConnectPendingRejected})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"TlsHandshakePendingAccepted", "<?php $ok=stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckTLSHandshakePendingAccepted})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"FtpPendingTransferAcceptedAsComplete", "<?php unknown();while(ftp_nb_continue($f)){break;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"FtpPendingTransferAcceptedAsComplete", "<?php if(ftp_nb_continue($f)){$g=function(){return true;};}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckFtpPendingTransferAcceptedAsComplete})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php function f(PDO $p){$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT :x+:x');}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare($sql);", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT $body$');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoGroupedFetchReadsRemovedColumn", "<?php echo $r['x'][0]['id'];", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoGroupedFetchReadsRemovedColumn", "<?php function f(PDO $p){$r=$p->query('SELECT id,value FROM x')->fetchAll(PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoGroupedFetchReadsRemovedColumn", "<?php function f(PDOStatement $s){$r=$s->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoGroupedFetchReadsRemovedColumn", "<?php function f(PDO $p){$r=$p->query($sql)->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0, []syntax.NodeKind{syntax.KArrayDimFetch}, CheckPdoGroupedFetchReadsRemovedColumn})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoFetchIntoRetainsSharedRows", "<?php $rows[]=$x;", 0, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PdoFetchIntoRetainsSharedRows", "<?php function f(PDOStatement $s){while($r=$s->fetch()){$rows[]=$r;}}", 0, []syntax.NodeKind{syntax.KAssign}, CheckPdoFetchIntoRetainsSharedRows})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliEscapedValueSurvivesCharsetChange", "<?php function f(mysqli $d){$d->query();}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliEscapedValueSurvivesCharsetChange", "<?php function f(mysqli $d){$d->query(function(){});}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliEscapedValueSurvivesCharsetChange", "<?php function f(mysqli $d,mysqli $e){$d->set_charset('utf8mb4');$v=$d->real_escape_string($x);if($d->set_charset('gbk')){$e->query(\"SELECT '$v'\");}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliEscapedValueSurvivesCharsetChange", "<?php function f(mysqli $d){$v=$d->real_escape_string($x);if($d->set_charset('gbk')){$d->query(\"SELECT '$v'\");}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliEscapedValueSurvivesCharsetChange", "<?php function f(mysqli $d){$d->set_charset($charset);$v=$d->real_escape_string($x);if($d->set_charset('gbk')){$d->query(\"SELECT '$v'\");}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliEscapedValueSurvivesCharsetChange})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliMultiQueryResultsNotDrained", "<?php function f(mysqli $d){$d->multi_query('SELECT 1;SELECT 2');$d->query('SELECT 3');}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliMultiQueryResultsNotDrained", "<?php function f(mysqli $d){if($d->multi_query($sql)){$d->query('SELECT 3');}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliMultiQueryResultsNotDrained", "<?php function f(mysqli $d){if($d->multi_query('SELECT $body$')){$d->query('SELECT 3');}}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"MysqliMultiQueryResultsNotDrained", "<?php function f(mysqli $d){if($d->multi_query('SELECT 1;SELECT 2')){$pending=$d->more_results();$d->query('SELECT 3');}}", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckMysqliMultiQueryResultsNotDrained})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PgReturnedRowsUsedAsAffectedRows", "<?php $r=pg_query('UPDATE x SET y=1');if($r!==false){pg_num_rows($r);}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PgReturnedRowsUsedAsAffectedRows", "<?php $r=pg_query($d,$sql);if($r!==false){pg_num_rows($r);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"PgReturnedRowsUsedAsAffectedRows", "<?php $r=pg_query($d,'SELECT $body$');if($r!==false){pg_num_rows($r);}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckPgReturnedRowsUsedAsAffectedRows})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php $x->createFunction('x',fn()=>time(),0,SQLITE3_DETERMINISTIC);", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteNondeterministicFunctionDeclaredDeterministic", "<?php function f(SQLite3 $d){$d->createFunction('x',fn()=>fn()=>time(),0,SQLITE3_DETERMINISTIC);}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteNondeterministicFunctionDeclaredDeterministic})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new PDO('sqlite::memory:');$d->beginTransaction();$d->commit();$d->exec('PRAGMA foreign_keys=ON');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new SQLite3(':memory:');$d->exec($sql);$d->exec('PRAGMA foreign_keys=ON');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new SQLite3(':memory:');$d->exec('SELECT $body$');$d->exec('PRAGMA foreign_keys=ON');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new SQLite3(':memory:');$d->exec($sql);", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"SqliteForeignKeySettingInsideTransaction", "<?php $d=new SQLite3(':memory:');$d->exec('SELECT $body$');", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckSqliteForeignKeySettingInsideTransaction})

	for i, tc := range cases {
		t.Run(tc.id+"/"+fmt.Sprint(i), func(t *testing.T) {
			p := expansionCryptoDBProbe{id: tc.id, kinds: tc.kinds, check: func(ctx *analysis.Context, n syntax.Node) { tc.check(ctx, n, "finding") }}
			engine, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := engine.Analyze(syntax.Parse("case.php", []byte(tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v\n%s", len(got), tc.want, got, tc.src)
			}
		})
	}
}

func TestExpansionCryptoDBSQLGrammar(t *testing.T) {
	for _, tc := range []struct {
		sql   string
		known bool
	}{
		{"SELECT 1 -- :ignored\n, :id # :ignored\n /* :ignored */", true},
		{"SELECT 'a\\b', 'a''b', `label`, 1::int", true},
		{"SELECT $body$", false},
		{"SELECT /* unfinished", false},
		{"SELECT /* /* nested */", false},
		{"SELECT 'unfinished", false},
		{"SELECT 'unfinished\\", false},
		{strings.Repeat("a", 65537), false},
	} {
		_, known := expansionDSQL(tc.sql)
		if known != tc.known {
			t.Fatalf("%q: known=%v", tc.sql, known)
		}
	}
	tokens, known := expansionDSQL("SELECT ':x', :x, :X /* :x */")
	if !known || strings.Join(tokens, " ") != "SELECT ':x , :x , :X" {
		t.Fatalf("incorrect marker tokens: %v", tokens)
	}
	for _, tc := range []struct {
		sql   string
		known bool
	}{
		{"UPDATE x SET y=1", false},
		{"SELECT", false},
		{"SELECT FROM x", false},
		{"SELECT x,y FROM t", true},
		{"SELECT 'x'", false},
		{"SELECT x AS 'alias'", false},
		{"SELECT x+1", false},
		{"SELECT `id`, value FROM t", true},
		{"SELECT x AS a,y AS a", false},
		{"SELECT fn(x)", false},
	} {
		_, known := expansionDProjection(tc.sql)
		if known != tc.known {
			t.Fatalf("projection %q: known=%v", tc.sql, known)
		}
	}
	for _, tc := range []struct {
		sql         string
		active, bad bool
	}{
		{";BEGIN;PRAGMA foreign_keys=ON;", true, true},
		{"BEGIN;ROLLBACK;PRAGMA foreign_keys(ON);", false, false},
		{"BEGIN;END;", false, false},
		{"BEGIN;PRAGMA cache_size=100;", true, false},
	} {
		tokens, _ := expansionDSQL(tc.sql)
		active, bad := expansionDTransaction(tokens, false)
		if active != tc.active || bad != tc.bad {
			t.Fatalf("state %q: %v %v", tc.sql, active, bad)
		}
	}
}

func TestExpansionCryptoDBLocalProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		same bool
	}{
		{"probe(1,1);", true},
		{"probe(1,2);", false},
		{"probe($a,$b);", false},
		{"function f($a){useValue($a);unset($a);probe($a);}", false},
		{"function f($a){useValue($a);$o->unknown($a);probe($a);}", false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			first := CallArgument(c.Args, 0, "a")
			second := CallArgument(c.Args, 1, "b")
			if second == nil {
				for _, ev := range expansionDCalls(ctx, c) {
					if prior, ok := ev.(*syntax.FuncCall); ok && ctx.Text(prior.Name) == "useValue" {
						second = CallArgument(prior.Args, 0, "")
					}
				}
			}
			if same := expansionDSame(ctx, first, second); same != tc.same {
				t.Fatalf("%s: same=%v", tc.src, same)
			}
			if expansionDSame(ctx, first, nil) {
				t.Fatal("nil proves identity")
			}
		})
	}
	probeNative(t, "probe(1);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		for _, x := range []syntax.Expr{nil, CallArgument(c.Args, 0, "")} {
			if _, ok := expansionDByteLength(ctx, x); ok {
				t.Fatal("non-string has known length")
			}
		}
	})
}

func TestExpansionCryptoDBBudgetsAndSuccessGuards(t *testing.T) {
	probeNative(t, "function f($a){"+strings.Repeat("other();", 513)+"probe($a,$a);}", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if expansionDSame(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("exhausted scope establishes identity")
		}
		if expansionDOutputs(ctx, c) != nil {
			t.Fatal("exhausted output budget retained proof")
		}
	})
	for _, src := range []string{
		`function f(mysqli $d){if(!$d->multi_query('SELECT 1;SELECT 2')){return;}$d->query('SELECT 3');}`,
		`function f(mysqli $d){if($d->multi_query('SELECT 1;SELECT 2')===true){$d->query('SELECT 3');}}`,
	} {
		p := expansionCryptoDBProbe{id: "MysqliMultiQueryResultsNotDrained", kinds: []syntax.NodeKind{syntax.KMethodCall}, check: func(ctx *analysis.Context, n syntax.Node) { CheckMysqliMultiQueryResultsNotDrained(ctx, n, "finding") }}
		e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("case.php", []byte("<?php "+src), syntax.Options{}))
		if len(got) != 1 {
			t.Fatalf("guard %s: %+v", src, got)
		}
	}
	p := expansionCryptoDBProbe{id: "SecretstreamPushAfterFinalTag", kinds: []syntax.NodeKind{syntax.KFuncCall}, check: func(ctx *analysis.Context, n syntax.Node) { CheckSecretstreamPushAfterFinalTag(ctx, n, "finding") }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
	if err != nil {
		t.Fatal(err)
	}
	if got := e.Analyze(syntax.Parse("case.php", []byte("<?php unknown();"), syntax.Options{})); len(got) != 0 {
		t.Fatal(got)
	}
}

func TestExpansionCryptoDBEscapedState(t *testing.T) {
	for _, tc := range []struct {
		id, src string
		kind    syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{
		{"SodiumSecretboxNonceReused", `function f($k,$n){sodium_crypto_secretbox('a',$n,$k);$cb=function()use(&$k){$k='changed';};$cb();sodium_crypto_secretbox('b',$n,$k);}`, syntax.KFuncCall, CheckSodiumSecretboxNonceReused},
		{"PdoGroupedFetchReadsRemovedColumn", `function f(PDO $p){$r=$p->query('SELECT x AS a,y AS b FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);$cb=function()use(&$r){$r=[];};$cb();echo $r['x'][0]['a'];}`, syntax.KArrayDimFetch, CheckPdoGroupedFetchReadsRemovedColumn},
	} {
		p := expansionCryptoDBProbe{id: tc.id, kinds: []syntax.NodeKind{tc.kind}, check: func(ctx *analysis.Context, n syntax.Node) { tc.check(ctx, n, "finding") }}
		e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
		if err != nil {
			t.Fatal(err)
		}
		if got := e.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.src), syntax.Options{})); len(got) != 0 {
			t.Fatal(got)
		}
	}
}
