---
id: SqliteResultUsedAfterFinalize
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SqliteResultUsedAfterFinalize

## Summary

Read the SQLite result before finalizing it.

## Detection

- D1. After SQLite3Result finalize on a proven local receiver, fetchArray, reset, columnName, columnType or numColumns is called on that same result without reassignment. Report the invalid use. finalize cannot fail, so its return value does not have to be tested. Two shapes prove that finalize ran first: a finalize call in an expression statement of the same or an enclosing statement list that always executes before the use, including when its result is stored (`$ok = $r->finalize();`); or a use inside the body of an `if` whose condition is the finalize call itself or its comparison with `true` (`==` or `===`, either operand order).
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. A finalize call that may be skipped proves nothing: one inside a conditional statement that does not contain the use, on the right of a short-circuit or null-coalescing operator, in a ternary branch, in a `match` arm, in the arguments, array key or dynamic member name of a null-safe chain, in a `try` block that the use follows, or in a nested closure.
- E3. No report when the receiver is a by-reference parameter: any call can replace the result behind it.
- E4. Both @custos-ignore SqliteResultUsedAfterFinalize and @noinspection SqliteResultUsedAfterFinalize suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Read the SQLite result before finalizing it.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f(SQLite3Result $r) { if($r->finalize()){ <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; } }
function g(SQLite3 $db) { $r = $db->query('select 1'); $r->finalize(); return <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; }
```

Valid case:

```php
<?php
function f(SQLite3Result $r) { $row=$r->fetchArray(); $r->finalize(); }
function g(SQLite3Result $r, SQLite3Stmt $s) { $r->finalize(); $r = $s->execute(); $r->fetchArray(); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
