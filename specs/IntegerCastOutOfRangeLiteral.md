---
id: IntegerCastOutOfRangeLiteral
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# IntegerCastOutOfRangeLiteral

## Summary

An integer string outside the supported integer range cannot be preserved by an integer cast. Keep it as a string or use arbitrary precision.

## Detection

- D1. An int cast receives a literal signed decimal integer string beyond signed 64-bit range. Compare digits textually; exclude exponent/fraction notation and values merely beyond 32-bit range.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Preserve integers that exceed the supported integer range.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$id=<warning descr="Preserve integers that exceed the supported integer range.">(int)"999999999999999999999999"</warning>;
```

Valid case:

```php
<?php
$v=(int)"123";
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
