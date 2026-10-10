---
id: CurlMultiSuccessAssumedPerTransfer
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlMultiSuccessAssumedPerTransfer

## Summary

Successful multi-transfer execution does not establish that each request succeeded. Inspect the completion result for every added transfer.

## Detection

- D1. A curl_multi_exec result compared to CURLM_OK and completed running count guards unconditional true return, with no curl_multi_info_read inspection on that handle. Match complete function-local transfer lifecycle, not arbitrary success-named consumers.
- D1a. Require at least one resolved curl_multi_add_handle on the same locally initialized multi handle before execution. Empty multi handles are valid and excluded. Any intervening curl_multi_remove_handle on that same multi handle invalidates the nonempty-transfer proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Inspect each completed transfer result.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
function bad(){$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(<warning descr="Inspect each completed transfer result.">curl_multi_exec($h,$running)</warning>===CURLM_OK && !$running){return true;}}
```

Valid case:

```php
<?php
function run($m){curl_multi_exec($m,$running);while($info=curl_multi_info_read($m)){if($info["result"]!==CURLE_OK){return false;}}return !$running;}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
