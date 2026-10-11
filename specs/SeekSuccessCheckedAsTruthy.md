---
id: SeekSuccessCheckedAsTruthy
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SeekSuccessCheckedAsTruthy

## Summary

Compare the seek result with zero.

## Detection

This is an opt-in policy. Truthiness, previous-state checks, raw wait-status comparisons and dispatch acknowledgements can be intentional; enabling this inspection requests explicit handling of the contract described below.

- D1. Report a resolved `fseek` call whose result is tested for truthiness as the condition of `if`, `elseif`, `else if`, `while`, `do … while` or `for` (only the last expression of a `for` condition list decides the loop). The result may reach the condition directly, through parentheses, through logical negation, through a loose or strict comparison with `true`/`false`, or through a plain `=` assignment whose assigned value is the call. These tests treat nonzero as success although fseek returns zero on success and minus one on failure. No claim is made about other positions such as ternary conditions, `&&`/`||` operands or branch bodies.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Strict result comparisons against 0 or -1 and merely storing the result are excluded. rewind returns boolean and must not be flagged.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore SeekSuccessCheckedAsTruthy` and `@noinspection SeekSuccessCheckedAsTruthy` suppress this inspection.
- E3. A first-class callable `fseek(...)` creates a closure and performs no seek, so it is never reported.
- E4. A compound assignment in the condition (`$x += fseek(…)`, `$x ??= fseek(…)`, …) tests the combined value rather than the seek result, and a call used inside the assignment target (`$a[fseek(…)] = 1`) is not the tested value; neither is reported.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Compare the seek result with zero.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
if (<warning descr="Compare the seek result with zero.">fseek($fp, 0)</warning>) { echo "seeked"; }
```

Valid case:

```php
<?php
if (fseek($fp, 0) === 0) { echo "seeked"; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
