---
id: EmptyClass
group: Architecture
kind: semantic
needs: [names, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# EmptyClass

## Summary

A named class (or trait, or enum) that declares nothing at all — no
properties, constants, methods, used traits or enum cases — is usually dead
weight or an unfinished stub.

## Detection

- **D1** A named class-like declaration: `class`, `trait` or `enum`
  (interfaces are never reported; anonymous classes are never reported).
- **D2** It declares, in its own body:
  - no properties (including promoted constructor properties — these imply a
    constructor anyway),
  - no class constants,
  - no methods,
  - no `use SomeTrait;` statements,
  - no enum cases.
  `extends`/`implements` clauses do not matter.
- **D3** It is not marked deprecated (its doc comment carries an
  `@deprecated` tag).
- **D4** If the declaration has an `extends` clause whose parent class
  resolves:
  - the parent is `abstract` → no report;
  - the class's inheritance closure (the class itself, all its resolved
    ancestor classes, all interfaces they implement transitively, and all
    traits they use transitively) contains a class whose FQN is exactly
    `\Exception` → no report.
  If the parent does not resolve, D4 does not apply (the class is reported).

## Exceptions (no report)

- **E1** Interfaces, anonymous classes.
- **E2** Classes with any member, trait use or enum case.
- **E3** `@deprecated` classes.
- **E4** Classes extending an abstract class.
- **E5** Classes that are (directly or indirectly) subclasses of
  `\Exception`. Subclasses of `\Error` or other throwables that do not reach
  `\Exception` are reported.
- **E6** (custos) Classes carrying an attribute or a docblock annotation
  (see Divergences).
- **E8** (custos) Classes whose doc comment defines a type alias with
  `@phpstan-type` or `@psalm-type`: other files import the alias through
  the class (`@phpstan-import-type Row from Rows`), which is its purpose.
- **E9** (custos) Classes in a test context: the file name ends with
  `Test.php`, `Spec.php` or `.phpt` or the path has a `/Fixtures/`
  directory, or the class name ends with `Test` or its namespace has a
  `Tests` or `Test` segment. Test doubles and fixtures are empty on purpose:
  the tests only need a distinct class name or type.
- **E7** (custos) Declarations nested in a block (`if (…) { class A {} }`):
  a conditional declaration is usually the empty fallback of a version
  check whose other branch declares the full class or trait.

## Report

- Range: the name identifier of the class/trait/enum.
- Severity: info.
- Message (class or trait): `This class declares no members; remove it or
  give it a purpose.`
- Message (enum): `This enum declares no cases or methods.`

## Fix

None.

## Options

None.

## PHP versions

No gating. Enums (8.1 syntax) are parsed and checked at every level.

## Examples

```php
<?php
class HasState  { private $count = 0; }
class HasLimit  { const MAX = 10; }
class HasAction { public function go() {} }
trait Greets    { public function hi() {} }
class UsesTrait { use Greets; }
interface Marker {}

class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Placeholder</weak_warning> {}
class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Tagged</weak_warning> implements Marker {}
trait <weak_warning descr="This class declares no members; remove it or give it a purpose.">Blank</weak_warning> {}
enum <weak_warning descr="This enum declares no cases or methods.">Mood</weak_warning> {}
enum Size { case Small; }

class ParseFailure extends \InvalidArgumentException {}
abstract class Shape { abstract public function area(); }
class Dot extends Shape {}
/** @deprecated use HasState */
class OldStub {}
class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Fatal</weak_warning> extends \Error {}
$x = new class {};
```

## Divergences

- **E9 — custos diverges from upstream.** Upstream also reports empty
  test doubles and fixtures.
- `#[\Deprecated]` attributes are not considered by upstream (docblock tag
  only). custos does not treat the attribute as a deprecation marker either,
  but a class carrying it is still not reported: any attribute exempts the
  class (see below).
- **Attribute classes (custos diverges).** A class carrying `#[Attribute]`
  (resolved through imports: `#[\Attribute]`, `use Attribute;
  #[Attribute(…)]`) is an attribute: it is used by its name alone
  (`#[WithoutRelations]`), so declaring nothing is its purpose. custos does
  not report it; upstream does.
- **Configured classes (custos diverges, extends the above).** A class
  carrying any attribute (`#[ApiResource]`, `#[ORM\Entity]`, `#[Get(…)]`,
  `#[AsEventListener]`, `#[DiscriminatorMap]`) is configured from outside:
  frameworks discover and use it through the attribute, so having no
  members is its purpose (7 of 18 sampled findings on API Platform,
  Sylius and Shopware). custos does not report it. This also exempts
  classes whose only attribute is metadata (Shopware `#[Package]`), an
  accepted loss.
- **Annotated classes (custos diverges, extends the above).** A docblock
  annotation, a tag whose name is not all lower-case (`@FieldFormatter(…)`,
  `@ContentEntityType(…)`, `@ORM\Entity`, `@Annotation`), registers the
  class the way an attribute does: Drupal discovers its plugins through
  them (11 of 15 sampled findings on a Drupal site). custos does not report
  such a class. Lower-case phpDoc tags (`@internal`, `@see`) do not count.
- **Type-alias holders (custos diverges, E8).** A class that only carries
  `@phpstan-type`/`@psalm-type` definitions exists so that other files can
  import them; upstream reports it as empty.
- **Conditional declarations (custos diverges, E7).** Upstream reports the
  empty `else` side of `if (version_compare(…)) { trait T { … } } else {
  trait T {} }` (Symfony's Redis proxy traits); removing it leaves `T`
  undeclared on the other versions, a fatal error for every class using it.
