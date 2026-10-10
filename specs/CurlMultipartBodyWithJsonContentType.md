---
id: CurlMultipartBodyWithJsonContentType
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlMultipartBodyWithJsonContentType

## Summary

An array supplied as cURL post fields generates multipart content. Encode the body explicitly before declaring it to be JSON.

## Detection

- D1. Same resolved curl handle has CURLOPT_POSTFIELDS set to literal array and final CURLOPT_HTTPHEADER list includes Content-Type: application/json. No later body/header replacement before execution. Highlight conflicting header configuration.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the final CURLOPT_HTTPHEADER argument list containing the conflicting content type.
- Severity: warning.
- Enabled by default: true.
- Message: Encode the request body as JSON before declaring JSON content.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=curl_init();curl_setopt($h,CURLOPT_POSTFIELDS,["id"=>1]);curl_setopt($h,CURLOPT_HTTPHEADER,<warning descr="Encode the request body as JSON before declaring JSON content.">["Content-Type: application/json"]</warning>);curl_exec($h);
```

Valid case:

```php
<?php
$h=curl_init($url);curl_setopt($h,CURLOPT_POSTFIELDS,json_encode(["id"=>1]));curl_setopt($h,CURLOPT_HTTPHEADER,["Content-Type: application/json"]);curl_exec($h);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
