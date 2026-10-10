---
id: StrtokStateClobbered
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# StrtokStateClobbered

## Summary

Detect starting a second two-argument strtok sequence between an initial sequence call and its one-argument continuation on the same reachable path.

## Detection

- D1. Report starting a second two-argument strtok sequence between an initial sequence call and its one-argument continuation on the same reachable path.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Completion/restart of the same intended sequence, no later continuation, or a wrapper known not to invoke strtok are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Keep strtok sequences separate.`

## Fix

No automatic fix: independent tokenizer state requires restructuring.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$first=strtok($a,','); strtok($b,':'); $next=<warning descr="Keep strtok sequences separate.">strtok(',')</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
