---
id: IntdivZeroDivisor
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# IntdivZeroDivisor

## Summary

Identify intdiv with second operand proven integer zero.

## Detection

- D1. Report intdiv with second operand proven integer zero.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. A variable divisor counts as zero only when the single assignment reaching the call is an integer zero and nothing else can have rewritten the variable before the call runs. Besides plain, compound and reference assignment, these writes cancel the proof when they sit earlier in the same function body (or the same file for top-level code), or anywhere in that scope when the call is inside a loop: `++` or `--` on the variable, a closure capturing it by reference, a `foreach` whose key or value target names it (destructuring included), a `catch` binding of that name, and a call to `extract()`. Code inside a nested function does not count, apart from that by-reference capture.
- D4. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. A counter that starts at zero and is stepped before the division is not reported, whether the step is in straight-line code, in an enclosing loop or in a loop that completed earlier. A step located after the call, outside any loop around it, does not cancel the report.
- E3. Both `@custos-ignore IntdivZeroDivisor` and `@noinspection IntdivZeroDivisor` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Supply a nonzero integer divisor.`

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
intdiv(14, 0);
```

Valid cases:

```php
<?php
intdiv(14, 2);

function averages(array $samples): void
{
    $count = 0;
    $sum = 0;
    foreach ($samples as $sample) {
        $sum += $sample;
        $count++;
        echo intdiv($sum, $count), "\n";
    }
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
