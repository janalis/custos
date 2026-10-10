---
id: OpenSslKeyLengthMismatch
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# OpenSslKeyLengthMismatch

## Summary

OpenSSL keys must match the chosen cipher's required byte length. Supply a correctly sized key instead of relying on padding or truncation.

## Detection

- D1. Resolved openssl_encrypt/decrypt uses a literal aes-128/192/256 cipher variant; key byte length is proven and differs from 16/24/32 respectively. Track literal string, str_repeat, random_bytes and unmodified local values. Unknown cipher/key produces no report.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the encryption/decryption key argument.
- Severity: warning.
- Enabled by default: true.
- Message: Provide the cipher-required key length.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
openssl_encrypt($data,"aes-256-gcm",<warning descr="Provide the cipher-required key length.">random_bytes(16)</warning>,OPENSSL_RAW_DATA,$iv,$tag);
```

Valid case:

```php
<?php
$c=openssl_encrypt($s,"aes-256-gcm",random_bytes(32),OPENSSL_RAW_DATA,$iv,$tag);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
