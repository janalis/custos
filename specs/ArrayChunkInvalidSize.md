---
id: ArrayChunkInvalidSize
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# ArrayChunkInvalidSize

## Summary

Detect array_chunk with a size proven an integer less than one.

## Detection

- D1. Report array_chunk with a size proven an integer less than one.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases or recursion cannot establish definite facts. When the proof is inconclusive, nothing is reported.
- D4. A size read from a variable is proven only when every change to that variable in the enclosing function (or top-level code) is a plain assignment the flow analysis follows. The same holds for every variable whose value an assignment in that scope copies into it, directly or through further copies (`$size = $seen;`). `++`/`--` on the variable, a closure capturing it by reference (`use (&$size)`), an array item holding a reference to it (`[&$size]`) and an `extract()` call in the same scope each make the size unknown, wherever they appear in that scope.

## Exceptions (no report)

- E1. Unknown sizes and values proven positive are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A size variable that is incremented, decremented, referenced from a closure or array item, or possibly rebound by `extract()` in the calling scope (D4) is unknown: `$size = 0; foreach ($headers as $h) { $size++; } array_chunk($items, $size);` is not reported. Bodies of nested functions and closures are separate scopes and do not count, nor does a by-value capture.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Supply a positive chunk size.`

## Fix

No automatic fix: a replacement size cannot be inferred.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Supply a positive chunk size.">array_chunk($items,0)</error>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
