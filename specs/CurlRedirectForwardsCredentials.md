---
id: CurlRedirectForwardsCredentials
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlRedirectForwardsCredentials

## Summary

Unrestricted authentication can forward credentials across redirected hosts. Keep credential forwarding within an explicit trusted redirect policy. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Same resolved curl handle enables CURLOPT_FOLLOWLOCATION and CURLOPT_UNRESTRICTED_AUTH and configures CURLOPT_USERPWD to a known nonempty string before execution. Empty or unknown credentials do not establish a finding. Highlight unrestricted-auth option; unknown URL does not excuse enabled credential forwarding.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the CURLOPT_UNRESTRICTED_AUTH option constant.
- Severity: warning.
- Enabled by default: false.
- Message: Restrict credential forwarding across redirects.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=curl_init();curl_setopt($h,CURLOPT_FOLLOWLOCATION,true);curl_setopt($h,<warning descr="Restrict credential forwarding across redirects.">CURLOPT_UNRESTRICTED_AUTH</warning>,true);curl_setopt($h,CURLOPT_USERPWD,"user:secret");curl_exec($h);
```

Valid case:

```php
<?php
$h=curl_init($url);curl_setopt($h,CURLOPT_USERPWD,$auth);curl_setopt($h,CURLOPT_FOLLOWLOCATION,true);curl_exec($h);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
