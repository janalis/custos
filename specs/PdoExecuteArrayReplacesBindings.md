---
id: PdoExecuteArrayReplacesBindings
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoExecuteArrayReplacesBindings

## Summary

Detect execute with a literal empty parameter array on a statement with previously established required bindValue/bindParam bindings.

## Detection

- D1. Report execute with a literal empty parameter array on a statement with previously established required bindValue/bindParam bindings. A binding is required when the prepared SQL is a known string containing at least one placeholder and the binding names one of them: a named placeholder (with or without the leading colon) or a 1-based position within the positional placeholders. Every earlier binding on the statement must be required in that sense.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. No prior bindings, SQL with no required markers, unknown binding state and deliberate replacement with complete values are excluded.
- E3. Any earlier binding whose placeholder is not proven (a non-constant name or position) or names a placeholder absent from the SQL (for example `bindValue(':id', 7)` on `SELECT 1`, or position 2 with a single `?`) suppresses the report: executing without the array would then fail on the stray binding, so the empty array is what keeps the call working.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Omit the empty parameter array to retain existing bindings.`

## Fix

Remove the empty execute argument to retain prior bindings; preserve any comments by withholding the fix when removal would discard them. Because of E3 the fix is only ever offered when all earlier bindings target placeholders of the statement, so it cannot turn a working execute into a failing one.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->bindValue(':id', 7); <warning descr="Omit the empty parameter array to retain existing bindings.">$s->execute([])</warning>; }
```

```php
<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->bindValue(':id', 7); $s->execute(); }
```

```php
<?php
// No report: the SQL has no placeholder, the empty array discards a stray binding.
function ok(PDO $pdo) { $s = $pdo->prepare('SELECT 1 AS v'); $s->bindValue(':id', 7); $s->execute([]); }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
