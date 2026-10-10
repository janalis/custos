---
id: ExecOutputArrayAccumulates
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ExecOutputArrayAccumulates

## Summary

exec appends lines to an existing output array. Reset the array when collecting the independent output of another command. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved exec is called twice with same output-array local, no reset/unset between, and second result is subsequently counted/iterated as a complete command output. Require local initialized [] before first exec; opt-in since aggregation may be intentional.
- D1c. The initial supported proof uses adjacent empty-array initialization and two exec calls, followed immediately by foreach or echo(count(output)) on that same local.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Reset the output array before collecting a separate command result.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$out=[];exec("printf a",$out);<warning descr="Reset the output array before collecting a separate command result.">exec("printf b",$out)</warning>;echo count($out);
```

Valid case:

```php
<?php
$out=[];exec("printf a",$out);$out=[];exec("printf b",$out);echo count($out);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
