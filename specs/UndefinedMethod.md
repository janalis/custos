---
id: UndefinedMethod
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types]
php: { min: "", max: "" }
---

# UndefinedMethod

## Summary

Report a method call on an object or a class that declares no method of
that name, judging the object only by what the code itself shows. PHP
throws an `Error` when such a call runs; the name usually has a typo, or
the method belongs to another class. Objects whose classes may answer
any name (magic `__call`, `@method` and `@mixin` tags) are left alone.

## Detection

- D1. A method call with a name written in the source: through an object
  (`$cart->add()`, `$cart?->add()`) or through a class name, `self`,
  `parent` or `static` (`Cart::make()`, `parent::boot()`).
- D2. The object's classes come from native declarations only (typed
  parameters, properties and returns, `new`, `instanceof` checks), as
  InvalidArgumentType defines it; null aside, every possible type of the
  object is a class, an interface or an enum. A static call names one
  class.
- D3. Each class is declared once, outside any condition or function, and
  so is every class-like in its hierarchy (parents, interfaces, traits,
  transitively), which the analysis knows completely.
- D4. No class-like in the hierarchy of any of the classes declares a
  method of that name (compared case-insensitively) for any PHP version,
  by a declaration, a trait import alias (`use T { m as alias; }`) or a
  `@method` tag, nor `__call` or `__callStatic`.
- D5. Unless the call names its class exactly (a class name, `self`,
  `parent`), the object may be of any class extending or implementing its
  type: no such class known to the analysis declares the method either, by
  the same terms as D4.

## Exceptions (no report)

- E1. Dynamic method names (`$cart->$name()`, `$cart->{'add'}()`) and
  dynamic class names (`$class::make()`).
- E2. Receivers whose type only PHPDoc gives, receivers of unknown,
  `mixed` or `object` type, unions with a non-class member (`Cart|int`),
  and nullsafe chains whose start is unknown.
- E3. A union or an intersection one of whose classes declares the method
  (D4), and a type a subclass or an implementation of which declares it
  (D5): code that relies on a narrower type than its declarations show
  (often through PHPDoc) still works.
- E4. `$this`, `self::`, `static::` and `parent::` inside a trait (it runs
  in the classes using it), and `$this`, `self::`, `static::` and
  `parent::` inside a closure or an arrow function, which may run bound to
  another object or scope.
- E5. Classes whose hierarchy carries a `@mixin` tag (or its `phpstan-` or
  `psalm-` variant), and builtin classes that answer undeclared names:
  `RecursiveIteratorIterator` forwards calls to its inner iterator, and
  `SimpleXMLElement`, `ArrayObject`, `ArrayIterator`, `PDORow`, `COM`,
  `VARIANT` and `DOTNET` resolve members at run time.
- E6. A class declared more than once (a polyfill next to PHP's own class,
  two variants of a file) or declared conditionally
  (`if (!class_exists(…)) { class … }`), in the hierarchy too, and a class
  with an ancestor the analysis does not know.
- E7. A method PHP declares only in another version than the configured
  one (`ReflectionProperty::hasDefaultValue()` before PHP 8.0): code meant
  for several versions guards it by version.
- E8. Calls guarded by a check of a method of that name: inside a branch
  whose condition calls `method_exists()` or `is_callable()` with that
  name as a literal string (also inside an array, as in
  `is_callable([$cart, 'add'])`, or after `::` in a string whose class
  part names the receiver's class, `self`, `static` or `parent`, as in
  `is_callable('Shop\Cart::add')`), or after such a check whose result
  guards the call: an early exit, a short circuit
  (`method_exists($cart, 'add') && $cart->add()`), an else arm.
- E9. `self::` and `parent::` outside a class, or `parent::` in a class
  without a parent.

## Report

- Range: the method name.
- Severity: error.
- Message: `Fix the method name: Shop\Cart declares no clear().`: the
  receiver's classes (`|`-separated for a union) and the method as
  written.

## Fix

None: the intended method cannot be guessed.

## Options

None.

## PHP versions

All versions. Methods PHP declares count at any version (E7).

## Examples

```php
<?php
namespace Shop;

final class Cart
{
    public function add(int $n): void {}
}

function fill(Cart $cart): void
{
    $cart-><error descr="Fix the method name: Shop\Cart declares no ad().">ad</error>(1);
}
```

Magic methods, PHPDoc-only receivers and guarded calls are not reported:

```php
<?php
namespace Shop;

class Proxy
{
    public function __call($name, $args) {}
}

final class Note {}

/** @param Note $note */
function play(Proxy $proxy, $note, Note $real): void
{
    $proxy->anything();
    $note->play();
    if (method_exists($real, 'play')) {
        $real->play();
    }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
