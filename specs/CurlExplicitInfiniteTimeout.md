---
id: CurlExplicitInfiniteTimeout
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlExplicitInfiniteTimeout

## Summary

An explicitly unlimited cURL timeout can keep workers occupied indefinitely. Set a finite limit when the application's timeout policy requires one. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved curl_setopt CURLOPT_TIMEOUT or CURLOPT_TIMEOUT_MS is explicit integer zero on a handle subsequently executed. Inspect final effective timeout configuration; positive timeout in either unit suppresses.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Set a finite transfer timeout.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=curl_init();curl_setopt($h,CURLOPT_TIMEOUT,0);<warning descr="Set a finite transfer timeout.">curl_exec($h)</warning>;
$g=curl_init();curl_setopt($g,CURLOPT_TIMEOUT_MS,0);<warning descr="Set a finite transfer timeout.">curl_exec($g)</warning>;
```

Valid case:

```php
<?php
$h=curl_init($url);curl_setopt($h,CURLOPT_TIMEOUT,10);curl_exec($h);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
