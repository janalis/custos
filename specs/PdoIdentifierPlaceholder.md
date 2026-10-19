---
id: PdoIdentifierPlaceholder
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoIdentifierPlaceholder

## Summary

Detect PDO prepare with a named or positional parameter marker in a table-name position following FROM, JOIN, UPDATE, INTO or TABLE in parsed literal SQL.

## Detection

- D1. Report PDO prepare when the parsed literal SQL has a parameter marker, named (`:name`) or positional (`?`), as the first token after one of the keywords FROM, JOIN, UPDATE, INTO or TABLE. Only whitespace and comments may separate the keyword from the marker; keywords are matched case-insensitively. The rule applies at any parenthesis depth, so a subquery's table reference counts.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. String literals, comments, valid value markers, unresolved SQL grammar and dynamic identifiers already selected from a validated allowlist are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A FROM that separates function arguments or belongs to a comparison does not open a table reference, so a marker after it is an ordinary value: `IS [NOT] DISTINCT FROM`, and a FROM written directly inside the parentheses of `EXTRACT`, `SUBSTRING`, `SUBSTR`, `MID`, `TRIM` or `OVERLAY`. A nested parenthesis inside such a call (a subquery) is judged on its own.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Select SQL identifiers from a trusted allowlist.`

## Fix

No automatic fix: permitted identifiers and quoting policy must be supplied by the caller.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo) { <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('SELECT * FROM :table')</error>; }
function alsoBad(PDO $pdo) { <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('DELETE FROM ? WHERE id = 1')</error>; }
function fine(PDO $pdo) { $pdo->prepare('SELECT EXTRACT(YEAR FROM :day) WHERE :a IS DISTINCT FROM :b'); }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
