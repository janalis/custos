---
id: SelfClassReferencing
group: Code style
kind: semantic
needs: [names]
php: { min: "", max: "" }
---

# SelfClassReferencing

## Summary
Inside a class's own methods, refer to the class consistently: by default with
`self` (and `__CLASS__` instead of `Name::class`), which survives renames; or,
when `PREFER_CLASS_NAMES` is on, with the explicit class name instead of `self`
/ `__CLASS__`.

## Detection
Scope (both modes):
- S1: only methods of a named class (class, enum, or any non-anonymous,
  non-trait class-like). Methods of traits and anonymous classes are skipped.
- S2: abstract methods are skipped (interface methods, having no body, are
  treated as abstract).
- S3: a candidate node counts only if its nearest enclosing function-like is
  the method itself — nodes inside closures, arrow functions, nested named
  functions, or methods of anonymous classes declared in the method are not
  considered. The method's signature (parameter types, return type) is part of
  the method.
- S4: code outside methods (constant values, property defaults, class-level
  attributes) is never inspected.

Default mode (`PREFER_CLASS_NAMES` = false), with `Name` = the class's short
name:
- D1: a class-name reference whose last segment is `Name` in any letter case
  (class names are case-insensitive in PHP), which resolves (through namespace and imports) to this very
  class. Class-name references include: `new Name`, `Name::CONST`,
  `Name::method()`, `Name::$prop`, `instanceof Name`, `catch (Name $e)`,
  parameter/return/property types (including inside nullable/union types).
  Qualified forms (`\Ns\Name`, `namespace\Name`) count when they resolve to the
  class. Docblock types are not references.
- D2: if the reference is the class part of `Name::class` (the keyword
  `class` in any letter case), the whole `Name::class` expression is reported
  instead (replacement `__CLASS__`), and not the reference alone — but only
  when the reference, as written, resolves to the class FQN with its exact
  declared letter case (E6).
- D3: otherwise the reference itself is reported (replacement `self`).

Reverse mode (`PREFER_CLASS_NAMES` = true):
- D4: a class reference written `self` in any letter case (`SELF`) that resolves
  to this class is reported (replacement: the class's short name). This
  includes `self::class` (only `self` is reported/replaced), `new self`, type
  declarations `self`, etc.
- D5: the magic constant `__CLASS__` in any letter case (`__class__`) is reported
  (replacement `Name::class`).

## Exceptions (no report)
- E1: references inside closures, arrow functions or nested functions (S3).
- E2: traits, anonymous classes, abstract/interface methods (S1, S2).
- E3: a reference whose direct parent is an `extends` clause (e.g.
  `new class() extends Name {}` written inside a method of `Name`).
- E4: references to a different class with the same short name (resolution
  fails to point at this class).
- E5: `static`, `parent`, `static::class` are never touched; in default mode
  `self` is fine; in reverse mode explicit `Name` is fine.
- E6: `Name::class` written with a letter case that differs from the
  declaration (`basket::class`, `\shop\Basket::class`): `::class` yields the
  name as written, so `__CLASS__` (or `self::class`) would change the
  resulting string. Not reported at all.

## Report
- Range: the class reference node (full text including any namespace
  qualifier) for D1/D3/D4; the whole `Name::class` expression for D2; the
  `__CLASS__` token for D5.
- Severity: info (weak warning).
- Message (default mode): "Refer to the class as '{replacement}' instead of '{found}'."
- Message (reverse mode): "Spell the class reference '{found}' as '{replacement}'."

## Fix
- F1 (D3): replace the reference with `self`.
- F2 (D2): replace the whole `Name::class` with `__CLASS__`.
- F3 (D4): replace `self` with the class's short name.
- F4 (D5): replace `__CLASS__` with `Name::class`.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| PREFER_CLASS_NAMES | bool | false | Switches the preferred style: false = use `self`/`__CLASS__`; true = use the explicit class name / `Name::class`. |

## PHP versions
No gating upstream (see Divergences for `::class` on PHP < 5.5).

## Examples

Default mode:

```php
<?php
namespace Shop;

class Basket
{
    public function merge(<weak_warning descr="Refer to the class as 'self' instead of 'Basket'.">Basket</weak_warning> $other): ?<weak_warning descr="Refer to the class as 'self' instead of 'Basket'.">Basket</weak_warning>
    {
        $copy = new <weak_warning descr="Refer to the class as 'self' instead of 'Basket'.">\Shop\Basket</weak_warning>();
        $max  = <weak_warning descr="Refer to the class as 'self' instead of 'Basket'.">Basket</weak_warning>::LIMIT;
        if ($other instanceof <weak_warning descr="Refer to the class as 'self' instead of 'Basket'.">Basket</weak_warning>) {}
        log_event(<weak_warning descr="Refer to the class as '__CLASS__' instead of 'Basket::class'.">Basket::class</weak_warning>);

        $make = fn() => new Basket();          // inside arrow function: skipped
        $anon = new class() extends Basket {}; // extends clause: skipped
        return self::empty();
    }
}

trait Countable2
{
    public function who() { return Countable2::class; }   // trait: skipped
}
```

```php
<?php
namespace Shop;

class Basket
{
    public function merge(self $other): ?self
    {
        $copy = new self();
        $max  = self::LIMIT;
        if ($other instanceof self) {}
        log_event(__CLASS__);

        $make = fn() => new Basket();          // inside arrow function: skipped
        $anon = new class() extends Basket {}; // extends clause: skipped
        return self::empty();
    }
}

trait Countable2
{
    public function who() { return Countable2::class; }   // trait: skipped
}
```

Reverse mode (`PREFER_CLASS_NAMES` = true):

```php
<?php
class Ledger
{
    public static function open(<weak_warning descr="Spell the class reference 'self' as 'Ledger'.">self</weak_warning> $base): <weak_warning descr="Spell the class reference 'self' as 'Ledger'.">self</weak_warning>
    {
        tag(<weak_warning descr="Spell the class reference '__CLASS__' as 'Ledger::class'.">__CLASS__</weak_warning>);
        tag(<weak_warning descr="Spell the class reference 'self' as 'Ledger'.">self</weak_warning>::class);
        return new <weak_warning descr="Spell the class reference 'self' as 'Ledger'.">self</weak_warning>;
    }

    public function plain() {
        return new Ledger;          // already explicit
    }
}

__CLASS__;                          // not inside a class method
```

```php
<?php
class Ledger
{
    public static function open(Ledger $base): Ledger
    {
        tag(Ledger::class);
        tag(Ledger::class);
        return new Ledger;
    }

    public function plain() {
        return new Ledger;          // already explicit
    }
}

__CLASS__;                          // not inside a class method
```

## Divergences
- Upstream also considers references inside an anonymous class declared in the
  method but outside that anonymous class's methods (e.g. its constant or
  property initialisers): `self` there would mean the anonymous class.
  Recommendation: skip any reference located inside an anonymous class body.
  Not covered by upstream fixtures.
- Reverse mode on PHP < 5.5: `Name::class` does not exist. Recommendation: do
  not report D5 (`__CLASS__`) when the configured PHP version is below 5.5.
- **Letter case (custos diverges).** Upstream matches `Name`, `self` and
  `__CLASS__` case-sensitively, so `new basket()`, `SELF::X` and `__class__`
  are skipped although PHP resolves them to the same class. custos matches
  them in any letter case (D1, D4, D5), except `Name::class` whose case
  differs from the declaration (E6), since rewriting it would change the
  string value. Reverse-mode messages quote the reference as written.
