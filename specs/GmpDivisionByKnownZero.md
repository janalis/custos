---
id: GmpDivisionByKnownZero
group: Probable bugs
kind: flow
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
- E2. A divisor that reads a local variable is not a known zero when that local cannot be trusted. That local is proven only while nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument a resolved callee binds by reference, other than the inspected argument itself); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it, or a by-reference capture by the closure that is the scope, removes the proof too. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, does as well. The same holds for every local whose value an assignment stores into it, whatever the assigned expression computes, directly or through further assignments. A read inside an arrow function also checks the scopes the arrow function captures from. A counter initialized to zero and incremented in a loop is ordinary code.
- E3. Both @custos-ignore GmpDivisionByKnownZero and @noinspection GmpDivisionByKnownZero suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Use a nonzero GMP divisor.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None.

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

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
