---
id: ArrayRandInvalidCount
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayRandInvalidCount

## Summary

Identify array_rand when count is a proven integer below one or greater than the exact cardinality of a known array; empty arrays with omitted count also violate.

## Detection

- D1. Report array_rand when count is a proven integer below one or greater than the exact cardinality of a known array; empty arrays with omitted count also violate.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values or escaped aliases invalidate proof. When the proof is inconclusive, nothing is reported.
- D4. A count or array read from a local variable is proven only while nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument a resolved callee binds by reference, other than the inspected argument itself); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it, or a by-reference capture by the closure that is the scope, removes the proof too. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, does as well. The same holds for every local whose value an assignment stores into it, whatever the assigned expression computes, directly or through further assignments. A read inside an arrow function also checks the scopes the arrow function captures from. A withdrawn count yields no finding; a withdrawn array keeps only the size-independent check that the count is at least one.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1b. A count or array read from a property (`$this->pool`,
  `self::$slaves`) is not proven: other methods may have changed it since
  its default or the constructor's write.
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
