---
id: TraitsPropertiesConflicts
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy]
php: { min: "", max: "" }
---

# TraitsPropertiesConflicts

## Summary

When a class and one of its traits (or its parent class and one of its
traits) declare the same property, PHP only accepts it if both declarations
are compatible, and even then the duplication is fragile. Point out these
overlaps; when the two declarations are incompatible (PHP refuses to compose
the class), it is a real conflict.

## Detection

Visit every class-like declaration (class, trait, enum) that directly uses at
least one trait (`use T1, T2;` inside its body). Let `Traits` be the resolved
traits in declaration order. Unresolvable trait names are ignored.

Helper predicates:

- *real property* of `X`: a property declared in `X`'s own body (not a class
  constant, not a `@property` docblock tag, not a property inherited from
  elsewhere, not a promoted-constructor quirk — promoted properties count as
  declared in the class).
- *default(p)*: the property's default value expression, or *none* when it
  has no initializer (`public $a;`). `= null` is a default (not *none*).
- *same default*: both *none*, or both present and equivalent (same token
  sequence ignoring whitespace/comments, or identical text). One *none* and
  one present → different.
- *compatibility* of two declarations `x` and `y` (PHP's rules for
  composing a trait property into a class). The pair is:
  - *incompatible* when the visibility (`public`/`protected`/`private`),
    the `static` flag or the `readonly` flag differs (a `readonly` class
    makes all its properties readonly), when exactly one side declares a
    type, when both declare types that differ (compared as resolved,
    normalised type sets, case-insensitively; `?T` equals `T|null`), or when
    the initial values are known to differ (below);
  - *compatible* when none of the above differs and the defaults are the
    *same* (or both evaluate to the same known value);
  - *undecided* otherwise: a type mentioning `self`/`static`/`parent`
    (meaning depends on the composing class), or defaults that differ in
    text but are not both known values.
  Known initial values: no initializer (equals `null` for an untyped
  property, "uninitialized" for a typed one), integer and float literals
  with an optional leading minus (compared by value and kind, so `1` and
  `1.0` differ), single- or double-quoted strings without backslashes or
  `$` (compared by content, so `'a'` equals `"a"`), and `true`/`false`/`null`
  in any case. Constants, arrays and other expressions are unknown. A
  constructor-promoted property never has an initial value of its own.
- *trait property lookup* `find(T, name)`: the property named `name` that is
  a real property of trait `T` (a property `T` obtains from a trait it
  itself uses may also be found; a `@property` docblock tag on `T` does not
  count).

### A. Own properties

- **D1** For each real property `p` of the class (named, non-constant,
  non-abstract):
  - skip if `p`'s docblock contains a tag whose name is not all-lowercase
    (an annotation such as `@ORM\Column`, `@Id`);
  - if `p` carries a PHP attribute (`#[...]`, also on a promoted
    constructor parameter), a *compatible* pair is not reported (custos
    diverges, see Divergences); incompatible pairs still are;
  - walk `Traits` in order and take the first trait `T` for which
    `find(T, p.name)` exists; stop there (later traits are not consulted);
  - *compatible* pair → report on `p`'s name (weak); *incompatible* pair →
    report on `p`'s name as an **error**; *undecided* → no report (custos
    diverges: upstream only compares defaults and stays silent when they
    differ).

### B. Parent properties

- **D2** Collect the trait references written in the class's `use` lists
  (including names inside a `{ ... insteadof ...; }` adaptation block); map
  each resolved trait to its first reference. Let `Parent` be the directly
  extended class (resolved). Stop if there is no parent or no resolved trait
  reference.
- **D3** For each real property `q` of `Parent` (declared in `Parent`'s own
  body only — properties `Parent` inherits or gets from its own traits are
  skipped), non-constant, not `private`, not abstract (annotations are not
  checked here):
  - take the first trait `T` in `Traits` for which `find(T, q.name)` exists;
    stop there;
  - if `T` has a mapped reference: report on that reference — **weak** when
    the defaults are the *same* and the pair is not *incompatible*,
    **error** otherwise (different defaults, or a visibility/`static`/
    `readonly`/type mismatch).

Checks A and B are independent; both may report for the same class.

## Exceptions (no report)

- **E1** Classes without trait usage.
- **E2** Own property and trait property whose compatibility cannot be
  decided (A): e.g. `= self::RED` against `= 'red'`.
- **E3** Annotated own properties (A); docblock `@property` declarations on
  either side; class constants.
- **E4** Private or inherited (non-own) parent properties (B).
- **E5** No parent class (B).

## Report

- A: range = the property's name token including `$` (e.g. `$same`, not the
  modifiers, not the default). Severity: info (weak warning).
- B: range = the trait name as written in the `use` list (the reference that
  resolved to `T`, with any namespace qualifier). Severity: error when the
  defaults differ, info (weak warning) when they are the same. Several
  properties may produce several reports on the same reference.
- Message (both): `{class} and trait {trait} both declare property ${name}.`
  where `{class}` is the inspected class's short name and `{trait}` the
  trait's short name.

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
trait HasColor {
    public $color = 'red';
    protected $shade = 1;
}
trait HasSize {
    public $size = 10;
    private $unit = 'px';
}

/**
 * @property $note
 */
trait HasMeta {
    /** @Column */
    public $tag = 'x';
    public $label = 'L';
    protected $weight = 2;
}

class Shape {
    public $color = 'red';
    protected $size = 12;
    private $unit = 'em';
}

/**
 * @property $weight
 */
class Box extends Shape {
    use <weak_warning descr="Box and trait HasColor both declare property $color.">HasColor</weak_warning>, <error descr="Box and trait HasSize both declare property $size.">HasSize</error>;
    use HasMeta;

    /** @Column */
    public $tag = 'x';
    public <weak_warning descr="Box and trait HasMeta both declare property $label.">$label</weak_warning> = 'L';
    protected <error descr="Box and trait HasMeta both declare property $weight.">$weight</error> = 3;
    public $note = 'n';
}
```

## Divergences

- **Attributed re-declarations (custos diverges):** upstream exempts only
  properties annotated in their docblock (`@ORM\Column`), so the PHP 8
  equivalent — re-declaring a trait property to attach `#[ORM\ManyToOne]`
  or similar attributes — is reported as a duplicate. custos does not report
  a *compatible* pair when the own property carries an attribute; an
  *incompatible* one is still an error (PHP refuses to compose the class
  regardless of attributes).
- **Full compatibility criteria (custos diverges).** Upstream decides
  between the weak and the error report from the default value alone, so a
  trait `protected int $hits = 0` against a class `public $hits = 0` is
  presented as a harmless duplicate although PHP refuses to compose the
  class. custos also compares visibility, `static`, `readonly` and the
  declared type (see *compatibility*) in both checks.
- **Incompatible own properties (custos diverges).** In check A upstream
  reports compatible duplicates (weak) and is silent on incompatible ones,
  which are the fatal case. custos reports them as errors, like check B;
  pairs it cannot decide (non-literal defaults with different text, types
  using `self`/`static`/`parent`) stay silent.
