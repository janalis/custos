---
id: NonBlockingEmptyReadTreatedAsEof
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# NonBlockingEmptyReadTreatedAsEof

## Summary

An empty nonblocking read can mean that no data is currently available. Check EOF separately before closing the stream.

## Detection

- D1. Resolved stream_set_blocking(handle,false) establishes nonblocking mode; condition strictly compares fread(sameHandle) result to empty string and branch closes that handle, without feof guard. Highlight equality.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Check EOF separately from an empty nonblocking read.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=tmpfile();stream_set_blocking($h,false);if(<warning descr="Check EOF separately from an empty nonblocking read.">fread($h,4096)===""</warning>){fclose($h);}
```

Valid case:

```php
<?php
stream_set_blocking($h,false);if(feof($h)){fclose($h);}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
