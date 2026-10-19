---
id: NonExhaustiveEnumMatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# NonExhaustiveEnumMatch

## Summary

Detect match over a resolved enum when reachable enum cases are missing and there is no default arm.

## Detection

- D1. Report match over a resolved enum when reachable enum cases are missing and there is no default arm.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. An arm condition written as a constant of the enum that aliases one of its cases (`const Fallback = self::Left;`, directly or through another such constant) handles that case.
- D5. The required cases are the reachable ones. Strict comparisons (`===`, `!==`) of the subject variable with a case of its enum (or an aliasing constant) narrow them when they hold on every path to the match. Conditions combine through `!`, a true `&&` and a false `||`; an equality that holds leaves one reachable case, one that fails removes that case. A comparison holds on every path when:
  - the match sits in the taken branch of an `if` / `elseif` / `else` chain or of a ternary, or in the right operand of `&&` / `||`;
  - an earlier statement of an enclosing statement list is an `if` chain without `else` whose every branch exits abruptly (return, throw, exit, break, continue): its conditions were false.
- D6. Comparison narrowing applies only to a subject that provably keeps one value for the whole enclosing function, method, closure or arrow function: a plain variable that is never assigned, unset, incremented, declared `global`/`static`, used as a loop or catch target, bound to a reference (by-reference parameter or closure capture, `&$var` in an assignment or array literal), yielded, or passed as an argument to anything but a resolved function whose matching parameter is by value. The scope must also be free of `goto`, `include`/`require`, `eval`, variable variables and `extract()`. Otherwise the comparisons are ignored and every case stays required.

## Exceptions (no report)

- E1. Exhaustive matches, flow-narrowed case sets, default arms and unknown subject types are excluded. A resolved by-reference call without an output contract invalidates the previous enum type; its parameter declaration constrains entry to the call, not the value returned through the reference.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A match with an arm condition that is not a constant of the subject's enum, or is a constant of it that cannot be traced to a case (a non-case value, an unknown name, a dynamic name), has an unknown arm set and is not reported.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Handle every reachable enum case.`

## Fix

No automatic fix: missing result expressions cannot be invented.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.1. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
enum Choice {case Left;case Right;} function label(Choice $c) {return <error descr="Handle every reachable enum case.">match($c) {Choice::Left=>'L'}</error>;}
```

No report: the guard returns before the match for the only case without an arm, and `Any` aliases a case.

```php
<?php
enum Side {case Up;case Down;const Any = self::Up;}
function sign(Side $s): int {if ($s === Side::Down) {return -1;} return match($s) {Side::Any => 1};}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: a single enum case established by a dominating local assignment narrows the required case set. Other enum cases are unreachable for that subject and are not reported.

Comparison narrowing is deliberately limited to the shapes of D5 and D6. Loose comparisons, `switch` labels, enclosing `match` arms, loop conditions, `assert()` and `in_array()` do not narrow, and a subject passed to a method or an unresolved callee is treated as possibly rebound; such matches stay reported unless they handle every case.
