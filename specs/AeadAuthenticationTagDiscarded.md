---
id: AeadAuthenticationTagDiscarded
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "7.1", max: "" }
---

# AeadAuthenticationTagDiscarded

## Summary

Authenticated encryption requires the generated authentication tag for later verification. Capture and store the tag with the ciphertext.

## Detection

- D1. Resolved openssl_encrypt uses literal aes-128/192/256-gcm or ccm and omits output tag argument, including named-argument mapping. Highlight call. Explicit tag locals that escape/storage receive no finding.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Capture the authenticated encryption tag.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 7.1 or later.

## Examples

```php
<?php
<warning descr="Capture the authenticated encryption tag.">openssl_encrypt($data,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv)</warning>;
```

Valid case:

```php
<?php
$c=openssl_encrypt($s,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
