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

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
