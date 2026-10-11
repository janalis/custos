---
id: OwnedStreamNotClosed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# OwnedStreamNotClosed

## Summary

Detect an owned successfully opened stream leaving its local scope without close, return or ownership transfer.

## Detection

- D1. Report an owned successfully opened stream leaving its local scope without close, return or ownership transfer.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. The report is made on a `return` statement whose expression only lends the handle: every mention of the local in the returned expression is an argument of a resolved builtin stream call that reads, writes, positions or inspects the stream without closing or keeping it (`fgets`, `fread`, `fwrite`, `ftell`, `feof`, `stream_get_contents`, ...).

## Exceptions (no report)

- E1. Returned handles, transferred ownership, finally cleanup, automatic-lifetime intentional policy and unknown external retention are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. The return expression closes the handle itself (`return fclose($h);`, `return fgets($h) . fclose($h);`).
- E4. The handle may outlive the scope. Anywhere in the function, a use of the local other than a truthiness or equality test, a plain rebinding of the local, or an argument of the borrowing builtin calls of D4 counts as a hand-over: an element of an array literal, a constructor, method, dynamic or user-function argument, an argument of any other builtin (such as `register_shutdown_function`), a copy into another variable, property, static or array slot, a reference binding, a ternary or coalesce operand, an unpacked or by-reference argument, and a capture by a closure `use` clause or an arrow function body.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Close the owned stream before leaving its scope.`

## Fix

No automatic fix: adding finally may alter exception/return interactions and resource ownership.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } <warning descr="Close the owned stream before leaving its scope.">return fgets($h);</warning> }
function lent($path) { $h = fopen($path, 'rb'); if ($h === false) { return 0; } <warning descr="Close the owned stream before leaving its scope.">return strlen(fgets($h)) + ftell($h);</warning> }
function closed($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return fclose($h); }
function wrapped($path) { $h = fopen($path, 'rb'); if ($h === false) { return null; } return new Reader($h); }
function paired($path) { $h = fopen($path, 'rb'); if ($h === false) { return null; } return [$h, filesize($path)]; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
