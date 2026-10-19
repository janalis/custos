---
id: DuplicateMatchCondition
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# DuplicateMatchCondition

## Summary

Detect a later match arm condition statically strictly identical to an earlier condition, including resolved references to the same enum case.

## Detection

- D1. Report a later match arm condition statically strictly identical to an earlier condition, including resolved references to the same enum case.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases or recursion cannot establish definite facts. When the proof is inconclusive, nothing is reported.
- D4. A variable condition takes part only when it is a local of a braced function, method or closure body bound by exactly one plain `=` assignment of the variable itself, the assigned expression reads no variable, and the resolved value of that assignment is known. Two conditions repeated inside the same arm are reported, but they do not make the arm unreachable.

## Exceptions (no report)

- E1. Unknown values, different strict types and effectful condition expressions are excluded.
- E3. A variable condition is ignored when the variable has any other binding in its scope: `++`/`--`, a compound, by-reference or destructuring assignment, a second assignment, a `catch`, `foreach`, `global`, `static` or `unset` binding, or a by-reference closure capture. It is ignored as well when its single assignment copies or computes from another variable (`$a = $b`), because that variable may itself have changed. It is also ignored in top-level code and arrow functions, and when the scope contains a variable-variable, `extract()`, `eval` or an include, since those can rebind a local by name.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Remove the repeated match condition.`

## Fix

Remove the arm, with one adjacent comma, only when the whole arm is unreachable: every one of its conditions has a known value already listed by an earlier arm. A condition repeated only within its own arm, or sharing its arm with a new or unknown condition, gets no fix. An arm with several conditions carries the removal on its first condition only. No fix is offered when the removed text contains a comment.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.0. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$v=match($input) {4=>'first',<warning descr="Remove the repeated match condition.">4</warning>=>'second',default=>'other'};
```

```php
<?php
$v=match($input) {4=>'first',default=>'other'};
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: variable conditions are accepted only under the single-assignment proof of D4/E3, because a value that changed after its literal assignment makes the arm reachable and its removal a behaviour change.

Audit constraint: arm removal is withheld when its byte range contains a PHP hash comment, as well as slash comments. The diagnostic remains available without a fix.
