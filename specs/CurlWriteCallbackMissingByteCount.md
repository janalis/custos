---
id: CurlWriteCallbackMissingByteCount
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlWriteCallbackMissingByteCount

## Summary

A cURL write callback must return the number of bytes it handled. Return the handled count on every successful callback path.

## Detection

- D1. Resolved curl_setopt CURLOPT_WRITEFUNCTION receives literal closure whose reachable paths lack a returned integer count; recognize no-return closure and explicit null/false return. A complete branch return analysis is required; unknown expressions excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Return the number of bytes handled by the write callback.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,<warning descr="Return the number of bytes handled by the write callback.">function($h,$data){echo $data;}</warning>);
curl_setopt($h,CURLOPT_WRITEFUNCTION,<warning descr="Return the number of bytes handled by the write callback.">function($h,$data){return null;}</warning>);
```

Valid case:

```php
<?php
$h=curl_init($url);curl_setopt($h,CURLOPT_WRITEFUNCTION,function($h,$chunk){echo $chunk;return strlen($chunk);});
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
