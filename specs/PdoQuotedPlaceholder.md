---
id: PdoQuotedPlaceholder
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoQuotedPlaceholder

## Summary

Detect execution supplying a named binding occurring only as a quoted SQL string literal rather than an actual marker.

## Detection

- D1. Report execution supplying a named binding occurring only as a quoted SQL string literal rather than an actual marker.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Quoted strings not supplied as bindings, legitimate literal colon text, unknown bindings/SQL and comments are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Remove SQL quotes around this parameter marker.`

## Fix

Remove surrounding SQL single quotes only when the complete literal is exactly one marker and no escaping or interpolation changes its meaning; preserve the PHP string's quoting.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo) { $s = $pdo->prepare("SELECT * FROM items WHERE label = ':label'"); <warning descr="Remove SQL quotes around this parameter marker.">$s->execute(['label' => 'new'])</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
