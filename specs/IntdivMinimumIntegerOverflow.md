---
id: IntdivMinimumIntegerOverflow
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# IntdivMinimumIntegerOverflow

## Summary

Identify intdiv whose dividend resolves to PHP_INT_MIN and divisor to -1; do not assume a particular machine integer width.

## Detection

- D1. Report intdiv whose dividend resolves to PHP_INT_MIN and divisor to -1; do not assume a particular machine integer width.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. A variable operand counts only when every binding of it that can run before the call is a plain `=` assignment, and the same holds for each variable its value is copied from. Bindings located after the call are ignored unless the call sits in a loop of the same scope.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore IntdivMinimumIntegerOverflow` and `@noinspection IntdivMinimumIntegerOverflow` suppress this native rule.
- E3. No report when a variable operand, or a variable it was copied from, is changed in any other way before the call: `++`/`--`, a compound assignment, a destructuring assignment, a `foreach` key or value target, `unset`, `global`, `static`, a `catch` variable or a by-reference assignment.
- E4. No report when such a variable may be reachable through a reference: `$r = &$v`, `[&$v]`, `use (&$v)`, a call-time `&$v`, an unpacked argument, an argument bound to a by-reference parameter of the resolved callee (for example `sscanf` or `preg_match` output arguments), or any argument of a call whose callee cannot be resolved.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Avoid dividing PHP_INT_MIN by negative one.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 7.0 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
intdiv(PHP_INT_MIN, -1);
```

Valid case:

```php
<?php
intdiv(PHP_INT_MIN, 1);
```

Valid case, the dividend is no longer the minimum integer:

```php
<?php
function next_after_minimum(): int
{
    $value = PHP_INT_MIN;
    $value++;

    return intdiv($value, -1);
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
