---
id: PasswordComparedWithFreshHash
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# PasswordComparedWithFreshHash

## Summary

Detect password_hash result compared for equality with a stored string hash to decide password acceptance.

## Detection

- D1. Report password_hash result compared for equality with a stored string hash to decide password acceptance.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. password_verify, hashes used only for storage, explicit non-password comparison and unknown builtin targets are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Verify the password against the stored hash.`

## Fix

Replace equality with password_verify(password, stored) only for a strict equality with resolved `PASSWORD_DEFAULT` or `PASSWORD_BCRYPT`, omitted or known empty options, valid argument binding and side-effect-free operands. Unknown algorithm or option validity withholds the fix.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 5.5. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($password, $stored) { if (<warning descr="Verify the password against the stored hash.">password_hash($password, PASSWORD_DEFAULT) === $stored</warning>) { return true; } return false; }
```

```php
<?php
function bad($password, $stored) { if (password_verify($password, $stored)) { return true; } return false; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
