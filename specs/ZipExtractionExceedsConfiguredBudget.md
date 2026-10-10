---
id: ZipExtractionExceedsConfiguredBudget
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ZipExtractionExceedsConfiguredBudget

## Summary

Validate ZIP extraction against configured limits. This opt-in rule applies the policy described in Detection.

## Detection

- D1. Enabled policy requires maxEntries and maxExpandedBytes. Report extractTo in a known local archive lifecycle unless a terminating numFiles > configured-count guard and a complete ascending index loop bound extraction. The supported loop starts an accumulator at zero immediately before traversal, reads statIndex for every index, separately rejects false metadata and negative size, rejects size > configured-byte-limit minus accumulated bytes before adding the size, and increments every index once. This proves both total size and overflow safety. Earlier direct statIndex reads alone are insufficient. Unknown external validation receiving the archive is excluded rather than certified safe; metadata budgets do not guarantee runtime decompressor resource limits.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Validate ZIP extraction against configured limits.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

- `maxExpandedBytes`: positive integer; default `20971520`.
- `maxEntries`: positive integer; default `1000`.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$z=new ZipArchive(); $z->open("bundle.zip"); <warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
```

Valid case:

```php
<?php
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){throw new RuntimeException();}
$total += $s["size"];
}
$z->extractTo("output");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
