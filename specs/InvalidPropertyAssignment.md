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

- D1. A write whose target is a property named in the source: an
  instance property through an object (`$this->total`, `$order->total`) or
  a static property through a class name, `self`, `parent` or `static`
  (`self::$cache`). These writes count:
  - plain assignments, compound assignments (`.=`, `+=`…) and `??=`;
  - by-reference assignments of a local variable (`$this->total = &$sum`):
    PHP checks the local's current value when binding;
  - destructuring (`[$this->from, $this->to] = $range`, `list(…)`, keyed
    and nested patterns), each item a separate write;
  - the key and the value a foreach loop binds (`foreach ($rows as
    $this->key => $this->row)`), the value possibly destructured;
  - increments and decrements (`$this->hits++`, `--$this->hits`).
- D2. The stored value is:
  - the right-hand side of a plain assignment or of `??=`, the result of
    the operation for any other compound assignment;
  - the current value of the local bound by reference;
  - for a destructured item, the value under its key (explicit or
    positional) of the destructured value: the type the value's shape
    gives that key, or the element type of a typed array, joined with
    null when the key may be missing (PHP then stores null); destructuring
    null or a scalar stores null in every item;
  - for a foreach value, the element type of the iterated value; for a
    foreach key, integer or string for an array (the kinds of its literal
    keys when known), the declared key type of a Traversable;
  - for `++` and `--`, the operand's value after the operation: `null`
    incremented becomes an integer, decremented stays null; an integer may
    become a float; a string may become a number.
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
- E2. Writes whose stored value cannot be told:
  - writes through a reference (`$r = &$this->p; $r = 'x';`, or a local
    bound to the property and written later), by-reference assignments of
    anything but a local variable (an absent element is created as null),
    by-reference foreach values and destructured items;
  - destructuring with a computed key, a key the value's shape does not
    list, a value of plain `array` type, or an object (ArrayAccess runs
    code);
  - foreach loops over an iterator whose key or value type is not
    declared (it may yield keys of any type);
  - `++` and `--` on arrays, objects and booleans (PHP fails before
    storing anything, or leaves the value unchanged), and the overflow of
    an integer property, which PHP reports with its own error.
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
  to that property (any form of D1) and the property not promoted
  (otherwise the write may be a re-initialization). A compound
  assignment, `++`, `--` or a by-reference assignment never initializes a
  readonly property.
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

- Range: the assigned value; the target for a destructured item or a
  foreach key or value; the whole `++`/`--` expression.
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

Destructuring, loops and increments store values too:

```php
<?php
declare(strict_types=1);

final class Range
{
    public int $from = 0;
    public null $unset = null;

    public function load(array $row)
    {
        [<error descr="Assign int to Range::$from, not string.">$this->from</error>] = ['first'];
        foreach (['a' => 1] as <error descr="Assign int to Range::$from, not string.">$this->from</error> => $value) {}
        <error descr="Assign null to Range::$unset, not int.">$this->unset++</error>;
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
