---
id: GmpDivisionByKnownZero
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GmpDivisionByKnownZero

## Summary

Use a nonzero GMP divisor.

## Detection

- D1. gmp_div_q, gmp_div_r, gmp_div_qr or gmp_mod has an integer zero or valid zero-valued integer string divisor. Include proven gmp_init zero values; exclude ambiguous bases and unknown operands.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. A divisor that reads a local variable is not a known zero when that local, or a local it was assigned from, is changed anywhere in the same scope by `++` or `--`, bound by a `catch` clause or imported by reference into a closure, or when the scope calls `extract()` or `parse_str()`. A counter initialized to zero and incremented in a loop is ordinary code.
- E3. Both @custos-ignore GmpDivisionByKnownZero and @noinspection GmpDivisionByKnownZero suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Use a nonzero GMP divisor.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
$n = <error descr="Use a nonzero GMP divisor.">gmp_div_q('23', '0')</error>;
```

Valid cases:

```php
<?php
$n = gmp_div_q('23', '4');

$count = 0;
foreach ($values as $value) {
    $count++;
}
$mean = gmp_div_q($sum, $count);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
