---
id: PdoReferenceBindingVariableReuse
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoReferenceBindingVariableReuse

## Summary

Detect two distinct PDO markers bound through bindParam to the same scalar local identity, with its value changed between bindings and before execute.

## Detection

- D1. Report two distinct PDO markers bound through bindParam to the same scalar local identity, with its value changed between bindings and before execute.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Separate variables, bindValue, deliberate identical bindings without intervening change, unknown aliases and references escaping invalidate proof.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Bind distinct parameters to distinct scalar identities.`

## Fix

No automatic fix: copying values at binding time may not preserve the intended first binding.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :a, :b'); $v = 0; $s->bindParam(':a', $v); $v = 1; $s->bindParam(':b', $v); $v = 2; <warning descr="Bind distinct parameters to distinct scalar identities.">$s->execute()</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
