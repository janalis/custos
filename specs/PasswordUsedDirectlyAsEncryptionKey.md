---
id: PasswordUsedDirectlyAsEncryptionKey
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PasswordUsedDirectlyAsEncryptionKey

## Summary

Encryption APIs do not derive a cryptographic key from a password. Apply an appropriate key derivation function before encrypting or decrypting. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved openssl_encrypt/decrypt key argument directly references request data from a configurable password key list, default [password, passphrase]. Require request-origin expression and explicit configured key; generic request fields are excluded. KDF output is not direct password data.
- D1e. Source provenance additionally requires an untouched lexical superglobal prefix. Exclude earlier explicit writes to or reference escapes of that superglobal and any earlier write or reference target involving GLOBALS. Reject all earlier object construction and unknown calls, including calls with no arguments, method calls and static calls; only resolved strlen, is_string, ctype_alnum, ctype_digit or in_array are supported pure-call exceptions. Traverse at most 4096 AST nodes, excluding nested variable scopes; exhaustion or unsupported source mutation discards the request-origin proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the password-derived key argument.
- Severity: warning.
- Enabled by default: false.
- Message: Derive an encryption key from the password.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

| Option | Type | Default | Effect |
| --- | --- | --- | --- |
| passwordKeys | list | `["password", "passphrase"]` | Configure the explicit policy evidence described in D1. |

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
openssl_encrypt($data,"aes-256-gcm",<warning descr="Derive an encryption key from the password.">$_POST["password"]</warning>,OPENSSL_RAW_DATA,$iv,$tag);
```

Valid case:

```php
<?php
$key=hash_pbkdf2("sha256",$_POST["password"],$salt,100000,32,true);$c=openssl_encrypt($s,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
