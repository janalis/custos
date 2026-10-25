---
id: ClassMethodNameMatchesFieldName
group: Confusing constructs
kind: semantic
needs: [names, hierarchy, index, types]
php: { min: "", max: "" }
---

# ClassMethodNameMatchesFieldName

## Summary

When a method and a property share a name, `$obj->name()` and
`($obj->name)()` mean different things. This is especially confusing when the
property holds a callable, and when the property's type is unknown the reader
cannot tell which one is meant.

## Detection

- **D1** A method declared in a class, abstract class, trait or enum (any
  visibility, static or not). Methods of interfaces are skipped.
- **D2** Look up a **property** (not a class constant) whose name equals the
  method name exactly (case-sensitive, without `$`) in the containing class,
  including properties inherited from ancestor classes and properties of used
  traits (any visibility, static or not). No match → nothing.
- **D3** Compute the property's type set: union of
  - its declared type (typed property),
  - the types of an `@var` tag in the comment immediately preceding the
    declaration — both doc comments `/** … */` **and** plain block comments
    `/* … */` are honoured,
  - the type of its default value, if any (e.g. `= 0` → int, `= null` → null,
    `= []` → array).
  Drop unknown/unresolvable entries.
- **D4** If the set is empty → report "type unknown" (D4 message).
- **D5** Otherwise, if any type in the set is `callable` or the class
  `\Closure` (compare case-insensitively; a bare `closure`/`Closure` in a
  docblock of a file in the global namespace resolves to `\Closure`) → report
  "callable" (D5 message). Any other known type → no report.

## Exceptions (no report)

- **E1** Interface methods.
- **E2** A same-named class constant (`const name`) is not a property.
- **E3** A same-named property with a known, non-callable type
  (e.g. `@var int`, `= 'x'`, `?string`).
- **E4** A same-named property inherited from a built-in (PHP core or
  extension) class when no type is known for it, e.g. `file()` or `line()`
  on an `\Exception` subclass. The "type unknown" report (D4) is withheld:
  the type of such a property is fixed by PHP, not missing, and the user
  cannot type it.

- **E5** (custos) The property found is `private` and declared in an
  ancestor class (or a trait that an ancestor uses): it is invisible from the
  method's class, where `$this->name` cannot name it.

## Report

- Range: the method name identifier.
- Severity: info (both cases).
- Messages:
  - D4: `A property with this name exists and its type is unknown; rename
    the method or type the property.`
  - D5: `A callable property with this name exists; rename the method (for
    example with a get/is/has prefix).`
- **R1** (custos) When the parent class or an interface of the method's
  class declares a method with this name, renaming the method would break
  that contract, so the messages advise renaming the property instead:
  - D4: `A property with this name exists and its type is unknown; the
    method name is inherited, so rename or type the property.`
  - D5: `A callable property with this name exists; the method name is
    inherited, so rename the property.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
class Account {
    private $owner;
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">owner</weak_warning>() {}
}

class Hooks {
    /* @var callable */
    protected $onSave;
    /** @var \Closure|null */
    protected $onLoad;
    public function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onSave</weak_warning>() {}
    public static function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onLoad</weak_warning>() {}
}

class Child extends Account {
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">owner</weak_warning>() {}
}

class Counter {
    private $total = 0;
    private ?string $label;
    const size = 3;
    public function total() {}
    public function label() {}
    public function size() {}
}

interface Named {
    public function name();
}
```

## Divergences

- Upstream takes the property type from the IDE's type engine; whether types
  inferred from `$this->prop = …` assignments inside the class also count is
  not observable in fixtures. custos uses only D3 sources (declared type,
  preceding `@var` comment, default value).
- **Promoted properties (custos diverges, D3).** For a constructor-promoted
  property, the constructor's `@param` type of the same name is part of
  the property's type set (`/** @param bool $isFinal */ __construct(private
  $isFinal)`): the doc comment of the constructor is where such a property
  is documented. Upstream (no promotion support) has no type and reports
  "type unknown" (three Mautic DTOs).
- **Built-in properties (custos diverges, E4).** A method named like an
  untyped-looking property of a built-in ancestor (`\Exception::$file`,
  `$line`, `$code`) is not reported as "type unknown": these properties are
  typed or documented by PHP itself and the advice to type the property
  cannot be followed.
- **Private ancestor properties (custos diverges, E5).** Upstream looks the
  property up through the ancestors whatever its visibility, so a subclass
  method is reported for a `private` property of its parent, which the
  subclass cannot read or call (Symfony's `ContainerBuilder::getEnv()`).
- **Inherited method names (custos, R1).** Same reports, but the advice to
  rename the method cannot be followed when an interface or the parent class
  imposes the name (`DataTransformerInterface::transform()`,
  `\Countable::count()`): the property is the one to rename.
