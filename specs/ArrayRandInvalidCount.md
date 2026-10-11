---
id: ArrayRandInvalidCount
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayRandInvalidCount

## Summary

Identify array_rand when count is a proven integer below one or greater than the exact cardinality of a known array; empty arrays with omitted count also violate.

## Detection

- D1. Report array_rand when count is a proven integer below one or greater than the exact cardinality of a known array; empty arrays with omitted count also violate.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. A count or array read from a plain local variable stays proven only while the variable's scope never rewrites it behind the tracked assignment. Any of these, anywhere in the scope and regardless of statement order, withdraws the proof for that variable: `++` or `--` applied to it or to one of its offsets (`$seen[$key]++`), a `foreach` whose key or value target is the variable or one of its offsets, a by-reference closure capture (`use (&$pool)`), a by-reference array slot holding it (`[&$pool]`), or a call to `extract()`. A variable whose value an assignment in the scope copies into the tracked one, directly or through further copies (`$pool = $seen;`), withdraws it the same way. A withdrawn count yields no finding; a withdrawn array keeps only the size-independent check that the count is at least one.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. An array filled by incrementing its offsets, or a counter advanced with `++`, is not treated as still holding its initial value (D4).
- E3. Both `@custos-ignore ArrayRandInvalidCount` and `@noinspection ArrayRandInvalidCount` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Request a count within the array size.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
array_rand(['oak', 'elm'], 4);
```

Valid cases:

```php
<?php
array_rand(['oak', 'elm'], 1);

function pickTally(array $words)
{
    $tally = [];
    foreach ($words as $word) {
        @$tally[$word]++;
    }
    return array_rand($tally);
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
