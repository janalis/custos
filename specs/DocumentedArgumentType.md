---
id: DocumentedArgumentType
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# DocumentedArgumentType

## Summary

Report an argument that breaks a parameter's contract when PHPDoc is taken
into account: the `@param` type of a project function rejects it, or the
declared type rejects a value whose type only PHPDoc describes. PHP may run
such a call, but the code and its documentation disagree. Mismatches the
code alone proves are InvalidArgumentType's.

## Detection

- D1. Calls, argument binding and callee resolution as InvalidArgumentType
  defines them.
- D2. The type of the argument is the one PHPDoc-aware inference gives:
  `@param`, `@var` and `@return` tags count.
- D3. Every declaration the call may bind to rejects every value of the
  argument's type, through the declared type of the parameter or, for a
  function of the project, its `@param` type. A null default makes either
  type nullable.
- D4. The weak-mode conversions and the exceptions of InvalidArgumentType
  apply alike.
- D5. Integer ranges and non-empty strings of a `@param` type count only
  against an argument whose possible values are all known (an integer
  literal such as `0`) or, for a string, whose byte lengths are (a string
  literal such as `''`); otherwise `@param positive-int` accepts any
  integer.

## Exceptions (no report)

- E1. Arguments InvalidArgumentType reports: the declared types reject
  what the code alone shows.
- E2. The `@param` type of a variadic parameter, of a parameter of PHP's
  own functions, a `@param` type naming templates, and a `@param` type the
  declared type rejects (a stale tag).
- E3. The other value details of a `@param` type: string lengths other
  than non-empty and literal string values, and every value detail against
  an argument whose values are not all known (see D5).
- E4. Every exception of InvalidArgumentType.
- E5. Values read from a variadic parameter whose `@param` type is an array
  or iterable (`@param string[] ...$names`): the tag may type the collected
  array or each value, so the values have no PHPDoc type.

## Report

- Range: the argument's value.
- Severity: warning.
- Message: `Pass Order to $order of archive(), not string.`, with the
  documented or declared type that rejects the value.

## Fix

None: either the call or the documentation is wrong.

## Options

None.

## PHP versions

All versions, as InvalidArgumentType.

## Examples

```php
<?php
declare(strict_types=1);

class Order {}
/** @param Order $order */
function archive($order) {}
archive(<warning descr="Pass Order to $order of archive(), not string.">'order-1'</warning>);

function charge(int $cents) {}
/** @param string $amount */
function bill($amount) { charge(<warning descr="Pass int to $cents of charge(), not string.">$amount</warning>); }

/** @param positive-int $quantity */
function restock(int $quantity) {}
restock(<warning descr="Pass positive-int to $quantity of restock(), not int.">0</warning>);
```

A template, or a variadic parameter whose tag types each value, is not
compared:

```php
<?php
/**
 * @template T
 * @param T $value
 */
function keep($value) {}
keep('x');
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
