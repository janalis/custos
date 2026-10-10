---
id: RepeatedHtmlEntityEncoding
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# RepeatedHtmlEntityEncoding

## Summary

Detect htmlspecialchars or htmlentities directly nesting an equivalent encoding call with matching flags/encoding and default or true double_encode.

## Detection

- D1. Report htmlspecialchars or htmlentities directly nesting an equivalent encoding call with matching flags/encoding and default or true double_encode.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Different contexts/options, double_encode false on the outer call and intentionally literal entity rendering are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Encode HTML entities at one intended output boundary.`

## Fix

No automatic fix: repeated encoding can intentionally display entity syntax.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
echo <warning descr="Encode HTML entities at one intended output boundary.">htmlspecialchars(htmlspecialchars($text))</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
