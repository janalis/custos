---
id: PossiblyNullArgument
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# PossiblyNullArgument

## Summary

Report an argument that the code itself shows may be null, passed to a
parameter whose declared type rejects null. PHP throws a `TypeError` when
the call runs with null. Null that comes only from PHPDoc is not reported.

## Detection

- D1. A function call, method call, static call or object construction
  whose callee resolves to known declarations, as InvalidArgumentType
  resolves them (overrides in indexed subtypes included).
- D2. Arguments bind as InvalidArgumentType binds them: positional
  arguments in order, named arguments to the parameter of that name,
  extra positional arguments to a variadic parameter.
- D3. The argument's type comes from native declarations only, narrowed by
  the conditions guarding the call, as PossiblyNullAccess types a value
  (its D2), and includes null alongside another type.
- D4. Every declaration of D1 has a native type on the parameter that
  rejects null: no `?`, `null` or `mixed` in it and no null default.

## Exceptions (no report)

- E1. The parameter's type rejects every value of the argument, not only
  null: InvalidArgumentType reports that.
- E2. null passed to a scalar parameter of a PHP function outside strict
  mode, which is only deprecated.
- E3. By-reference parameters; arguments after an unpacked one; named
  arguments to PHP's own functions; uncertain callees, as for
  InvalidArgumentType.
- E4. The exceptions E4 to E8 and E11 of PossiblyNullAccess: unknown or
  `mixed` values, null ruled out by PHPDoc or `assert()`, properties
  checked before a call, checked identical calls and their holders,
  `$this` in closures (also as the receiver of the call), and the inner
  iterator of PHP's iterator wrappers.
- E5. PHP's own functions below PHP 8.0, functions of extensions not
  bundled with PHP, and builtins whose stub types differ from PHP's.

## Report

- Range: the argument's value.
- Severity: warning.
- Message: `Check this value for null before passing it to $order of
  ship(): it is Order|null.` with the parameter, the callee and the
  argument's type.

## Fix

None: whether to check, default or throw is the author's choice.

## Options

None.

## PHP versions

All versions. Builtin callees are checked from PHP 8.0.

## Examples

```php
<?php
class Order {}
function ship(Order $order) {}
function find(): ?Order { return null; }
ship(<warning descr="Check this value for null before passing it to $order of ship(): it is Order|null.">find()</warning>);
```

A check, or a parameter that accepts null, is silent; so is null passed
to a scalar parameter of a PHP function outside strict mode:

```php
<?php
class Order {}
function ship(Order $order) {}
function keep(?Order $order) {}
function find(): ?Order { return null; }
function label(): ?string { return null; }
$order = find();
keep($order);
if ($order !== null) {
    ship($order);
}
strlen(label());
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
NullPointerException, when it runs together with this rule, leaves to it
the arguments this rule reports.
