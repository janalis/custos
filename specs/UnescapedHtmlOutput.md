---
id: UnescapedHtmlOutput
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnescapedHtmlOutput

## Summary

Detect request-derived text emitted inside a proven HTML text-node context without context-appropriate escaping.

## Detection

- D1. Report request-derived text emitted inside a proven HTML text-node context without context-appropriate escaping.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. JSON/plain-text responses, trusted literals, htmlspecialchars/htmlentities with correct charset/flags and unknown output context are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

- E3. Known boolean or numeric scalar output and resolved builtin `md5`, `sha1`, `hash` or `hash_hmac` hexadecimal output cannot inject the target text syntax. Raw binary or unknown output modes retain source taint. These narrow text proofs do not establish URL or filesystem-path authorization.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Escape untrusted text for its HTML output context.`

## Fix

Wrap only a proven HTML text-node interpolation in htmlspecialchars with explicit UTF-8 and ENT_QUOTES/ENT_SUBSTITUTE on PHP 5.4+; no fix for attributes, scripts, URLs or unknown charset.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $_GET['label'] . '</p>';</error>
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
