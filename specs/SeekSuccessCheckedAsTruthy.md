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

- D1. Report resolved fseek or rewind-result misuse only for fseek: direct condition, logical negation, loose boolean comparison or assignment-in-condition treating nonzero as success. Preserve no semantic claims for arbitrary branch bodies: flag truthiness tests because fseek uses zero success and minus-one failure. Include comparisons to true/false, not strict comparisons to integer zero.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Strict result comparisons against 0 or -1 and merely storing the result are excluded. rewind returns boolean and must not be flagged.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore SeekSuccessCheckedAsTruthy` and `@noinspection SeekSuccessCheckedAsTruthy` suppress this inspection.

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
