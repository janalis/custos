---
id: PriorityQueueMinOrderAssumption
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# PriorityQueueMinOrderAssumption

## Summary

Reverse priorities for smallest-first extraction. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy requires smallest numeric priority first. Report first extraction from an exact builtin SplPriorityQueue with at least two distinct proven literal numeric priorities since construction, using unmodified inherited comparator. Highlight extract/top/current call. The queue inherently returns the largest priority first.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Exclude a single priority, tied priorities, unknown priorities, subclasses with custom compare, proven negated priorities and invalidated queue state. Enabling states the numeric ordering policy; labels and comments are not evidence.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore PriorityQueueMinOrderAssumption` and `@noinspection PriorityQueueMinOrderAssumption` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Reverse priorities for smallest-first extraction.`

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
$q = new SplPriorityQueue(); $q->insert("first", 2); $q->insert("second", 8); <warning descr="Reverse priorities for smallest-first extraction.">$q->extract()</warning>;
```

Valid case:

```php
<?php
$q = new SplPriorityQueue(); $q->insert("first", -2); $q->insert("second", -8); $q->extract();
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
