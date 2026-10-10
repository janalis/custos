---
id: ZipOpenTruthyResult
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ZipOpenTruthyResult

## Summary

ZIP opening returns true for success and integer codes for errors. Check for true explicitly instead of using truthiness.

## Detection

- D1. Resolved ZipArchive::open call is used directly as an if/while truthiness condition or direct logical negation. Its error integers are nonzero; success is boolean true. Require receiver proven ZipArchive. Highlight call.
- D1b. Report a direct logical negation even outside a branch condition, but offer a fix only for direct if/while truthiness or its negation. Withhold the fix if replacing its condition span would discard comments.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Accept only true from ZipArchive::open.

## Fix

F1. In the proven truthiness condition replace the call with (originalCall === true); preserve all original argument bytes and call evaluation count. For negated condition use originalCall !== true. No fix in numeric comparisons or assignments whose value is used elsewhere.

Comments introduced by `#` are comments too: suppress a fix whenever its replacement would discard them, just as for line and block comments.

## Options

None.

## PHP versions

Requires PHP 5.3 or later.

## Examples

```php
<?php
$z=new ZipArchive(); if(<warning descr="Accept only true from ZipArchive::open.">$z->open($path)</warning>){echo "ready";}
```

```php
<?php
$z=new ZipArchive(); if(($z->open($path)===true)){echo "ready";}
```

Valid case:

```php
<?php
$z=new ZipArchive(); if($z->open($path)===true){echo "ready";}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
