---
id: FileStatCacheAfterExternalMutation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FileStatCacheAfterExternalMutation

## Summary

Cached file metadata can survive changes made by an external command. Clear the stat cache before reading that metadata again. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Repeated filesize/filemtime of same known literal path straddles resolved exec/system command with statically parseable literal touch/truncate/rm on that exact path. No intervening clearstatcache; unsupported shell quoting/operators excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

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

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

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
