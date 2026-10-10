---
id: BcryptPasswordTruncation
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# BcryptPasswordTruncation

## Summary

bcrypt supports only a limited password byte length. Apply an explicit supported password-length policy before hashing.

## Detection

- D1. Resolved password_hash uses explicit PASSWORD_BCRYPT and a password of known byte length greater than 72. Literal/str_repeat provenance only; do not flag PASSWORD_DEFAULT because its algorithm can change.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Keep bcrypt input within its supported byte length.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 5.5 or later.

## Examples

```php
<?php
password_hash(<warning descr="Keep bcrypt input within its supported byte length.">str_repeat("p",80)</warning>,PASSWORD_BCRYPT);
```

Valid case:

```php
<?php
$h=password_hash(str_repeat("p",64),PASSWORD_BCRYPT);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
