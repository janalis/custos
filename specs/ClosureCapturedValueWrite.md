---
id: ClosureCapturedValueWrite
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ClosureCapturedValueWrite

## Summary

Detect a closure writing a scalar captured by value when that closure is invoked and the outer variable is subsequently read without an intervening assignment.

## Detection

- D1. Report a closure writing a scalar captured by value when that closure is invoked and the outer variable is subsequently read without an intervening assignment.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases or recursion cannot establish definite facts. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- Ignore writes or transformations after unconditional termination or inside a literal false if branch; these statements cannot establish a missing update or mapped value.

- E1. Unknown invocations, object member mutation, explicit references, and deliberate inner-local results are excluded.
- E3. The closure uses its copy as a working local, so capturing by reference would change behaviour. No report when, in reachable closure code, any of these holds:
  - the captured variable is mentioned anywhere other than as the target of a write to itself (for example echoed, interpolated, passed to a call, or captured or used by a nested function or class body);
  - the value of a write to it is consumed rather than discarded, i.e. the assignment or increment/decrement is not a whole expression statement (`while ($n-- > 0)`, `echo $n = 4`, a `for` step);
  - the closure returns a value.

  A statement-level write whose right-hand side reads the same variable (`$n = $n + 1;`, `$s .= 'x';`) is still a lost update and stays reported when nothing else reads the copy. This deliberately gives up closures that update the copy and then also read it (`++$n; echo $n;`): intent cannot be told apart there.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Capture the scalar by reference when updating outer state.`

## Fix

No automatic fix: changing capture semantics can affect lifetime and subsequent calls.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$n = 0; $f = <warning descr="Capture the scalar by reference when updating outer state.">function () use ($n) { ++$n; }</warning>; $f(); echo $n;
```

```php
<?php
// The copy is a working local consumed inside the closure: no report.
$left = 3; $tick = function () use ($left) { while ($left-- > 0) { echo '.'; } }; $tick(); echo $left;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
