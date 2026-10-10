---
id: ShallowCloneNestedMutation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ShallowCloneNestedMutation

## Summary

Detect cloning an object followed by mutation of a nested object member where the nested identity is proven shared with the original.

## Detection

- D1. Report cloning an object followed by mutation of a nested object member where the nested identity is proven shared with the original.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Custom __clone that replaces the nested object, scalar members, unknown identities and escaped aliases are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Clone the nested object before mutating independent state.`

## Fix

No automatic fix: recursive cloning and sharing policies are application-specific.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$a=new stdClass(); $a->child=new stdClass(); $b=clone $a; <warning descr="Clone the nested object before mutating independent state.">$b->child->label</warning>='changed';
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: PHP 8.5 clone property overrides, computed child names, unresolved inheritance and magic or hooked child access do not prove shared nested storage and are excluded.
