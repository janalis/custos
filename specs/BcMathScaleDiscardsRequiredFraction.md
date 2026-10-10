---
id: BcMathScaleDiscardsRequiredFraction
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# BcMathScaleDiscardsRequiredFraction

## Summary

Retain fractional digits in decimal arithmetic. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. For bcadd, bcsub or bcmul with constant decimal strings and explicit constant scale, exact decimal evaluation proves nonzero digits will be discarded. Enabling selects a policy prohibiting this truncation; do not infer expected scale from comments.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore BcMathScaleDiscardsRequiredFraction and @noinspection BcMathScaleDiscardsRequiredFraction suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Retain fractional digits in decimal arithmetic.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Invalid-input error forms vary by PHP version; the BCMath operand contract is unchanged.

## Examples

```php
<?php
$x = <warning descr="Retain fractional digits in decimal arithmetic.">bcadd('2.125', '0.50', 1)</warning>;
```

Valid case:

```php
<?php
$x = bcadd('2.125', '0.50', 3);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
