---
id: BcMathExponentNotation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# BcMathExponentNotation

## Summary

Use ordinary decimal notation for BCMath operands.

## Detection

- D1. A numeric operand of bcadd, bcsub, bcmul, bcdiv, bcmod, bcpow, bcpowmod, bccomp or bcsqrt is a constant string containing exponent notation, with an otherwise valid signed decimal mantissa and integer exponent.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore BcMathExponentNotation and @noinspection BcMathExponentNotation suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Use ordinary decimal notation for BCMath operands.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Invalid-input error forms vary by PHP version; the BCMath operand contract is unchanged.

## Examples

```php
<?php
$x = <error descr="Use ordinary decimal notation for BCMath operands.">bcadd('2e4', '3', 0)</error>;
```

Valid case:

```php
<?php
$x = bcadd('20000', '3', 0);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
