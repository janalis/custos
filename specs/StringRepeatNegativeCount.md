---
id: StringRepeatNegativeCount
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# StringRepeatNegativeCount

## Summary

Identify str_repeat when times is a proven negative integer.

## Detection

- D1. Report str_repeat when times is a proven negative integer.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Integer intervals report only when every value is negative. A bound that changes inside a loop is widened after the loop's second pass, and integer comparisons narrow the values seen in their body. Arithmetic that may overflow PHP integers loses its integer proof. Straight-line code from the start of the scope, without loops, can also establish a negative first read before a later rewrite; unknown branches, prior calls, references and unsupported operations end that additional proof. Fresh assignments supersede earlier steps or extract calls, and merely creating an uninvoked capture does not execute its body. Unknown or crossing intervals give no finding without a concrete entry witness.
- E3. Both `@custos-ignore StringRepeatNegativeCount` and `@noinspection StringRepeatNegativeCount` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Supply a nonnegative repetition count.`

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
str_repeat('cd', -3);
```

Valid case:

```php
<?php
str_repeat('cd', 3);

$depth = -1;
foreach ($nodes as $node) {
    $depth++;
    echo str_repeat('  ', $depth), $node, "\n";
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
