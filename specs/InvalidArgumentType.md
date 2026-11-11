---
id: InvalidArgumentType
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# InvalidArgumentType

## Summary

Report an argument that the parameter's declared type can never accept,
judging the argument only by what the code itself shows. PHP throws a
`TypeError` when such a call runs. Mismatches that rest on PHPDoc are left
to DocumentedArgumentType.

## Detection

- D1. A function call, method call, static call or object construction
  whose callee resolves to known declarations: the single declaration of the
  function, the constructor of a concrete class, or the method as every
  class the receiver may be resolves it, its overrides in the indexed
  subtypes included unless the method is private or final or its class is
  final.
- D2. Arguments bind as PHP binds them: positional arguments in order,
  named arguments to the parameter of that name, extra positional
  arguments to a variadic parameter, each checked against that parameter's
  type.
- D3. The type of the argument comes from native declarations only (typed
  parameters, properties and returns, literals, operations) and is unknown
  for a local that may change without an assignment showing it: bound by
  reference, captured by reference, a `global` or `static` local, or passed
  where it may be written by reference.
- D4. Every declaration of D1 has a native type on the parameter, and that
  type accepts no value of the argument's type. A null default (`null` or
  `\null`, in any case) makes the parameter type nullable.
- D5. Without `declare(strict_types=1)` in the calling file, PHP converts
  scalars to the scalar type expected: only values that no conversion
  accepts count (arrays, objects, null for a type without null).
- D6. A builtin's parameter types are those it declares at the configured
  PHP version.

## Exceptions (no report)

- E1. Arguments after an unpacked one (`...$args`), first-class callables
  (`f(...)`), named arguments to a parameter that does not exist or to a
  variadic parameter, and named arguments to PHP's own functions.
- E2. By-reference parameters.
- E3. Values some of which are accepted: a union with an accepted member,
  `mixed`, `object`, an unknown value, a class that a subclass may turn
  into the expected type (a non-final class and an interface, a parent
  class of the expected one), a string or array where a callable is
  expected, an object with `__toString` for a string outside strict mode.
- E4. Template types, `self`, `static` and classes whose hierarchy is not
  fully known.
- E5. null passed to a scalar parameter of a PHP function outside strict
  mode, which is only deprecated.
- E6. Uncertain callees: dynamic names, receivers of unknown, mixed or
  mixed-class types, methods served by `__call` or `__callStatic` or
  declared by a `@method` tag, trait methods called through `$this`, `self`
  or `static`, functions declared more than once or named like a function
  PHP declares in some version (a polyfill), classes declared more than
  once.
- E7. An override in a subtype that accepts the value.
- E8. `$this` read inside a closure or arrow function, as an argument or
  as the receiver of the method called, since the closure may run bound
  to another object.
- E9. PHP's own functions below PHP 8.0, functions of extensions not
  bundled with PHP, and the builtins whose stub declares other parameter
  types than PHP enforces (overloaded forms such as `strtr`, option arrays
  such as `setcookie`).

## Report

- Range: the argument's value.
- Severity: error.
- Message: `Pass int to $cents of charge(), not string.`: the declared type,
  the parameter, the callee and the argument's type, without value details.

## Fix

None: the intended value cannot be guessed.

## Options

None.

## PHP versions

All versions. Builtin callees are checked from PHP 8.0, against the
signature they declare at the configured version.

## Examples

```php
<?php
declare(strict_types=1);

function charge(int $cents) {}
charge(<error descr="Pass int to $cents of charge(), not string.">'1200'</error>);
strlen(<error descr="Pass string to $string of strlen(), not array.">[]</error>);
```

Without strict types, PHP converts the string; a union may hold an
accepted value:

```php
<?php
function charge(int $cents) {}
charge('1200');
function pay(int|string $amount) { charge($amount); }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
