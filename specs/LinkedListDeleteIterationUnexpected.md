---
id: LinkedListDeleteIterationUnexpected
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# LinkedListDeleteIterationUnexpected

## Summary

Preserve linked-list elements during iteration. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy requires non-destructive linked-list traversal. Track setIteratorMode flags for resolved builtin SplDoublyLinkedList (and inherited SplQueue/SplStack behavior); report foreach iterable when the current mode includes IT_MODE_DELETE.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.
- D4. The mode is the one set by the last `setIteratorMode` call that runs unconditionally before the loop on the same list. It only counts as proven when no other mode change can run between that call and the first execution of the loop. Mode changes placed after the loop do not alter that first execution and are ignored.

## Exceptions (no report)

- E1. Exclude IT_MODE_KEEP, unknown modes, clones and custom iteration behavior. Direction bits alone do not imply deletion.
- E3. No report when, after the last unconditional mode call and before the loop, a `setIteratorMode` call (or a method call with a dynamic name) sits where it is not guaranteed to run or cannot be tied to the list: inside an `if`/`else`/`switch` branch, a `try` block, a loop header or body, a ternary or short-circuit operand, or on a nullsafe or unproven receiver. The mode is then unknown rather than `IT_MODE_DELETE`.
- E4. No report when, anywhere before the loop, a closure or arrow function contains a `setIteratorMode` call or refers to the list variable (including `use` captures), or the list variable is placed in an array literal such as a `[$list, 'setIteratorMode']` callable. The list has escaped and its mode can be changed out of sight.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore LinkedListDeleteIterationUnexpected` and `@noinspection LinkedListDeleteIterationUnexpected` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Preserve linked-list elements during iteration.`

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
$l = new SplDoublyLinkedList(); $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE); foreach (<warning descr="Preserve linked-list elements during iteration.">$l</warning> as $item) {}
```

Valid case:

```php
<?php
$l = new SplDoublyLinkedList(); $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP); foreach ($l as $item) {}
```

Valid case, the mode is reset on the only path that reaches the loop:

```php
<?php
$l = new SplQueue();
$l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
try {
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
} catch (Exception $e) {
}
foreach ($l as $item) {
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
