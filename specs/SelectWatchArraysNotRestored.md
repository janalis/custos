---
id: SelectWatchArraysNotRestored
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# SelectWatchArraysNotRestored

## Summary

Restore stream watch arrays before each selection. This opt-in rule applies the policy described in Detection.

## Detection

- D1. Enabled policy requires preserving the initial watch set. Report stream_select inside a while loop when a nonempty literal read/write/except array assignment in the same scope precedes the loop and nothing in the loop restores that watch variable between two selections. Null/empty watch sets and restored arrays are excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

Find the latest dominating assignment in the selection operation’s own variable scope. A later null, unknown or conditional assignment discards the initial-array proof. When that search exceeds the analysis budget, nothing is reported.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. The watch variable counts as restored when the loop condition or body, before or after the selection, assigns it (plainly or as a slot of a `[...]`/`list(...)` destructuring pattern), rebinds it as a foreach key or value target, or passes it as an argument to a method, constructor, or function that is not a proven builtin (such code may refill it by reference). Reading it, including through builtin calls, restores nothing.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Restore stream watch arrays before each selection.`

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
$read=[$a,$b]; $write=null; $except=null; while ($running) { <warning descr="Restore stream watch arrays before each selection.">stream_select($read,$write,$except,1)</warning>; }
```

Valid case:

```php
<?php
$all=[$a,$b]; $write=null; $except=null; while ($running) { $read=$all; stream_select($read,$write,$except,1); }
```

Also valid, restored at the end of the body:

```php
<?php
$read=[$a,$b]; $write=null; $except=null; while ($running) { stream_select($read,$write,$except,1); $read=[$a,$b]; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
