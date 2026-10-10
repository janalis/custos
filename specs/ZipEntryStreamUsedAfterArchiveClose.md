---
id: ZipEntryStreamUsedAfterArchiveClose
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ZipEntryStreamUsedAfterArchiveClose

## Summary

An entry stream depends on its containing ZIP archive remaining open. Consume the stream before closing the archive.

## Detection

- D1. A resource local is assigned getStream on a proven ZipArchive receiver; that same receiver is closed and stream_get_contents/fread/fgets subsequently consumes the resource. Track direct locals and invalidate escape/reassignment.
- D1a. The initial supported proof is consecutive local getStream assignment, close call, and stream consumer statements.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Read the entry stream before closing its archive.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$z=new ZipArchive(); $z->open($p); $h=$z->getStream("a.txt"); $z->close(); echo <warning descr="Read the entry stream before closing its archive.">stream_get_contents($h)</warning>;
```

Valid case:

```php
<?php
$z=new ZipArchive(); $z->open($p); $h=$z->getStream("a.txt"); echo stream_get_contents($h); $z->close();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
