---
id: FractionalArrayKeyTruncation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FractionalArrayKeyTruncation

## Summary

PHP converts float array keys to integers and loses their fractional part. Use a string key when the complete value identifies the entry.

## Detection

- D1. A literal array key is a finite float literal with a nonzero fractional part; highlight key. Negative unary literals are included. Whole-valued floats and computed unknown keys are excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use a string key to preserve the fractional value.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$a=[<warning descr="Use a string key to preserve the fractional value.">4.8</warning>=>1,<warning descr="Use a string key to preserve the fractional value.">-2.5</warning>=>2];
```

Valid case:

```php
<?php
$a=["4.8"=>"discount"];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
