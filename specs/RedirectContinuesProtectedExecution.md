---
id: RedirectContinuesProtectedExecution
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RedirectContinuesProtectedExecution

## Summary

Detect a conditional unauthorized branch emitting a Location header and falling through to an operation marked protected by the same authorization predicate.

## Detection

- D1. Report a conditional unauthorized branch emitting a Location header and falling through to an operation marked protected by the same authorization predicate.
- D4. A protected operation must be a resolved project callable explicitly documented with `@custos-protected`; callable names alone never establish protection. The redirect branch tests the same authorization value guarding access to that operation.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. A terminating exit/return/throw, mutually exclusive protected branch, deliberate unconditional redirection with no protected continuation and unknown authorization relationship are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Terminate the unauthorized branch after redirecting.`

## Fix

No automatic fix: exit versus return depends on the containing scope and response contract.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
/** @custos-protected */ function deleteRecord() {}
function bad(bool $allowed) { if (!$allowed) { header('Location: /signin'); } <warning descr="Terminate the unauthorized branch after redirecting.">deleteRecord()</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
