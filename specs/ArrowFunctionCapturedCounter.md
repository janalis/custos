---
id: ArrowFunctionCapturedCounter
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.4", max: "" }
---

# ArrowFunctionCapturedCounter

## Summary

Detect an arrow function updating a captured scalar counter when the same callable is invoked at least twice without reassignment.

## Detection

- D1. Report an arrow function updating a captured scalar counter when the same callable is invoked at least twice without reassignment.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Parameters, object member updates, single calls, and unknown captured types are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Use persistent shared state for this counter.`

## Fix

No automatic fix: choosing reference capture and state lifetime requires intent.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 7.4. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$n = 0; $next = <warning descr="Use persistent shared state for this counter.">fn() => ++$n</warning>; echo $next(), $next();
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
