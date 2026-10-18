---
id: NonVoidFunctionFallsThrough
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# NonVoidFunctionFallsThrough

## Summary

Identify the return type of an explicitly nonvoid nongenerator function when structural control flow proves a reachable normal exit without return or throw. Exclude mixed, void, never and untyped returns; a bare return is covered separately. Restrict proof to finite straight-line blocks and if/else branches; loops, switch, try, goto, parser recovery and all calls on a candidate exit path exclude because they may terminate, throw or exit.

## Detection

- D1. Report the return type of an explicitly nonvoid nongenerator function when structural control flow proves a reachable normal exit without return or throw. Exclude mixed, void, never and untyped returns; a bare return is covered separately. Restrict proof to finite straight-line blocks and if/else branches; loops, switch, try, goto, parser recovery and all calls on a candidate exit path exclude because they may terminate, throw or exit.
- D1a. The proof is path-aware for repeated tests. While walking a block, remember each condition whose outcome is fixed on the current path: inside a branch its own condition holds or fails, and after an `if` only one side of which continues, the condition has the value of the continuing side. A later `if` whose condition is the same test, or its exact logical complement, is decided by that fact instead of being treated as open. Complement pairs are `==`/`!=`, `===`/`!==`, `<`/`>=` and `>`/`<=`, with operands in either order. Only conditions made of a plain variable, or a comparison between a plain variable and another plain variable or scalar literal, are remembered. A bare variable and a loose comparison with `true` or `false` are the same truthiness test; a strict comparison with `true` or `false` counts as one only for a parameter declared exactly `bool`. Any assignment or increment on the continuing path forgets every remembered condition.
- D1b. An expression statement that cannot complete ends its path like a `throw` statement: a throw expression, a plain `=` assignment whose value is one, and a `match` whose subject throws or whose every arm body throws (a subject matching no arm raises an error as well).
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E3. Sibling branches that together cover every value, such as `if ($n > 0) { return 1; } if ($n <= 0) { return -1; }`, leave no proven exit. Ordering complements are trusted even though NAN or uncomparable arrays fail both tests: reaching the end then depends on unknown runtime values, which is not proof. Tests that merely mention the same variable (`> 0` then `< 0`) remain a finding.
- E4. A statement or condition that may run code (a call anywhere in it, `eval`, `include`, `clone`, a shell command, `exit`) or divide may leave the function, so no fall-through is proven past it.
- E2. Both `@custos-ignore NonVoidFunctionFallsThrough` and `@noinspection NonVoidFunctionFallsThrough` suppress this native rule.

## Report

- Range: the declared return type. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Return a value on every reachable path.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 7.0 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
function countItems(bool $ready): int { if ($ready) return 7; }
```

Valid case:

```php
<?php
function countItems(bool $ready): int { if ($ready) return 7; return 0; }
```

Valid case, exhaustive sibling tests:

```php
<?php
function direction(int $delta): string { if ($delta >= 0) { return 'up'; } if ($delta < 0) { return 'down'; } }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
