---
id: UndefinedProperty
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types]
php: { min: "", max: "" }
---

# UndefinedProperty

## Summary

Report a read of a property that the object's class does not declare,
judging the object only by what the code itself shows, and any access to an
undeclared static property. PHP warns and reads `null` for the first, and
throws an `Error` for the second; the name usually has a typo. Objects
that accept any property (magic `__get`, `@property` and `@mixin` tags,
`stdClass`, `#[AllowDynamicProperties]`) are left alone, and so are writes,
which create a dynamic property (see UndeclaredDynamicProperty).

## Detection

- D1. A read of an instance property with a name written in the source
  (`$cart->total`, `$cart?->total`), or an access to a static property
  through a class name, `self`, `parent` or `static` (`Cart::$lines`),
  read or written.
- D2. An instance property counts where it is read: as a value, an
  operand, the target of a compound assignment other than `??=` or of an
  increment or decrement (`$cart->total += 1`, `$cart->lines[0]++`),
  which read it before writing, an argument PHP passes by value, an array
  item, the subject a `foreach` iterates by value, a key of a
  destructuring target (`[$cart->key => $value] = $row`), or the base of
  an element or property read (`$cart->items[0]`, `$cart->owner->name`).
- D3. The object's classes come from native declarations only, as
  UndefinedMethod defines them (D2 to D3 there).
- D4. No class-like in the hierarchy of any of the classes declares a
  property of that name (compared case-sensitively) by a declaration, a
  promoted constructor parameter, a trait or a `@property`,
  `@property-read` or `@property-write` tag; none declares `__get` or
  `__isset`; none is `stdClass` or carries `#[AllowDynamicProperties]`.
- D5. Unless the access names its class exactly (a class name, `self`,
  `parent`), the object may be of any class extending or implementing its
  type: no such class known to the analysis declares the property or a
  handler either, by the same terms as D4.

## Exceptions (no report)

- E1. Writes to an instance property and accesses that never warn: the
  target of a plain or `??=` assignment, also through an element
  (`$cart->items[] = 1`), a by-reference binding (`&$cart->total`, an
  argument taken by reference or passed to a callee the analysis cannot
  resolve), `unset()` (also of an element or a property below it, as in
  `unset($cart->meta->key)`), a `foreach` target or a subject it iterates
  by reference, a destructuring target, and a subject of `isset()`,
  `empty()` or the left operand of `??`, also through elements or
  properties of it. Static properties are only exempt under `isset()`,
  `empty()` and `??`.
- E2. An instance property of a name the analyzed file writes anywhere
  (E1, except an `unset()` of a property below it, which creates
  nothing): a dynamic property created there may be read elsewhere. This
  holds at every PHP version, since PHP 8.2 only deprecates dynamic
  properties.
- E3. Dynamic property names (`$cart->$name`, `$cart->{'total'}`,
  `Cart::${'lines'}`) and dynamic class names (`$class::$lines`).
- E4. The exceptions of UndefinedMethod on receivers, unions, subclasses,
  traits, closures, `@mixin` tags, builtin classes answering undeclared
  names, classes declared more than once or conditionally, and unresolved
  hierarchies (E2 to E6 there).
- E5. Accesses guarded by a check of a property of that name: inside a
  branch whose condition calls `property_exists()` with that name as a
  literal string, or holds `isset()` of a property of that name, or after
  such a call whose result guards the read (an early exit, a short circuit,
  an else arm).

## Report

- Range: the property name (the identifier, or the `$name` variable of a
  static access).
- Severity: error.
- Message: `Fix the property name: Shop\Cart declares no $totl.`: the
  receiver's classes (`|`-separated for a union) and the property as
  written.

## Fix

None: the intended property cannot be guessed.

## Options

None.

## PHP versions

All versions.

## Examples

```php
<?php
namespace Shop;

final class Cart
{
    public int $total = 0;
}

function show(Cart $cart): int
{
    return $cart-><error descr="Fix the property name: Shop\Cart declares no $totl.">totl</error>;
}
```

Writes, checks and objects accepting any property are not reported:

```php
<?php
namespace Shop;

final class Cart {}

function fill(Cart $cart, \stdClass $row): void
{
    $cart->note = 'gift';
    echo $cart->note, $row->id;
    echo isset($cart->label) ? 'yes' : 'no';
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
