---
id: ScandirFailureIterated
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ScandirFailureIterated

## Summary

Identify foreach iterable directly produced by scandir or its unchanged local result without a dominating false-rejecting guard.

## Detection

- D1. Report foreach iterable directly produced by scandir or its unchanged local result without a dominating false-rejecting guard.
- D1a. A guard rejects false when it tests the same unchanged local with: strict or loose comparison against `false` (`=== false`, `!== false`, `== false`, `!= false`, either operand order), plain truthiness (`$entries`, `!$entries`), `empty($entries)`, or the builtin `is_array($entries)`, each optionally negated or parenthesized. Truthiness tests count because an empty listing gives the loop nothing to do.
- D1b. A guard dominates the loop in two positions. Inside a branch: the loop sits in an `if`/`elseif` body whose condition proves non-false when true, or in a later `elseif`/`else` branch after a condition that proves non-false when false. After an early exit: an earlier statement of the same or an enclosing statement list is an `if` whose body always leaves the path (return, throw, exit, break, continue) and whose condition proves non-false when false, or an `if`/`else` without `elseif` whose `else` body always leaves and whose condition proves non-false when true.
- D1c. Composition is directional. A condition known to hold proves non-false through any `&&`/`and` operand but only through every `||`/`or` operand; a condition known to fail proves it through any `||`/`or` operand but only through every `&&`/`and` operand. So `if (!$entries || $skip) { return; }` guards the loop while `if (!$entries && $skip) { return; }` does not.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1a. A guard located outside a loop that assigns the iterated local is ignored: later iterations may see a different value.
- E1b. `assert()` is not a guard. Assertions can be compiled out, so the loop still receives false in production and the finding stays.
- E2. Both `@custos-ignore ScandirFailureIterated` and `@noinspection ScandirFailureIterated` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Check directory enumeration before iterating.`

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
foreach (scandir($directory) as $entry) { echo $entry; }
```

Valid case:

```php
<?php
$entries = scandir($directory); if ($entries !== false) { foreach ($entries as $entry) { echo $entry; } }
```

Also valid, guarded by an early exit on a falsy listing:

```php
<?php
$entries = scandir($directory);
if (!$entries) {
    return;
}
foreach ($entries as $entry) { echo $entry; }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
