---
id: TlsHandshakePendingAccepted
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# TlsHandshakePendingAccepted

## Summary

Identify stream_socket_enable_crypto result tested only !== false and followed by fwrite in accepted branch, with same stream proven nonblocking through successful stream_set_blocking(false). Only true confirms established handshake.

## Detection

- D1. Report stream_socket_enable_crypto result tested only !== false and followed by fwrite in accepted branch, with same stream proven nonblocking through successful stream_set_blocking(false). Only true confirms established handshake.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. The write must sit in the accepted branch itself, outside nested conditionals, loops, `switch`, `try` and nested functions, and before any completion guard (E3). Switching the stream back to blocking mode does not finish a pending handshake, so it never clears the finding on its own.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore TlsHandshakePendingAccepted` and `@noinspection TlsHandshakePendingAccepted` suppress this native rule.
- E3. A write placed after a completion guard is not reported. A completion guard is a test, written before the write in the accepted branch, proving that a further `stream_socket_enable_crypto` call on the same stream with a literal truthy `enable` argument returned `true` on every path to the write: `if (<handshake> !== true) { ... }` whose body always leaves (return, throw, exit, break, continue), `!<handshake>` and `<handshake> != true` (each rejects both `false` and the pending `0`), `<handshake> === true or exit;`, the body of `if (<handshake> === true)`, or any of these on a variable holding the result while it is not reassigned. A test of `=== false` or `!== false`, a guard whose body can fall through, a guard on another stream or with a non-literal or falsy `enable` argument prove nothing. Writes located before the guard are still reported.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Require a completed TLS handshake before writing.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
if (stream_set_blocking($s, false)) { if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) { fwrite($s, $secret); } }
```

Valid case:

```php
<?php
if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) === true) { fwrite($s, $secret); }
```

Valid case with a completion guard:

```php
<?php
function send($s, string $secret): bool
{
    if (!stream_set_blocking($s, false)) { return false; }
    if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) {
        if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== true) {
            return false;
        }
        fwrite($s, $secret);
        return true;
    }
    return false;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
