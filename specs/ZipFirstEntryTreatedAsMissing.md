---
id: ZipFirstEntryTreatedAsMissing
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ZipFirstEntryTreatedAsMissing

## Summary

Distinguish a missing ZIP entry from index zero.

## Detection

- D1. Report direct truthiness, negation or assignment-in-condition of resolved ZipArchive::locateName. Zero is a valid index. Highlight the truthiness expression containing an assignment or negation; strict false comparisons and later separate comparisons are excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

Preserve assignment precedence in negated conditions: the repaired comparison surrounds an unparenthesized assignment before comparing its result with false.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Distinguish a missing ZIP entry from index zero.`

## Fix

Replace !call with call === false; positive call truthiness with call !== false; == false with === false and != false with !== false. Parenthesize as required, preserve evaluation count. Fix local-result conditions only when that local is proven unmodified.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$z=new ZipArchive(); if (<warning descr="Distinguish a missing ZIP entry from index zero.">!$z->locateName("settings.json")</warning>) { echo "missing"; }
```

Fixed result:

```php
<?php
$z=new ZipArchive(); if ($z->locateName("settings.json") === false) { echo "missing"; }
```

Valid case:

```php
<?php
$z=new ZipArchive(); if ($z->locateName("settings.json") === false) { echo "missing"; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
