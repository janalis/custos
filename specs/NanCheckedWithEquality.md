---
id: NanCheckedWithEquality
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# NanCheckedWithEquality

## Summary

NaN does not equal itself, so equality comparisons cannot recognize it. Use is_nan when testing whether a numeric value is NaN.

## Detection

- D1. An equality or inequality comparison has one operand resolved NAN and the other an arbitrary expression. == and === are always false; != and !== always true for NaN. Highlight the comparison.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Shadowed NAN constants and already-correct is_nan calls are excluded.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use is_nan to test for NaN.

## Fix

F1. Replace equality with \is_nan(otherOperand), and inequality with !\is_nan(otherOperand). Preserve the operand bytes and evaluate it once. Do not fix operands with unknown nonnumeric type: is_nan can reject/coerce them. Offer only when type evidence proves int or float, or leave the finding without a fix. If replacing the comparison would discard comments outside the retained operand, suppress the fix.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5. The scalar-type declaration in the example needs PHP 7.0; numeric expression inference can establish a safe fix on earlier versions.

## Examples

```php
<?php
function f(float $value) { if(<warning descr="Use is_nan to test for NaN.">$value===NAN</warning>) { echo "bad"; } }
```

```php
<?php
function f(float $value) { if(\is_nan($value)) { echo "bad"; } }
```

Valid case:

```php
<?php
if(is_nan($value)){echo "bad";}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
