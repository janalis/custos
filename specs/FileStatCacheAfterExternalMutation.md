---
id: FileStatCacheAfterExternalMutation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "8.2" }
---

# FileStatCacheAfterExternalMutation

## Summary

On PHP 8.2 and earlier, cached file metadata can survive changes made by a command run through exec or system. Clear the stat cache before reading that metadata again. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Repeated filesize/filemtime of same known literal path straddles resolved exec/system command with statically parseable literal touch/truncate/rm on that exact path. No intervening clearstatcache; unsupported shell quoting/operators excluded.
- D2. Only targets up to PHP 8.2 are inspected: from PHP 8.3 on, exec and system drop the stat cache themselves, so the second read is already fresh.
- D3. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D4. Emit one finding per violating operation, not per supporting evidence statement. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Do not report when the target is PHP 8.3 or later.
- E3. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Clear cached file metadata after external file changes.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Reported for targets PHP 5.3–8.2. On those versions exec and system leave the stat cache untouched, so a repeated filesize/filemtime returns the value cached before the command ran. Starting with PHP 8.3 both functions clear the cache on every call (observed on 8.3 and 8.5), which makes the repeated read fresh and a finding a false positive; the rule is silent there. Process channels that still bypass the cache on newer versions (popen, proc_open) are outside this rule's detection contract.

## Examples

With a PHP 8.2 or earlier target:

```php
<?php
$before=filesize("/tmp/cache.bin");exec("touch /tmp/cache.bin");$after=<warning descr="Clear cached file metadata after external file changes.">filesize("/tmp/cache.bin")</warning>;
```

Valid case:

```php
<?php
$a=filesize("/tmp/custos-demo");exec("truncate -s 0 /tmp/custos-demo");clearstatcache(true,"/tmp/custos-demo");$b=filesize("/tmp/custos-demo");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
