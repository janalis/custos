---
id: ZipEntryNameCollision
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ZipEntryNameCollision

## Summary

Make ZIP entry replacement explicit. This opt-in rule applies the policy described in Detection.

## Detection

- D1. Enabled policy forbids implicit ZIP entry replacement. Report resolved addFromString of a proven literal name in the successful true branch of an earlier addFromString on the same receiver with that identical name, unless close/deleteName invalidates the preceding entry. Explicit flags and unknown names are excluded; other insertion patterns remain outside this guarded subset.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

Only the successful branch establishes the earlier entry; exclude fallback additions in the failed branch.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Make ZIP entry replacement explicit.`

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
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFromString("note.txt","one")) { <warning descr="Make ZIP entry replacement explicit.">$z->addFromString("note.txt","two")</warning>; }
```

Valid case:

```php
<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); $z->addFromString("first.txt","one"); $z->addFromString("second.txt","two");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
