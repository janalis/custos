---
id: Utf8ByteTruncation
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# Utf8ByteTruncation

## Summary

Detect substr on a literal valid UTF-8 string with statically known offsets selecting an incomplete code-point byte sequence.

## Detection

- D1. Report substr on a literal valid UTF-8 string with statically known offsets selecting an incomplete code-point byte sequence.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. ASCII, boundaries preserving complete code points, binary-data contracts and unknown strings are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Keep UTF-8 code points intact when truncating text.`

## Fix

No automatic fix: byte offsets cannot generally be reinterpreted as character offsets.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$initial=<warning descr="Keep UTF-8 code points intact when truncating text.">substr('ñame',0,1)</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
