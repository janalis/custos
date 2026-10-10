---
id: SodiumSecretboxNonceReused
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "7.2", max: "" }
---

# SodiumSecretboxNonceReused

## Summary

Identify the second secretbox encryption on same proven key and nonce when plaintexts are proven distinct. Identical plaintexts, unknown equality, new nonce and new key exclude.

## Detection

- D1. Report the second secretbox encryption on same proven key and nonce when plaintexts are proven distinct. Identical plaintexts, unknown equality, new nonce and new key exclude.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore SodiumSecretboxNonceReused` and `@noinspection SodiumSecretboxNonceReused` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Use a fresh nonce for each distinct message.`

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
$n = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES); sodium_crypto_secretbox('alpha', $n, $k); sodium_crypto_secretbox('beta', $n, $k);
```

Valid case:

```php
<?php
sodium_crypto_secretbox('alpha', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k); sodium_crypto_secretbox('beta', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
