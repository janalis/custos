---
id: OpenSslVerifyTruthyResult
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# OpenSslVerifyTruthyResult

## Summary

OpenSSL signature verification returns 1 for success, while an error result can also be truthy. Accept only the explicit success result.

## Detection

- D1. Resolved openssl_verify is directly used as a truthiness condition or direct logical negation. Success is integer 1, while error -1 is truthy. Highlight call.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Accept only verification result 1.

## Fix

F1. Replace direct truthiness condition with (originalCall === 1); preserve all original argument bytes and evaluation count. For direct negation use originalCall !== 1. Do not rewrite a stored result assignment or a general value-producing use. Preserve parentheses and comments surrounding the call; if the chosen replacement span would discard comments, suppress the fix.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
if(<warning descr="Accept only verification result 1.">openssl_verify($s,$sig,$key,OPENSSL_ALGO_SHA256)</warning>){echo "accepted";}
```

```php
<?php
if(openssl_verify($s,$sig,$key,OPENSSL_ALGO_SHA256) === 1){echo "accepted";}
```

Valid case:

```php
<?php
if(openssl_verify($s,$sig,$key,OPENSSL_ALGO_SHA256)===1){echo "accepted";}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
