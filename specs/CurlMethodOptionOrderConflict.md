---
id: CurlMethodOptionOrderConflict
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlMethodOptionOrderConflict

## Summary

cURL method options can override previously selected request modes. Configure one consistent final method before executing the transfer.

## Detection

- D1. Same curl handle sets CURLOPT_POST=true then CURLOPT_HTTPGET=true before one curl_exec, or vice versa, without intervening execution. Highlight later option; definite opposing mode values only.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the later conflicting option constant.
- Severity: warning.
- Enabled by default: true.
- Message: Configure a consistent final HTTP request method.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=curl_init();curl_setopt($h,CURLOPT_POST,true);curl_setopt($h,<warning descr="Configure a consistent final HTTP request method.">CURLOPT_HTTPGET</warning>,true);curl_exec($h);
$g=curl_init();curl_setopt($g,CURLOPT_HTTPGET,true);curl_setopt($g,<warning descr="Configure a consistent final HTTP request method.">CURLOPT_POST</warning>,true);curl_exec($g);
```

Valid case:

```php
<?php
$h=curl_init($url);curl_setopt($h,CURLOPT_HTTPGET,true);curl_exec($h);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
