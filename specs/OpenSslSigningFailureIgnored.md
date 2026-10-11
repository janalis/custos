---
id: OpenSslSigningFailureIgnored
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# OpenSslSigningFailureIgnored

## Summary

Identify openssl_sign whose status is discarded and whose output signature is subsequently encoded, emitted or stored on an unconditional path.

## Detection

- D1. Report openssl_sign whose status is discarded and whose output signature is subsequently encoded, emitted or stored on an unconditional path.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. No report when every later mention of a plain signature variable is either part of a success test or reached only after one. A failed `openssl_sign` leaves the variable untouched, so the test must match what the variable could hold before the call:
  - `isset($signature)` proves success when the variable was unset or null: it is assigned `null` by the immediately preceding statement, or it is a local of a function-like scope with no earlier mention (not a parameter, `global`, `static` or captured variable), the call is not inside a loop, and the scope has no label, `include`/`require` or `eval`.
  - `empty($signature)`, `!$signature` and plain truthiness additionally prove success when the immediately preceding statement assigns `''`.
  - `openssl_verify(..., $signature, ...)` (optionally through a `(string)` cast) compared with `=== 1` / `!== 1` proves success whatever the variable held.
  - A test protects a mention located in the branch it selects (`if` body, `else` without `elseif`, ternary arm, right operand of `&&`), or located after an `if` without `else`/`elseif` whose body always leaves the path when the test fails. Tests combine through `!`, `&&` (success) and `||` (failure). The test must be evaluated after the signing call.
  - Mentions that are the direct operand of `isset`, `empty`, `!` or a truthiness condition (parentheses and `&&`/`||` aside), or that sit in the arguments of `openssl_verify`, are part of a test; a value computed from the signature (`!deliver(base64_encode($signature))`) is a use, not a test. Any other unprotected mention, a stale or unknown earlier value (file scope, loop iteration, earlier write), or a signature target that is not a plain variable keeps D1.
- E3. Both `@custos-ignore OpenSslSigningFailureIgnored` and `@noinspection OpenSslSigningFailureIgnored` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Check signing success before publishing the signature.`

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
openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256); echo base64_encode($signature);
```

Valid case:

```php
<?php
if (openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256)) { echo base64_encode($signature); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
