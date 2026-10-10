---
id: BasicAuthenticationOverPlainHttp
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# BasicAuthenticationOverPlainHttp

## Summary

Basic authentication credentials sent over plain HTTP are exposed in transit. Use HTTPS for the configured request.

## Detection

- D1. Resolved curl_init/curl_setopt URL is literal http:// and same curl handle configures CURLOPT_USERPWD with a known nonempty credential string before curl_exec, without known secure URL replacement. Highlight credentials option.
- D1c. The initial supported execution proof is an immediately following curl_exec on the same owned handle, with known prior URL replacements.
- D1d. Require CURLOPT_USERPWD to be a statically known nonempty credential string. Authentication must remain at its Basic default or be explicitly configured to exactly CURLAUTH_BASIC; Digest, Negotiate, NTLM, mixed masks and unknown policies do not establish Basic credentials on the wire. Unknown setters, bulk curl_setopt_array configuration, handle escapes and unknown same-handle operations invalidate the final request proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Protect basic authentication with HTTPS.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$ch=curl_init("http://api.example.test"); <warning descr="Protect basic authentication with HTTPS.">curl_setopt($ch,CURLOPT_USERPWD,"user:secret")</warning>; curl_exec($ch);
```

Valid case:

```php
<?php
$h=curl_init("https://api.example.test");curl_setopt($h,CURLOPT_USERPWD,$credentials);curl_exec($h);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
