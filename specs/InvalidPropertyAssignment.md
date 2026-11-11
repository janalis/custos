---
id: InvalidPropertyAssignment
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types]
php: { min: "", max: "" }
---

# InvalidPropertyAssignment

## Summary

Report a value stored in a typed property whose declared type can never
accept it, judging the value only by what the code itself shows. PHP throws
a `TypeError` when such an assignment runs. Mismatches that rest on PHPDoc
are left to DocumentedPropertyAssignment.

## Detection

- D1. An assignment whose target is a property named in the source: an
  instance property through an object (`$this->total`, `$order->total`) or
  a static property through a class name, `self`, `parent` or `static`
  (`self::$cache`). Plain assignments, compound assignments (`.=`, `+=`…)
  and `??=` count.
- D2. The stored value is the right-hand side of a plain assignment or of
  `??=`, and the result of the operation for any other compound
  assignment.
- D3. The written class comes from native declarations only: the object's
  type names only classes, each declared once with every ancestor known.
  The property is the declaration that class resolves, by inheritance or
  through a trait, and, unless it is private or final or the class is
  final, its declaration in every known subclass, since the object may be
  any of them; `static::` likewise reaches the subclasses.
- D4. The type of the value comes from native declarations only (typed
  parameters, properties and returns, literals, operations) and is unknown
  for a local that may change without an assignment showing it, as
  InvalidArgumentType defines it.
- D5. Every declaration of D3 has a native type, and that type accepts no
  value of the stored type.
- D6. Without `declare(strict_types=1)` in the writing file, PHP converts
  scalars to the scalar type of the property: only values that no
  conversion accepts count (arrays, objects, null for a type without
  null).

## Exceptions (no report)

- E1. Targets that are not plainly named: dynamic property names
  (`$this->$name`, `$this->{'a'}`, `$cls::$x`, `static::$$name`), nullsafe
  fetches, array element writes (`$this->items[] = …`).
- E2. By-reference assignments (`$this->p = &$v`) and writes through a
  reference (`$r = &$this->p; $r = 'x';`).
- E3. Properties with hooks, virtual or abstract properties, properties
  declared only by a `@property` tag, undeclared properties, a static
  property written as an instance property or the reverse, and properties
  PHP itself declares.
- E4. A class or subclass of the receiver that declares `__set`, which may
  intercept the write (an unset property, for instance).
- E5. A write the scope may not perform, which fails before any type check:
  a private property outside its class, a protected property outside its
  hierarchy, a non-public property from a closure or arrow function (it may
  run in another scope), and a readonly property other than through
  `$this` in its own class's constructor, where it must be the only write
  to that property and the property not promoted (otherwise the write may
  be a re-initialization).
- E6. Values some of which are accepted: a union with an accepted member,
  `mixed`, `object`, an unknown value, a class that a subclass may turn
  into the expected type, an object with `__toString` for a string
  property outside strict mode.
- E7. Untyped properties; template types, `self`, `static`; classes whose
  hierarchy is not fully known, declared more than once, or traits (a
  trait method runs in the classes using it).
- E8. Receivers whose type only PHPDoc gives, receivers of unknown, mixed
  or mixed-class types, and `$this` read inside a closure or arrow
  function, which may run bound to another object.

## Report

- Range: the assigned value.
- Severity: error.
- Message: `Assign int to Cart::$count, not string.`: the declared type,
  the property and the stored type, without value details.

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

final class Cart
{
    public int $count = 0;

    public function fill()
    {
        $this->count = <error descr="Assign int to Cart::$count, not string.">'3'</error>;
    }
}
```

Without strict types, PHP converts the string; a hooked property runs its
own code:

```php
<?php
final class Cart
{
    public int $count = 0;
    public float $ratio = 0.0 {
        set(float|string $value) { $this->ratio = (float) $value; }
    }

    public function fill()
    {
        $this->count = '3';
        $this->ratio = 'half';
    }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
