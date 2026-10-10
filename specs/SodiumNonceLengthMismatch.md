---
id: SodiumNonceLengthMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.2", max: "" }
---

# SodiumNonceLengthMismatch

## Summary

Secretbox requires its defined nonce byte length. Construct the nonce using the extension's required-length constant.

## Detection

- D1. Resolved sodium_crypto_secretbox second argument has known byte length other than 24. Recognize literal strings, str_repeat of literal strings and literal counts, and random_bytes of a positive literal count. Highlight nonce.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use the required secretbox nonce length.

## Fix

F1. When the nonce is a direct resolved random_bytes call with a positive literal length, replace only that length node with \SODIUM_CRYPTO_SECRETBOX_NONCEBYTES. Preserve comments and argument formatting outside the literal span. Do not fix stored random-byte results, dynamic or variable counts, string literals, or other nonce constructors.

## Options

None.

## PHP versions

Requires PHP 7.2 or later.

## Examples

```php
<?php
$c=sodium_crypto_secretbox($message,<warning descr="Use the required secretbox nonce length.">random_bytes(12)</warning>,$key);
```

```php
<?php
$c=sodium_crypto_secretbox($message,random_bytes(\SODIUM_CRYPTO_SECRETBOX_NONCEBYTES),$key);
```

Valid case:

```php
<?php
$c=sodium_crypto_secretbox($message,random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES),$key);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
