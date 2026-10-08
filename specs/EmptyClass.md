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
- `#[\Deprecated]` attributes are not considered by upstream (docblock tag
  only). Recommendation: same.
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
