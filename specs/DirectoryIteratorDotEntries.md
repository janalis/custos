---
id: DirectoryIteratorDotEntries
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DirectoryIteratorDotEntries

## Summary

Detect a filesystem mutation on DirectoryIterator entry paths inside foreach without a dominating isDot exclusion.

## Detection

- D1. Report a filesystem mutation on DirectoryIterator entry paths inside foreach without a dominating isDot exclusion.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. FilesystemIterator with SKIP_DOTS, isDot continue/guard, explicit dot-name rejection and nonmutating entry handling are excluded.
- E3. Reassignment, unsetting or reuse of the element variable before its destructive use leaves its directory-entry identity incomplete. Writes inside an uninvoked separate closure do not establish outer-variable reassignment.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Skip dot entries before modifying directory entries.`

## Fix

No automatic fix: loop control structure and entry-processing policy require deliberate handling.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($dir) { foreach (new DirectoryIterator($dir) as $entry) { <warning descr="Skip dot entries before modifying directory entries.">unlink($entry->getPathname())</warning>; } }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
