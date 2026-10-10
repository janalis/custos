---
id: FastDigestUsedForPasswordStorage
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FastDigestUsedForPasswordStorage

## Summary

Detect a password input positively established by an HTML password-field mapping or explicit annotated password parameter reaching md5/sha1/hash and then a persisted credential field.

## Detection

- D1. Report a password input positively established by an HTML password-field mapping or explicit annotated password parameter reaching md5/sha1/hash and then a persisted credential field.
- D4. Recognize password inputs only through an exact `@param password-string $parameter` annotation on their declaring callable or a positively resolved HTML password-field mapping. Recognize persistence only through a resolved project callable explicitly documented with `@custos-credential-store`. Do not infer either contract from identifiers.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. General integrity checks, hash-based message authentication, password_hash, variable-name-only guesses and unknown persistence are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Store passwords with a password hashing API.`

## Fix

No automatic fix: migrating existing credential storage and verification requires compatibility policy.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
/** @custos-credential-store */ function saveCredential($hash) {}
/** @param password-string $password */ function bad($password) { <warning descr="Store passwords with a password hashing API.">saveCredential(md5($password))</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
