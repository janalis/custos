---
id: HeapIterationConsumesCollection
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# HeapIterationConsumesCollection

## Summary

Iterate a clone to preserve the heap. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy requires heap preservation. Report foreach directly over a resolved builtin SplHeap subtype or SplPriorityQueue object; each iterator step extracts from the heap. Highlight the foreach iterable expression. A clone expression satisfies preservation, written inline or bound to a local as described in E3.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Exclude clone iterables and classes with overridden iterator operations, IteratorAggregate wrappers and unknown concrete heap type. Policy is supplied by enabling the rule, never inferred from later variable names.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore HeapIterationConsumesCollection` and `@noinspection HeapIterationConsumesCollection` suppress this inspection.
- E3. Exclude a plain local variable iterable that provably holds a private copy. The proof is syntactic and limited to the variable's own scope (the whole file for top-level code): the variable has exactly one write, a plain `$copy = clone <expr>;` expression statement that runs on every ordinary path to the loop, and every other mention of the variable is a foreach iterable, the receiver of a method call or the argument of the builtin `count()`. Any other mention keeps the report: a second or conditional assignment, a compound or by-reference assignment, a clone nested in a larger expression, passing or storing the variable, a closure capture, `global`/`static`, or use as a loop target. Variable variables, `$GLOBALS`, `extract()`, `parse_str()`, `include`/`require` and `eval` anywhere in the scope also keep the report. Properties and array elements are never treated as private copies.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Iterate a clone to preserve the heap.`

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
$h = new SplMinHeap(); $h->insert(4); foreach (<warning descr="Iterate a clone to preserve the heap.">$h</warning> as $item) {}
```

Valid case:

```php
<?php
$h = new SplMinHeap(); $h->insert(4); foreach (clone $h as $item) {}
```

A local bound once to a clone is equally valid:

```php
<?php
function show(SplMinHeap $heap): void {
    $copy = clone $heap;
    foreach ($copy as $item) { echo $item; }
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
