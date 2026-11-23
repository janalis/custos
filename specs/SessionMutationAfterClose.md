---
id: SessionMutationAfterClose
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SessionMutationAfterClose

## Summary

Detect a write to $_SESSION after successful session_write_close and before a proven session_start reopening.

## Detection

- D1. Report a write to $_SESSION after successful session_write_close and before a proven session_start reopening.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or code too complex to analyze cannot establish definite facts.

## Exceptions (no report)

- E1. Reads, reopened sessions, failed close, externally persisted session wrappers and unknown session state are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. Assigning an empty array to the whole `$_SESSION` (`$_SESSION = [];`, `$_SESSION = array();`) clears the in-memory copy of a closed session; no change is meant to persist. Writes of a key, compound assignments and non-empty arrays stay reported.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Persist session changes before closing the session.`

## Fix

No automatic fix: moving writes or reopening a session changes locking and failure behavior.

## Options

None.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within the supported PHP versions). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
session_start(); session_write_close(); <warning descr="Persist session changes before closing the session.">$_SESSION['visits']++</warning>;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
