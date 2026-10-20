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
  - when it was unset or null, any test ruling null out proves success (`isset($signature)`, truthiness, `!empty()`, `!== null`): it is assigned `null` by the immediately preceding statement, or it is a local of a function-like scope with no earlier mention (not a parameter, `global`, `static` or captured variable), the call is not inside a loop, and the scope has no label, `include`/`require` or `eval`;
  - when the immediately preceding statement assigns `''`, any test ruling the empty string out proves success (truthiness, `!empty()`, `!== ''`); `isset()` does not;
  - `openssl_verify(..., $signature, ...)` (optionally through a `(string)` cast) proven to have returned `1` proves success whatever the variable held.
  - A test protects a mention where the conditions guarding it hold only after success: the branch it selects (`if`/`elseif`/`else` body, ternary or `match (true)` arm, right operand of `&&`/`||`), or code after an `if` whose branch for a failure always leaves, with or without other arms. Tests combine through `!`, `&&` and `||`, compare with booleans, or are kept in a local; they must be evaluated after the signing call, and a jump label between a test and the mention voids it.
  - Mentions that are the direct operand of `isset`, `empty`, `!`, a comparison with a constant or a truthiness condition (parentheses and `&&`/`||` aside), or that sit in the arguments of `openssl_verify`, are part of a test; a value computed from the signature (`!deliver(base64_encode($signature))`) is a use, not a test. A capture by a closure is a use where the closure is created. Any other unprotected mention, a stale or unknown earlier value (file scope, loop iteration, earlier write), or a signature target that is not a plain variable keeps D1.
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
