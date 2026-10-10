---
id: EmptySplCollectionExtraction
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# EmptySplCollectionExtraction

## Summary

Check that the collection contains an element.

## Detection

- D1. Track exact builtin SplQueue, SplStack, SplDoublyLinkedList, SplHeap and SplPriorityQueue size from construction, insert/push/unshift/enqueue and extract/pop/shift/dequeue. Report a removal or top/bottom read when the collection is proven empty. Recognize a local isEmpty/count guard.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

Only a fully modeled, bounded history proves emptiness. Array writes, unmodeled methods, conditional insertions and calls receiving the collection discard that proof.

- E1. Unknown initial collection, aliases escaping or custom method behavior prevent empty proof. Merely potentially empty collections are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore EmptySplCollectionExtraction` and `@noinspection EmptySplCollectionExtraction` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Check that the collection contains an element.`

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
$q = new SplQueue(); <error descr="Check that the collection contains an element.">$q->dequeue()</error>;
```

Valid case:

```php
<?php
$q = new SplQueue(); $q->enqueue("one"); $q->dequeue();
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
