---
id: OpenSslRawCiphertextOptionMismatch
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# OpenSslRawCiphertextOptionMismatch

## Summary

Raw ciphertext and base64-encoded ciphertext require different OpenSSL options. Use matching encoding options when decrypting an encryption result.

## Detection

- D1. A local assigned resolved openssl_encrypt with definite OPENSSL_RAW_DATA is supplied to resolved openssl_decrypt with absent raw-data bit, same proven cipher/key/IV, and no intervening base64_encode. Also detect inverse encode/decode mismatch with definite flag facts.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Match raw-ciphertext options when decrypting.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$key="0123456789012345"; $iv="0123456789012345"; $c=openssl_encrypt($data,"aes-128-cbc",$key,OPENSSL_RAW_DATA,$iv); <warning descr="Match raw-ciphertext options when decrypting.">openssl_decrypt($c,"aes-128-cbc",$key,0,$iv)</warning>;
```

Valid case:

```php
<?php
$c=openssl_encrypt($s,"aes-256-cbc",$key,OPENSSL_RAW_DATA,$iv);$p=openssl_decrypt($c,"aes-256-cbc",$key,OPENSSL_RAW_DATA,$iv);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
