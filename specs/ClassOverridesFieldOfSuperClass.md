---
id: ClassOverridesFieldOfSuperClass
group: Architecture
kind: semantic
needs: [names, hierarchy, index]
php: { min: "", max: "" }
---

# ClassOverridesFieldOfSuperClass

## Summary

Re-declaring a property that an ancestor already declares is usually
accidental: for protected/public properties it merely duplicates the
declaration (often with a diverging default), and for an ancestor's private
property it creates a second, unrelated property with the same name.

## Detection

- **D1** A property declaration in a class body (each property of a
  multi-property declaration is checked on its own), where:
  - the containing class is not in a test context (T below);
  - the property is not static (and is not a class constant);
  - it is a real declaration in the class body: virtual properties from
    `@property` doc tags and promoted constructor parameters are not
    considered.
- **D2** The property's doc comment (the comment attached to the declaration
  statement) has no tag whose name contains an uppercase letter
  (e.g. `@ORM\Column`, `@Assert\NotBlank`, `@Inject`); lowercase tags such as
  `@var`, `@internal` are fine. Attributes (`#[…]`) are not considered.
- **D3** Let `P` be the resolved direct parent class (`extends`). If `P` has
  a constructor — its own `__construct` or one it inherits — that is `final`
  or `private` → no report.
- **D4** Look up a property with the same name (case-sensitive) in `P`
  *including everything `P` inherits*: its ancestors' properties of any
  visibility (private included) and properties of traits used anywhere in
  that chain. Let `H` be the class or trait that declares the found property.
  No `P` or no match → nothing.
- **D5** If the found property is **private** → report "private
  redefinition" (only when `REPORT_PRIVATE_REDEFINITION` is on), regardless
  of the own property's visibility. Stop.
- **D6** Otherwise (found property is protected or public): report
  "duplicate" unless the own property's visibility is *more permissive* than
  the found one. Visibility order: public > protected > private.
  - protected over protected, public over public, private/protected over
    public, private over protected → report;
  - public over protected → no report (visibility deliberately widened).
  A `var` property is public.
- **D6b** (custos, see Divergences) No "duplicate" report when the own
  declaration's default differs from the found property's default:
  defaults are compared as source text with whitespace removed; `null`
  (any case) and a missing default on an untyped property are the same;
  a missing default on a typed property (uninitialised) differs from any
  default.

**T — test context**: the file path ends with `Test.php`, `Spec.php` or
`.phpt`, or contains `/Fixtures/`; or the class FQN ends with `Test`, or
contains `\Tests\` or `\Test\`.

## Exceptions (no report)

- **E1** Static properties and constants.
- **E2** Test classes (T).
- **E3** Properties annotated with framework-style tags (uppercase letter in
  the tag name).
- **E4** Parent class (or its ancestors) with a final or private
  constructor.
- **E5** Widening protected → public.
- **E6** `@property` virtual properties; promoted constructor properties.
- **E7** Private redefinitions when `REPORT_PRIVATE_REDEFINITION` is false.

## Report

- Range: the own property name including `$` (e.g. `$items`), not the
  modifiers or default value.
- Severity: **info** (weak warning) for both cases. Upstream registers both
  with a weak-warning level although the catalogue default is `warning`; the
  fixtures expect weak warnings.
- Messages:
  - D6: `Property '{name}' is already declared in {H}; drop this
    re-declaration.` (`{H}` = FQN with leading `\`)
  - D5: `{H} already has a private property with this name; consider a
    different name.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `REPORT_PRIVATE_REDEFINITION` | bool | true | Also report re-declarations of an ancestor's (or ancestor trait's) private property (D5). |

## PHP versions

No gating.

## Examples

Default options:

```php
<?php
trait Tracks { private $trail; }

class Model {
    use Tracks;
    private $secret;
    protected $table;
    protected $visible;
    public $id;
    protected static $registry;
    const KIND = 'model';
}
class Record extends Model {
    protected $dirty;
}

/** @property string $virtual */
class Invoice extends Record {
    protected <weak_warning descr="Property 'table' is already declared in \Model; drop this re-declaration.">$table</weak_warning> = 'invoices';
    protected <weak_warning descr="Property 'dirty' is already declared in \Record; drop this re-declaration.">$dirty</weak_warning>;
    public <weak_warning descr="Property 'id' is already declared in \Model; drop this re-declaration.">$id</weak_warning>;
    public <weak_warning descr="\Model already has a private property with this name; consider a different name.">$secret</weak_warning>;
    private <weak_warning descr="\Tracks already has a private property with this name; consider a different name.">$trail</weak_warning>;

    public $visible;
    protected static $registry;
    const KIND = 'invoice';
}

class InvoiceTest extends Record {
    protected $table;
}

class Entity { protected $uuid; }
class Order extends Entity {
    /** @Column(type="guid") */
    protected $uuid;
}

class Locked {
    protected $state;
    final protected function __construct($seed = 1) {}
}
class Door extends Locked {
    protected $state;
}
```

## Divergences

- **Default overrides (custos diverges).** Upstream reports any
  re-declaration, including `protected $table = 'invoices';` over a parent's
  `protected $table;` (Eloquent models) or `protected bool $skipScalars =
  true;` over `false` (Symfony compiler passes). Re-declaring is the way to
  override a default; dropping it as advised changes behaviour. custos only
  reports re-declarations with the same default (D6b).
- **Documented narrowing (custos diverges).** A re-declaration whose `@var`
  type differs from the inherited property's documented type (or, without
  one, its declared type) narrows the type for static analysis and
  completion (`/** @var ReportModel|null */ protected $model;` over the
  parent's `@var Model|null`); native property types are invariant, so
  the doc comment is the only way to refine it. Dropping it changes nothing
  at runtime but loses the type, so custos does not report it (Mautic API
  controllers, Akeneo `OptionValue::$data`). Re-declarations repeating the
  same type are still reported.
