---
id: ZipCloseFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ZipCloseFailureUnchecked

## Summary

Writing an archive is not complete until finalization succeeds. Check the close result before returning success. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. A proven ZipArchive receiver has a successful checked open, then addFromString/addFile mutation, and a discarded close result followed by unconditional literal true return from the containing function. Highlight close; no success-name heuristics.
- D1a. The initial supported proof is consecutive checked-open rejection guard, archive mutation, discarded close, and return true statements.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Check archive finalization before returning success.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
function archive($p){$z=new ZipArchive();if($z->open($p,ZipArchive::CREATE)!==true){return false;}$z->addFromString("a.txt","a");<warning descr="Check archive finalization before returning success.">$z->close()</warning>;return true;}
```

Valid case:

```php
<?php
function archive($p){$z=new ZipArchive();if($z->open($p,ZipArchive::CREATE)!==true){return false;}$z->addFromString("a.txt","a");return $z->close();}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
