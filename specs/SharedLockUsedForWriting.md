---
id: SharedLockUsedForWriting
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SharedLockUsedForWriting

## Summary

A shared lock does not provide exclusive access for a write. Acquire an exclusive lock before changing the protected stream.

## Detection

- D1. Resolved flock(handle,LOCK_SH optionally OR LOCK_NB) precedes fwrite/fputs/ftruncate on same simple-local handle before unlock/exclusive-lock change. Unknown handle escape invalidates proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Acquire an exclusive lock before writing.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=tmpfile();flock($h,LOCK_SH);<warning descr="Acquire an exclusive lock before writing.">fwrite($h,$data)</warning>;
$g=tmpfile();flock($g,LOCK_NB|LOCK_SH);<warning descr="Acquire an exclusive lock before writing.">ftruncate($g,0)</warning>;
```

Valid case:

```php
<?php
flock($h,LOCK_EX);fwrite($h,$data);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
