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

- D1. Track exact builtin SplQueue, SplStack, SplDoublyLinkedList, SplHeap and SplPriorityQueue size from construction, insert/push/unshift/enqueue and extract/pop/shift/dequeue. Report a removal or top/bottom read when the collection is proven empty.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias or unknown call involving the tracked value. When the proof is inconclusive, nothing is reported.
- D4. The history of a call inside a `while`, `do`, `for` or `foreach` loop covers the whole body of the outermost enclosing loop that does not itself construct the collection, because code written after the call runs before it on the next pass. Within that replayed part an insertion or any unmodeled use discards the proof; an earlier-turn removal may also prevent an exact ordered replay. Unknown receivers of matching mutators prevent a proof when they cannot be distinguished from the tracked collection. A statement label before the call extends the history to the end of the scope in the same way.

## Exceptions (no report)

Only a fully modeled, bounded history proves emptiness. Array writes, unmodeled methods, conditional insertions and calls receiving the collection discard that proof.

- E1. Unknown initial collection, aliases escaping or custom method behavior prevent empty proof. Merely potentially empty collections are excluded. The collection variable may appear only as the receiver of a method call or as the target of a plain assignment: any other occurrence (call argument, assigned value stored in a variable, property or array element, array literal item, by-reference binding, array-style write, `return`, `yield`, iteration) counts as an escape.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore EmptySplCollectionExtraction` and `@noinspection EmptySplCollectionExtraction` suppress this inspection.
- E3. No report when an `isEmpty()` or `count()` test on the same receiver, earlier in the scope, controls whether the call runs and holds only for a collection with an element: `isEmpty()` false, or `count()` compared with an integer so that zero is excluded (`> 0`, `>= 1`, `!== 0`, `=== 1`, `0 <`, a truthy `count()`, and the failing side of `< 1`, `<= 0` or `=== 0`). The test may be negated, parenthesized, compared with `true`/`false`, combined through `&&`/`and` and `||`/`or` (a conjunction that holds or a disjunction that failed), or kept in a local read later. It controls the call:
  - in the matching branch of an `if`/`elseif`/`else` chain, of a ternary (including the short `?:` form) or of a `match (true)` arm;
  - in the right operand of `&&`/`and` or `||`/`or`;
  - in the body or step of a `while` or `for` loop whose condition (the last one of a `for`) is the test;
  - after an `if` whose branch for an empty collection always leaves, with or without other arms.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Check that the collection contains an element.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

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

```php
<?php
$pending = new SplStack();
foreach ($tokens as $token) {
    if ($token === 'close') {
        $pending->pop(); // filled by an earlier pass
    } else {
        $pending->push($token);
    }
}
$jobs = new SplQueue();
$registry->jobs = $jobs; // escaped
$next = $jobs->isEmpty() ? null : $jobs->dequeue();
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
