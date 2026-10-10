---
id: SodiumSecretboxKeyLengthMismatch
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "7.2", max: "" }
---

# SodiumSecretboxKeyLengthMismatch

## Summary

Identify sodium_crypto_secretbox or open when key byte length is proven other than SODIUM_CRYPTO_SECRETBOX_KEYBYTES; unknown length excludes.

## Detection

- D1. Report sodium_crypto_secretbox or open when key byte length is proven other than SODIUM_CRYPTO_SECRETBOX_KEYBYTES; unknown length excludes.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore SodiumSecretboxKeyLengthMismatch` and `@noinspection SodiumSecretboxKeyLengthMismatch` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Supply a correctly sized secretbox key.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 7.2 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
sodium_crypto_secretbox($message, $nonce, random_bytes(16));
```

Valid case:

```php
<?php
sodium_crypto_secretbox($message, $nonce, sodium_crypto_secretbox_keygen());
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
