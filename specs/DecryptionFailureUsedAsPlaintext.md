---
id: DecryptionFailureUsedAsPlaintext
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DecryptionFailureUsedAsPlaintext

## Summary

Decryption can fail and return false. Check success before treating the result as plaintext.

## Detection

- D1. Resolved openssl_decrypt result is passed to a known string sink file_put_contents data argument or strlen without a false-result guard. Track direct nesting or unmodified locals; arbitrary named consumers are excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Check decryption success before using the plaintext.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$plain=openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag); file_put_contents($path,<warning descr="Check decryption success before using the plaintext.">$plain</warning>);
```

Valid case:

```php
<?php
$p=openssl_decrypt($c,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);if($p!==false){file_put_contents($path,$p);}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
