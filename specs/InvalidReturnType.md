---
id: InvalidReturnType
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# InvalidReturnType

## Summary

Report a returned value that the function's declared return type can never
accept, judging the value only by what the code itself shows. PHP throws a
`TypeError` when such a `return` runs. Mismatches that rest on PHPDoc are
left to DocumentedReturnType.

## Detection

- D1. A `return` statement with a value, checked against the return type
  declared by the function, method, closure or arrow function that owns
  it. The body of an arrow function is its returned value.
- D2. Only the function's own declaration counts, never one inherited
  from a parent class or an interface.
- D3. The type of the value comes from native declarations only (typed
  parameters, properties and returns, literals, operations) and is unknown
  for a local that may change without an assignment showing it: bound by
  reference, captured by reference, a `global` or `static` local, or passed
  where it may be written by reference.
- D4. The declared type accepts no value of the value's type. `self`
  stands for the class that declares the method; `static` accepts nothing
  that `self` rejects.
- D5. Without `declare(strict_types=1)` in the file, PHP converts scalars
  to the scalar type expected: only values that no conversion accepts
  count (arrays, objects, null for a type without null).

## Exceptions (no report)

- E1. Generators: a function whose body yields returns its completion
  value, not the value the call gives.
- E2. `void` and `never` return types, and a `return` without a value.
- E3. A `return` inside a closure or an arrow function belongs to that
  closure, not to the function around it.
- E4. Values some of which are accepted: a union with an accepted member,
  `mixed`, `object`, an unknown value, a class that a subclass may turn
  into the expected type (a non-final class and an interface, a parent
  class of the expected one), a string or array where a callable is
  expected, an object where a string is expected outside strict mode.
- E5. `self` and `static` in a trait, whose methods run as members of the
  class using it, and `$this` read inside a closure or arrow function,
  which may run bound to another object, unless the value's type does not
  depend on that object.
- E6. Template types and classes whose hierarchy is not fully known.
- E7. Code outside any function and property hooks.

## Report

- Range: the returned value.
- Severity: error.
- Message: `Return int from total(), not string.`: the declared type, the
  function (`Class::method()`, `the closure`, `the arrow function`) and
  the value's type, without value details.

## Fix

None: the intended value cannot be guessed.

## Options

None.

## PHP versions

All versions.

## Examples

```php
<?php
declare(strict_types=1);

function total(): int
{
    return <error descr="Return int from total(), not string.">'1200'</error>;
}
$label = fn(): string => <error descr="Return string from the arrow function, not int.">42</error>;
```

Without strict types, PHP converts the string; a generator's `return`
gives its completion value:

```php
<?php
function total(): int
{
    return '1200';
}

function ids(): Generator
{
    yield 1;
    return 'done';
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
