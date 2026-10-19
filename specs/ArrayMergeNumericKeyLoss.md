---
id: ArrayMergeNumericKeyLoss
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# ArrayMergeNumericKeyLoss

## Summary

Detect array_merge with a literal input whose integer key is renumbered by the merge; report the merge call.

## Detection

- D1. Report array_merge with a literal array input holding an integer key that the merge provably moves to another index. Within one input, array_merge renumbers integer keys from zero in entry order: only integer-keyed entries take an index, so string-keyed items are skipped and a repeated key overwrites its earlier entry without taking a new one. An item without a key takes the next index. A key is renumbered when it differs from the index its entry receives: `['name' => 'x', 1 => 'y']` is reported (1 becomes 0), `['name' => 'x', 0 => 'y']` is not. Negative integer keys are always renumbered. Canonical decimal strings such as `"42"` are integer keys.
- D1a. Each literal input is judged on its own; the offset added by earlier arguments is ignored, so concatenating lists is never reported.
- D1b. An item whose key is not a known integer or string may or may not add an integer entry. After such an item, report only a key that is negative or greater than the highest index any resolution of the unknown keys could give it.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases or recursion cannot establish definite facts. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report inputs whose integer keys already equal their renumbered indexes (lists, lists mixed with string keys, duplicate keys that collapse to a list) or string-only keys; do not guess that an unknown array holds identifiers. Arrays with spread or by-reference items are unknown.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Preserve numeric identifier keys when combining arrays.`

## Fix

No automatic fix: array union and replacement have different collision semantics.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge([42=>"a"],[87=>"b"])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge(['label' => 'x', 1 => 'y'], $more)</warning>;
array_merge(['label' => 'x', 0 => 'y'], $more);
array_merge(['a', 'k' => 'b', 1 => 'c'], $more);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
