---
id: SelectFailureTreatedAsReadiness
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# SelectFailureTreatedAsReadiness

## Summary

Handle selection failure before processing ready streams.

## Detection

- D1. Report a direct or unchanged-local stream_select result compared strictly !== 0. False passes this readiness test. Loose != 0 excludes false and is not reported; explicit three-way handling and other comparison forms are outside this subset.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. No report when a strict comparison of the same stored result with false has already ruled failure out on every path reaching the readiness test: an immediately preceding `if ($r === false)` whose body always exits, an enclosing `if ($r !== false)` body or the else branch of `if ($r === false)`, an elseif arm that follows a leading test or earlier arm `$r === false`, or the left operand of a short-circuit (`$r !== false && ...`, `$r === false || ...`, also with `and`/`or`). Inside such a guard condition, a conjunction counts when either operand excludes false on its true outcome and a disjunction when either operand does on its false outcome. The guard must read the same unchanged variable; a result compared directly from the call expression is never guarded, because a second call yields a different result.
- E4. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Handle selection failure before processing ready streams.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$r=stream_select($read,$write,$except,1); if (<warning descr="Handle selection failure before processing ready streams.">$r!==0</warning>) { echo "ready"; }
```

Valid case:

```php
<?php
$r=stream_select($read,$write,$except,1); if ($r===false) { throw new RuntimeException(); } if ($r>0) { echo "ready"; }
```

Failure excluded first:

```php
<?php
$r=stream_select($read,$write,$except,1); if ($r!==false && $r!==0) { echo "ready"; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
