---
id: AeadDecryptionTagLengthUnchecked
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "7.1", max: "" }
---

# AeadDecryptionTagLengthUnchecked

## Summary

Authenticated decryption needs a tag of the expected length. Validate request-supplied tag lengths before calling the decryptor. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved openssl_decrypt uses literal GCM cipher and tag directly derived from request superglobal, without a dominating strlen(tag)===16 guard. Restrict to configured expected GCM tag length 16 by default; CCM and known constant tags excluded.
- D1e. Source provenance additionally requires an untouched lexical superglobal prefix. Exclude earlier explicit writes to or reference escapes of that superglobal and any earlier write or reference target involving GLOBALS. Reject all earlier object construction and unknown calls, including calls with no arguments, method calls and static calls; only resolved strlen, is_string, ctype_alnum, ctype_digit or in_array are supported pure-call exceptions. Traverse at most 4096 AST nodes, excluding nested variable scopes; exhaustion or unsupported source mutation discards the request-origin proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the unchecked authentication tag argument.
- Severity: warning.
- Enabled by default: false.
- Message: Validate the authentication tag length before decrypting.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

| Option | Type | Default | Effect |
| --- | --- | --- | --- |
| tagLength | int | `16` | Configure the explicit policy evidence described in D1. |

## PHP versions

Requires PHP 7.1 or later.

## Examples

```php
<?php
openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,<warning descr="Validate the authentication tag length before decrypting.">$_POST["tag"]</warning>);
```

Valid case:

```php
<?php
$tag=$_POST["tag"];if(strlen($tag)!==16){throw new RuntimeException("Bad tag");}$p=openssl_decrypt($c,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
