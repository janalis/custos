---
id: GmpDivisionPairUsedAsNumber
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GmpDivisionPairUsedAsNumber

## Summary

Select a quotient or remainder before using the GMP result.

## Detection

- D1. The array result of gmp_div_qr directly or through an unchanged local variable reaches a numeric-operand slot of gmp_strval, gmp_add, gmp_sub, gmp_mul, gmp_div_q, gmp_div_r, gmp_div_qr, gmp_mod, gmp_cmp, gmp_abs, gmp_neg, gmp_sqrt or gmp_pow (first argument only for gmp_pow) without indexing.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore GmpDivisionPairUsedAsNumber and @noinspection GmpDivisionPairUsedAsNumber suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Select a quotient or remainder before using the GMP result.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
echo <error descr="Select a quotient or remainder before using the GMP result.">gmp_strval(gmp_div_qr('23', '4'))</error>;
```

Valid case:

```php
<?php
$pair = gmp_div_qr('23', '4'); echo gmp_strval($pair[0]);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
