---
id: RedirectContinuesProtectedExecution
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RedirectContinuesProtectedExecution

## Summary

Detect a conditional unauthorized branch emitting a Location header and falling through to an operation marked protected by the same authorization predicate.

## Detection

- D1. Report a conditional unauthorized branch emitting a Location header and falling through to an operation marked protected by the same authorization predicate.
- D4. A protected operation must be a resolved project callable explicitly documented with `@custos-protected`; callable names alone never establish protection. The redirect branch tests the same authorization value guarding access to that operation.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D5. The redirect branch is an `if (!$v)` statement without `elseif`/`else` whose body does not terminate and emits a literal `Location:` header in its own scope; a header call inside a nested closure or function body does not count. The branch must precede the protected call in a statement list enclosing that call, either inside the same function-like scope or in top-level code. A protected call inside a nested closure or arrow function is not related to a redirect of the outer scope.
- D6. The protected call must stay reachable while `$v` is falsy. `$v` is proven truthy at the call, and nothing is reported, when the call sits in an arm entered only if `$v` is truthy (`if`/`elseif`/`while` body, `else`/later `elseif` arm of a condition that is true whenever `$v` is falsy, ternary arm, right operand of `&&`/`and`/`||`/`or`), or when a terminating `if` without `else` located between the redirect branch and the call exits whenever `$v` is falsy. Conditions are understood through parentheses, `!`, true conjunctions and false disjunctions of plain variables only; any other condition proves nothing and the report is kept.
- D7. Any write to `$v` between the start of the redirect branch and the protected call (assignment, compound assignment, destructuring, `unset`, `++`/`--`, loop or catch binding) makes the authorization relationship unknown: no report.

## Exceptions (no report)

- E1. A terminating exit/return/throw, mutually exclusive protected branch, deliberate unconditional redirection with no protected continuation and unknown authorization relationship are excluded.
- E3. Same-predicate guards per D6: `if (!$v) { header('Location: …'); } if ($v) { op(); }`, a later `if (!$v) { return; }` before the call, `$v ? op() : null`, `$v && op()`, `!$v || op()` and the `else` arm of a second `if (!$v)`.
- E4. The predicate is rebound before the call (D7), the header is only emitted by a nested closure, or the protected call is only made by a nested closure or arrow function.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Terminate the unauthorized branch after redirecting.`

## Fix

No automatic fix: exit versus return depends on the containing scope and response contract.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
/** @custos-protected */ function deleteRecord() {}
function bad(bool $allowed) { if (!$allowed) { header('Location: /signin'); } <warning descr="Terminate the unauthorized branch after redirecting.">deleteRecord()</warning>; }
```

No report: the protected call only runs when the predicate holds.

```php
<?php
/** @custos-protected */ function purgeArchive() {}
function exclusive(bool $member) { if (!$member) { header('Location: /join'); } if ($member) { purgeArchive(); } }
function guarded(bool $member) { if (!$member) { header('Location: /join'); } if (!$member) { return; } purgeArchive(); }
function inline(bool $member) { if (!$member) { header('Location: /join'); } return $member ? purgeArchive() : null; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
