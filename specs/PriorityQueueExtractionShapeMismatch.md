---
id: PriorityQueueExtractionShapeMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# PriorityQueueExtractionShapeMismatch

## Summary

Match the priority queue extraction mode.

## Detection

- D1. Track builtin SplPriorityQueue mode (default EXTR_DATA) and literal inserted data. Report indexing extracted/current/top values by data or priority when the proven extraction mode does not produce that field and the inserted data type cannot provide it. EXTR_BOTH provides both fields; EXTR_PRIORITY provides only priority itself. Report the indexed expression.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Unknown inserted data (which may itself be an array), overridden extraction behavior or mode invalidated by unknown calls prevent a finding.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore PriorityQueueExtractionShapeMismatch` and `@noinspection PriorityQueueExtractionShapeMismatch` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Match the priority queue extraction mode.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 5.3. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
$q = new SplPriorityQueue(); $q->insert("task", 5); <warning descr="Match the priority queue extraction mode.">$q->extract()["priority"]</warning>;
```

Valid case:

```php
<?php
$q = new SplPriorityQueue(); $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH); $q->insert("task", 5); $q->extract()["priority"];
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
