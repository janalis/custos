---
id: ReflectionCompositeTypeAssumedNamed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# ReflectionCompositeTypeAssumedNamed

## Summary

Inspect the composite reflection type members.

## Detection

- D1. Resolve getType on a ReflectionParameter, ReflectionProperty or getReturnType on a ReflectionFunctionAbstract for a known union type (PHP 8.0+) or intersection type (PHP 8.1+). Report getName invoked on that result, because only ReflectionNamedType provides it. Follow local assignments and indexed parameter selection.
- D1a. A declared type is composite when it has an intersection (`&`, including a parenthesised group inside a union) or when at least two union members remain after discarding a leading `?` and any `null` member. `?T`, `T|null` and `null|T` are a single nullable named type and are never reported; `int|string|null` is.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Calls on members returned by getTypes and unknown reflected declarations are excluded. Nullable single named types remain ReflectionNamedType (D1a).
- E1a. No report when an instanceof test on the same receiver value dominates the call. A condition proves a named type when it is `$t instanceof ReflectionNamedType`, or when it is false and is an instanceof test for the composite class the declared type reflects as (ReflectionUnionType for a union, ReflectionIntersectionType for a pure intersection). Negation with `!`, parentheses, `&&`/`and` and `||`/`or` combine these the usual way: a true conjunction or a false disjunction needs one proving operand, the opposite needs both. Recognised positions:
  - the body of an `if`/`elseif` whose condition proves it when true, and any later `elseif` condition, `elseif` body or `else` body of a chain in which an earlier condition proves it when false;
  - the matching arm of a ternary (including the right side of the short form);
  - the right operand of `&&`/`and` (left proves when true) or `||`/`or` (left proves when false);
  - a `match (true)` arm whose every condition proves it when true;
  - any statement after an earlier sibling, in the same or an enclosing statement list of the scope (top-level code included), that is either an `if` whose terminating main body is the branch where the type is not proven (`if (!$t instanceof ReflectionNamedType) { return; }`, with or without later `elseif`/`else` arms), an `if` without `elseif` whose `else` leaves on that branch, or a builtin `assert()` of a proving condition.

  The test must concern the same unmodified receiver value: reassigning the variable between the test and the call, a test on another value, a dynamic class operand, a guard in an outer function scope or a branch that falls through keep the report. The wrong side of a guard (`if ($t instanceof ReflectionUnionType) { $t->getName(); }`, the `else` of a ReflectionNamedType test) is still reported.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ReflectionCompositeTypeAssumedNamed` and `@noinspection ReflectionCompositeTypeAssumedNamed` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Inspect the composite reflection type members.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 8.0. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
function consume(int|string $x) {} $r = new ReflectionFunction("consume"); $t = $r->getParameters()[0]->getType(); <error descr="Inspect the composite reflection type members.">$t->getName()</error>;
```

Valid case:

```php
<?php
function consume(int|string $x) {} $r = new ReflectionFunction("consume"); $t = $r->getParameters()[0]->getType(); foreach ($t->getTypes() as $part) { $part->getName(); }
```

### Assertion assumptions

Only plain variable tests, literals, boolean operators, strict comparisons,
instanceof, isset/empty and native type predicates supply assertion guards.
A direct native flock result can prove that operation succeeded.
Conditions with property/array accesses, evaluated code, conversions or other
effects supply none; an earlier test may precede a rewrite.

A builtin assertion supplies a guard only with the resolved `flow.assertions`
setting enabled (default false). Its argument must bind the native assertion
parameter without unpacking or duplicate names. A later description that may
run callbacks or rewrite the tested value prevents that guard. Enabling the
setting assumes assertions execute and failed assertions terminate.

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
