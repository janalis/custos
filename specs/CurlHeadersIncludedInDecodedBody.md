---
id: CurlHeadersIncludedInDecodedBody
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlHeadersIncludedInDecodedBody

## Summary

Including response headers in a cURL return value makes the combined data unsuitable for JSON decoding. Capture headers separately from the body.

## Detection

- D1. Same resolved curl handle has definite CURLOPT_HEADER=true and CURLOPT_RETURNTRANSFER=true; curl_exec result enters json_decode directly or through unmodified local without splitting headers. Highlight decode call.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Keep HTTP headers separate from the decoded JSON body.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);curl_setopt($h,CURLOPT_RETURNTRANSFER,true);<warning descr="Keep HTTP headers separate from the decoded JSON body.">json_decode(curl_exec($h))</warning>;
```

Valid case:

```php
<?php
$h=curl_init($url);curl_setopt($h,CURLOPT_HEADER,false);curl_setopt($h,CURLOPT_RETURNTRANSFER,true);$data=json_decode(curl_exec($h),true);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
