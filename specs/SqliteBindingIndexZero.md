---
id: SqliteBindingIndexZero
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SqliteBindingIndexZero

## Summary

Start SQLite binding positions at one.

## Detection

- D1. Resolved SQLite3Stmt bindValue or bindParam receives constant integer zero as the parameter argument. String names and positive positions are excluded.
  A variable argument counts only while it provably still holds the literal zero it was assigned (see E3).
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore SqliteBindingIndexZero and @noinspection SqliteBindingIndexZero suppress this native rule.
- E3. A variable position is not a proven zero once the variable may have been stepped or rewritten behind the
  assignment: an increment or decrement (`$i++`, `++$i`, `$i--`, `--$i`) or a closure capturing it by reference
  (`use (&$i)`) that appears before the call, or anywhere inside a loop that also contains the call. The same
  constructs placed after a call that no loop repeats do not affect it, and a by-value capture never does.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Start SQLite binding positions at one.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f(SQLite3Stmt $s) { <error descr="Start SQLite binding positions at one.">$s->bindValue(0, 'ready')</error>; }
```

Valid case:

```php
<?php
function f(SQLite3Stmt $s) { $s->bindValue(1, 'ready'); }
```

Counter advanced before use (E3):

```php
<?php
function g(SQLite3Stmt $s, array $values) {
    $position = 0;
    foreach ($values as $value) {
        $position++;
        $s->bindValue($position, $value);
    }
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
