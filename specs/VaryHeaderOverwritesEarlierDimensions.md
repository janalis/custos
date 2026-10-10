---
id: VaryHeaderOverwritesEarlierDimensions
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# VaryHeaderOverwritesEarlierDimensions

## Summary

Retain the earlier Vary dimensions. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy requires preserving declared cache variation dimensions. Track literal Vary field-name tokens; report a replacement Vary call whose token set omits at least one prior effective dimension. Compare case-insensitively. A replacement with star retains all dimensions.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Appending with replace=false, identical/superset replacement, header_remove, unknown token values and a preceding star followed by star are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore VaryHeaderOverwritesEarlierDimensions` and `@noinspection VaryHeaderOverwritesEarlierDimensions` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Retain the earlier Vary dimensions.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Enabling the rule declares the policy stated in Detection; the rule is disabled by default.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding with the rule explicitly enabled:

```php
<?php
header("Vary: Accept-Encoding"); <warning descr="Retain the earlier Vary dimensions.">header("Vary: Accept-Language")</warning>;
```

Valid case:

```php
<?php
header("Vary: Accept-Encoding"); header("Vary: Accept-Encoding, Accept-Language");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
