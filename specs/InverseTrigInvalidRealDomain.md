---
id: InverseTrigInvalidRealDomain
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# InverseTrigInvalidRealDomain

## Summary

Keep the inverse trigonometric argument within minus one and one.

## Detection

- D1. asin or acos receives a proven real numeric argument less than -1 or greater than 1. Exclude unknown/non-numeric inputs.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore InverseTrigInvalidRealDomain and @noinspection InverseTrigInvalidRealDomain suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Keep the inverse trigonometric argument within minus one and one.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
$n = <warning descr="Keep the inverse trigonometric argument within minus one and one.">asin(1.25)</warning>;
```

Valid case:

```php
<?php
$n = asin(0.25);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
