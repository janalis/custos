---
id: CredentialedCorsUsesWildcardOrigin
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CredentialedCorsUsesWildcardOrigin

## Summary

Browsers reject a wildcard CORS origin for credentialed requests. Return an explicit allowed origin instead.

## Detection

- D1. Resolved header calls in same response scope definitely set Access-Control-Allow-Origin to * and Access-Control-Allow-Credentials to true, case-insensitive. Account for later replacement/removal and unknown control flow.
- D1c. The supported initial proof uses a contiguous block of known header and header_remove operations. Account for the final replacement/removal state, and do not treat unknown later operations as proof that the conflicting headers remain final.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use an explicit origin for credentialed CORS.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
header("Access-Control-Allow-Origin: *"); <warning descr="Use an explicit origin for credentialed CORS.">header("Access-Control-Allow-Credentials: true")</warning>;
```

Valid case:

```php
<?php
header("Access-Control-Allow-Origin: https://client.example.test");header("Access-Control-Allow-Credentials: true");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
