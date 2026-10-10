---
id: SplFixedArrayShrinkDiscardsValues
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# SplFixedArrayShrinkDiscardsValues

## Summary

Preserve populated entries before shrinking. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy forbids discarding populated fixed-array slots. Track known size and assigned slots of builtin SplFixedArray; report setSize with smaller nonnegative literal size when at least one populated slot is outside the new bounds. A null value still counts as populated when assignment is proven.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Fresh unpopulated arrays, explicit unset of all discarded slots, enlargement, unchanged size and unknown occupancy are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore SplFixedArrayShrinkDiscardsValues` and `@noinspection SplFixedArrayShrinkDiscardsValues` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Preserve populated entries before shrinking.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Enabling the rule declares the policy stated in Detection; the rule is disabled by default.

## PHP versions

Minimum PHP 5.3. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding with the rule explicitly enabled:

```php
<?php
$a = SplFixedArray::fromArray(["east", "west"]); <warning descr="Preserve populated entries before shrinking.">$a->setSize(1)</warning>;
```

Valid case:

```php
<?php
$a = SplFixedArray::fromArray(["east", "west"]); $a->setSize(3);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
