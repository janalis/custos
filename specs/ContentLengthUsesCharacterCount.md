---
id: ContentLengthUsesCharacterCount
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ContentLengthUsesCharacterCount

## Summary

Detect Content-Length built from mb_strlen/grapheme_strlen of the exact response body later emitted without transformation.

## Detection

- D1. Report Content-Length built from mb_strlen/grapheme_strlen of the exact response body later emitted without transformation.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. ASCII-proven bodies, post-length compression/transformation making either length invalid, byte strlen and unknown body identity are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Measure response content length in bytes.`

## Fix

Replace the character-length call with strlen of the same body, removing encoding arguments, only when arguments bind correctly, the encoding is omitted or a known supported encoding, and no later transformation occurs. Require one exact emitted body; multiple output chunks and intervening or subsequent calls leave body identity incomplete.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($body) { <warning descr="Measure response content length in bytes.">header('Content-Length: ' . mb_strlen($body, 'UTF-8'))</warning>; echo $body; }
```

```php
<?php
function bad($body) { header('Content-Length: ' . strlen($body)); echo $body; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
