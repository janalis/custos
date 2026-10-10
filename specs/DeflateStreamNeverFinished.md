---
id: DeflateStreamNeverFinished
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# DeflateStreamNeverFinished

## Summary

Finish the compressed stream before publishing it. This opt-in rule applies the policy described in Detection.

## Detection

- D1. Enabled policy requires complete standalone compression streams. Report file_put_contents publishing an unchanged local/direct deflate_add result whose explicit flush constant is ZLIB_SYNC_FLUSH, ZLIB_NO_FLUSH or ZLIB_FULL_FLUSH. Concatenated outputs and publication through other consumers remain outside this local subset.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Finish the compressed stream before publishing it.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 7.0 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$c=deflate_init(ZLIB_ENCODING_GZIP); $out=deflate_add($c,"sample",ZLIB_SYNC_FLUSH); <warning descr="Finish the compressed stream before publishing it.">file_put_contents("result.gz",$out)</warning>;
```

Valid case:

```php
<?php
$c=deflate_init(ZLIB_ENCODING_GZIP); $out=deflate_add($c,"sample",ZLIB_FINISH); file_put_contents("result.gz",$out);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
